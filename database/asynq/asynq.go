package asynq

import (
	"context"
	"encoding/json"
	"time"

	"github.com/deividaspetraitis/ledger"

	"github.com/deividaspetraitis/go/errors"
	"github.com/deividaspetraitis/go/log"
	"github.com/deividaspetraitis/go/validator"

	"github.com/hibiken/asynq"
)

type TaskType string

const (
	TaskCreateWallet      TaskType = "create:wallet"
	TaskCreateTransaction TaskType = "create:transaction"
)

func Unmarshal[T any](t *asynq.Task) (T, error) {
	var req T
	if err := json.Unmarshal(t.Payload(), &req); err != nil {
		return *new(T), errors.Wrapf(err, "json.Unmarshal failed: %w", asynq.SkipRetry)
	}

	v, ok := any(req).(validator.Validator)
	if !ok {
		return req, nil
	}

	return req, v.Validate()
}


// TODO
func GetTask(ctx context.Context, asynqi *asynq.Inspector, id string) (*ledger.Task, error) {
	info, err := asynqi.GetTaskInfo("default", id) // TODO: queue is hardcoded
	if err != nil {
		if errors.Is(err, asynq.ErrTaskNotFound) {
			return nil, ledger.ErrEntryNotFound
		}
		return nil, errors.Wrapf(err, "asynqi.GetTaskInfo failed")
	}
	return &ledger.Task{
		ID:     id,
		Status: info.State.String(),
	}, nil
}

func ScheduleTask[T any](ctx context.Context, asynqc *asynq.Client, t TaskType, req T) (string, error) {
	task, err := newTask(t, req)
	if err != nil {
		return "", errors.Wrap(err, "asynq.newTask failed")
	}

	info, err := asynqc.EnqueueContext(ctx, task)
	if err != nil {
		return "", errors.Wrap(err, "asynq.EnqueueContext failed")
	}
	log.Printf("enqueued task: id=%s queue=%s", info.ID, info.Queue)

	return info.ID, nil

}

// newTask constructs and returns new task with req payload.
// If req implements validator.Validator req will be validated first.
func newTask[T any](task TaskType, req T) (*asynq.Task, error) {
	v, ok := any(req).(validator.Validator)
	if ok {
		if err := validator.Validate(v); err != nil {
			return nil, err
		}
	}

	payload, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	return asynq.NewTask(
		string(task),
		payload,
		asynq.MaxRetry(5),
		asynq.Timeout(20*time.Minute),
		asynq.Retention(5*time.Minute),
	), nil
}
