package controller

import (
	"strconv"

	"github.com/Aminorsh/sztu-trip-planner-backend/internal/dto"
	apperrors "github.com/Aminorsh/sztu-trip-planner-backend/internal/errors"
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/middleware"
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/response"
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/service"
	"github.com/gin-gonic/gin"
)

type RouteController struct {
	routeService *service.RouteService
}

func NewRouteController(routeService *service.RouteService) *RouteController {
	return &RouteController{
		routeService: routeService,
	}
}

// PlanRoute 规划路线
func (rc *RouteController) PlanRoute(ctx *gin.Context) {
	var req dto.RoutePlanRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(ctx, apperrors.NewInvalidRequestError(err.Error()))
		return
	}

	resp, err := rc.routeService.PlanRoute(ctx.Request.Context(), &req)
	if err != nil {
		middleware.HandleError(ctx, err)
		return
	}

	response.Success(ctx, resp)
}

// GetRoute 获取路线详情
func (rc *RouteController) GetRoute(ctx *gin.Context) {
	routeID, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		middleware.HandleError(ctx, apperrors.NewInvalidRequestError("无效的路线ID"))
		return
	}

	route, err := rc.routeService.GetRouteByID(ctx.Request.Context(), routeID)
	if err != nil {
		middleware.HandleError(ctx, err)
		return
	}

	if route == nil {
		middleware.HandleError(ctx, apperrors.NewInvalidRequestError("路线不存在"))
		return
	}

	response.Success(ctx, route)
}

// GetTripRoutes 获取行程的所有路线
func (rc *RouteController) GetTripRoutes(ctx *gin.Context) {
	tripID, err := strconv.ParseUint(ctx.Query("trip_id"), 10, 64)
	if err != nil {
		middleware.HandleError(ctx, apperrors.NewInvalidRequestError("无效的行程ID"))
		return
	}

	routes, err := rc.routeService.GetRoutesByTripID(ctx.Request.Context(), tripID)
	if err != nil {
		middleware.HandleError(ctx, err)
		return
	}

	response.Success(ctx, routes)
}
