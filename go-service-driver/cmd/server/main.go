package main

import (
	"github.com/Aminorsh/sztu-trip-planner-backend/config"
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/database"
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/route"
)

func main() {
	config.LoadConfig()
	database.InitDB()
	database.InitRedis()

	r := route.InitRoute(database.DB)

	r.Run(config.AppConfig.Server.Port)
}
