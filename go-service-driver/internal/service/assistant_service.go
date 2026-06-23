package service

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Aminorsh/sztu-trip-planner-backend/config"
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/dto"
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/errors"
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/repository"
)

const (
	recentTurnsLimit = 6
	// summaryLockTTL 远大于 DeepSeek 30s 调用上限,留冗余防 GC/网络抖动;
	// 同时保持足够短,进程崩溃后键能尽快自然过期。
	summaryLockTTL = 60 * time.Second
)

func NewAssistantService(memoryRepo repository.AssistantRepository, deepseekAPIKey string, deepseekAPIURL string) *AssistantService {
	return &AssistantService{
		memoryRepo:     memoryRepo,
		deepseekAPIKey: deepseekAPIKey,
		deepseekAPIURL: deepseekAPIURL,
	}
}

type AssistantService struct {
	memoryRepo     repository.AssistantRepository
	deepseekAPIKey string
	deepseekAPIURL string
}

type chatOptions struct {
	model        string
	stream       bool
	with_summary bool
}

type ChatOption func(*chatOptions)

func WithModel(model string) ChatOption {
	return func(co *chatOptions) {
		co.model = model
	}
}

func WithStream(stream bool) ChatOption {
	return func(co *chatOptions) {
		co.stream = stream
	}
}

func WithSummary(with_summary bool) ChatOption {
	return func(co *chatOptions) {
		co.with_summary = with_summary
	}
}

func (s *AssistantService) Chat(ctx context.Context, userID uint64, messages []dto.AssistantMessage, opts ...ChatOption) (string, error) {
	co := &chatOptions{
		model:        "deepseek-chat",
		stream:       false,
		with_summary: false,
	}
	for _, opt := range opts {
		opt(co)
	}
	if co.model == "" {
		co.model = "deepseek-chat"
	}
	if len(messages) == 0 {
		return "", errors.NewInvalidRequestError("messages cannot be empty")
	}

	summary := ""
	var history []repository.AssistantTurn
	if co.with_summary {
		summary, _ = s.memoryRepo.GetSummary(ctx, userID)
		history, _ = s.memoryRepo.GetRecentTurns(ctx, userID, recentTurnsLimit)
	}
	systemPrompt := buildSystemPrompt(summary)
	finalMessages := []dto.AssistantMessage{
		{
			Role:    "system",
			Content: systemPrompt,
		},
	}
	for _, turn := range history {
		finalMessages = append(finalMessages,
			dto.AssistantMessage{Role: "user", Content: turn.User},
			dto.AssistantMessage{Role: "assistant", Content: turn.Assistant},
		)
	}
	finalMessages = append(finalMessages, messages...)

	reply, err := s.callDeepseek(co.model, finalMessages, co.stream)
	if err != nil {
		return "", err
	}

	if co.with_summary {
		lastUserMsg := messages[len(messages)-1].Content
		// 同步写 history:下一轮 Chat 必须看得到这一轮,不能丢
		if appendErr := s.memoryRepo.AppendTurn(ctx, userID, repository.AssistantTurn{
			User:      lastUserMsg,
			Assistant: reply,
		}, recentTurnsLimit); appendErr != nil {
			log.Printf("assistant: append turn failed uid=%d: %v", userID, appendErr)
		}
		// 异步摘要:并发用 Redis 锁去重
		go s.tryUpdateSummary(context.Background(), userID, summary, lastUserMsg, reply)
	}

	return reply, nil
}

func buildSystemPrompt(summary string) string {
	base := `You are an AI assistant designed to help users plan trips effectively. Your role is to provide accurate, concise, and helpful information based on user queries related to travel planning.

Guidelines:
1. Understand User Intent: Carefully analyze the user's questions or statements to grasp their travel needs and preferences.
2. Provide Relevant Information: Offer suggestions on destinations, accommodations, transportation, activities, and dining options that align with the user's interests.
3. Be Concise and Clear: Deliver information in a straightforward manner, avoiding unnecessary jargon or complexity.
4. Stay Updated: Ensure that the information provided is current and reflects the latest travel advisories, trends, and options.
5. Personalize Responses: Tailor your recommendations based on any specific details the user provides about their travel plans, such as budget, duration, and interests.
6. Encourage Engagement: Prompt users for additional details if needed to refine your suggestions and enhance their trip planning experience.
7. If a question falls outside the scope of travel planning, politely inform the user that your expertise is focused on trip planning and offer to assist with any travel-related queries.

Remember, your goal is to assist users in creating memorable and well-organized travel experiences. Always strive to be helpful, informative, and user-centric in your responses.`

	// if config.IsTestMode() {
	// 	return base
	// }

	if summary == "" {
		return base
	}

	return base + "\nHere is the current user context:\n" + summary
}

func buildSummaryPrompt(
	oldSummary string,
	userMessage string,
	assistantReply string,
) string {
	return fmt.Sprintf(`
	You are maintaining a short summary for a travel assistant.

	Previous Summary:
	%s

	New interaction:
	User: %s
	Assistant: %s

	Update the summary in 3-5 concise sentences, focusing on relevant details about the user's travel preferences, plans, and interests.
	`, oldSummary, userMessage, assistantReply)
}

// tryUpdateSummary 在 Redis 锁保护下更新 rolling summary。
// 拿不到锁(已有摘要任务在跑)直接跳过——每次 Chat 都会重试,跳过几轮可接受。
// 释放策略:
//  1. defer ReleaseSummaryLock 覆盖所有正常 / 错误返回路径(token 校验防误删);
//  2. defer recover 兜底 panic,且 LIFO 顺序确保 Release 仍然执行;
//  3. Redis TTL 兜底进程被 SIGKILL 等极端情况。
func (s *AssistantService) tryUpdateSummary(
	ctx context.Context,
	userID uint64,
	oldSummary string,
	userMessage string,
	assistantReply string,
) {
	// LIFO:Release 先 defer、recover 后 defer,recover 先执行后再走 Release
	token, err := newLockToken()
	if err != nil {
		log.Printf("assistant: lock token gen failed uid=%d: %v", userID, err)
		return
	}

	ok, err := s.memoryRepo.AcquireSummaryLock(ctx, userID, token, summaryLockTTL)
	if err != nil {
		log.Printf("assistant: acquire summary lock failed uid=%d: %v", userID, err)
		return
	}
	if !ok {
		return
	}
	defer func() {
		if relErr := s.memoryRepo.ReleaseSummaryLock(ctx, userID, token); relErr != nil {
			log.Printf("assistant: release summary lock failed uid=%d: %v", userID, relErr)
		}
	}()
	defer func() {
		if r := recover(); r != nil {
			log.Printf("assistant: summarize panic uid=%d: %v", userID, r)
		}
	}()

	prompt := buildSummaryPrompt(oldSummary, userMessage, assistantReply)

	newSummary, err := s.callDeepseek(
		"deepseek-chat",
		[]dto.AssistantMessage{
			{
				Role:    "system",
				Content: prompt,
			},
		},
		false,
	)
	if err != nil {
		return
	}

	newSummary = strings.TrimSpace(newSummary)
	if newSummary == "" {
		return
	}

	if err := s.memoryRepo.SetSummary(ctx, userID, newSummary); err != nil {
		log.Printf("assistant: set summary failed uid=%d: %v", userID, err)
	}
}

func newLockToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (s *AssistantService) callDeepseek(model string, messages []dto.AssistantMessage, stream bool) (string, error) {
	if s.deepseekAPIKey == "" || s.deepseekAPIURL == "" {
		return "", errors.NewAssistantConfigError()
	}

	reqBody := dto.DeepseekChatRequest{
		Model:    model,
		Messages: messages,
		Stream:   stream,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", errors.NewSerializationError(err)
	}

	baseURL := fmt.Sprintf("%s/chat/completions", s.deepseekAPIURL)

	if config.IsTestMode() {
		log.Printf("Calling Deepseek API at %s with body: %s", baseURL, string(bodyBytes))
	}

	req, err := http.NewRequest(http.MethodPost, baseURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", errors.NewInternalServerError(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.deepseekAPIKey))

	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", errors.NewDeepseekAPIError(err)
	}
	defer resp.Body.Close()

	if config.IsTestMode() {
		log.Printf("Deepseek API response status: %d", resp.StatusCode)
		log.Printf("Deepseek API response headers: %v", resp.Header)
		log.Printf("Deepseek API response content-length: %d", resp.ContentLength)
	}

	if resp.StatusCode != http.StatusOK {
		return "", errors.MapDeepseekErrorResponse(resp.StatusCode)
	}

	if config.IsTestMode() {
		log.Printf("Starting to read response body...")
	}

	respBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		if config.IsTestMode() {
			log.Printf("Error reading response body: %v", err)
		}
		return "", errors.NewParseResponseError(err)
	}

	if config.IsTestMode() {
		log.Printf("Response body read successfully, length: %d bytes", len(respBodyBytes))
		log.Printf("Deepseek API response: %s", string(respBodyBytes))
	}

	var result dto.DeepseekChatResponse
	if err := json.Unmarshal(respBodyBytes, &result); err != nil {
		return "", errors.NewParseResponseError(err)
	}

	if len(result.Choices) == 0 {
		return "", errors.NewDeepseekAPIError(fmt.Errorf("Empty response from Deepseek API"))
	}

	return result.Choices[0].Message.Content, nil
}
