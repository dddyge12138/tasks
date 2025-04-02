package redis_db

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"task/config"
	"task/internal/model"
	"task/pkg/Constants"
	"task/pkg/logger"

	"github.com/redis/go-redis/v9"
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
* 在redis中获取某个时间段需要执行的任务
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
		taskIdInt64, err := strconv.ParseInt(taskId, 10, 64)
		if err != nil {
			logger.Logger.WithField("task_id", taskId).Error("转换taskId错误，请检查")
			continue
		}
		taskKeyArr = append(taskKeyArr, fmt.Sprintf(Constants.TaskInfoKey, taskIdInt64))
	}
	mGetCmd := RedisDb.MGet(ctx, taskKeyArr...)
	if mGetCmd.Err() != nil {
		return tasks, mGetCmd.Err()
	}
	var tmpTask *model.Task
	for _, task := range mGetCmd.Val() {
		if task != nil {
			err := json.Unmarshal([]byte(task.(string)), &tmpTask)
			if err != nil {
				logger.Logger.WithError(err).Error("从redis中获取到无法反序列化的任务,请检查")
				continue
			}
			tasks = append(tasks, tmpTask)
		}
	}
	return tasks, nil
}

func RemoveSlotTasks(ctx context.Context, taskId int64) error {
	pipe := RedisDb.Pipeline()
	pipe.ZRem(ctx, Constants.TaskSlotKey, taskId)
	pipe.Del(ctx, fmt.Sprintf(Constants.TaskInfoKey, taskId))
	_, err := pipe.Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}
