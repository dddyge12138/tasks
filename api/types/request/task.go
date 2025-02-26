package request

type CreateTaskRequest struct {
	Name   string      `json:"name" binding:"required"`
	Cron   *string    `json:"cron"`
	Params interface{} `json:"params" binding:"required"`
}
