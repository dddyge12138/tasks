package handler

import (
	"errors"
	"net/http"
	"task/api/types/request"
	"task/api/types/response"
	"task/internal/service"
	"task/pkg/Constants"

	"github.com/gin-gonic/gin"
)

type TaskHandler struct {
	taskService service.TaskService
}

func NewTaskHandler(taskService service.TaskService) *TaskHandler {
	return &TaskHandler{
		taskService: taskService,
	}
}

// CreateTask godoc
// @Summary Create a new task
// @Description Create a new task with optional cron expression
// @Tags tasks
// @Accept json
// @Produce json
// @Param task body request.CreateTaskRequest true "Task info"
// @Success 200 {object} response.CreateTaskResponse
// @Router /api/v1/createTask [post]
func (h *TaskHandler) CreateTask(c *gin.Context) {
	var req request.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task, err := h.taskService.CreateTask(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.CreateTaskResponse{
		TaskID: task.ID,
	})
}

func (h *TaskHandler) RemoveTask(c *gin.Context) {
	var req request.RemoveTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := h.taskService.RemoveTask(c.Request.Context(), req)
	if errors.Is(err, Constants.ErrTaskNotFound) {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Task removed successfully"})
	return
}

// TODO 内网接口，后面移动到Internal
func (h *TaskHandler) LoadTask(c *gin.Context) {
	err := h.taskService.LoadTask(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	c.JSON(http.StatusOK, gin.H{"message": "Task loaded successfully"})
}
