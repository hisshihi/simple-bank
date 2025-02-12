package sqlc

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

	// запускаем n горутин
	n := 5
	// сумма перевода
	amount := int64(10)

	errs := make(chan error)
	results := make(chan TransferTxResult)

	// запускаем n горутин для передачи средств между счетами
	for range n {
		go func() {
			result, err := store.TransferTx(context.Background(), TransferTxParams{
				FromAccountID: account1.ID,
				ToAccountID:   account2.ID,
				Amount:        amount,
			})

			errs <- err
			results <- result

		}()
	}

	// проверяем результаты
	for range n {
		existed := make(map[int]bool)

		err := <-errs
		require.NoError(t, err)

		result := <-results
		require.NotEmpty(t, result)

		// проверяем транзакцию
		transfer := result.Transfer
		require.NotEmpty(t, transfer)
		require.Equal(t, account1.ID, transfer.FromAccountID)
		require.Equal(t, account2.ID, transfer.ToAccountID)
		require.Equal(t, amount, transfer.Amount)
		require.NotZero(t, transfer.ID)
		require.NotZero(t, transfer.CreatedAt)

		// проверяем существование записи в базе данных
		_, err = testQueries.GetTransfer(context.Background(), transfer.ID)
		require.NoError(t, err)

		// проверяем отправителя
		fromEntry := result.FromEntry
		require.NotEmpty(t, fromEntry)
		require.Equal(t, account1.ID, fromEntry.AccountID)
		require.Equal(t, -amount, fromEntry.Amount)
		require.NotZero(t, fromEntry.ID)
		require.NotZero(t, fromEntry.CreatedAt)

		// проверяем существование записи в базе данных
		_, err = testQueries.GetEntry(context.Background(), fromEntry.ID)
		require.NoError(t, err)

		// проверяем получателя
		toEntry := result.ToEntry
		require.NotEmpty(t, toEntry)
		require.Equal(t, account2.ID, toEntry.AccountID)
		require.Equal(t, amount, toEntry.Amount)
		require.NotZero(t, toEntry.ID)
		require.NotZero(t, toEntry.CreatedAt)

		// проверяем существование записи в базе данных
		_, err = testQueries.GetEntry(context.Background(), toEntry.ID)
		require.NoError(t, err)

		// проверяем аккаунты
		fromAccount := result.FromAccount
		require.NotEmpty(t, fromAccount)
		require.Equal(t, account1.ID, fromAccount.ID)

		toAccount := result.ToAccount
		require.NotEmpty(t, toAccount)
		require.Equal(t, account2.ID, toAccount.ID)

		fmt.Println(">>tx:", fromAccount.Balance, toAccount.Balance)
		// проверка балансов
		diff1 := account1.Balance - fromAccount.Balance
		diff2 := toAccount.Balance - account2.Balance
		require.Equal(t, diff1, diff2)
		require.True(t, diff1 > 0)
		require.True(t, diff1%amount == 0) // 1 * amount, 2 * amount, 3 * amount, ..., n * amount

		k := int(diff1 / amount)           // количество переводов
		require.True(t, k >= 1 && k <= n)  // количество переводов должно быть в диапазоне от 1 до n
		require.NotContains(t, existed, k) // k не должно быть в мапе
		existed[k] = true
	}

	// проверяем финальное обновление балансов
	updatedAccount1, err := testQueries.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)

	updatedAccount2, err := testQueries.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err)

	fmt.Println(">>afters:", updatedAccount1.Balance, updatedAccount2.Balance)
	require.Equal(t, account1.Balance-int64(n)*amount, updatedAccount1.Balance) // баланс аккаунта 1 должен быть равен начальному балансу минус количество переводов умноженное на сумму перевода
	require.Equal(t, account2.Balance+int64(n)*amount, updatedAccount2.Balance) // баланс аккаунта 2 должен быть равен начальному балансу плюс количество переводов умноженное на сумму перевода

}
