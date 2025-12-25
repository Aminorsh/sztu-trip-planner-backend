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
	r.Static("/static/avatars", "./uploads/avatars")
	r.Static("/static/trips", "./uploads/trips")

	r.Static("/static/system/avatars", "./assets/avatars")
	r.Static("/static/system/trips", "./assets/trips")

	userRepo := repository.NewUserRepository(db)
	placeRepo := repository.NewPlaceRepository(db)
	amapCacheRepo := repository.NewAmapPoiCacheRepository(db)
	routeRepo := repository.NewRouteRepository(db)
	tripRepo := repository.NewTripRepository(db)
	assistantRepo := repository.NewAssistantRepository(database.RedisClient)

	userService := service.NewUserService(userRepo)
	assistantService := service.NewAssistantService(
		assistantRepo,
		config.GetDeepseekAPIKey(),
		config.GetDeepseekAPIURL(),
	)
	placeService := service.NewPlaceService(
		config.GetAmapAPIKey(),
		config.GetAmapAPIURL(),
		database.RedisClient,
		placeRepo,
		amapCacheRepo,
		assistantService,
	)
	routeService := service.NewRouteService(
		routeRepo,
		tripRepo,
		config.GetAmapAPIKey(),
		config.GetAmapAPIURL(),
	)
	tripService := service.NewTripService(
		tripRepo,
		placeRepo,
	)

	userController := controller.NewUserController(userService)
	placeController := controller.NewPlaceController(placeService)
	routeController := controller.NewRouteController(routeService)
	tripController := controller.NewTripController(tripService)
	assistantController := controller.NewAssistantController(assistantService)

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
				users.PUT("/update-avatar", userController.UpdateAvatar)
			}

			places := v2.Group("/places")
			{
				places.GET("/search", placeController.SearchPlaces)
				places.GET("/:placeId/detail", placeController.GetPlaceDetail)
				places.POST("/:placeId/ai-description", placeController.GenerateAIDescription)
			}

			routes := v2.Group("/routes")
			{
				routes.POST("/plan", routeController.PlanRoute)
				routes.GET("/:id", routeController.GetRoute)
				routes.GET("", routeController.GetTripRoutes)
			}
		}

		v3 := api.Group("/v3")
		v3.Use(middleware.AuthMiddleware())
		{
			v3.GET("/health", func(ctx *gin.Context) {
				ctx.JSON(200, gin.H{
					"status": "ok",
				})
			})

			trips := v3.Group("/trips")
			{
				trips.POST("", tripController.CreateTrip)           // 创建行程（新增）
				trips.GET("", tripController.GetTrips)              // 获取行程列表（新增）
				trips.DELETE("/:tripId", tripController.DeleteTrip) // 删除行程（新增）
				trips.GET("/:tripId", tripController.GetTrip)
				trips.PUT("/:tripId", tripController.UpdateTrip)
				trips.POST("/:tripId/days/:dayId/items", tripController.AddTripItems)
				trips.DELETE("/:tripId/days/:dayId/items/:itemId", tripController.DeleteTripItem)
				trips.PUT("/:tripId/days/:dayId/items/:itemId", tripController.UpdateTripItem)
				trips.POST("/:tripId/days", tripController.AddTripDay)
				trips.DELETE("/:tripId/days/:dayId", tripController.DeleteTripDay)
				trips.PUT("/:tripId/cover", tripController.UpdateTripCover)
			}

			assistants := v3.Group("/assistants")
			{
				assistants.POST("/chat", assistantController.Chat)
			}
		}
	}

	return r
}
