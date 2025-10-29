package main

import (
	"github.com/Aminorsh/sztu-trip-planner-backend/config"
	"github.com/Aminorsh/sztu-trip-planner-backend/route"
)

func main() {
	config.InitDB()

	r := route.InitRoute(config.DB)

	r.Run(":8080")
}
