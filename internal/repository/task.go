package repository

import (
	"context"

	"gorm.io/gorm"
	"e/Private/goSdk/goProject/yytest/task/internal/model"
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
