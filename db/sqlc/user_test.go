package sqlc

import (
	"context"
	"testing"
	"time"

	"github.com/hisshihi/simple-bank-go/util"
	"github.com/stretchr/testify/require"
)

func createRandomUser(t *testing.T) User {
	arg := CreateUserParams{
		Username:       util.RandomOwner(),
		HashedPassword: "secret",
		FullName:       util.RandomOwner(),
		Email:          util.RandomEmail(),
	}

	user, err := testQueries.CreateUser(context.Background(), arg) // создаем аккаунт
	require.NoError(t, err)                                        // проверяем, что нет ошибок
	require.NotEmpty(t, user)                                      // проверяем, что не пустой

	require.Equal(t, arg.Username, user.Username)             // проверяем, что Username совпадает
	require.Equal(t, arg.HashedPassword, user.HashedPassword) // проверяем, что HashedPassword совпадает
	require.Equal(t, arg.FullName, user.FullName)             // проверяем, что full_name совпадает
	require.Equal(t, arg.Email, user.Email)                   // проверяем, что email совпадает

	require.NotZero(t, user.CreatedAt) // проверяем, что created_at не равен 0

	require.True(t, user.PasswordChangeAt.IsZero())

	return user
}

func TestCreateUser(t *testing.T) {
	createRandomUser(t)
}

func TestGetUser(t *testing.T) {
	// создаем аккаунт
	user1 := createRandomUser(t)
	user2, err := testQueries.GetUser(context.Background(), user1.Username)
	require.NoError(t, err)
	require.NotEmpty(t, user2)

	require.Equal(t, user1.Username, user2.Username)
	require.Equal(t, user1.HashedPassword, user2.HashedPassword)
	require.Equal(t, user1.FullName, user2.FullName)
	require.Equal(t, user1.Email, user2.Email)
	require.WithinDuration(t, user1.PasswordChangeAt, user2.PasswordChangeAt, time.Second) // проверяем, что PasswordChangeAt совпадает
	require.WithinDuration(t, user1.CreatedAt, user2.CreatedAt, time.Second) // проверяем, что created_at совпадает
}
