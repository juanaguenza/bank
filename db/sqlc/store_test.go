package db

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransferTx(t *testing.T) {
	store := NewStore(testDB)

	fromAccount := createRandomAccount(t)
	toAccount := createRandomAccount(t)
	fmt.Println(">> before:", fromAccount.Balance, toAccount.Balance)

	// run n concurrent transfer transaction (concurrent goroutines)
	n := 5
	amount := int64(10)

	// Use channels to connect our concurrent goroutines
	// One channel to receive errors, another channel to receive the TransferTxResults
	errs := make(chan error)
	results := make(chan TransferTxResult)

	for i := 0; i < n; i++ {
		// Use go keyword to start a new routine
		go func() {
			ctx := context.Background()
			result, err := store.TransferTx(ctx, TransferTxParams{
				FromAccountID: fromAccount.ID,
				ToAccountID:   toAccount.ID,
				Amount:        amount,
			})

			// Send our error to the errors channel
			// Send our result to the results channel
			errs <- err
			results <- result
		}()
	}

	// Check results
	existed := make(map[int]bool)
	for i := 0; i < n; i++ {
		// Receive the error from the channel
		err := <-errs
		require.NoError(t, err)

		result := <-results
		require.NotEmpty(t, result)

		// Check transfer
		transfer := result.Transfer
		require.NotEmpty(t, transfer)
		require.Equal(t, fromAccount.ID, transfer.FromAccountID)
		require.Equal(t, toAccount.ID, transfer.ToAccountID)
		require.Equal(t, amount, transfer.Amount)
		require.NotZero(t, transfer.ID)
		require.NotZero(t, transfer.CreatedAt)

		// Get the transfer from the database to ensure it was created
		// As the Queries object is inside the store... the GetTransfer query is also available inside the store
		_, err = store.GetTransfer(context.Background(), transfer.ID)
		require.NoError(t, err)

		// Check entries
		fromEntry := result.FromEntry
		require.NotEmpty(t, fromEntry)
		require.Equal(t, fromAccount.ID, fromEntry.AccountID)
		require.Equal(t, -amount, fromEntry.Amount)
		require.NotZero(t, fromEntry.ID)
		require.NotZero(t, fromEntry.CreatedAt)

		// Get the FromEntry from the database to ensure it was created
		_, err = store.GetEntry(context.Background(), fromEntry.ID)
		require.NoError(t, err)

		// Check entries
		toEntry := result.ToEntry
		require.NotEmpty(t, toEntry)
		require.Equal(t, toAccount.ID, toEntry.AccountID)
		require.Equal(t, amount, toEntry.Amount)
		require.NotZero(t, toEntry.ID)
		require.NotZero(t, toEntry.CreatedAt)

		// Get the ToEntry from the database to ensure it was created
		_, err = store.GetEntry(context.Background(), toEntry.ID)
		require.NoError(t, err)

		// Check accounts
		fromAccountResult := result.FromAccount
		require.NotEmpty(t, fromAccountResult)
		require.Equal(t, fromAccount.ID, fromAccountResult.ID)

		toAccountResult := result.ToAccount
		require.NotEmpty(t, toAccountResult)
		require.Equal(t, toAccount.ID, toAccountResult.ID)

		// Check accounts' balance
		fmt.Println(">> tx:", fromAccountResult.Balance, toAccountResult.Balance)
		fromAccountDiff := fromAccount.Balance - fromAccountResult.Balance
		toAccountDiff := toAccountResult.Balance - toAccount.Balance
		require.Equal(t, fromAccountDiff, toAccountDiff)
		require.True(t, fromAccountDiff > 0)
		require.True(t, fromAccountDiff%amount == 0) // amount, 2 * amount, 3 * amount, ...

		k := int(fromAccountDiff / amount)
		require.True(t, k >= 1 && k <= n)
		require.NotContains(t, existed, k)
		existed[k] = true
	}

	// Check the final updated balances
	updatedFromAccount, err := testQueries.GetAccount(context.Background(), fromAccount.ID)
	require.NoError(t, err)

	updatedToAccount, err := testQueries.GetAccount(context.Background(), toAccount.ID)
	require.NoError(t, err)

	fmt.Println(">> after:", updatedFromAccount.Balance, updatedToAccount.Balance)
	require.Equal(t, fromAccount.Balance-int64(n)*amount, updatedFromAccount.Balance)
	require.Equal(t, toAccount.Balance+int64(n)*amount, updatedToAccount.Balance)
}
