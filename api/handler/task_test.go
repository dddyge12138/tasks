package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"task/internal/model"
	"testing"
)

// MockTaskService 创建一个mock的service
type MockTaskService struct {
	mock.Mock
}

// CreateTask mock方法
func (m *MockTaskService) CreateTask(ctx context.Context, name string, cronExpr string, params interface{}) (*model.Task, error) {
	args := m.Called(ctx, name, cronExpr, params)
	return args.Get(0).(*model.Task), args.Error(1)
}

// RemoveTask mock方法
func (m *MockTaskService) RemoveTask(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func setupTestRouter(th *TaskHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	v1 := r.Group("/api/v1")
	{
		v1.POST("/createTask", th.CreateTask)
		v1.POST("/removeTask", th.RemoveTask)
	}

	return r
}

func TestTaskHandler_CreateTask(t *testing.T) {
	// 创建mock service
	mockService := new(MockTaskService)

	// 设置mock期望
	mockService.On("CreateTask", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&model.Task{
		ID:     1,
		Name:   "Test Task",
		Cron:   "1/* * * * *",
		Status: 1,
	}, nil)

	// 创建handler
	th := NewTaskHandler(mockService)

	// 设置路由
	r := setupTestRouter(th)

	// 创建测试数据
	testTask := map[string]interface{}{
		"name":        "Test Task",
		"description": "Test Description",
		"task_id":     1,
		"params": map[string]interface{}{
			"key": "value",
		},
	}

	jsonData, _ := json.Marshal(testTask)

	// 创建请求
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/createTask", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	// 执行请求
	r.ServeHTTP(w, req)

	// 验证结果
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestTaskHandler_RemoveTask(t *testing.T) {
	// 创建mock service
	mockService := new(MockTaskService)

	// 设置mock期望
	mockService.On("RemoveTask", mock.Anything).Return(nil)

	// 创建handler
	th := NewTaskHandler(mockService)

	// 设置路由
	r := setupTestRouter(th)

	// 创建测试数据
	removeRequest := map[string]interface{}{
		"task_id": "1",
	}
	jsonData, _ := json.Marshal(removeRequest)

	// 创建请求
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/removeTask", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	// 执行请求
	r.ServeHTTP(w, req)

	// 验证结果
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}
