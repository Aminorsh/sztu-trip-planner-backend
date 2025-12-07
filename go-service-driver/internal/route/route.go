package route

import (
	"github.com/Aminorsh/sztu-trip-planner-backend/config"
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/controller"
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/database"
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/middleware"
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/repository"
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func InitRoute(db *gorm.DB) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.CORS())

	userRepo := repository.NewUserRepository(db)

	userService := service.NewUserService(userRepo)
	placeService := service.NewPlaceService(
		config.GetAmapAPIKey(),
		config.GetAmapAPIURL(),
		database.RedisClient,
		db,
	)
	tripService := service.NewTripService(db)

	userController := controller.NewUserController(userService)
	placeController := controller.NewPlaceController(placeService)
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

			places := v2.Group("/places")
			{
				places.POST("/search", placeController.SearchPlaces)
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
