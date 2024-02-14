package sql

import (
	"context"
	"fmt"
	"strings"

	"github.com/deividaspetraitis/ledger"

	"github.com/deividaspetraitis/go/database/sql"
	"github.com/deividaspetraitis/go/errors"
)

// StoreWallet persists wallet w into database db.
func StoreWallet(ctx context.Context, db *sql.DB, w *ledger.WalletAggregate) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// persist wallet
	if err := storeWallet(ctx, tx, &w.Wallet, int(w.Version())); err != nil {
		return err
	}

	return tx.Commit()
}

// storeWallet creates a new wallet.
func storeWallet(ctx context.Context, tx *sql.Tx, w *ledger.Wallet, revision int) error {
	return tx.QueryRowContext(ctx,
		`INSERT INTO wallets (
			id,
			name,
			balance,
			version
		) 
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE 
		SET
			name = excluded.name,
			balance = excluded.balance,
			version = excluded.version
		RETURNING id`,
		w.ID,
		w.Name,
		w.Balance,
		revision,
	).Scan(&w.ID)
}

// updateWallet updates given wallets information.
func updateWallet(ctx context.Context, tx *sql.Tx, w *ledger.Wallet) error {
	if _, err := tx.ExecContext(ctx, `
		UPDATE 
			wallets
		SET 
			name = $1,
		    balance = $2
		WHERE 
			id = $3
	`,
		w.Name,
		w.Balance,
		w.ID,
	); err != nil {
		return err
	}

	fmt.Println("balance", w)
	return nil
}

// GetWallet retrieves a wallet from the database.
func GetWallet(ctx context.Context, db *sql.DB, id string) (*ledger.WalletAggregate, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	wallet, err := getWallet(ctx, tx, id)
	if err != nil {
		return nil, err
	}

	aggregate, err := ledger.NewWallet(&ledger.CreateWalletRequest{
		ID: wallet.ID,
		Name: wallet.Name,
		Balance: wallet.Balance,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "ledger.NewWallet failed: %#v", wallet)
	}

	return aggregate, tx.Commit()
}

func getWallet(ctx context.Context, tx *sql.Tx, id string) (*ledger.Wallet, error) {
	wallets, n, err := findWallets(ctx, tx, &walletsFilter{ID: &id})
	if err != nil {
		return nil, err
	} else if n == 0 {
		return nil, ledger.ErrEntryNotFound
	}
	return wallets[0], nil
}

type walletsFilter struct {
	ID *string
}

func findWallets(ctx context.Context, tx *sql.Tx, filter *walletsFilter) (_ []*ledger.Wallet, n int, err error) {
	where, args := []string{"1 = 1"}, []interface{}{}
	if v := filter.ID; v != nil {
		where, args = append(where, "id = $1"), append(args, *v)
	}

	rows, err := tx.QueryContext(ctx, `
		SELECT 
		    id,
		    name,
			balance,
		    COUNT(*) OVER()
		FROM
			wallets
		WHERE 
			`+strings.Join(where, " AND ")+`
		ORDER BY
			id DESC`,
		args...,
	)
	if err != nil {
		return nil, n, err
	}
	defer rows.Close()

	wallets := make([]*ledger.Wallet, 0)
	for rows.Next() {
		var w ledger.Wallet
		if err := rows.Scan(
			&w.ID,
			&w.Name,
			&w.Balance,
			&n,
		); err != nil {
			return nil, 0, err
		}
		wallets = append(wallets, &w)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return wallets, n, nil
}
