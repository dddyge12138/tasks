package model

import (
	"github.com/lib/pq"
	"time"
)

type Task struct {
	ID              int64         `json:"id" gorm:"primaryKey"`
	Name            string        `json:"name"`
	Status          int           `json:"status"`
	Cron            string        `json:"cron"`
	NextPendingTime time.Time     `json:"next_pending_time"`
	Params          []byte        `json:"params" gorm:"type:jsonb"`
	CronTaskIds     pq.Int64Array `json:"cron_task_ids" gorm:"type:bigint[]"`
	IsDeleted       int           `json:"is_deleted"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
}

type TaskExecution struct {
	ID            int64     `json:"id" gorm:"primaryKey"`
	TaskID        int64     `json:"task_id"`
	ExecutionTime time.Time `json:"execution_time"`
	Status        int       `json:"status"`
	ErrorMessage  string    `json:"error_message"`
	RetryCount    int       `json:"retry_count"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type TaskResult struct {
	ID         int64     `json:"id" gorm:"primaryKey"`
	TaskID     int64     `json:"task_id"`
	CronTaskID int64     `json:"cron_task_id"`
	Status     int       `json:"status"`
	Result     []byte    `json:"result" gorm:"type:jsonb"`
	CreatedAt  time.Time `json:"created_at"`
}
