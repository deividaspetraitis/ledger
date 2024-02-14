package asynq

import (
	"context"

	"github.com/deividaspetraitis/ledger"

	"github.com/deividaspetraitis/go/database"
	"github.com/deividaspetraitis/go/errors"
	"github.com/deividaspetraitis/go/log"

	"github.com/hibiken/asynq"
)

// ScheduleCreateTransactionTask creates asynq.Task for handling TransactionRequest and returns its ID.
func ScheduleCreateTransactionTask(ctx context.Context, asynqc *asynq.Client, req *ledger.TransactionRequest) (string, error) {
	return ScheduleTask(ctx, asynqc, TaskCreateTransaction, req)
}

func NewCreateTransactionProcessor(logger log.Logger, get database.GetAggregateFunc[*ledger.WalletAggregate], save database.SaveAggregateFunc) *CreateTransactionProcessor {
	return &CreateTransactionProcessor{
		logger:        logger,
		saveAggregate: save,
		getAggregate:  get,
	}
}

// CreateTransactionProcessor repesents CreateTransaction task processor.
type CreateTransactionProcessor struct {
	logger        log.Logger
	saveAggregate database.SaveAggregateFunc
	getAggregate  database.GetAggregateFunc[*ledger.WalletAggregate]
}

// ProcessTask implements asynq.Handler.
func (processor *CreateTransactionProcessor) ProcessTask(ctx context.Context, t *asynq.Task) error {
	req, err := Unmarshal[ledger.TransactionRequest](t)
	processor.logger.WithError(err).Println("processing task")
	if err != nil {
		return errors.Wrapf(err, "asynq.Unmarshal failed: %w", asynq.SkipRetry)
	}

	if _, err := ledger.CreateTransaction(ctx, processor.saveAggregate, processor.getAggregate, &req); err != nil {
		processor.logger.WithError(err).Println("processing task")
		return errors.Wrapf(err, "legder.CreateTransaction failed: %w", asynq.SkipRetry)
	}

	return nil
}
