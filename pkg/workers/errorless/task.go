package errorless

import "fmt"

type TaskNotFoundError struct {
	ID string
}

func (e *TaskNotFoundError) Error() string {
	return fmt.Sprintf("task: %s not found in current pool", e.ID)
}

func TaskNotFound(id string) *TaskNotFoundError {
	return &TaskNotFoundError{id}
}

type TaskAlreadyExistsError struct {
	ID string
}

func (e *TaskAlreadyExistsError) Error() string {
	return fmt.Sprintf("task: %s already exists in current pool, skipping it...", e.ID)
}

func TaskAlreadyExists(id string) *TaskAlreadyExistsError {
	return &TaskAlreadyExistsError{id}
}
