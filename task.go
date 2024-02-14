package ledger

import (
	"context"
	// "github.com/deividaspetraitis/go/database"
)

// Task represents scheduled operation to be completed asynchronously.
type Task struct {
	ID     string // Unique identifier
	Status string // Status of the job
}

// GetTaskFunc retrieves task from underlying database store.
// If task is not found an ErrEntityNotFound will be returned.
// TODO
type GetTaskFunc func(ctx context.Context, id string) (*Task, error)

// GetTask retrieves existing task based on given ID.
// If task does not exist in the system database.ErrEntityNotFound will be returned.
func GetTask(ctx context.Context, getTask GetTaskFunc, id string) (*Task, error) {
	task, err := getTask(ctx, id)
	if err != nil {
		return nil, err
	}
	return task, nil
}
