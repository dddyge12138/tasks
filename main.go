package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"log"
	"task/config"
	"task/pkg/database"
)

func main() {
	// Load configuration
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("config")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file: %s", err)
	}

	var cfg config.Config
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("Unable to decode config into struct: %s", err)
	}

	// Initialize database
	err := database.NewPostgresDB(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize database: %s", err)
	}

	// Initialize Redis
	//redisClient, err := redis.NewRedisClient(cfg.Redis)
	//if err != nil {
	//	log.Fatalf("Failed to initialize Redis: %s", err)
	//}

	// Initialize Gin router
	r := gin.Default()

	// TODO: Setup routes

	// Start server
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %s", err)
	}
}
