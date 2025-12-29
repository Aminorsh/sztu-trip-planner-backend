package main

import (
	"log"

	"github.com/Aminorsh/sztu-trip-planner-backend/config"
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/database"
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/route"
)

func main() {
	config.LoadConfig()
	database.InitDB()
	if err := database.AutoMigrate(database.DB); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}
	database.InitRedis()

	r := route.InitRoute(database.DB)

	r.Run(config.AppConfig.Server.Port)
}
