package db

import (
	"context"
	"database/sql"
	"fmt"
)

// Store provides all functions to execute db queries and transactions
// it also stores all combinations which will be using in transactions
// Queries struct does not support transactions
// we want extand it about this functionality with
// adding it inside Store struct it is called composition
// It is way to extand Queries functionality
// more prefered than inharitance
type Store struct {
	*Queries
	db *sql.DB
}

// Function to create new store object

func NewStore(db *sql.DB) *Store {
	return &Store{
		db:      db,
		Queries: New(db),
	}
}

// Function to execute transaction on the created store object
// it takes context as agument and function which creates Queries object and returns error
// it call callback func with created Queries and commit or rollback the transaction
// base on the error returned by that function
// funkce je loweCase protože ji nebudeme chtít exportovat s package jinak
// místo toho budu exportovat funkce pro každou specifickou transakci nad databází
func (store *Store) execTx(ctx context.Context, fn func(*Queries) error) error {
	// txOptions je možnosti jak customizovat některé věci pro konkrétní transakci
	// pokud nic nedefinuji použijí de default
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// voláme novou funkci s vytvořenou transakcí
	// je to stejná funkce jako ve store jen tentokrát nepředáváme db ale tsx object
	// to funguje protože funkce New přijmá DBTX interface
	q := New(tx)
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		// pokud rollback proběhne vrátí to použe originální error z transakce
		return err
	}
	// pokud to žádný error neshodí, změny se potvrdí pomocí commit

	return tx.Commit()
}

// TransferTxParams type contains the input parameters of the transfer transaction
// which are neccessary to transfer money between accounts

type TransferTxParams struct {
	FromAccountID int64 `json:"from_account_id"`
	ToAccountID   int64 `json:"to_account_id"`
	Amount        int64 `json:"ammount"`
}

// TransferTxResult type contains the result after execute money transfer db transaction
type TransferTxResult struct {
	// napopulovanou struktura Transfer, která nese údaje o tom odkud kam, kolik peněz a jaký moment se pohybovalo
	Transfer Transfer `json:"transfer"`
	// nový stav úču ze kterého peníze odešli
	FromAccount Account `json:"from_account"`
	// nový stav účtu kam peníze dorazily
	ToAccount Account `json:"to_account"`
	// záznamy v tabulce entries pro každý účet ukazující pohyby mezi účty
	FromEntry Entry `json:"from_entry"`
	ToEntry   Entry `json:"to_entry"`
}

// první exportovaná funkce s konkrétní transakcí ( reprezentuje trasfer peněz  mezi dvěma účty )
// běží v prasakci protože při transferu pěněz se v databází děje větší množství operací nad
// různými tabulkami
// It creates a transfer record, add account entries, and update accounts' balance within a single database transaction.
func (store *Store) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
	// inicializace prázdného result
	var result TransferTxResult
	// vytvoření nově db transakce
	err := store.execTx(ctx, func(q *Queries) error {
		var err error
		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams{
			FromAccountID: arg.FromAccountID,
			ToAccountID:   arg.ToAccountID,
			Amount:        arg.Amount,
		})
		if err != nil {
			return err
		}
		// záznam o transakci pro účet ze kterého peníze odešli
		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.FromAccountID,
			Amount:    -arg.Amount,
		})
		if err != nil {
			// pokud vrátím error provede se roll back
			return err
		}
		// záznam o transakci pro účet na který se peníze přidaly
		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.ToAccountID,
			Amount:    arg.Amount,
		})
		if err != nil {
			// pokud vrátím error provede se roll back
			return err
		}
		// update balance zahrnuje nutnost prevence potenscionálních deadlock v databází
		// ToDo: update accounts balance

		return nil
	})

	return result, err

}
