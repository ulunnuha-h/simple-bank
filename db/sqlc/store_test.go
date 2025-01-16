package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/ulunnuha-h/simple_bank/util"
)

func TestTransaction(t *testing.T) {
	store := NewStore(testDB)

	account1 := CreateRandomAccount(t)
	account2 := CreateRandomAccount(t)

	var amount int64 = util.RandomMoney()
	n := 5

	errs := make(chan error, n)
	results := make(chan TransferTxResult, n)

	for i := 0; i < n; i++ {
		go func() {
			result, err := store.TransferTx(context.Background(), TransferTxParams{
				FromAccountId: account1.ID,
				ToAccountId:   account2.ID,
				Amount:        amount,
			})
			require.NotEmpty(t, result)
			require.NoError(t, err)

			results <- result
			errs <- err
		}()
	}

	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)

		result := <-results
		require.NotEmpty(t, result)

		transfer := result.Transfer
		require.NotEmpty(t, transfer)
		require.Equal(t, amount, transfer.Amount)
		require.Equal(t, account1.ID, transfer.FromAccountID)
		require.Equal(t, account2.ID, transfer.ToAccountID)
		require.NotEmpty(t, transfer.ID)
		require.NotEmpty(t, transfer.CreatedAt)

		_, err = store.GetTransfer(context.Background(), transfer.ID)
		require.NoError(t, err)

		fromEntry := result.FromEntry
		require.NotEmpty(t, fromEntry)
		require.Equal(t, account1.ID, fromEntry.AccountID)
		require.Equal(t, -amount, fromEntry.Amount)
		require.NotEmpty(t, fromEntry.CreatedAt)
		require.NotEmpty(t, fromEntry.ID)

		_, err = store.GetEntry(context.Background(), fromEntry.ID)
		require.NoError(t, err)

		toEntry := result.ToEntry
		require.NotEmpty(t, toEntry)
		require.Equal(t, account2.ID, toEntry.AccountID)
		require.Equal(t, amount, toEntry.Amount)
		require.NotEmpty(t, toEntry.CreatedAt)
		require.NotEmpty(t, toEntry.ID)

		_, err = store.GetEntry(context.Background(), toEntry.ID)
		require.NoError(t, err)
	}
}
