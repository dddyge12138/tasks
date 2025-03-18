package Constants

import "errors"

var ErrTaskNotFound = errors.New("task not found")

const (
	TaskSlotKey = "tasks:slot"
)
