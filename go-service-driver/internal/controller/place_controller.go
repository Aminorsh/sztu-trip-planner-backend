package controller

import (
	"log"

	"github.com/Aminorsh/sztu-trip-planner-backend/internal/dto"
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/errors"
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/middleware"
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/response"
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/service"
	"github.com/gin-gonic/gin"
)

type PlaceController struct {
	PlaceService *service.PlaceService
}

func NewPlaceController(placeService *service.PlaceService) *PlaceController {
	return &PlaceController{
		PlaceService: placeService,
	}
}

func (pc *PlaceController) SearchPlaces(ctx *gin.Context) {
	var req dto.PlaceSearchRequest

	// 使用 ShouldBindQuery 从 URL query parameters 绑定
	if err := ctx.ShouldBindQuery(&req); err != nil {
		log.Println("Error binding query params:", err)
		middleware.HandleError(ctx, errors.NewInvalidRequestError(err.Error()))
		return
	}

	resp, err := pc.PlaceService.SearchPlaces(ctx.Request.Context(), &req)
	if err != nil {
		middleware.HandleError(ctx, errors.NewInternalServerError(err))
		return
	}

	response.Success(ctx, resp)
}
