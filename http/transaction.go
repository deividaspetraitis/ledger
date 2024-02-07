package http

import (
	"context"
	"net/http"

	"github.com/deividaspetraitis/ledger"
	"github.com/deividaspetraitis/ledger/pkg/api/v1"

	libhttp "github.com/deividaspetraitis/go/http"
	"github.com/deividaspetraitis/go/log"
)

// createTransactionFunc decouples actual check implementation and allows easily test HTTP handler.
type createTransactionFunc func(context.Context, *ledger.TransactionRequest) (*ledger.Wallet, error)

// CreateTransaction handles HTTP requests for creating a new transaction.
func CreateTransaction(createTransaction createTransactionFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// It's always json.
		w.Header().Set("Content-Type", "application/json")

		var request api.CreateTransactionRequest
		if err := libhttp.UnmarshalRequest(r, &request); err != nil {
			log.WithError(err).WithFields(log.Fields{
				"handler": "transaction",
				"method":  "CreateTransaction",
			}).Println("unable to unmarshal request data")

			w.WriteHeader(http.StatusBadRequest)
			return
		}

		_, err := createTransaction(r.Context(), request.Parse())
		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"handler": "transaction",
				"method":  "CreateTransaction",
			}).Println("error to processing transaction")

			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		if err := libhttp.Marshal(w, api.NewCreateTransactionResponse(&request)); err != nil {
			log.WithError(err).WithFields(log.Fields{
				"handler": "transaction",
				"method":  "CreateTransaction",
			}).Println("unable to marshal response data")

			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}
