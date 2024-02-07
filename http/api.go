package http

import (
	"context"
	"net/http"
	"os"

	"github.com/deividaspetraitis/ledger"
	db "github.com/deividaspetraitis/ledger/database/esdb"

	"github.com/deividaspetraitis/go/database/esdb"
	"github.com/deividaspetraitis/go/es"
	libhttp "github.com/deividaspetraitis/go/http"
	"github.com/deividaspetraitis/go/log"

	"github.com/gorilla/mux"
)

// API constructs an http.Handler with all application routes defined.
func API(shutdown chan os.Signal, cfg *Config, logger log.Logger, esclient *esdb.Client) http.Handler {
	// =========================================================================
	// Construct the web app api which holds all routes as well as common Middleware.

	api := libhttp.NewApp(shutdown)

	// =========================================================================
	// Construct and attach relevant handlers to web app api

	// POST /wallet creates a wallet.
	api.API.HandleFunc("/wallets", CreateWallet(func(ctx context.Context, req *ledger.CreateWalletRequest) (*ledger.Wallet, error) {
		return ledger.CreateWallet(ctx, func(ctx context.Context, aggregate es.Aggregate) error {
			return db.Save(ctx, esclient, aggregate)
		}, req)
	})).Methods(http.MethodPost)

	// GET /wallet/{id} retrieves a wallet.
	api.API.HandleFunc("/wallets/{id}", GetWallet(func(ctx context.Context, id string) (*ledger.Wallet, error) {
		return ledger.GetWallet(ctx, func(ctx context.Context, aggregate es.Aggregate, id string) (*ledger.WalletAggregate, error) {
			return db.Get[*ledger.WalletAggregate](ctx, esclient, aggregate, id)
		}, id)
	})).Methods(http.MethodGet)

	// POST /transactions creates a new transaction.
	api.API.HandleFunc("/transactions", CreateTransaction(func(ctx context.Context, req *ledger.TransactionRequest) (*ledger.Wallet, error) {
		return ledger.CreateTransaction(ctx, func(ctx context.Context, aggregate es.Aggregate) error {
			return db.Save(ctx, esclient, aggregate)
		}, func(ctx context.Context, aggregate es.Aggregate, id string) (*ledger.WalletAggregate, error) {
			return db.Get[*ledger.WalletAggregate](ctx, esclient, aggregate, id)
		}, req)
	})).Methods(http.MethodPost)

	router := mux.NewRouter()

	router.PathPrefix("/").Handler(api.API)

	return router
}
