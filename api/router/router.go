package router

import (
	"github.com/gin-gonic/gin"
	"task/api/handler"
)

func RegisterRoutes(r *gin.Engine, taskHandler *handler.TaskHandler) {
	v1 := r.Group("/api/v1")
	{
		v1.POST("/createTask", taskHandler.CreateTask) // 添加任务
		v1.POST("/removeTask", taskHandler.RemoveTask) // 删除任务
	}
}
