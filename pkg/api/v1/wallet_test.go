package api

import (
	"testing"

	"github.com/deividaspetraitis/go/validator"
	"github.com/deividaspetraitis/ledger"
)

func TestCreateWalletRequest(t *testing.T) {
	var testcases = []struct {
		Name  string
		Error error
	}{
		{"", ledger.ErrNotValidWalletName},
		{"Random Name", nil},
	}

	for _, v := range testcases {
		req := CreateWalletRequest{
			Name: v.Name,
		}

		err := validator.Validate(&req)
		if err != v.Error {
			t.Errorf("got %v, want %v", err, v.Error)
		}
	}
}

func TestGetWalletRequest(t *testing.T) {
	var testcases = []struct {
		ID    string
		Error error
	}{
		{"", ledger.ErrNotValidWalletID},
		{"60f76d5d-f5d0-405b-b996-11c082d4e644", nil},
	}

	for _, v := range testcases {
		req := GetWalletRequest{
			ID: v.ID,
		}

		err := validator.Validate(&req)
		if err != v.Error {
			t.Errorf("got %v, want %v", err, v.Error)
		}
	}
}
