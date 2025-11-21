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

	userService := service.NewUserService(db)
	userController := controller.NewUserController(userService)

	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/send-code", userController.SendVerificationCode)
			auth.POST("/register", userController.RegisterUser)
			auth.POST("/login", userController.LoginUser)
			auth.POST("/login-email", userController.LoginUserByEmail)
			auth.POST("/send-forget-code", userController.SendForgetPasswordCode)
			auth.POST("/send-forget-code-by-username", userController.SendForgetPasswordCodeByUsername)
			auth.POST("/verify-forget-password", userController.VerifyForgetPassword)
		}

		trips := api.Group("/trips")
		trips.Use(middleware.AuthMiddleware())
		{
			// // Trip-related routes would go here
			// trips.GET("/")    // Get trip list
			// trips.GET("/")    // Search trips
			// trips.GET("/")    // Get trip details
			// trips.POST("/")   // Create a new trip
			// trips.DELETE("/") // Delete a trip
		}
	}

	return r
}
