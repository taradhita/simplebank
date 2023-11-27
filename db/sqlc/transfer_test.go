package db

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/taradhita/simplebank/util"
)

func createRandomTransfer(t *testing.T, firstAccount, secondAccount Account) Transfer {
	arg := CreateTransferParams{
		FromAccountID: firstAccount.ID,
		ToAccountID:   secondAccount.ID,
		Amount:        util.RandomMoney(),
	}

	transfer, err := testQueries.CreateTransfer(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, transfer)

	require.Equal(t, arg.FromAccountID, transfer.FromAccountID)
	require.Equal(t, arg.ToAccountID, transfer.ToAccountID)
	require.Equal(t, arg.Amount, transfer.Amount)

	require.NotZero(t, transfer.ID)
	require.NotZero(t, transfer.CreatedAt)

	return transfer
}

func TestCreateTransfer(t *testing.T) {
	firstAccount := createRandomAccount(t)
	secondAccount := createRandomAccount(t)
	createRandomTransfer(t, firstAccount, secondAccount)
}

func TestGetTransfer(t *testing.T) {
	firstAccount := createRandomAccount(t)
	secondAccount := createRandomAccount(t)
	firstTransfer := createRandomTransfer(t, firstAccount, secondAccount)
	secondTransfer, err := testQueries.GetTransfer(context.Background(), firstTransfer.ID)
	require.NoError(t, err)
	require.NotEmpty(t, secondTransfer)

	require.Equal(t, firstTransfer.ID, secondTransfer.ID)
	require.Equal(t, firstTransfer.FromAccountID, secondTransfer.FromAccountID)
	require.Equal(t, firstTransfer.ToAccountID, secondTransfer.ToAccountID)
	require.Equal(t, firstTransfer.Amount, secondTransfer.Amount)
	require.WithinDuration(t, firstTransfer.CreatedAt, secondTransfer.CreatedAt, time.Second)
}

func TestListTransfers(t *testing.T) {
	firstAccount := createRandomAccount(t)
	secondAccount := createRandomAccount(t)

	for i := 0; i < 5; i++ {
		createRandomTransfer(t, firstAccount, secondAccount)
		createRandomTransfer(t, secondAccount, firstAccount)
	}

	arg := ListTransferParams{
		FromAccountID: firstAccount.ID,
		ToAccountID:   firstAccount.ID,
		Limit:         5,
		Offset:        5,
	}

	transfers, err := testQueries.ListTransfer(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, transfers, 5)

	for _, transfer := range transfers {
		require.NotEmpty(t, transfer)
		require.True(t, transfer.FromAccountID == firstAccount.ID || transfer.ToAccountID == firstAccount.ID)
	}
}
