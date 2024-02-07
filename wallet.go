package ledger

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/deividaspetraitis/go/database"
	"github.com/deividaspetraitis/go/errors"
	"github.com/deividaspetraitis/go/es"
	"github.com/deividaspetraitis/go/validator"
)

// init initialises program state.
// register supported aggregates along their events.
func init() {
	es.RegisterAggregateEvent(&WalletAggregate{}, func() es.MarshalUnmarshaler {
		return &WalletInitialized{}
	})
	es.RegisterAggregateEvent(&WalletAggregate{}, func() es.MarshalUnmarshaler {
		return &Deposit{}
	})
	es.RegisterAggregateEvent(&WalletAggregate{}, func() es.MarshalUnmarshaler {
		return &Withdraw{}
	})
}

// CreateWalletRequest represents a request for creating a new wallet.
type CreateWalletRequest struct {
	Name string
}

// Validate implements validator.Validator.
func (r *CreateWalletRequest) Validate() error {
	if len(r.Name) == 0 {
		return ErrNotValidWalletName
	}
	return nil
}

// WalletInitialized represents an event emitted when a wallet is created.
type WalletInitialized struct {
	ID      string // Unique wallet identifier
	Name    string // Wallet name
	Balance int    // Wallet balance in cents
}

func (w *WalletInitialized) UnmarshalJSON(b []byte) error {
	type wallet WalletInitialized
	var temp wallet
	if err := json.Unmarshal(b, &temp); err != nil {
		return err
	}
	*w = WalletInitialized(temp)
	return nil
}

func (w *WalletInitialized) MarshalJSON() ([]byte, error) {
	type wallet WalletInitialized
	temp := wallet(*w)
	return json.Marshal(temp)
}

// NewWallet creates and returns a new named wallet with defaults.
// It validates CreateWalletRequest and if it does not pass, error will be returned instead.
func NewWallet(req *CreateWalletRequest) (*WalletAggregate, error) {
	if err := validator.Validate(req); err != nil {
		return nil, err
	}

	var aggregate WalletAggregate

	id := newID()

	err := (&aggregate).Apply(es.NewEvent(id, &aggregate, &WalletInitialized{
		ID:      id,
		Name:    req.Name,
		Balance: 0,
	}))
	if err != nil {
		return nil, err
	}

	return &aggregate, nil
}

// newWallet constructs a new instance of Wallet.
func newWallet(id string, name string, balance int) *Wallet {
	return &Wallet{
		ID:      id,
		Name:    name,
		Balance: balance,
	}
}

// WalletAggregate represents Wallet's aggregate.
type WalletAggregate struct {
	es.AggregateRoot
	Wallet
}

// Wallet represents current state of the wallet.
type Wallet struct {
	ID      string // Unique wallet identifier
	Name    string // Wallet name
	Balance int    // Wallet balance in cents
}

func (w *WalletAggregate) Deposit(tx *Transaction) error {
	return w.Apply(es.NewEvent(w.ID, w, &Deposit{
		WalletID: tx.WalletID,
		Amount:   tx.Amount,
	}))
}

var ErrInsufficientBalance = errors.New("insufficient balance")

func (w *WalletAggregate) Withdraw(tx *Transaction) error {
	if (w.Balance - tx.Amount) < 0 {
		return ErrInsufficientBalance
	}

	return w.Apply(es.NewEvent(w.ID, w, &Withdraw{
		WalletID: tx.WalletID,
		Amount:   tx.Amount,
	}))
}

// Reply implements es.Aggregate.
func (w *WalletAggregate) Reply(event []*es.Event) error {
	if err := w.Root().Reply(event); err != nil {
		return err
	}
	for _, v := range event {
		if err := w.on(v); err != nil {
			return err
		}
	}
	return nil
}

// Apply implements es.Aggregate.
func (w *WalletAggregate) Apply(event *es.Event) error {
	if err := w.AggregateRoot.Apply(event); err != nil {
		return err
	}
	return w.on(event)
}

// On applies given event to the wallet to update its state.
func (w *Wallet) on(event *es.Event) error {
	switch e := event.Data.(type) {
	case *WalletInitialized:
		*w = *newWallet(e.ID, e.Name, e.Balance)
	case *Deposit:
		w.Balance += e.Amount
	case *Withdraw:
		w.Balance -= e.Amount
	default:
		return errors.Newf("unsupported event: %#v", e)
	}

	return nil
}

// ProcessTransaction applies transaction.
func (w *WalletAggregate) ProcessTransaction(tx *Transaction) error {
	if err := validator.Validate(tx); err != nil {
		return err
	}

	switch strings.ToUpper(tx.Type) {
	case TransactionDeposit:
		return w.Deposit(tx)
	case TransactionWithdraw:
		return w.Withdraw(tx)
	default:
		return ErrNotValidTransaction
	}
}

// CreateWallet creates a new wallet with initialized defaults and information based on CreateWalletRequest.
// On failure error will be returned.
func CreateWallet(ctx context.Context, saveAggregate database.SaveAggregateFunc, req *CreateWalletRequest) (*Wallet, error) {
	if err := validator.Validate(req); err != nil {
		return nil, err
	}

	wallet, err := NewWallet(req)
	if err != nil {
		return nil, err
	}

	if err := saveAggregate(ctx, wallet); err != nil {
		return nil, err
	}

	return &wallet.Wallet, nil
}

// GetWallet retrieves existing wallet based on given wallet ID.
// If Wallet does not exist in the system database.ErrEntityNotFound will be returned.
func GetWallet(ctx context.Context, getWallet database.GetAggregateFunc[*WalletAggregate], id string) (*Wallet, error) {
	wallet, err := getWallet(ctx, &WalletAggregate{}, id)
	if err != nil {
		return nil, err
	}
	return &wallet.Wallet, nil
}
