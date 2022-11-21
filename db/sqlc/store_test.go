package db

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransferTx(t *testing.T) {
	store := NewStore(testDB)

	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)
	fmt.Println(">>before:", account1.Balance, account2.Balance)
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
		txName := fmt.Sprintf("tx %d", i +1)
		go func() {
			ctx := context.WithValue(context.Background(), txKey, txName) 
			// není vhodné kontrolovat errors funkcí které spouští v go rutinách
			// přímo v go rutině, není pak garantováno, že pokud se to neprovede,
			// opravdu ten test failne, je nutné posílat výsledky errorů v go rutinách
			// pomocí chanels zpátky do hlavní go rutiny, kde běží náš test
			result, err := store.TransferTx(ctx, TransferTxParams{
				FromAccountID: account1.ID,
				ToAccountID:   account2.ID,
				Amount:        amount,
			})

			errs <- err
			results <- result
		}()
	}

	// checks results
	existed := make(map[int]bool)
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

		// These next lines will be test the new accounts states base on test operations.
		fromAccount := result.FromAccount
		require.NotEmpty(t, fromAccount)
		require.Equal(t, account1.ID, fromAccount.ID)

		toAccount := result.ToAccount
		require.NotEmpty(t, toAccount)
		require.Equal(t, account2.ID, toAccount.ID)

		// This test will check new balances on each account after transaction is commited. 
		fmt.Println(">> tx:", fromAccount.Balance, toAccount.Balance)
		diff1 := account1.Balance - fromAccount.Balance
		diff2 := toAccount.Balance - account2.Balance
		// If there are correct changes for each account thesee two variables should be equal.
		fmt.Println(">> diff1: ", diff1, "diff2: ", diff2)
		require.Equal(t, diff1, diff2)
		require.True(t, diff1 > 0)
		// This test mean that the balance of account 1 will be decreased by 1 times amount each transaction
		require.True(t, diff1%amount == 0) 

		k := int(diff1 / amount)
		require.True(t, k >=1 && k <= n )
		require.NotContains(t, existed, k)
		existed[k] = true 
	}

	// check the final updated balances
	updatedAccount1, err := testQueries.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)

	updatedAccount2, err := testQueries.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err)
	fmt.Println(">>after:", updatedAccount1.Balance, updatedAccount2.Balance)
	require.Equal(t, account1.Balance - int64(n) * amount, updatedAccount1.Balance)
	require.Equal(t, account2.Balance + int64(n) * amount, updatedAccount2.Balance)

}


// This test the situation where are concurent transactions trying completed operations in reverse direction. 
func TestTransferTxReverse(t *testing.T) {
	store := NewStore(testDB)

	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)
	fmt.Println(">>before:", account1.Balance, account2.Balance)
	// The best way to test if our transaction works well
	// is to run it with several concurrent go routines.
	// n - number of concurren transactions creating in go routines
	n := 10
	// The amount of money moving in transactions
	amount := int64(10)

	// channels jsou navržené pro propojení konkurntních go rutin
	errs := make(chan error)
	// jeden kanál bude dostávat chyby a druhý výsledky z transakcí
	
	for i := 0; i < n; i++ {
		fromAcountID := account1.ID
		toAccountID := account2.ID
		
		// po splnění této podmínky se směr pohybu prostředků na účtech otočí
		// takže for loop odpálí celkem 10 konkuretních transakcí 5 z účtu A přidává prostředky na B
		// dalších 5 odebírá z B stejnou částu a přidává ji na A (reverse direction)
		if i % 2 == 1 {
			fromAcountID = account2.ID
			toAccountID = account1.ID
		}
		go func() {
			// není vhodné kontrolovat errors funkcí které spouští v go rutinách
			// přímo v go rutině, není pak garantováno, že pokud se to neprovede,
			// opradu ten test failne, je nutné posílat výsledky errorů v go rutinách
			// pomocí chanels zpátky do hlavní go rutiny, kde běží náš test
			_, err := store.TransferTx(context.Background(), TransferTxParams{
				FromAccountID: fromAcountID,
				ToAccountID:   toAccountID,
				Amount:        amount,
			})

			errs <- err
	
		}()
	}

	// checks results
	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)

	}

	// check the final updated balances
	updatedAccount1, err := testQueries.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)

	updatedAccount2, err := testQueries.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err)

	// na konci testu se kontroluje jestli je balance na obou účtech stejná jako před zahájením testu
	fmt.Println(">>after:", updatedAccount1.Balance, updatedAccount2.Balance)
	require.Equal(t, account1.Balance, updatedAccount1.Balance)
	require.Equal(t, account2.Balance, updatedAccount2.Balance)

}