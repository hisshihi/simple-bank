package sqlc

import (
	"context"
	"testing"

	"github.com/hisshihi/simple-bank-go/util"
	"github.com/stretchr/testify/require"
)

func TestCreateAccount(t *testing.T) {
	arg := CreateAccountParams{
		Owner:    util.RandomOwner(),
		Balance:  util.RandomMoney(),
		Currency: util.RandomCurrency(),
	}

	account, err := testQueries.CreateAccount(context.Background(), arg) // создаем аккаунт
	require.NoError(t, err) // проверяем, что нет ошибок
	require.NotEmpty(t, account) // проверяем, что не пустой

	require.Equal(t, arg.Owner, account.Owner) // проверяем, что owner совпадает
	require.Equal(t, arg.Balance, account.Balance) // проверяем, что balance совпадает
	require.Equal(t, arg.Currency, account.Currency) // проверяем, что currency совпадает

	require.NotZero(t, account.ID) // проверяем, что id не равен 0
	require.NotZero(t, account.CreatedAt) // проверяем, что created_at не равен 0
}