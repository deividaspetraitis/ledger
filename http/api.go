package http

import (
	"context"
	"net/http"
	"os"

	"github.com/deividaspetraitis/ledger"
	taskdb "github.com/deividaspetraitis/ledger/database/asynq"

	"github.com/deividaspetraitis/go/database/esdb"
	"github.com/deividaspetraitis/go/es"
	libhttp "github.com/deividaspetraitis/go/http"
	"github.com/deividaspetraitis/go/log"

	"github.com/gorilla/mux"
	"github.com/hibiken/asynq"
)

// API constructs an http.Handler with all application routes defined.
func API(shutdown chan os.Signal, cfg *Config, logger log.Logger, esclient *esdb.Client, asynqc *asynq.Client, asynqi *asynq.Inspector, cache *ledger.WithCache[*ledger.WalletAggregate]) http.Handler {
	// =========================================================================
	// Construct the web app api which holds all routes as well as common Middleware.

	api := libhttp.NewApp(shutdown)

	// =========================================================================
	// Construct and attach relevant handlers to web app api

	// POST /wallet creates a wallet.
	api.API.HandleFunc("/wallets", CreateWallet(func(ctx context.Context, req *ledger.CreateWalletRequest) (string, error) {
		return taskdb.ScheduleCreateWalletTask(ctx, asynqc, req)
	})).Methods(http.MethodPost)

	// GET /wallet/{id} retrieves a wallet.
	api.API.HandleFunc("/wallets/{id}", GetWallet(func(ctx context.Context, id string) (*ledger.Wallet, error) {
		aggregate, err := ledger.GetWallet(ctx, func(ctx context.Context, aggregate es.Aggregate, id string) (*ledger.WalletAggregate, error) {
			return cache.Get(id)
		}, id)
		return &aggregate.Wallet, err
	})).Methods(http.MethodGet)

	// POST /transactions creates a new transaction.
	api.API.HandleFunc("/transactions", CreateTransaction(func(ctx context.Context, req *ledger.TransactionRequest) (string, error) {
		return taskdb.ScheduleCreateTransactionTask(ctx, asynqc, req)
	})).Methods(http.MethodPost)

	// GET /jobs/{id} retrieves a job status.
	api.API.HandleFunc("/tasks/{id}", GetTask(func(ctx context.Context, id string) (*ledger.Task, error) {
		return ledger.GetTask(ctx, func(ctx context.Context, id string) (*ledger.Task, error) {
			return taskdb.GetTask(ctx, asynqi, id)
		}, id)
	})).Methods(http.MethodGet)

	router := mux.NewRouter()

	router.PathPrefix("/").Handler(api.API)

	return router
}
