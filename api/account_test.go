package api

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	mockdb "github.com/hisshihi/simple-bank-go/db/mock"
	"github.com/hisshihi/simple-bank-go/db/sqlc"
	"github.com/hisshihi/simple-bank-go/token"
	"github.com/hisshihi/simple-bank-go/util"
	"github.com/stretchr/testify/require"
)

func TestGetAccountAPI(t *testing.T) {
	user := randomUser()
	account := randomAccount(user.Username)

	// Создаём тестовые случаи
	testCases := []struct {
		name          string
		accountID     int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker) // Функция настройки авторизации запроса.
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:      "OK",
			accountID: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				// Ожидаем вызов GetAccount с любым контекстом и аргументом, равным account.ID
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(account, nil)
			},
			// Проверяет ответ
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccount(t, recorder.Body, account)
			},
		},
		{
			name:      "UnauthorizedUser",
			accountID: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, "unauthorized", time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				// Ожидаем вызов GetAccount с любым контекстом и аргументом, равным account.ID
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(account, nil)
			},
			// Проверяет ответ
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:      "NoAuthorization",
			accountID: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				// Ожидаем вызов GetAccount с любым контекстом и аргументом, равным account.ID
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			// Проверяет ответ
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:      "NotFound",
			accountID: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				// Ожидаем вызов GetAccount с любым контекстом и аргументом, равным account.ID
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(sqlc.Account{}, sql.ErrNoRows)
			},
			// Проверяет ответ
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:      "InternalServerError",
			accountID: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				// Ожидаем вызов GetAccount с любым контекстом и аргументом, равным account.ID
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(sqlc.Account{}, sql.ErrConnDone)
			},
			// Проверяет ответ
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:      "InvalidID",
			accountID: 0,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				// Ожидаем вызов GetAccount с любым контекстом и аргументом, равным account.ID
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			// Проверяет ответ
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			// Создаём контроллер для мок-объектов
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Создаём мок-объект для Store
			store := mockdb.NewMockStore(ctrl)
			// Строит мок-объект
			tc.buildStubs(store)

			// Создаём новый HTTP-сервер с мок-объектом в качестве аргумента
			server := newTestServer(t, store)
			recorder := httptest.NewRecorder() // Записывает ответы сервера

			// Указываем параметр равный каждому тестовому случаю
			url := fmt.Sprintf("/accounts/%d", tc.accountID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)

			// Запускает сервер и обрабатывает запрос
			server.router.ServeHTTP(recorder, request)

			// Проверяет ответ
			tc.checkResponse(recorder)
		})
	}
}

// Тест API для создания аккаунта. Этот тест проверяет, что при отправке POST запроса на создание аккаунта,
// сервер корректно обрабатывает запрос и возвращает созданный аккаунт с нужными данными.
func TestCreateAccountAPI(t *testing.T) {
	// Создаем случайный аккаунт для тестирования. Функция randomAccount генерирует аккаунт с рандомными данными.
	user := randomUser()
	account := randomAccount(user.Username)

	// Определяем набор тестовых сценариев, каждый из которых описывает конкретный случай:
	// - name: Имя тестового сценария.
	// - body: Тело запроса в JSON, которое будет отправлено на сервер.
	// - buildStubs: Функция для настройки мок-объекта (имитации базы данных) с ожидаемым поведением.
	// - checkResponse: Функция для проверки корректности ответа сервера.
	testCases := []struct {
		name          string
		body          json.RawMessage
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker) // Функция настройки авторизации запроса.
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			// Сценарий "OK" — успешное создание аккаунта.
			name: "OK",
			// Формируем JSON строку с параметрами "owner" и "currency", используя данные из сгенерированного аккаунта.
			body: json.RawMessage(fmt.Sprintf(`{"owner": "%s", "currency": "%s"}`, account.Owner, account.Currency)),
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			// buildStubs настраивает мок-объект базы данных так, чтобы при вызове CreateAccount возвращался наш аккаунт.
			buildStubs: func(store *mockdb.MockStore) {
				// Создаем аргументы для метода CreateAccount. Здесь balance всегда 0,
				// а owner и currency берутся из сгенерированного аккаунта.
				arg := sqlc.CreateAccountParams{
					Owner:    account.Owner,
					Balance:  0,
					Currency: account.Currency,
				}
				// Ожидаем, что метод CreateAccount будет вызван один раз с любым контекстом и с аргументом, равным arg.
				// Возвращаем account и nil, чтобы имитировать успешное создание аккаунта.
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(account, nil)
			},
			// checkResponse проверяет, что ответ сервера имеет статус 200 (OK) и тело ответа соответствует созданному аккаунту.
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccount(t, recorder.Body, account)
			},
		},
		{
			name: "BadRequest",
			body: json.RawMessage(fmt.Sprintf(`{"owner": "%s", "currency": "%s"}`, account.Owner, "YEN")),
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				// Т.к. поле "owner" пустое, валидация должна не допустить вызова метода CreateAccount,
				// поэтому ожидаем, что этот метод не будет вызван вовсе.
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InternalServerError",
			body: json.RawMessage(fmt.Sprintf(`{"owner": "%s", "currency": "%s"}`, account.Owner, account.Currency)),
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Any()).
					Times(1).
					Return(account, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
				requireBodyMatchAccount(t, recorder.Body, sqlc.Account{})
			},
		},
	}

	// Перебираем все тестовые сценарии.
	for i := range testCases {
		tc := testCases[i]

		// t.Run запускает под-тест с именем, указанным в тестовом сценарии.
		t.Run(tc.name, func(t *testing.T) {
			// Создаем контроллер для управления мок-объектами.
			ctrl := gomock.NewController(t)
			// defer ctrl.Finish() гарантирует, что после выполнения теста будут проверены все ожидания мока.
			defer ctrl.Finish()

			// Создаем мок-объект для Store, который эмулирует работу с базой данных.
			store := mockdb.NewMockStore(ctrl)
			// Настраиваем мок-объект согласно сценарию теста (описываем ожидаемое поведение).
			tc.buildStubs(store)

			// Создаем новый сервер с нашим мок-объектом Store.
			server := newTestServer(t, store)
			// Создаем объект recorder, который перехватывает ответ HTTP сервера.
			recorder := httptest.NewRecorder()

			// Определяем URL для запроса создания аккаунта.
			url := "/accounts"
			// Создаем новый HTTP запрос с методом POST.
			// bytes.NewReader(tc.body) используется для передачи тела запроса, содержащего JSON с данными аккаунта.
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(tc.body))
			// Проверяем, что при создании запроса не возникло ошибок.
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)

			// Отправляем запрос через роутер нашего сервера. Это симулирует обработку запроса, как если бы он поступил по HTTP.
			server.router.ServeHTTP(recorder, request)

			// Вызываем функцию проверки ответа, которая сверяет фактический ответ с ожидаемым.
			tc.checkResponse(recorder)
		})
	}
}

// Функция TestListAccountsAPI тестирует API для получения списка аккаунтов.
func TestListAccountsAPI(t *testing.T) {
	user := randomUser()

	n := 5
	accounts := make([]sqlc.Account, n)
	for v := range accounts {
		accounts[v] = randomAccount(user.Username)
	}

	testCases := []struct {
		name          string
		queryParams   string
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker) // Функция настройки авторизации запроса.
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:        "OK",
			queryParams: "?page_id=1&page_size=5",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				// Вычисляем аргументы для метода ListAccounts.
				arg := sqlc.ListAccountsParams{
					Owner:  user.Username,
					Limit:  5,
					Offset: 0, // (page_id - 1) * page_size = (1-1)*5 = 0
				}
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(accounts, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccounts(t, recorder.Body, accounts)
			},
		},
		{
			name:        "BadRequest",
			queryParams: "?page_id=0&page_size=12",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:        "InternalServerError",
			queryParams: "?page_id=1&page_size=5",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Any()).
					Times(1).
					Return(nil, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			url := "/accounts" + tc.queryParams
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)

			server.router.ServeHTTP(recorder, request)

			tc.checkResponse(recorder)
		})
	}
}

// Тест API для обновления баланса аккаунта.
func TestUpdateAccountAPI(t *testing.T) {
	user := randomUser()
	account := randomAccount(user.Username)
	newBalance := util.RandomMoney()

	updatedAccount := account
	updatedAccount.Balance = newBalance

	t.Logf("Исходный баланс: %d, Новый баланс: %d", account.Balance, newBalance)

	testCases := []struct {
		name          string
		accountID     int64
		body          json.RawMessage
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker) // Функция настройки авторизации запроса.
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:      "OK",
			accountID: account.ID,
			body:      json.RawMessage(fmt.Sprintf(`{"balance": %d}`, newBalance)),
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := sqlc.UpdateAccountParams{
					ID:      account.ID,
					Balance: newBalance,
				}
				store.EXPECT().
					UpdateAccount(gomock.Any(), gomock.Eq(arg)).
					Do(func(ctx context.Context, params sqlc.UpdateAccountParams) {
						// Выводим в лог параметры, с которыми вызывается метод обновления
						fmt.Printf("Вызов UpdateAccount с параметрами: %+v\n", params)
					}).
					Times(1).
					Return(updatedAccount, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccount(t, recorder.Body, updatedAccount)
			},
		},
		{
			name:      "NotFound",
			accountID: account.ID,
			body:      json.RawMessage(fmt.Sprintf(`{"balance": %d}`, newBalance)),
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := sqlc.UpdateAccountParams{
					ID:      account.ID,
					Balance: newBalance,
				}
				store.EXPECT().
					UpdateAccount(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(sqlc.Account{}, sql.ErrNoRows)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:      "InternalServerError",
			accountID: account.ID,
			body:      json.RawMessage(fmt.Sprintf(`{"balance": %d}`, newBalance)),
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := sqlc.UpdateAccountParams{
					ID:      account.ID,
					Balance: newBalance,
				}
				store.EXPECT().
					UpdateAccount(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(sqlc.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:      "BadRequestUri",
			accountID: 0,
			body:      json.RawMessage(fmt.Sprintf(`{"balance": %d}`, newBalance)),
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := sqlc.UpdateAccountParams{
					ID:      account.ID,
					Balance: newBalance,
				}
				store.EXPECT().
					UpdateAccount(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:      "BadRequestBody",
			accountID: account.ID,
			body:      json.RawMessage(fmt.Sprintf(`{"balance": "not a number"}`)),
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := sqlc.UpdateAccountParams{
					ID:      account.ID,
					Balance: newBalance,
				}
				store.EXPECT().
					UpdateAccount(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/accounts/%d", tc.accountID)
			request, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(tc.body))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)

			server.router.ServeHTTP(recorder, request)

			// Выводим статус и тело ответа для отладки
			fmt.Printf("Статус ответа: %d\n", recorder.Code)
			fmt.Printf("Тело ответа: %s\n", recorder.Body.String())

			tc.checkResponse(recorder)
		})
	}
}

// Тест API для удаления аккаунта.
func TestDeleteAccountAPI(t *testing.T) {
	user := randomUser()
	account := randomAccount(user.Username)

	testCases := []struct {
		name          string
		accountID     int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:      "OK",
			accountID: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNoContent, recorder.Code)
			},
		},
		{
			name:      "BadRequest",
			accountID: 0,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:      "NotFound",
			accountID: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(sql.ErrNoRows)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:      "InternalServerError",
			accountID: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			// Создаём контроллер для мок-объектов
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Создаём мок-объект для Store
			store := mockdb.NewMockStore(ctrl)
			// Строит мок-объект
			tc.buildStubs(store)

			// Создаём новый HTTP-сервер с мок-объектом в качестве аргумента
			server := newTestServer(t, store)
			recorder := httptest.NewRecorder() // Записывает ответы сервера

			// Указываем параметр равный каждому тестовому случаю
			url := fmt.Sprintf("/accounts/%d", tc.accountID)
			request, err := http.NewRequest(http.MethodDelete, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)

			// Запускает сервер и обрабатывает запрос
			server.router.ServeHTTP(recorder, request)

			// Проверяет ответ
			tc.checkResponse(recorder)
		})
	}
}

func randomAccount(owner string) sqlc.Account {
	return sqlc.Account{
		ID:       util.RandomInt(1, 1000),
		Owner:    owner,
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

func requireBodyMatchAccounts(t *testing.T, body *bytes.Buffer, accounts []sqlc.Account) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotAccounts []sqlc.Account
	err = json.Unmarshal(data, &gotAccounts)
	require.NoError(t, err)
	require.Equal(t, accounts, gotAccounts)
}
