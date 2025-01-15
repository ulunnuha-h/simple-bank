package db

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/ulunnuha-h/simple_bank/util"
)

func CreateRandomEntry(t *testing.T, account_id int64) Entry {
	args := CreateEntryParams{
		AccountID: account_id,
		Amount:    util.RandomMoney(),
	}

	entry, err := testQuery.CreateEntry(context.Background(), args)
	require.NoError(t, err)
	require.NotEmpty(t, entry)

	require.Equal(t, account_id, entry.AccountID)
	require.Equal(t, args.Amount, entry.Amount)

	require.NotEmpty(t, entry.CreatedAt)

	return entry
}

func TestCreateEntry(t *testing.T) {
	account := CreateRandomAccount(t)
	CreateRandomEntry(t, account.ID)
}

func TestGetEntry(t *testing.T) {
	account := CreateRandomAccount(t)
	entry := CreateRandomEntry(t, account.ID)

	entry2, err := testQuery.GetEntry(context.Background(), entry.ID)
	require.NoError(t, err)
	require.NotEmpty(t, entry2)

	require.Equal(t, account.ID, entry2.AccountID)
	require.Equal(t, entry.Amount, entry2.Amount)

	require.WithinDuration(t, entry.CreatedAt, entry2.CreatedAt, time.Second)
}

func TestListEntries(t *testing.T) {
	account := CreateRandomAccount(t)
	for i := 0; i < 10; i++ {
		CreateRandomEntry(t, account.ID)
	}

	args := ListEntriesParams{
		AccountID: account.ID,
		Limit:     5,
		Offset:    5,
	}

	entries, err := testQuery.ListEntries(context.Background(), args)
	require.NoError(t, err)
	require.NotEmpty(t, entries)

	for _, entry := range entries {
		require.NotEmpty(t, entry)
	}
}
