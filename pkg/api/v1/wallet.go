package api

import (
	"encoding/json"
	"net/http"

	"github.com/deividaspetraitis/ledger"

	"github.com/gorilla/mux"
)

// Wallet represents API response Wallet entity.
type Wallet struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Balance int    `json:"balance"`
}

// NewWalletResponse constructs and returns response Wallet entity.
func NewWalletResponse(w *ledger.Wallet) *Wallet {
	return &Wallet{
		ID:      w.ID,
		Name:    w.Name,
		Balance: w.Balance,
	}
}

// MarshalHTTP implements http.Marshaler.
func (r *Wallet) MarshalHTTP(w http.ResponseWriter) error {
	return json.NewEncoder(w).Encode(r)
}

// CreateWalletRequest represents HTTP request for creating a new wallet.
type CreateWalletRequest struct {
	Name string `json:"name"` // Wallet name
}

// Validate validates request data and returns an error if it's not a valid.
// TODO: implement more sophisticated rule
func (r *CreateWalletRequest) Validate() error {
	if len(r.Name) == 0 {
		return ledger.ErrNotValidWalletName
	}
	return nil
}

// UnmarshalHTTP implements http.RequestUnmarshaler.
func (r *CreateWalletRequest) UnmarshalHTTPRequest(req *http.Request) error {
	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&r); err != nil {
		return err
	}
	return r.Validate()
}

// Parse constructs and returns *ledger.CreateWalletRequest populated with information from the request.
func (r *CreateWalletRequest) Parse() *ledger.CreateWalletRequest {
	return &ledger.CreateWalletRequest{
		Name: r.Name,
	}
}

// GetWalletRequest represents HTTP request for retrieving a wallet.
type GetWalletRequest struct {
	ID string `json:"id"` // Wallet ID
}

// Validate validates request data and returns an error if it's not a valid.
// Validate implements validator.Validator.
// TODO: implement more sophisticated rule
func (r *GetWalletRequest) Validate() error {
	if len(r.ID) < 3 {
		return ledger.ErrNotValidWalletID
	}
	return nil
}

// UnmarshalHTTP implements http.RequestUnmarshaler.
func (r *GetWalletRequest) UnmarshalHTTPRequest(req *http.Request) error {
	r.ID = mux.Vars(req)["id"]
	return r.Validate()
}

// Parse parses and returns Wallet ID from the request.
func (r *GetWalletRequest) Parse() string {
	return r.ID
}
