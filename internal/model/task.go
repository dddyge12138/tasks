package model

import (
	"time"

	"github.com/lib/pq"
)

type Task struct {
	ID               int64         `json:"id" gorm:"primaryKey;autoIncrement"`
	TaskId           int64         `json:"task_id"`
	Name             string        `json:"name"`
	Status           int           `json:"status"`
	Cron             string        `json:"cron"`
	NextPendingTime  int64         `json:"next_pending_time"`
	Params           []byte        `json:"params" gorm:"type:jsonb"`
	CronTaskIds      pq.Int64Array `json:"cron_task_ids" gorm:"type:bigint[]"`
	IsDeleted        int           `json:"is_deleted"`
	Version          int64         `json:"version"`
	TaskProduceCount int64         `json:"task_produce_count"`
	CreatedAt        time.Time     `json:"created_at"`
	UpdatedAt        time.Time     `json:"updated_at"`
}

type TaskExecution struct {
	ID            int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	TaskId        int64     `json:"task_id"`
	ExecutionTime time.Time `json:"execution_time"`
	Status        int       `json:"status"`
	ErrorMessage  string    `json:"error_message"`
	RetryCount    int       `json:"retry_count"`
	Version       int64     `json:"version"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type TaskResult struct {
	ID         int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	TaskId     int64     `json:"task_id"`
	CronTaskId int64     `json:"cron_task_id"`
	Status     int       `json:"status"`
	Version    int64     `json:"version"`
	Result     []byte    `json:"result" gorm:"type:jsonb"`
	CreatedAt  time.Time `json:"created_at"`
}

// pulsar中消息的格式
type TaskMessage struct {
	TaskId      int64   `json:"task_id"`
	Params      []byte  `json:"params"`
	Version     int64   `json:"version"`
	CronTaskIds []int64 `json:"cron_task_ids"`
}
