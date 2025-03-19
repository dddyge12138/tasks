package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/robfig/cron/v3"
	"strconv"
	"task/api/types/request"
	"task/internal/model"
	"task/internal/repository"
	"task/pkg/Constants"
	"task/pkg/redis"
	"time"
)

type TaskService interface {
	CreateTask(ctx context.Context, req request.CreateTaskRequest) (*model.Task, error)
	RemoveTask(ctx context.Context, req request.RemoveTaskRequest) error
	LoadTask(ctx context.Context) error
}

type taskService struct {
	taskRepo repository.TaskRepository
	parser   cron.Parser
}

func NewTaskService(taskRepo repository.TaskRepository) TaskService {
	return &taskService{
		taskRepo: taskRepo,
		parser:   cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow),
	}
}

func (s *taskService) RemoveTask(ctx context.Context, req request.RemoveTaskRequest) error {
	if _, err := s.taskRepo.GetTaskById(ctx, req.TaskId); err != nil {
		return Constants.ErrTaskNotFound
	}
	return s.taskRepo.RemoveTask(ctx, &model.Task{
		TaskId: req.TaskId,
	})
}

func (s *taskService) CreateTask(ctx context.Context, req request.CreateTaskRequest) (*model.Task, error) {
	// 将参数转换为JSON
	paramsBytes, err := json.Marshal(req.Params)
	if err != nil {
		return nil, err
	}

	task := &model.Task{
		TaskId: req.TaskId,
		Name:   req.Name,
		Status: 1, // 待执行状态
		Cron:   req.Cron,
		Params: paramsBytes,
	}

	// 如果是定时任务，解析cron表达式并设置下次执行时间
	if len(req.Cron) != 0 {
		schedule, err := s.parser.Parse(req.Cron)
		if err != nil {
			return nil, err
		}
		task.NextPendingTime = int64(schedule.Next(time.Now()).Second())
	} else {
		// 一次性任务，直接设置为当前时间
		task.NextPendingTime = int64(time.Now().Second())
	}

	// 保存到数据库
	if err := s.taskRepo.CreateTask(ctx, task); err != nil {
		return nil, err
	}

	return task, nil
}

func (s *taskService) LoadTask(ctx context.Context) error {
	// 1 => 查询出未来两个小时内触发的任务
	now := time.Now()
	var tasks []*model.Task
	var err error
	tasks, err = s.taskRepo.GetTasksByTime(ctx, int64(now.Second()), int64(now.Add(2*time.Hour).Second()))
	if err != nil {
		return nil
	}
	// 2 => 加入到redis的有序集合
	var members []redis.Z
	pipe := redis_db.RedisDb.Pipeline()
	for _, task := range tasks {
		members = append(members, redis.Z{
			Score:  float64(task.NextPendingTime),
			Member: task.TaskId,
		})
		// 单独对每个task存到哈希表，键名就是task_id, 使用管道一次性写入
		pipe.Set(ctx, fmt.Sprintf(Constants.TaskInfoKey, strconv.FormatInt(task.TaskId, 10)), task, time.Hour)
	}
	pipe.ZAdd(ctx, Constants.TaskSlotKey, members...)
	_, err = pipe.Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}
