package repository

import (
	"context"
	"task/internal/model"

	"gorm.io/gorm"
)

type TaskRepository interface {
	CreateTask(ctx context.Context, task *model.Task) error
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
	return r.db.WithContext(ctx).Where("id = ?", task.ID).Update("is_deleted", 1).Error
}

func (r *taskRepository) GetTaskById(ctx context.Context, taskId int64) (*model.Task, error) {
	var task model.Task
	if err := r.db.WithContext(ctx).First(&task, taskId).Error; err != nil {
		return nil, err
	}
	return &task, nil
}
