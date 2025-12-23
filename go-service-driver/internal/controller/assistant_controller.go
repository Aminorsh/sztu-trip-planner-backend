package controller

import (
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/dto"
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/errors"
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/middleware"
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/response"
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/service"
	"github.com/gin-gonic/gin"
)

type AssistantController struct {
	assistantService *service.AssistantService
}

func NewAssistantController(s *service.AssistantService) *AssistantController {
	return &AssistantController{
		assistantService: s,
	}
}

func (ac *AssistantController) Chat(ctx *gin.Context) {
	userID := ctx.GetUint64("userID")

	var req dto.AssistantChatRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(ctx, errors.NewInvalidRequestError(err.Error()))
		return
	}

	messages := []dto.AssistantMessage{
		{
			Role:    "user",
			Content: req.Messages,
		},
	}

	reply, err := ac.assistantService.Chat(ctx.Request.Context(), userID, req.Model, messages, req.Stream)
	if err != nil {
		middleware.HandleError(ctx, err)
		return
	}

	response.Success(ctx, reply)
}
