package dto

type DeepseekChatRequest struct {
	Model    string             `json:"model"`
	Messages []AssistantMessage `json:"messages"`
	Stream   bool               `json:"stream"`
}

type DeepseekChatResponse struct {
	Choices []struct {
		Message AssistantMessage `json:"message"`
	} `json:"choices"`
}
