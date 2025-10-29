package main

import (
	"github.com/Aminorsh/sztu-trip-planner-backend/config"
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/database"
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/route"
)

func main() {
	database.InitDB()

	r := route.InitRoute(database.DB)

	port := config.GetServerPort()
	r.Run(":" + port)
}
