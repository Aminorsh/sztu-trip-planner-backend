package controller

import (
	"strconv"

	"github.com/Aminorsh/sztu-trip-planner-backend/internal/dto"
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/service"
	"github.com/gin-gonic/gin"
)

type TripController struct {
	TripService *service.TripService
}

func NewTripController(service *service.TripService) *TripController {
	return &TripController{
		TripService: service,
	}
}

func (tc *TripController) CreateTrip(c *gin.Context) {
	var req dto.CreateTripRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	userIDValue, exists := c.Get("userID")
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}
	userID := userIDValue.(int)

	tripResponse, err := tc.TripService.CreateTrip(userID, req)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, tripResponse)
}

func (tc *TripController) ListTrips(c *gin.Context) {
	userIDValue, exists := c.Get("userID")
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}
	userID := userIDValue.(int)

	search := c.Query("search")

	tripListResponse, err := tc.TripService.ListTrips(userID, search)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, tripListResponse)
}

func (tc *TripController) DeleteTrip(c *gin.Context) {
	userIDValue, exists := c.Get("userID")
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}
	userID := userIDValue.(int)

	tripIDParam := c.Param("id")
	tripID, err := strconv.Atoi(tripIDParam)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid trip ID"})
		return
	}

	if err := tc.TripService.DeleteTrip(userID, tripID); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Trip deleted successfully"})
}
