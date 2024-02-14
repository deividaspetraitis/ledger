package http

import (
	"context"
	"net/http"

	"github.com/deividaspetraitis/ledger"
	"github.com/deividaspetraitis/ledger/pkg/api/v1"

	libhttp "github.com/deividaspetraitis/go/http"
	"github.com/deividaspetraitis/go/log"
)

// createTaskFunc decouples actual check implementation and allows easily test HTTP handler.
type createTaskFunc[T any] func(context.Context, T) (string, error)

// CreateWallet handles HTTP requests for creating a new wallet.
// TODO
func CreateTask[T any, req api.Request[T]](createTask createTaskFunc[T]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// It's always json.
		w.Header().Set("Content-Type", "application/json")

		var request req
		if err := libhttp.UnmarshalRequest(r, request); err != nil {
			log.WithError(err).WithFields(log.Fields{
				"handler": "wallet",
				"method":  "CreateWallet",
			}).Println("unable to unmarshal request data")

			w.WriteHeader(http.StatusBadRequest)
			return
		}

		request = *new(req)
		log.Printf("request - - - %#v", request)

		id := ""
		var err error
		// id, err := createTask(r.Context(), &(request.Parse())
		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"handler": "wallet",
				"method":  "CreateWallet",
			}).Println("unable to create a wallet")

			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusAccepted)
		if err := libhttp.Marshal(w, api.NewCreateTaskResponse(id)); err != nil {
			log.WithError(err).WithFields(log.Fields{
				"handler": "wallet",
				"method":  "CreateWallet",
			}).Println("unable to marshal response data")

			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

// getTaskFunc decouples actual check implementation and allows easily test HTTP handler.
type getTaskFunc func(ctx context.Context, id string) (*ledger.Task, error)

// GetTask handles HTTP requests for retrieving a task by ID.
func GetTask(getTask getTaskFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// It's always json.
		w.Header().Set("Content-Type", "application/json")

		var request api.GetTaskRequest
		if err := libhttp.UnmarshalRequest(r, &request); err != nil {
			log.WithError(err).WithFields(log.Fields{
				"handler": "task",
				"method":  "GetTask",
			}).Println("unable to unmarshal request data")

			w.WriteHeader(http.StatusBadRequest)
			return
		}

		result, err := getTask(r.Context(), request.Parse())
		if err != nil {
			switch err {
			case ledger.ErrEntryNotFound:
				w.WriteHeader(http.StatusNotFound)
			default:
				log.WithError(err).WithFields(log.Fields{
					"handler": "task",
					"method":  "GetTask",
				}).Println("unable to retrieve a task")
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}

		w.WriteHeader(http.StatusOK)
		if err := libhttp.Marshal(w, api.NewTaskResponse(result)); err != nil {
			log.WithError(err).WithFields(log.Fields{
				"handler": "task",
				"method":  "GetTask",
			}).Println("unable to marshal response data")

			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}
