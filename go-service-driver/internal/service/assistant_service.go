package service

import (
	"fmt"

	"github.com/Aminorsh/sztu-trip-planner-backend/internal/dto"
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/repository"
)

type AssistantService interface {
	Chat(model string, messages []dto.AssistantMessage, stream bool) (string, error)
}

func NewAssistantService(ar repository.AssistantRepository) AssistantService {
	return &assistantService{
		ar: ar,
	}
}

type assistantService struct {
	ar repository.AssistantRepository
}

func (s *assistantService) Chat(model string, messages []dto.AssistantMessage, stream bool) (string, error) {
	/*
			usage:
			curl https://api.deepseek.com/chat/completions \
		  -H "Content-Type: application/json" \
		  -H "Authorization: Bearer ${DEEPSEEK_API_KEY}" \
		  -d '{
		        "model": "deepseek-chat",
		        "messages": [
		          {"role": "system", "content": "You are a helpful assistant."},
		          {"role": "user", "content": "Hello!"}
		        ],
		        "stream": false
		      }'
	*/
	return "This is a placeholder response from the AssistantService.", nil
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
