package repository

import (
	"context"
	"fmt"
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
	//err = db.Exec("TRUNCATE tasks RESTART IDENTITY CASCADE;").Error
	err = db.Exec("Update tasks set is_deleted = 1;").Error
	if err != nil {
		t.Fatalf("failed to reset sequence: %v", err)
	}

	return db
}

func TestCreateTask(t *testing.T) {
	db := setupTestDB(t)
	repo := NewTaskRepository(db)

	now := time.Now()
	task := &model.Task{
		TaskId:          1111231233,
		Name:            "Test Task",
		Status:          0,
		Cron:            "*/5 * * * *",
		NextPendingTime: now.Unix(),
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

func TestRemoveTask(t *testing.T) {
	db := setupTestDB(t)
	repo := NewTaskRepository(db)

	now := time.Now()
	// 创建一个测试任务
	task := &model.Task{
		TaskId:          2222222,
		Name:            "Test Task",
		Status:          0,
		Cron:            "*/5 * * * *",
		NextPendingTime: now.Unix(),
		Params:          []byte(`{"key": "value"}`),
		CronTaskIds:     []int64{1, 2, 3},
		IsDeleted:       0,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	err := db.Create(task).Error
	assert.NoError(t, err)

	// 测试删除
	err = repo.RemoveTask(context.Background(), task)
	assert.NoError(t, err)

	// 验证是否已被标记为删除
	var updatedTask model.Task
	err = db.First(&updatedTask, task.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, 1, updatedTask.IsDeleted)
}

func TestGetTaskById(t *testing.T) {
	db := setupTestDB(t)
	repo := NewTaskRepository(db)

	now := time.Now()
	// 创建一个测试任务
	expectedTask := &model.Task{
		TaskId:          1111111,
		Name:            "Test Task",
		Status:          0,
		Cron:            "*/5 * * * *",
		NextPendingTime: now.Unix(),
		Params:          []byte(`{"key": "value"}`),
		CronTaskIds:     []int64{1, 2, 3},
		IsDeleted:       0,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	err := db.Create(expectedTask).Error
	assert.NoError(t, err)

	// 测试获取
	task, err := repo.GetTaskById(context.Background(), expectedTask.ID)
	assert.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, expectedTask.Name, task.Name)

	// 测试获取不存在的任务
	task, err = repo.GetTaskById(context.Background(), 9999)
	assert.Error(t, err)
	assert.Nil(t, task)
}

func TestCreateCronTasksForEvening(t *testing.T) {
	db := setupTestDB(t)
	repo := NewTaskRepository(db)

	// 获取当前日期
	now := time.Now()

	// 创建日期（yyyy-mm-dd格式）
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// 设置任务名称前缀
	taskNamePrefix := "Evening Task"

	// 创建晚上6点到12点的cron任务
	cronTimes := []struct {
		hour   int
		minute int
		cron   string
	}{
		{18, 0, "0 0 18 * * *"}, // 晚上6点
		{19, 0, "0 0 19 * * *"}, // 晚上7点
		{20, 0, "0 0 20 * * *"}, // 晚上8点
		{21, 0, "0 0 21 * * *"}, // 晚上9点
		{22, 0, "0 0 22 * * *"}, // 晚上10点
		{23, 0, "0 0 23 * * *"}, // 晚上11点
		{0, 0, "0 0 0 * * *"},   // 晚上12点
	}

	// 创建10个任务，包括上面定义的7个不同时间的cron，以及3个随机时间
	for i := 0; i < 7; i++ {
		var cronStr string
		var taskTime time.Time

		// 使用预定义的cron时间
		cronStr = cronTimes[i].cron

		// 设置下一次执行时间为今天晚上的对应时间点
		taskTime = today.Add(time.Duration(cronTimes[i].hour) * time.Hour).Add(time.Duration(cronTimes[i].minute) * time.Minute)

		// 创建任务
		task := &model.Task{
			TaskId:          int64(1000 + i),
			Name:            fmt.Sprintf("%s %d", taskNamePrefix, i+1),
			Cron:            cronStr,
			NextPendingTime: taskTime.Unix(),
			Params:          []byte(fmt.Sprintf(`{"task_index": %d}`, i+1)),
			CreatedAt:       now,
			UpdatedAt:       now,
		}

		// 保存到数据库
		err := repo.CreateTask(context.Background(), task)
		assert.NoError(t, err)

		t.Logf("Created task %d: %s, cron: %s, next execution: %s",
			i+1, task.Name, task.Cron, time.Unix(task.NextPendingTime, 0).Format("2006-01-02 15:04:05"))
	}

	// 验证是否创建了10个任务
	var count int64
	err := db.Model(&model.Task{}).Where("name LIKE ?", taskNamePrefix+"%").Count(&count).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(7), count)

}
