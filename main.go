package main

import (
	"context"
	"fmt"
	"log"
	"task/api/router"
	"task/config"
	"task/internal/daemon"
	"task/pkg/database"
	"task/pkg/logger"
	"task/pkg/pulsar_queue"
	redis_db "task/pkg/redis"
	"task/pkg/worker_pool"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
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
		log.Fatalf("Failed to start server: %s", err)
	}
}

func InitAllComponents(ctx context.Context) (config.Config, *gin.Engine) {
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
	err = redis_db.NewRedisClient(cfg.Redis)
	if err != nil {
		log.Fatalf("Failed to initialize Redis: %s", err)
	}

	// Initialize Pulsar
	_, err = pulsar_queue.NewPulsarClient(cfg.Pulsar)
	if err != nil {
		log.Fatalf("Failed to initialize Pulsar: %s", err)
	}

	// Initialize logger
	err = logger.InitLogger(cfg.Log.Path)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %s", err)
	}

	// Initialize Gin router
	r := gin.Default()

	taskHandler, err := InitTaskHandler(database.Db)
	if err != nil {
		log.Fatalf("Failed to initialize task handler: %s", err)
	}
	router.RegisterRoutes(r, taskHandler)
	return cfg, r
}

func InitTaskProducer(ctx context.Context) {
	// 初始化pulsar客户端
	worker_pool.InitProducer()
	// 初始化生产者的协程池
	daemon.InitProducerWorkerPool(ctx)
	// 启动任务生产者
	go daemon.TaskProducerRun(ctx)
	// 主程序结束时关闭工作池
	defer daemon.ProducerWorkerPoolInstance.Stop()
}
