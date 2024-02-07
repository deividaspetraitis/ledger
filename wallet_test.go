package ledger

import (
	"testing"

	"github.com/deividaspetraitis/go/es"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

// idDecorator is an wrapper allowing to set ID to the T.
type idDecorator[T any] func(id string, agg es.Aggregate) T

// TestWalletOn runs various events on Wallet entity and tests resulting state.
// Note: This test does not check AggregrateRoot state.
func TestWalletOn(t *testing.T) {
	var testcases = []struct {
		name string

		walletID string
		events   []idDecorator[*es.Event]

		wallet idDecorator[*Wallet]

		err error
	}{
		{
			name:     "initialisation/deposit/withdraw operations",
			walletID: newID(),
			events: []idDecorator[*es.Event]{
				func(id string, agg es.Aggregate) *es.Event {
					return es.NewEvent(id, agg,
						&WalletInitialized{
							ID:      id,
							Name:    "test wallet",
							Balance: 100,
						})
				},

				func(id string, agg es.Aggregate) *es.Event {
					return es.NewEvent(id, agg,
						&Deposit{
							WalletID: id,
							Amount:   50,
						})
				}, func(id string, agg es.Aggregate) *es.Event {
					return es.NewEvent(id, agg,
						&Withdraw{
							WalletID: id,
							Amount:   15,
						})
				},
			},
			wallet: func(id string, aggregate es.Aggregate) *Wallet {
				return newWallet(id, "test wallet", 135)
			},
		},
	}

	for i, tt := range testcases {

		t.Run(tt.name, func(t *testing.T) {
			var wallet WalletAggregate
			for _, v := range tt.events {
				if err := (&wallet).on(v(tt.walletID, &wallet)); err != tt.err {
					t.Errorf("#%d got %v, want %v", i, err, tt.err)
				}
			}

			if !cmp.Equal(&wallet.Wallet, tt.wallet(tt.walletID, &wallet), cmpopts.IgnoreUnexported(Wallet{})) {
				t.Errorf("#%d got %v, want %v", i, &wallet.Wallet, tt.wallet(tt.walletID, &wallet))
			}
		})
	}
}

// TestSync tests that aggregate state is synchronised properly.
func TestSync(t *testing.T) {
	var testcases = []struct {
		name string

		walletID string

		events []idDecorator[*es.Event]
		sync   []idDecorator[*es.Event]

		wallet idDecorator[*Wallet]

		err error
	}{
		{
			name:     "successfully sync recently applied events",
			walletID: newID(),
			events: []idDecorator[*es.Event]{
				func(id string, agg es.Aggregate) *es.Event {
					return es.NewEvent(id, agg,
						&WalletInitialized{
							ID:      id,
							Name:    "test wallet",
							Balance: 100,
						})
				},

				func(id string, agg es.Aggregate) *es.Event {
					return es.NewEvent(id, agg,
						&Deposit{
							WalletID: id,
							Amount:   50,
						})
				}, func(id string, agg es.Aggregate) *es.Event {
					return es.NewEvent(id, agg,
						&Withdraw{
							WalletID: id,
							Amount:   15,
						})
				},
			},
			wallet: func(id string, aggregate es.Aggregate) *Wallet {
				return newWallet(id, "test wallet", 135)
			},
		},
		{
			name:     "sync recently applied events",
			walletID: newID(),
			events: []idDecorator[*es.Event]{
				func(id string, agg es.Aggregate) *es.Event {
					return es.NewEvent(id, agg,
						&WalletInitialized{
							ID:      id,
							Name:    "test wallet",
							Balance: 100,
						})
				},

				func(id string, agg es.Aggregate) *es.Event {
					return es.NewEvent(id, agg,
						&Deposit{
							WalletID: id,
							Amount:   50,
						})
				}, func(id string, agg es.Aggregate) *es.Event {
					return es.NewEvent(id, agg,
						&Withdraw{
							WalletID: id,
							Amount:   15,
						})
				},
			},
			wallet: func(id string, aggregate es.Aggregate) *Wallet {
				return newWallet(id, "test wallet", 135)
			},
		},
	}

	for i, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			var wallet WalletAggregate
			var events []*es.Event
			for i, v := range tt.events {
				events = append(events, v(tt.walletID, &wallet))
				if err := (&wallet).Apply(events[i]); err != tt.err {
					t.Errorf("#%d got %v, want %v", i, err, tt.err)
				}
			}

			// sync
			for _, v := range events {
				if err := (&wallet).Sync(v); err != tt.err {
					t.Errorf("#%d got %v, want %v", i, err, tt.err)
				}
			}

			if !cmp.Equal(&wallet.Wallet, tt.wallet(tt.walletID, &wallet), cmpopts.IgnoreUnexported(Wallet{})) {
				t.Errorf("#%d got %v, want %v", i, &wallet.Wallet, tt.wallet(tt.walletID, &wallet))
			}
		})
	}
}
