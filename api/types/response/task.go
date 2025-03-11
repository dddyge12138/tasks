package response

type CreateTaskResponse struct {
	TaskID int64 `json:"task_id"`
}

type RemoveTaskResponse struct {
	Msg     string
	Success bool
}
