package controller

import (
	"strconv"

	"github.com/Aminorsh/sztu-trip-planner-backend/internal/dto"
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/errors"
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/middleware"
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/response"
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/service"
	"github.com/gin-gonic/gin"
)

type TripController struct {
	tripService *service.TripService
}

func NewTripController(tripService *service.TripService) *TripController {
	return &TripController{
		tripService: tripService,
	}
}

// CreateTrip 创建新行程
// POST /api/trips
func (c *TripController) CreateTrip(ctx *gin.Context) {
	var req dto.CreateTripRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(ctx, errors.NewInvalidRequestError(err.Error()))
		return
	}

	// 从上下文获取用户ID（需要auth middleware）
	userID := ctx.GetUint64("userID")

	trip, err := c.tripService.CreateTrip(ctx.Request.Context(), userID, &req)
	if err != nil {
		middleware.HandleError(ctx, err)
		return
	}

	response.Created(ctx, trip)
}

// GetTrips 获取用户的行程列表
// GET /api/trips
func (c *TripController) GetTrips(ctx *gin.Context) {
	userID := ctx.GetUint64("userID")

	search := ctx.Query("search")

	trips, err := c.tripService.GetTrips(ctx.Request.Context(), userID, search)
	if err != nil {
		middleware.HandleError(ctx, err)
		return
	}

	response.Success(ctx, trips)
}

// GetTrip 获取行程详情
// GET /api/trips/:tripId
func (c *TripController) GetTrip(ctx *gin.Context) {
	tripID := ctx.Param("tripId")

	trip, err := c.tripService.GetTrip(ctx.Request.Context(), tripID)
	if err != nil {
		middleware.HandleError(ctx, err)
		return
	}

	response.Success(ctx, trip)
}

// UpdateTrip 更新行程
// PUT /api/trips/:tripId
func (c *TripController) UpdateTrip(ctx *gin.Context) {
	tripID := ctx.Param("tripId")

	var req dto.UpdateTripRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(ctx, errors.NewInvalidRequestError(err.Error()))
		return
	}

	if err := c.tripService.UpdateTrip(ctx.Request.Context(), tripID, &req); err != nil {
		middleware.HandleError(ctx, err)
		return
	}

	response.Success(ctx, gin.H{
		"success": true,
		"message": "行程已保存",
	})
}

// AddTripItems 添加行程项
// POST /api/trips/:tripId/days/:dayId/items
func (c *TripController) AddTripItems(ctx *gin.Context) {
	tripID := ctx.Param("tripId")
	dayID := ctx.Param("dayId")

	dayNumber, err := strconv.Atoi(dayID)
	if err != nil {
		middleware.HandleError(ctx, errors.NewInvalidRequestError("invalid day ID"))
		return
	}

	var req dto.AddTripItemRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(ctx, errors.NewInvalidRequestError(err.Error()))
		return
	}

	itemIDs, err := c.tripService.AddTripItems(ctx.Request.Context(), tripID, dayNumber, &req)
	if err != nil {
		middleware.HandleError(ctx, err)
		return
	}

	response.Success(ctx, gin.H{
		"success": true,
		"itemId":  itemIDs[0], // 返回第一个项的ID
	})
}

// UpdateTripItem 更新行程项
// PUT /api/trips/:tripId/days/:dayId/items/:itemId
func (c *TripController) UpdateTripItem(ctx *gin.Context) {
	tripID := ctx.Param("tripId")
	dayID := ctx.Param("dayId")
	itemID := ctx.Param("itemId")

	var req dto.UpdateTripItemRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(ctx, errors.NewInvalidRequestError(err.Error()))
		return
	}

	if err := c.tripService.UpdateTripItem(ctx.Request.Context(), tripID, dayID, itemID, &req); err != nil {
		middleware.HandleError(ctx, err)
		return
	}

	response.Success(ctx, gin.H{
		"success": true,
	})
}

// DeleteTripItem 删除行程项
// DELETE /api/trips/:tripId/days/:dayId/items/:itemId
func (c *TripController) DeleteTripItem(ctx *gin.Context) {
	tripID := ctx.Param("tripId")
	dayID := ctx.Param("dayId")
	itemID := ctx.Param("itemId")

	if err := c.tripService.DeleteTripItem(ctx.Request.Context(), tripID, dayID, itemID); err != nil {
		middleware.HandleError(ctx, err)
		return
	}

	response.Success(ctx, gin.H{
		"success": true,
	})
}

// AddTripDay 添加新的一天
// POST /api/trips/:tripId/days
func (c *TripController) AddTripDay(ctx *gin.Context) {
	tripID := ctx.Param("tripId")

	var req dto.AddTripDayRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(ctx, errors.NewInvalidRequestError(err.Error()))
		return
	}

	dayCount, err := c.tripService.AddTripDay(ctx.Request.Context(), tripID, &req)
	if err != nil {
		middleware.HandleError(ctx, err)
		return
	}

	response.Success(ctx, gin.H{
		"success":  true,
		"dayCount": dayCount,
	})
}

// DeleteTripDay 删除指定天
// DELETE /api/trips/:tripId/days/:dayId
func (c *TripController) DeleteTripDay(ctx *gin.Context) {
	tripID := ctx.Param("tripId")
	dayID := ctx.Param("dayId")

	dayNumber, err := strconv.Atoi(dayID)
	if err != nil {
		middleware.HandleError(ctx, errors.NewInvalidRequestError("invalid day ID"))
		return
	}

	if err := c.tripService.DeleteTripDay(ctx.Request.Context(), tripID, dayNumber); err != nil {
		middleware.HandleError(ctx, err)
		return
	}

	response.Success(ctx, gin.H{
		"success": true,
	})
}

// DeleteTrip 删除行程
// DELETE /api/trips/:tripId
func (c *TripController) DeleteTrip(ctx *gin.Context) {
	tripID := ctx.Param("tripId")
	userID := ctx.GetUint64("userID")

	if err := c.tripService.DeleteTrip(ctx.Request.Context(), tripID, userID); err != nil {
		middleware.HandleError(ctx, err)
		return
	}

	response.Success(ctx, gin.H{
		"success": true,
	})
}

func (c *TripController) UpdateTripCover(ctx *gin.Context) {
	tripID := ctx.Param("tripId")

	file, err := ctx.FormFile("cover_image")
	if err != nil {
		middleware.HandleError(ctx, errors.NewInvalidRequestError("failed to get cover image file"))
		return
	}

	coverURL, err := c.tripService.UpdateTripCoverImage(ctx, tripID, file)
	if err != nil {
		middleware.HandleError(ctx, err)
		return
	}

	response.Success(ctx, gin.H{
		"cover_image_url": coverURL,
	})
}

// CheckoutTripItem 打卡行程项（标记为已完成）
// POST /api/trips/:tripId/days/:dayId/items/:itemId/checkout
func (c *TripController) CheckoutTripItem(ctx *gin.Context) {
	tripID := ctx.Param("tripId")
	dayID := ctx.Param("dayId")
	itemID := ctx.Param("itemId")

	if err := c.tripService.CheckoutTripItem(ctx.Request.Context(), tripID, dayID, itemID); err != nil {
		middleware.HandleError(ctx, err)
		return
	}

	response.Success(ctx, gin.H{
		"success": true,
		"message": "打卡成功",
	})
}

// UncheckoutTripItem 取消打卡行程项（标记为未完成）
// POST /api/trips/:tripId/days/:dayId/items/:itemId/uncheckout
func (c *TripController) UncheckoutTripItem(ctx *gin.Context) {
	tripID := ctx.Param("tripId")
	dayID := ctx.Param("dayId")
	itemID := ctx.Param("itemId")

	if err := c.tripService.UncheckoutTripItem(ctx.Request.Context(), tripID, dayID, itemID); err != nil {
		middleware.HandleError(ctx, err)
		return
	}

	response.Success(ctx, gin.H{
		"success": true,
		"message": "已取消打卡",
	})
}
