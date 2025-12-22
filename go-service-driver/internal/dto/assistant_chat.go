package dto

type AssistantMessage struct {
	Role    string `json:"role" binding:"required"`
	Content string `json:"content" binding:"required"`
}

type AssistantChatRequest struct {
	Model    string             `json:"model" binding:"required"`
	Messages []AssistantMessage `json:"messages" binding:"required"`
	Stream   bool               `json:"stream"`
}

type AssistantChatResponse struct {
	Reply string `json:"reply"`
}
