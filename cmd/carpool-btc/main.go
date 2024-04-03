package main

import (
	"carpool-btc/internal/app/api/btc"
	"carpool-btc/internal/app/api/users"
	"carpool-btc/internal/app/middlewares"
	"carpool-btc/internal/app/utils"
	"github.com/gin-gonic/gin"
	"log"
	"os"
)

func init() {
	utils.LoadEnvironmentVariable()
	utils.ConnectToDb()
}

func main() {
	r := gin.Default()
	v1 := r.Group("api/v1/")

	userRoute := v1.Group("users")
	{
		userRoute.POST("/signup", users.Register)
		userRoute.POST("/login", users.Login)
	}

	authProtected := r.Group("/api/v1")
	authProtected.Use(middlewares.RequireAuth)

	authProtected.POST("deposit", btc.Deposit)
	authProtected.GET("list_unconfirmed", btc.ListUnConfirmed)

	authProtected.POST("withdraw", btc.Withdraw)

	err := r.Run(":" + os.Getenv("PORT"))

	if err != nil {
		log.Fatal("Failed to run server")
	}
}
