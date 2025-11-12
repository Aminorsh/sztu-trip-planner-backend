package main

import (
	"github.com/Aminorsh/sztu-trip-planner-backend/config"
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/database"
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/utils"
)

func main() {
	config.LoadConfig()
	database.InitDB()

	code := utils.GenerateCode()
	println(code)
	expiry := utils.GenerateVerificationCodeExpiry()
	println("Expiry time:", expiry.String())

	err := utils.SendVerificationEmail("aminorsh@gmail.com", code)
	if err != nil {
		println("Error sending verification email:", err.Error())
	}

	// r := route.InitRoute(database.DB)

	// r.Run(config.AppConfig.Server.Port)
}
