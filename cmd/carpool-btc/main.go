package main

import (
	"carpool-btc/internal/app/api/users"
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
	}

	err := r.Run(":" + os.Getenv("PORT"))

	if err != nil {
		log.Fatal("Failed to run server")
	}
}
