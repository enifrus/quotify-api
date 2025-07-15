package routes

import (
	"quote-voting-backend/controllers"
	"quote-voting-backend/middleware"

	"github.com/gin-gonic/gin"
)

func QuoteRoutes(r *gin.Engine) {
	quote := r.Group("/quotes")
	quote.Use(middleware.AuthMiddleware())
	{
		quote.POST("", controllers.CreateQuote)
		quote.GET("", controllers.GetQuotes)
		quote.PUT("/:id", controllers.UpdateQuote)
		quote.POST("/:quoteId/vote", controllers.VoteQuote)
		quote.GET("/today", controllers.GetTodayVotes)
		quote.GET("/qotd", controllers.GetQuoteOfTheDay)
	}
}
