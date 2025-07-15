package main

import (
	"log"
	"quote-voting-backend/config"
	"quote-voting-backend/models"
	"quote-voting-backend/routes"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file, using default values")
	}
	config.ConnectDB()

	// AutoMigrate
	config.DB.AutoMigrate(
		&models.User{},
		&models.Quote{},
		&models.Vote{},
		&models.QuoteOfTheDay{},
	)

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"}, // อนุญาต origin ของ frontend
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	routes.RegisterAuthRoutes(r)
	routes.QuoteRoutes(r)

	r.Run(":8080")
}
