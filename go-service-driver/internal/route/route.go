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

	tripService := service.NewTripService(db)
	tripController := controller.NewTripController(tripService)

	api := r.Group("/api")
	{
		api.GET("/health", func(ctx *gin.Context) {
			ctx.JSON(200, gin.H{
				"status": "ok",
			})
		})

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

		v2 := api.Group("/v2")
		v2.Use(middleware.AuthMiddleware())
		{
			v2.GET("/health", func(ctx *gin.Context) {
				ctx.JSON(200, gin.H{
					"status": "ok",
				})
			})

			users := v2.Group("/users")
			{
				users.GET("/profile", userController.GetUserProfile)
				users.PUT("/profile", userController.UpdateUserProfile)
				users.PUT("/change-password", userController.ChangePassword)
				users.DELETE("/delete-account", userController.DeleteAccount)
			}

			trips := v2.Group("/trips")
			{
				trips.POST("", tripController.CreateTrip)
				trips.GET("", tripController.ListTrips)
				trips.DELETE("/:id", tripController.DeleteTrip)
			}
		}
	}

	return r
}
