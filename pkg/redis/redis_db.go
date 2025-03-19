package redis_db

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"strconv"
	"task/config"
	"task/internal/model"
	"task/pkg/Constants"
)

var RedisDb *redis.Client

func NewRedisClient(cfg config.RedisConfig) error {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// Test the connection
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to redis: %w", err)
	}
	RedisDb = client
	return nil
}

/*
*
在redis中获取某个时间段需要执行的任务
*/
func GetTasksList(ctx context.Context, startTime, endTime int64) ([]*model.Task, error) {
	var tasks []*model.Task
	zRangeCmd := RedisDb.ZRangeByScore(ctx, Constants.TaskSlotKey, &redis.ZRangeBy{
		Min: strconv.FormatInt(startTime, 10),
		Max: strconv.FormatInt(endTime, 10),
	})
	if zRangeCmd.Err() != nil {
		return tasks, zRangeCmd.Err()
	}
	taskKeyArr := []string{}
	for _, taskId := range zRangeCmd.Val() {
		taskKeyArr = append(taskKeyArr, fmt.Sprintf(Constants.TaskInfoKey, taskId))
	}
	mGetCmd := RedisDb.MGet(ctx, taskKeyArr...)
	if mGetCmd.Err() != nil {
		return tasks, mGetCmd.Err()
	}
	for _, task := range mGetCmd.Val() {
		if task != nil {
			tasks = append(tasks, task.(*model.Task))
		}
	}
	return tasks, nil
}
