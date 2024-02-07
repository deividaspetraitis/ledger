package http

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/deividaspetraitis/ledger"

	"github.com/deividaspetraitis/go/errors"

	"github.com/gorilla/mux"
)

func TestCreateWallet(t *testing.T) {
}

func TestGetWallet(t *testing.T) {
	var testcases = []struct {
		id        string
		getWallet getWalletFunc

		response   string
		statusCode int
	}{
		// not a valid address
		{
			id: "-",
			getWallet: func(ctx context.Context, id string) (*ledger.Wallet, error) {
				return nil, nil
			},
			statusCode: http.StatusBadRequest,
		},
		// not found
		{
			id: "7b9db2f1-6777-4a9e-ac3d-efe0f7456a44",
			getWallet: func(ctx context.Context, id string) (*ledger.Wallet, error) {
				return nil, ledger.ErrEntryNotFound
			},
			response:   "",
			statusCode: http.StatusNotFound,
		},
		// found
		{
			id: "60c6d3f2-ada5-4723-b509-65ce0d595c33",
			getWallet: func(ctx context.Context, id string) (*ledger.Wallet, error) {
				return &ledger.Wallet{
					ID:      "60c6d3f2-ada5-4723-b509-65ce0d595c33",
					Name:    "test",
					Balance: 100,
				}, nil
			},
			response:   `{"id":"60c6d3f2-ada5-4723-b509-65ce0d595c33","name":"test","balance":100}`,
			statusCode: http.StatusOK,
		},
		// service error
		{
			id: "90cbd66a-4ba0-407d-8762-c8d4043cd680",
			getWallet: func(ctx context.Context, id string) (*ledger.Wallet, error) {
				return nil, errors.New("serivce error")
			},
			response:   "",
			statusCode: http.StatusInternalServerError,
		},
	}

	for i, tt := range testcases {
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost/wallet/%s", tt.id), nil)
		w := httptest.NewRecorder()

		// To add the vars to the context we need to create a router through which we can pass the request.
		// TODO: tests should be not aware of routing mechanism.
		router := mux.NewRouter()
		router.HandleFunc("/wallet/{id}", GetWallet(tt.getWallet))

		router.ServeHTTP(w, req)

		if statusCode := w.Result().StatusCode; statusCode != tt.statusCode {
			t.Errorf("#%d HTTP status got %v, want %v", i, statusCode, tt.statusCode)
		}

		// we do apply TrimSpace to clean up response coming from HTTP protocol
		if response := strings.TrimSpace(w.Body.String()); response != tt.response {
			t.Errorf("#%d HTTP status got %v, want %s", i, response, tt.response)
		}
	}
}
