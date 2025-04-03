package repository

import (
	"context"
	"github.com/lib/pq"
	"gorm.io/gorm"
	"task/internal/model"
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
	if err := r.db.WithContext(ctx).Table("tasks").Where("next_pending_time >= ? AND next_pending_time <= ?", startTime, endTime).Where("is_deleted = ?", 0).Find(&tasks).Error; err != nil {
		return tasks, err
	}
	return tasks, nil
}

func (r *taskRepository) UpdateTaskAfterProduce(ctx context.Context, task *model.Task) error {
	if err := r.db.WithContext(ctx).Exec("Update tasks set cron_task_ids = ?, task_produce_count = ?, next_pending_time = ?, is_deleted = ? where task_id = ?", pq.Int64Array(task.CronTaskIds), task.TaskProduceCount, task.NextPendingTime, task.IsDeleted, task.TaskId).Error; err != nil {
		return err
	}
	return nil
}
