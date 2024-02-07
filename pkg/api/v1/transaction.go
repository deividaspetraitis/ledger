package api

import (
	"encoding/json"
	"net/http"

	"github.com/deividaspetraitis/ledger"
)

// CreateTransactionRequest represents HTTP request for creating a wallet.
type CreateTransactionRequest struct {
	ID       int    `json:"id"`
	Type     string `json:"transaction"`
	WalletID string `json:"wallet_id"`
	Amount   int    `json:"amount"`
}

// Validate parses request fields and returns whether they contain valid data.
// Validate implements validator.Validator.
// TODO: implement more sophisticated rule
func (r *CreateTransactionRequest) Validate() error {
	if len(r.Type) < 3 {
		return ledger.ErrNotValidTransaction
	}

	if len(r.WalletID) < 3 {
		return ledger.ErrNotValidWalletID
	}

	if r.Amount == 0 {
		return ledger.ErrNotValidAmount
	}

	return nil
}

// Parse constructs and returns *ledger.TransactionRequest populated with information from the request.
func (r *CreateTransactionRequest) Parse() *ledger.TransactionRequest {
	return &ledger.TransactionRequest{
		Type:     r.Type,
		WalletID: r.WalletID,
		Amount:   r.Amount,
	}
}

// UnmarshalHTTP implements http.RequestUnmarshaler.
func (r *CreateTransactionRequest) UnmarshalHTTPRequest(req *http.Request) error {
	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&r); err != nil {
		return err
	}

	return r.Validate()
}

// NewCreateTransactionResponse constructs and returns CreateTransactionResponse.
func NewCreateTransactionResponse(req *CreateTransactionRequest) *CreateTransactionResponse {
	return &CreateTransactionResponse{
		Type:     req.Type,
		WalletID: req.WalletID,
		Amount:   req.Amount,
	}
}

// CreateTransactionResponse represents transaction response.
type CreateTransactionResponse struct {
	Type     string `json:"transaction"`
	WalletID string `json:"wallet_id"`
	Amount   int    `json:"amount"`
}

// MarshalHTTP implements http.Marshaler.
func (r *CreateTransactionResponse) MarshalHTTP(w http.ResponseWriter) error {
	return json.NewEncoder(w).Encode(r)
}
