package http

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/medods/auth-service/internal/handler"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRoutes настраивает маршруты
func SetupRoutes(r *gin.Engine, authHandler *handler.AuthHandler) {

	r.Use(cors.Default()) // тупо для работы сваггера, на проде так нельзя)

	auth := r.Group("/auth")
	{
		auth.POST("/tokens", authHandler.GenerateTokens)
		auth.POST("/refresh", authHandler.RefreshTokens)
	}
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
