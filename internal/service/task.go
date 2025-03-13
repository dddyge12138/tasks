package service

import (
	"context"
	"encoding/json"
	"github.com/robfig/cron/v3"
	"task/api/types/request"
	"task/internal/model"
	"task/internal/repository"
	"task/pkg/Constants"
	"time"
)

type TaskService interface {
	CreateTask(ctx context.Context, req request.CreateTaskRequest) (*model.Task, error)
	RemoveTask(ctx context.Context, req request.RemoveTaskRequest) error
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
