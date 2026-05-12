package main

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/xpartla/pharmacy-webapi/api"
	"github.com/xpartla/pharmacy-webapi/internal/db_service"
	"github.com/xpartla/pharmacy-webapi/internal/pharmacy_product"
)

func main() {
	log.Logger = zerolog.New(os.Stdout).With().
		Str("service", "pharmacy-product").
		Timestamp().
		Caller().
		Logger()

	logLevelStr := os.Getenv("LOG_LEVEL")
	level, err := zerolog.ParseLevel(strings.ToLower(logLevelStr))
	if err != nil {
		log.Warn().Str("LOG_LEVEL", logLevelStr).Msg("invalid log level, defaulting to info")
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	log.Info().Msg("Server started")
	port := os.Getenv("PHARMACY_API_PORT")
	if port == "" {
		port = "8080"
	}
	environment := os.Getenv("PHARMACY_API_ENVIRONMENT")
	if !strings.EqualFold(environment, "production") {
		gin.SetMode(gin.DebugMode)
	}

	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "PUT", "POST", "DELETE", "PATCH"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
		ExposeHeaders:    []string{""},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	dbService := db_service.NewMongoService[pharmacy_product.Pharmacy](db_service.MongoServiceConfig{})
	defer dbService.Disconnect(context.Background())
	engine.Use(func(ctx *gin.Context) {
		ctx.Set("db_service", dbService)
		ctx.Next()
	})

	handlers := &pharmacy_product.ApiHandleFunctions{
		PharmaciesAPI:         pharmacy_product.NewPharmaciesApi(),
		PharmacyCategoriesAPI: pharmacy_product.NewPharmacyCategoriesApi(),
		PharmacyProductsAPI:   pharmacy_product.NewPharmacyProductsApi(),
	}
	pharmacy_product.NewRouterWithGinEngine(engine, *handlers)
	engine.GET("/openapi", api.HandleOpenApi)

	if err := engine.Run(":" + port); err != nil {
		log.Fatal().Err(err).Msg("server stopped")
	}
}
