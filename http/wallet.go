package http

import (
	"context"
	"net/http"

	"github.com/deividaspetraitis/ledger"
	"github.com/deividaspetraitis/ledger/pkg/api/v1"

	libhttp "github.com/deividaspetraitis/go/http"
	"github.com/deividaspetraitis/go/log"
)

// createWalletFunc decouples actual check implementation and allows easily test HTTP handler.
type createWalletFunc func(context.Context, *ledger.CreateWalletRequest) (*ledger.Wallet, error)

// CreateWallet handles HTTP requests for creating a new wallet.
func CreateWallet(createWallet createWalletFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// It's always json.
		w.Header().Set("Content-Type", "application/json")

		var request api.CreateWalletRequest
		if err := libhttp.UnmarshalRequest(r, &request); err != nil {
			log.WithError(err).WithFields(log.Fields{
				"handler": "wallet",
				"method":  "CreateWallet",
			}).Println("unable to unmarshal request data")

			w.WriteHeader(http.StatusBadRequest)
			return
		}

		wallet, err := createWallet(r.Context(), request.Parse())
		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"handler": "wallet",
				"method":  "CreateWallet",
			}).Println("unable to create a wallet")

			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		if err := libhttp.Marshal(w, api.NewWalletResponse(wallet)); err != nil {
			log.WithError(err).WithFields(log.Fields{
				"handler": "wallet",
				"method":  "CreateWallet",
			}).Println("unable to marshal response data")

			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

type getWalletFunc func(ctx context.Context, id string) (*ledger.Wallet, error)

// GetWallet handles HTTP requests for retrieving a wallet by ID.
func GetWallet(getWallet getWalletFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// It's always json.
		w.Header().Set("Content-Type", "application/json")

		var request api.GetWalletRequest
		if err := libhttp.UnmarshalRequest(r, &request); err != nil {
			log.WithError(err).WithFields(log.Fields{
				"handler": "wallet",
				"method":  "GetWallet",
			}).Println("unable to unmarshal request data")

			w.WriteHeader(http.StatusBadRequest)
			return
		}

		result, err := getWallet(r.Context(), request.Parse())
		if err != nil {
			switch err {
			case ledger.ErrEntryNotFound:
				w.WriteHeader(http.StatusNotFound)
			default:
				log.WithError(err).WithFields(log.Fields{
					"handler": "wallet",
					"method":  "GetWallet",
				}).Println("unable to retrieve a wallet")
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}

		w.WriteHeader(http.StatusOK)
		if err := libhttp.Marshal(w, api.NewWalletResponse(result)); err != nil {
			log.WithError(err).WithFields(log.Fields{
				"handler": "wallet",
				"method":  "GetWallet",
			}).Println("unable to marshal response data")

			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}
