package asynq

import (
	"context"

	"github.com/deividaspetraitis/ledger"

	"github.com/deividaspetraitis/go/database"
	"github.com/deividaspetraitis/go/errors"
	"github.com/deividaspetraitis/go/log"

	"github.com/hibiken/asynq"
)

// ScheduleCreateWalletTask creates asynq.Task for handling CreateWalletRequest and returns its ID.
func ScheduleCreateWalletTask(ctx context.Context, asynqc *asynq.Client, req *ledger.CreateWalletRequest) (string, error) {
	return ScheduleTask(ctx, asynqc, TaskCreateWallet, req)
}

func NewCreateWalletProcessor(logger log.Logger, db database.SaveAggregateFunc) *CreateWalletProcessor {
	return &CreateWalletProcessor{
		logger: logger,
		db:     db,
	}
}

// CreateWalletProcessor repesents CreateWallet task processor.
type CreateWalletProcessor struct {
	logger log.Logger
	db     database.SaveAggregateFunc
}

// ProcessTask implements asynq.Handler.
func (processor *CreateWalletProcessor) ProcessTask(ctx context.Context, t *asynq.Task) error {
	processor.logger.Println("processing task")
	req, err := Unmarshal[ledger.CreateWalletRequest](t)
	if err != nil {
		return errors.Wrapf(err, "asynq.Unmarshal failed: %w", asynq.SkipRetry)
	}

	if _, err := ledger.CreateWallet(ctx, processor.db, &req); err != nil {
		return errors.Wrapf(err, "legder.CreateWallet failed: %w", asynq.SkipRetry)
	}
	return nil
}
