package db

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestTransferTx(t *testing.T) {
	store := NewStore(testDB)

	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)
	// The best way to test if our transaction works well
	// is to run it with several concurrent go routines.
	// n - number of concurren transactions creating in go routines
	n := 5
	// The amount of money moving in transactions
	amount := int64(10)

	// channels jsou navržené pro propojení konkurntních go rutin
	errs := make(chan error)
	// jeden kanál bude dostávat chyby a druhý výsledky z transakcí
	results := make(chan TransferTxResult)

	for i := 0; i < n; i++ {
		go func() {
			// není vhodné kontrolovat errors funkcí které spouští v go rutinách
			// přímo v go rutině, není pak garantováno, že pokud se to neprovede,
			// opravdu ten test failne, je nutné posílat výsledky errorů v go rutinách
			// pomocí chanels zpátky do hlavní go rutiny, kde běží náš test
			result, err := store.TransferTx(context.Background(), TransferTxParams{
				FromAccountID: account1.ID,
				ToAccountID:   account2.ID,
				Amount:        amount,
			})

			errs <- err
			results <- result
		}()
	}

	// checks results
	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)

		result := <-results
		require.NotEmpty(t, result)

		// check transfer
		transfer := result.Transfer
		// objeck obsahujicí detaily o transakci nemá být prázný
		require.NotEmpty(t, transfer)
		// id účtů se musí shodovat se záznamem v tabulce která úkládá trasfery financí
		require.Equal(t, account1.ID, transfer.FromAccountID)
		require.Equal(t, account2.ID, transfer.ToAccountID)
		// množství peněz v transferu se musí shodovat se zadaným množstvím pro převod
		require.Equal(t, amount, transfer.Amount)
		// ID transfer nesmí být 0 protože je autoincrement, který se má výplnit v databázi sám
		require.NotZero(t, transfer.ID)
		// CreatedAt nesmí být 0 protože databázové schéma má nastavené, aby si aktuální čas databáze tabulka
		// doplnila sama
		require.NotZero(t, transfer.CreatedAt)

		// check jestli je transfer v databázi opravdu vytvořený
		// udělá query do databáze, aby ověřil, že záznal s tímto ID opravdu existuje
		_, err = store.GetTransfer(context.Background(), transfer.ID)
		require.NoError(t, err)

		// check fromEntries
		fromEntry := result.FromEntry
		require.NotEmpty(t, fromEntry)
		require.Equal(t, account1.ID, fromEntry.AccountID)
		require.Equal(t, -amount, fromEntry.Amount)
		require.NotZero(t, fromEntry.ID)
		require.NotZero(t, fromEntry.CreatedAt)
		// query do databáze ověří, že tento record opravdu existuje
		_, err = store.GetEntry(context.Background(), fromEntry.ID)
		require.NoError(t, err)

		// check toEntries
		// check fromEntries
		toEntry := result.ToEntry
		require.NotEmpty(t, toEntry)
		require.Equal(t, account2.ID, toEntry.AccountID)
		require.Equal(t, amount, toEntry.Amount)
		require.NotZero(t, toEntry.ID)
		require.NotZero(t, toEntry.CreatedAt)
		// query do databáze ověří, že tento record opravdu existuje
		_, err = store.GetEntry(context.Background(), toEntry.ID)
		require.NoError(t, err)

		// TODO: check accounts balances

	}
}
