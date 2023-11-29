package db

import (
	"context"
	"testing"
	"time"

	"github.com/juanaguenza/bank/util"
	"github.com/stretchr/testify/require"
)

func createRandomEntry(t *testing.T, account Account) Entry {
	arg := CreateEntryParams{
		AccountID: account.ID,
		Amount:    util.RandomMoney(),
	}

	// Create the entry
	entry, err := testQueries.CreateEntry(context.Background(), arg)

	// Check that no error occurred
	// Entry isn't empty
	// The account id's match for the queried entry and arg
	// Amount matches as well
	// ID and CreatedAt are not zero
	require.NoError(t, err)
	require.NotEmpty(t, entry)

	require.Equal(t, arg.AccountID, entry.AccountID)
	require.Equal(t, arg.Amount, entry.Amount)

	require.NotZero(t, entry.ID)
	require.NotZero(t, entry.CreatedAt)

	return entry
}

func TestCreateEntry(t *testing.T) {
	account := createRandomAccount(t)
	createRandomEntry(t, account)
}

func TestGetEntry(t *testing.T) {
	account := createRandomAccount(t)
	entry := createRandomEntry(t, account)

	queriedEntry, err := testQueries.GetEntry(context.Background(), entry.ID)

	require.NoError(t, err)
	require.NotEmpty(t, queriedEntry)

	require.Equal(t, entry.ID, queriedEntry.ID)
	require.Equal(t, entry.AccountID, queriedEntry.AccountID)
	require.Equal(t, entry.Amount, queriedEntry.Amount)
	require.WithinDuration(t, entry.CreatedAt, queriedEntry.CreatedAt, time.Second)
}

func TestListEntries(t *testing.T) {
	account := createRandomAccount(t)

	for i := 0; i < 10; i++ {
		createRandomEntry(t, account)
	}

	arg := ListEntriesParams {
		AccountID: account.ID,
		Limit: 5,
		Offset: 5,
	}

	entries, err := testQueries.ListEntries(context.Background(), arg)

	require.NoError(t, err)
	require.Len(t, entries, 5)

	for _, entry := range entries {
		require.NotEmpty(t, entry)
	}
	
}