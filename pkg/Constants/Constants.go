package Constants

import "errors"

var ErrTaskNotFound = errors.New("task not found")

const (
	TaskSlotKey = "tasks:slot"
	TaskInfoKey = "tasks:%d"
	// 任务版本key, %s填充task_id
	TaskVersionKey = "tasks:version:%s"
)
