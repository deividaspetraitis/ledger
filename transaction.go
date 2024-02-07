package ledger

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/deividaspetraitis/go/database"
	"github.com/deividaspetraitis/go/errors"
	"github.com/deividaspetraitis/go/validator"

	"golang.org/x/exp/slices"
)

// Common transaction types.
const (
	TransactionDeposit  = "DEPOSIT"
	TransactionWithdraw = "WITHDRAW"
)

// Deposit represents wallet deposit transaction event.
type Deposit struct {
	WalletID string
	Amount   int
}

// Implements es.MarshalUnmarshaler
func (d *Deposit) UnmarshalJSON(b []byte) error {
	type deposit Deposit
	temp := deposit(*d)
	if err := json.Unmarshal(b, &temp); err != nil {
		return err
	}
	*d = Deposit(temp)
	return nil
}

// Implements es.MarshalUnmarshaler
func (d *Deposit) MarshalJSON() ([]byte, error) {
	type deposit Deposit
	temp := deposit(*d)
	return json.Marshal(temp)
}

// Withdraw represents wallet deposit transaction event.
type Withdraw struct {
	WalletID string
	Amount   int
}

// Implements es.MarshalUnmarshaler
func (w *Withdraw) UnmarshalJSON(b []byte) error {
	type withdraw Withdraw
	temp := withdraw(*w)
	if err := json.Unmarshal(b, &temp); err != nil {
		return err
	}
	*w = Withdraw(temp)
	return nil
}

// Implements es.MarshalUnmarshaler
func (w Withdraw) MarshalJSON() ([]byte, error) {
	type withdraw Withdraw
	temp := withdraw(w)
	return json.Marshal(temp)
}

// TransactionRequest represents a request for creating a new transaction.
type TransactionRequest struct {
	Type     string // Describes transaction type, see docs for supported common transaction types.
	WalletID string // Wallet identifier for the transaction.
	Amount   int    // Amount for the transaction.
}

// Validate implements validator.Validator.
// TODO: implement more sophisticated validation rules.
func (tx *TransactionRequest) Validate() error {
	if !isValidID(tx.WalletID) {
		return ErrNotValidWalletID
	}

	if len(tx.Type) == 0 {
		return ErrNotValidTransaction
	}

	if tx.Amount == 0 {
		return ErrNotValidAmount
	}

	return nil
}

// Transaction represents wallet transaction command.
type Transaction struct {
	Type     string // Describes transaction type, see docs for supported common transaction types.
	WalletID string // Wallet identifier for the transaction.
	Amount   int    // Amount for the transaction.
}

// Validate implements validator.Validator.
func (tx *Transaction) Validate() error {
	op := strings.ToUpper(tx.Type)
	if !slices.Contains([]string{TransactionDeposit, TransactionWithdraw}, op) {
		return ErrNotValidTransaction
	}
	return nil
}

// CreateTransaction creates a new transaction for the given wallet.
func CreateTransaction(ctx context.Context, saveAggregate database.SaveAggregateFunc, getWallet database.GetAggregateFunc[*WalletAggregate], req *TransactionRequest) (*Wallet, error) {
	// transaction must be a valid
	if err := validator.Validate(req); err != nil {
		return nil, err
	}

	// verify that such wallet exist
	wallet, err := getWallet(ctx, &WalletAggregate{}, req.WalletID)
	if err != nil {
		return nil, err
	}

	err = wallet.ProcessTransaction(&Transaction{
		Type:     req.Type,
		WalletID: req.WalletID,
		Amount:   req.Amount,
	})
	if err != nil {
		return nil, err
	}

	if err := saveAggregate(ctx, wallet); err != nil {
		return nil, errors.Wrap(err, "unable to persist transaction")
	}

	return &wallet.Wallet, nil
}
