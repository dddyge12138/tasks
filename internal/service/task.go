package service

import (
	"context"
	"encoding/json"
	"github.com/robfig/cron/v3"
	"task/internal/model"
	"task/internal/repository"
	"time"
)

type TaskService interface {
	CreateTask(ctx context.Context, name string, cronExpr *string, params interface{}) (*model.Task, error)
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

func (s *taskService) CreateTask(ctx context.Context, name string, cronExpr *string, params interface{}) (*model.Task, error) {
	// 将参数转换为JSON
	paramsBytes, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	task := &model.Task{
		Name:   name,
		Status: 1, // 待执行状态
		Cron:   cronExpr,
		Params: paramsBytes,
	}

	// 如果是定时任务，解析cron表达式并设置下次执行时间
	if cronExpr != nil {
		schedule, err := s.parser.Parse(*cronExpr)
		if err != nil {
			return nil, err
		}
		task.NextPendingTime = schedule.Next(time.Now())
	} else {
		// 一次性任务，直接设置为当前时间
		task.NextPendingTime = time.Now()
	}

	// 保存到数据库
	if err := s.taskRepo.CreateTask(ctx, task); err != nil {
		return nil, err
	}

	return task, nil
}
