package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"task/api/router"
	"task/config"
	"task/pkg/database"
	"task/pkg/logger"
	"task/pkg/pulsar_queue"
	redis_db "task/pkg/redis"
)

func main() {
	// Load configuration
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, r := InitAllComponents(ctx)
	// 初始化任务生产者
	InitTaskProducer(ctx)

	// defer延迟执行的回调写在这里, 因为函数结束就会执行defer
	defer logger.CloseLogger()

	// Start server
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	if err := r.Run(addr); err != nil {
		logger.Logger.WithError(err).Fatal("Failed to start server")
	}
}

func InitAllComponents(ctx context.Context) (config.Config, *gin.Engine) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("config")

	if err := viper.ReadInConfig(); err != nil {
		panic("Error reading config file")
	}

	var cfg config.Config
	if err := viper.Unmarshal(&cfg); err != nil {
		panic("Unable to decode config into struct")
	}

	// Initialize logger
	err := logger.InitLogger(cfg.Log.Path)
	if err != nil {
		logger.Logger.WithError(err).Fatal("Failed to initialize logger")
	}

	// Initialize database
	err = database.NewPostgresDB(cfg.Database)
	if err != nil {
		logger.Logger.WithError(err).Fatal("Failed to initialize database")
	}

	// Initialize Redis
	err = redis_db.NewRedisClient(cfg.Redis)
	if err != nil {
		logger.Logger.WithError(err).Fatal("Failed to initialize Redis")
	}

	// Initialize Pulsar
	_, err = pulsar_queue.NewPulsarClient(cfg.Pulsar)
	if err != nil {
		logger.Logger.WithError(err).Fatal("Failed to initialize Pulsar")
	}

	// Initialize Gin router
	r := gin.Default()

	taskHandler, err := InitTaskHandler(database.Db)
	if err != nil {
		logger.Logger.WithError(err).Fatal("Failed to initialize task handler")
	}
	router.RegisterRoutes(r, taskHandler)
	return cfg, r
}

func InitTaskProducer(ctx context.Context) {

}
