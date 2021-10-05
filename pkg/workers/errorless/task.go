package errorless

import "fmt"

type TaskNotFoundError struct {
	Id string
}

func (e *TaskNotFoundError) Error() string {
	return fmt.Sprintf("task: %s not found in current pool", e.Id)
}

func TaskNotFound(id string) *TaskNotFoundError {
	return &TaskNotFoundError{id}
}

type TaskAlreadyExistsError struct {
	Id string
}

func (e *TaskAlreadyExistsError) Error() string {
	return fmt.Sprintf("task: %s already exists in current pool, skipping it...", e.Id)
}

func TaskAlreadyExists(id string) *TaskAlreadyExistsError {
	return &TaskAlreadyExistsError{id}
}
