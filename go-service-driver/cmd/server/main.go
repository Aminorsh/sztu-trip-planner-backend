package main

import (
	"github.com/Aminorsh/sztu-trip-planner-backend/config"
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/database"
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/route"
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/utils"
)

func main() {
	config.LoadConfig()
	database.InitDB()
	database.InitRedis()

	code := utils.GenerateVerificationCode()
	println(code)

	// err := utils.SendVerificationEmail("aminorsh@gmail.com", code)
	// if err != nil {
	// 	println("Error sending verification email:", err.Error())
	// }

	r := route.InitRoute(database.DB)

	r.Run(config.AppConfig.Server.Port)
}
