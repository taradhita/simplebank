package db

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransferTx(t *testing.T) {
	store := NewStore(testDB)

	firstAccount := createRandomAccount(t)
	secondAccount := createRandomAccount(t)
	fmt.Println(">> before:", firstAccount.Balance, secondAccount.Balance)

	// run n concurrent transfer trx
	n := 2
	amount := int64(10)

	errs := make(chan error)
	results := make(chan TransferTxResult)

	for i := 0; i < n; i++ {
		txName := fmt.Sprintf("tx %d", i+1)
		go func() {
			ctx := context.WithValue(context.Background(), txKey, txName)
			result, err := store.TransferTx(ctx, TransferTxParams{
				FromAccountID: firstAccount.ID,
				ToAccountID:   secondAccount.ID,
				Amount:        amount,
			})
			errs <- err
			results <- result
		}()
	}

	existed := make(map[int]bool)

	// check results
	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)

		result := <-results
		require.NotEmpty(t, result)

		transfer := result.Transfer
		require.NotEmpty(t, transfer)
		require.Equal(t, firstAccount.ID, transfer.FromAccountID)
		require.Equal(t, secondAccount.ID, transfer.ToAccountID)
		require.Equal(t, amount, transfer.Amount)
		require.NotZero(t, transfer.ID)
		require.NotZero(t, transfer.CreatedAt)

		_, err = store.GetTransfer(context.Background(), transfer.ID)
		require.NoError(t, err)

		fromEntry := result.FromEntry
		require.NotEmpty(t, fromEntry)
		require.Equal(t, firstAccount.ID, fromEntry.AccountID)
		require.Equal(t, -amount, fromEntry.Amount)
		require.NotZero(t, fromEntry.ID)
		require.NotZero(t, fromEntry.CreatedAt)

		_, err = store.GetEntry(context.Background(), fromEntry.ID)
		require.NoError(t, err)

		toEntry := result.ToEntry
		require.NotEmpty(t, toEntry)
		require.Equal(t, secondAccount.ID, toEntry.AccountID)
		require.Equal(t, amount, toEntry.Amount)
		require.NotZero(t, toEntry.ID)
		require.NotZero(t, toEntry.CreatedAt)

		_, err = store.GetEntry(context.Background(), toEntry.ID)
		require.NoError(t, err)

		// check account
		fromAccount := result.FromAccount
		require.NotEmpty(t, fromAccount)
		require.Equal(t, firstAccount.ID, fromAccount.ID)

		toAccount := result.ToAccount
		require.NotEmpty(t, toAccount)
		require.Equal(t, toAccount.ID, toAccount.ID)

		// check balance
		fmt.Println(">> tx:", result.FromAccount.Balance, result.ToAccount.Balance)

		firstDiff := firstAccount.Balance - fromAccount.Balance
		secondDiff := toAccount.Balance - secondAccount.Balance
		require.Equal(t, firstDiff, secondDiff)
		require.True(t, firstDiff > 0)
		require.True(t, firstDiff%amount == 0)

		k := int(firstDiff / amount)
		require.True(t, k >= 1 && k <= n)
		require.NotContains(t, existed, k)
		existed[k] = true
	}

	// check final updated balance
	updatedFirstAcc, err := testQueries.GetAccount(context.Background(), firstAccount.ID)
	require.NoError(t, err)

	updatedSecondAcc, err := testQueries.GetAccount(context.Background(), secondAccount.ID)
	require.NoError(t, err)

	fmt.Println(">> after:", updatedFirstAcc.Balance, updatedSecondAcc.Balance)

	require.Equal(t, firstAccount.Balance-int64(n)*amount, updatedFirstAcc.Balance)
	require.Equal(t, secondAccount.Balance+int64(n)*amount, updatedSecondAcc.Balance)
}

func TestTransferTxDeadlock(t *testing.T) {
	store := NewStore(testDB)

	firstAccount := createRandomAccount(t)
	secondAccount := createRandomAccount(t)
	fmt.Println(">> before:", firstAccount.Balance, secondAccount.Balance)

	// run n concurrent transfer trx
	n := 10
	amount := int64(10)

	errs := make(chan error)

	for i := 0; i < n; i++ {
		fromAccountID := firstAccount.ID
		toAccountID := secondAccount.ID

		if i%2 == 1 {
			fromAccountID = secondAccount.ID
			toAccountID = firstAccount.ID
		}

		go func() {
			_, err := store.TransferTx(context.Background(), TransferTxParams{
				FromAccountID: fromAccountID,
				ToAccountID:   toAccountID,
				Amount:        amount,
			})
			errs <- err
		}()
	}
	// check results
	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)
	}

	// check final updated balance
	updatedFirstAcc, err := testQueries.GetAccount(context.Background(), firstAccount.ID)
	require.NoError(t, err)

	updatedSecondAcc, err := testQueries.GetAccount(context.Background(), secondAccount.ID)
	require.NoError(t, err)

	fmt.Println(">> after:", updatedFirstAcc.Balance, updatedSecondAcc.Balance)

	require.Equal(t, firstAccount.Balance, updatedFirstAcc.Balance)
	require.Equal(t, secondAccount.Balance, updatedSecondAcc.Balance)
}
