package route

import (
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/controller"
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/middleware"
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func InitRoute(db *gorm.DB) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.CORS())

	// r.GET("/", func(c *gin.Context) {
	// 	c.JSON(http.StatusOK, gin.H{
	// 		"message": "SZTU Trip Planner Backend is running 🚀",
	// 	})
	// })

	userService := service.NewUserService(db)
	userController := controller.NewUserController(userService)

	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/send-code", userController.SendVerificationCode)
			auth.POST("/register", userController.RegisterUser)
<<<<<<< HEAD
=======
			auth.POST("/login", userController.LoginUser)
>>>>>>> ef157c0 (Reinitialize repository and keep local changes)
		}
	}

	return r
}
