package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	mockdb "github.com/hisshihi/simple-bank/db/mock"
	"github.com/hisshihi/simple-bank/db/sqlc"
	"github.com/hisshihi/simple-bank/util"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestGetAccountAPI(t *testing.T) {
	// Создаём тестовый аккаунт с случайными данными
	account := randomAccount()

	testCases := []struct {
		name          string
		accountID     int64
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:      "OK",
			accountID: account.ID,
			buildStubs: func(store *mockdb.MockStore) {
				// Настраиваем ожидаемое поведение мока:
				store.EXPECT().
					// Ожидаем вызов метода GetAccount
					GetAccount(
						// gomock.Any() означает, что контекст может быть любым
						gomock.Any(),
						// Ожидаем точное совпадение ID аккаунта
						gomock.Eq(account.ID),
					).
					// Метод должен быть вызван ровно 1 раз
					Times(1).
					// При вызове должен вернуть наш тестовый аккаунт и nil ошибку
					Return(account, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				// Проверяем, что код ответа равен 200 (OK)
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccount(t, recorder.Body, account)
			},
		},
		{
			name:      "NotFound",
			accountID: account.ID,
			buildStubs: func(store *mockdb.MockStore) {
				// Настраиваем ожидаемое поведение мока:
				store.EXPECT().
					// Ожидаем вызов метода GetAccount
					GetAccount(
						// gomock.Any() означает, что контекст может быть любым
						gomock.Any(),
						// Ожидаем точное совпадение ID аккаунта
						gomock.Eq(account.ID),
					).
					// Метод должен быть вызван ровно 1 раз
					Times(1).
					// При вызове должен вернуть наш тестовый аккаунт и nil ошибку
					Return(sqlc.Account{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				// Проверяем, что код ответа равен 200 (OK)
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:      "InternalError",
			accountID: account.ID,
			buildStubs: func(store *mockdb.MockStore) {
				// Настраиваем ожидаемое поведение мока:
				store.EXPECT().
					// Ожидаем вызов метода GetAccount
					GetAccount(
						// gomock.Any() означает, что контекст может быть любым
						gomock.Any(),
						// Ожидаем точное совпадение ID аккаунта
						gomock.Eq(account.ID),
					).
					// Метод должен быть вызван ровно 1 раз
					Times(1).
					// При вызове должен вернуть наш тестовый аккаунт и nil ошибку
					Return(sqlc.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				// Проверяем, что код ответа равен 200 (OK)
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:      "InvalidID",
			accountID: 0,
			buildStubs: func(store *mockdb.MockStore) {
				// Настраиваем ожидаемое поведение мока:
				store.EXPECT().
					// Ожидаем вызов метода GetAccount
					GetAccount(
						// gomock.Any() означает, что контекст может быть любым
						gomock.Any(),
						// Ожидаем точное совпадение ID аккаунта
						gomock.Any(),
					).
					// Метод должен быть вызван ровно 1 раз
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				// Проверяем, что код ответа равен 200 (OK)
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			// Создаём новый контроллер для управления моками
			// gomock - это библиотека для создания mock-объектов в Go
			ctrl := gomock.NewController(t)
			// Освобождаем ресурсы после завершения теста
			defer ctrl.Finish()

			// Создаём мок хранилища данных
			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			// Создаём тестовый HTTP сервер с нашим мок-хранилищем
			server := NewServer(store)
			// Создаём ResponseRecorder для записи ответа сервера
			recorder := httptest.NewRecorder()

			// Формируем URL для запроса с ID аккаунта
			url := fmt.Sprintf("/accounts/%d", tc.accountID)
			// Создаём новый GET запрос
			request, err := http.NewRequest(http.MethodGet, url, nil)
			// Проверяем, что запрос создан без ошибок
			require.NoError(t, err)

			// Отправляем запрос на тестовый сервер
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func randomAccount() sqlc.Account {
	return sqlc.Account{
		ID:       util.RandomInt(1, 1000),
		Owner:    util.RandomOwner(),
		Balance:  util.RandomMoney(),
		Currency: util.RandomCurrency(),
	}
}

func requireBodyMatchAccount(t *testing.T, body *bytes.Buffer, account sqlc.Account) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotAccount sqlc.Account
	err = json.Unmarshal(data, &gotAccount)
	require.NoError(t, err)
	require.Equal(t, account, gotAccount)
}
