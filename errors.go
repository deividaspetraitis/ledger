package ledger

import "github.com/deividaspetraitis/go/errors"

// Common service errors.
var (
	ErrEntryNotFound = errors.New("entry not found")

	ErrNotValidWalletName = errors.New("given wallet name is not a valid name")
	ErrNotValidWalletID   = errors.New("given wallet name is not a valid ID")

	ErrNotValidTransaction = errors.New("given transaction type is not available")
	ErrNotValidAmount      = errors.New("given transaction amount is not valid")
)
