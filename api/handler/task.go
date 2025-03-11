package handler

import (
	"net/http"
	"task/api/types/request"
	"task/api/types/response"
	"task/internal/service"

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
// @Router /api/v1/:id [post]
func (h *TaskHandler) CreateTask(c *gin.Context) {
	var req request.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task, err := h.taskService.CreateTask(c.Request.Context(), req.Name, req.Cron, req.Params)
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

}
