package repository

import (
	"context"
	"task/internal/model"

	"gorm.io/gorm"
)

type TaskRepository interface {
	CreateTask(ctx context.Context, task *model.Task) error
	RemoveTask(ctx context.Context, task *model.Task) error
	GetTaskById(ctx context.Context, taskId int64) (*model.Task, error)
	GetTasksByTime(ctx context.Context, startTime, endTime int64) ([]*model.Task, error)
	UpdateTaskAfterProduce(ctx context.Context, task *model.Task) error
}

type taskRepository struct {
	db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) TaskRepository {
	return &taskRepository{db: db}
}

func (r *taskRepository) CreateTask(ctx context.Context, task *model.Task) error {
	return r.db.WithContext(ctx).Create(task).Error
}

func (r *taskRepository) RemoveTask(ctx context.Context, task *model.Task) error {
	return r.db.WithContext(ctx).Table("tasks").Where("id = ?", task.ID).Update("is_deleted", 1).Error
}

func (r *taskRepository) GetTaskById(ctx context.Context, taskId int64) (*model.Task, error) {
	var task model.Task
	if err := r.db.WithContext(ctx).First(&task, taskId).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *taskRepository) GetTasksByTime(ctx context.Context, startTime, endTime int64) ([]*model.Task, error) {
	var tasks []*model.Task
	if err := r.db.WithContext(ctx).Where("next_pending_time >= ? AND next_pending_time <= ?", startTime, endTime).Find(&tasks).Error; err != nil {
		return tasks, err
	}
	return tasks, nil
}

func (r *taskRepository) UpdateTaskAfterProduce(ctx context.Context, task *model.Task) error {
	updateMap := map[string]interface{}{
		"version":            task.Version,
		"next_pending_time":  task.NextPendingTime,
		"task_produce_count": task.TaskProduceCount,
		"cron_task_ids":      task.CronTaskIds,
	}
	if err := r.db.WithContext(ctx).Where("task_id = ?", task.TaskId).Updates(updateMap).Error; err != nil {
		return err
	}
	return nil
}
