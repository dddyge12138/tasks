package request

type CreateTaskRequest struct {
	TaskId int64       `json:"task_id" binding:"required"`
	Name   string      `json:"name" binding:"required"`
	Cron   string      `json:"cron"`
	Params interface{} `json:"params" binding:"required"`
}

type RemoveTaskRequest struct {
	TaskId int64 `json:"task_id" binding:"required"`
}
