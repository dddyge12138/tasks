package repository

import (
	"context"
	"task/config"
	"task/internal/model"
	"task/pkg/database"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	cfg := config.DatabaseConfig{
		Host:     "localhost",
		Port:     15432,
		User:     "root",
		Password: "123456",
		DBName:   "postgres",
	}

	err := database.NewPostgresDB(cfg)
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}

	db := database.Db
	// 清理测试数据
	err = db.Exec("UPDATE tasks SET is_deleted = 1").Error
	if err != nil {
		t.Fatalf("failed to clean test data: %v", err)
	}

	return db
}

func TestCreateTask(t *testing.T) {
	db := setupTestDB(t)
	repo := NewTaskRepository(db)

	now := time.Now()
	task := &model.Task{
		ID:              2,
		Name:            "Test Task",
		Status:          0,
		Cron:            "*/5 * * * *",
		NextPendingTime: now,
		Params:          []byte(`{"key": "value"}`),
		CronTaskIds:     []int64{1, 2, 3},
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	err := repo.CreateTask(context.Background(), task)
	assert.NoError(t, err)
	assert.NotZero(t, task.ID)

	// 验证数据是否正确保存
	var savedTask model.Task
	err = db.First(&savedTask, task.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, task.ID, savedTask.ID)
	assert.Equal(t, task.Name, savedTask.Name)
}

//func TestRemoveTask(t *testing.T) {
//	db := setupTestDB(t)
//	repo := NewTaskRepository(db)
//
//	now := time.Now()
//	// 创建一个测试任务
//	task := &model.Task{
//		Name:            "Test Task",
//		Status:          0,
//		Cron:            "*/5 * * * *",
//		NextPendingTime: now,
//		Params:          []byte(`{"key": "value"}`),
//		CronTaskIds:     []int64{1, 2, 3},
//		IsDeleted:       0,
//		CreatedAt:       now,
//		UpdatedAt:       now,
//	}
//	err := db.Create(task).Error
//	assert.NoError(t, err)
//
//	// 测试删除
//	err = repo.RemoveTask(context.Background(), task)
//	assert.NoError(t, err)
//
//	// 验证是否已被标记为删除
//	var updatedTask model.Task
//	err = db.First(&updatedTask, task.ID).Error
//	assert.NoError(t, err)
//	assert.Equal(t, int32(1), updatedTask.IsDeleted)
//}

//func TestGetTaskById(t *testing.T) {
//	db := setupTestDB(t)
//	repo := NewTaskRepository(db)
//
//	now := time.Now()
//	// 创建一个测试任务
//	expectedTask := &model.Task{
//		Name:            "Test Task",
//		Status:          0,
//		Cron:            "*/5 * * * *",
//		NextPendingTime: now,
//		Params:          []byte(`{"key": "value"}`),
//		CronTaskIds:     []int64{1, 2, 3},
//		IsDeleted:       0,
//		CreatedAt:       now,
//		UpdatedAt:       now,
//	}
//	err := db.Create(expectedTask).Error
//	assert.NoError(t, err)
//
//	// 测试获取
//	task, err := repo.GetTaskById(context.Background(), expectedTask.ID)
//	assert.NoError(t, err)
//	assert.NotNil(t, task)
//	assert.Equal(t, expectedTask.Title, task.Title)
//	assert.Equal(t, expectedTask.Content, task.Content)
//
//	// 测试获取不存在的任务
//	task, err = repo.GetTaskById(context.Background(), 9999)
//	assert.Error(t, err)
//	assert.Nil(t, task)
//}
