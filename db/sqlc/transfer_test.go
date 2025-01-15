package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/ulunnuha-h/simple_bank/util"
)

func CreateRandomTransfer(t *testing.T, accountId1 int64, accountId2 int64) Transfer {
	args := CreateTransferParams{
		FromAccountID: accountId1,
		ToAccountID:   accountId2,
		Amount:        util.RandomMoney(),
	}

	transfer, err := testQuery.CreateTransfer(context.Background(), args)
	require.NoError(t, err)
	require.NotEmpty(t, transfer)

	require.Equal(t, accountId1, transfer.FromAccountID)
	require.Equal(t, accountId2, transfer.ToAccountID)
	require.Equal(t, args.Amount, transfer.Amount)

	require.NotEmpty(t, transfer.CreatedAt)

	return transfer
}

func TestCreateTransfer(t *testing.T) {
	account1 := CreateRandomAccount(t)
	account2 := CreateRandomAccount(t)
	CreateRandomTransfer(t, account1.ID, account2.ID)
}

func TestGetTransfer(t *testing.T) {
	account1 := CreateRandomAccount(t)
	account2 := CreateRandomAccount(t)
	transfer := CreateRandomTransfer(t, account1.ID, account2.ID)

	transfer2, err := testQuery.GetTransfer(context.Background(), transfer.ID)
	require.NoError(t, err)
	require.NotEmpty(t, transfer2)

	require.Equal(t, transfer.Amount, transfer2.Amount)
	require.Equal(t, account1.ID, transfer2.FromAccountID)
	require.Equal(t, account2.ID, transfer2.ToAccountID)
	require.NotEmpty(t, transfer2.CreatedAt)
}

func TestListTransfers(t *testing.T) {
	account1 := CreateRandomAccount(t)
	account2 := CreateRandomAccount(t)
	for i := 0; i < 10; i++ {
		CreateRandomTransfer(t, account1.ID, account2.ID)
	}

	args := ListTransfersParams{
		FromAccountID: account1.ID,
		ToAccountID:   account2.ID,
		Offset:        5,
		Limit:         5,
	}

	transfers, err := testQuery.ListTransfers(context.Background(), args)
	require.NoError(t, err)
	require.NotEmpty(t, transfers)

	for _, transfer := range transfers {
		require.NotEmpty(t, transfer)
	}
}
