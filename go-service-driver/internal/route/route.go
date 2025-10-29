package route

import (
	"net/http"

	"github.com/Aminorsh/sztu-trip-planner-backend/internal/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func InitRoute(db *gorm.DB) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.CORS())

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "SZTU Trip Planner Backend is running 🚀",
		})
	})

	return r
}
