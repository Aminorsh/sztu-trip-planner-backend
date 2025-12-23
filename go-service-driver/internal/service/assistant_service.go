package service

import (
	"bytes"
	"context"
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

func (s *AssistantService) Chat(ctx context.Context, userID uint64, model string, messages []dto.AssistantMessage, stream bool) (string, error) {
	if model == "" {
		model = "deepseek-chat"
	}
	if len(messages) == 0 {
		return "", errors.NewInvalidRequestError("messages cannot be empty")
	}

	summary, _ := s.memoryRepo.GetSummary(ctx, userID)
	systemPrompt := buildSystemPrompt(summary)
	finalMessages := []dto.AssistantMessage{
		{
			Role:    "system",
			Content: systemPrompt,
		},
	}
	finalMessages = append(finalMessages, messages...)

	reply, err := s.callDeepseek(model, finalMessages, stream)
	if err != nil {
		// to be implement...
		return "", err
	}

	// if !config.IsTestMode() {
	go s.updateSummary(
		context.Background(),
		userID,
		summary,
		messages[len(messages)-1].Content,
		reply,
	)
	// }

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

func (s *AssistantService) updateSummary(
	ctx context.Context,
	userID uint64,
	oldSummary string,
	userMessage string,
	assassistantReply string,
) {
	prompt := buildSummaryPrompt(oldSummary, userMessage, assassistantReply)

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

	_ = s.memoryRepo.SetSummary(ctx, userID, newSummary)
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
