package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	mockdb "github.com/hisshihi/simple-bank-go/db/mock"
	"github.com/hisshihi/simple-bank-go/db/sqlc"
	"github.com/hisshihi/simple-bank-go/token"
	"github.com/hisshihi/simple-bank-go/util"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func TestGetUserAPI(t *testing.T) {
	user := randomUser()

	testCases := []struct {
		name          string
		username      string
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:     "OK",
			username: user.Username,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(user.Username)).
					Times(1).
					Return(user, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchUser(t, recorder.Body, user)
			},
		},
		{
			name:     "NotFound",
			username: user.Username,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(user.Username)).
					Times(1).
					Return(sqlc.User{}, sql.ErrNoRows)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:     "BadRequest",
			username: "hi%23ss",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:     "InternalError",
			username: user.Username,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(user.Username)).
					Times(1).
					Return(sqlc.User{}, sql.ErrConnDone)
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
			url := fmt.Sprintf("/users/%s", tc.username)
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

func TestCreateUserAPI(t *testing.T) {
	user := randomUser()

	testCases := []struct {
		name          string
		body          json.RawMessage
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: json.RawMessage(fmt.Sprintf(`{"username": "%s", "full_name": "%s", "email": "%s", "password": "%s"}`, user.Username, user.FullName, user.Email, "secret123")),
			buildStubs: func(store *mockdb.MockStore) {
				arg := sqlc.CreateUserParams{
					Username: user.Username,
					FullName: user.FullName,
					Email:    user.Email,
				}
				store.EXPECT().
					CreateUser(gomock.Any(), eqCreateUserParamsMatcher{
						arg:           arg,
						plainPassword: "secret123",
					}).
					Times(1).
					Return(user, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchUser(t, recorder.Body, user)
			},
		},
		{
			name: "BadRequest",
			body: json.RawMessage(fmt.Sprintf(`{"username": "%s", "full_name": "%s", "email": "%s", "password": "%s"}`, user.Username, user.FullName, "hiss", "secret123")),
			buildStubs: func(store *mockdb.MockStore) {
				arg := sqlc.CreateUserParams{
					Username: user.Username,
					FullName: user.FullName,
					Email:    "hiss",
				}
				store.EXPECT().
					CreateUser(gomock.Any(), eqCreateUserParamsMatcher{
						arg:           arg,
						plainPassword: "secret123",
					}).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InternalError",
			body: json.RawMessage(fmt.Sprintf(`{"username": "%s", "full_name": "%s", "email": "%s", "password": "%s"}`, user.Username, user.FullName, user.Email, "secret123")),
			buildStubs: func(store *mockdb.MockStore) {
				arg := sqlc.CreateUserParams{
					Username: user.Username,
					FullName: user.FullName,
					Email:    user.Email,
				}
				store.EXPECT().
					CreateUser(gomock.Any(), eqCreateUserParamsMatcher{
						arg:           arg,
						plainPassword: "secret123",
					}).
					Times(1).
					Return(sqlc.User{}, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "InternalErrorPassword",
			body: json.RawMessage(fmt.Sprintf(`{"username": "%s", "full_name": "%s", "email": "%s", "password": "%s"}`, user.Username, user.FullName, user.Email, "errorPassword")),
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "StatusForbidden",
			body: json.RawMessage(fmt.Sprintf(`{"username": "%s", "full_name": "%s", "email": "%s", "password": "%s"}`, user.Username, user.FullName, user.Email, "secret123")),
			buildStubs: func(store *mockdb.MockStore) {
				arg := sqlc.CreateUserParams{
					Username: user.Username,
					FullName: user.FullName,
					Email:    user.Email,
				}
				store.EXPECT().
					CreateUser(gomock.Any(), eqCreateUserParamsMatcher{
						arg:           arg,
						plainPassword: "secret123",
					}).
					Times(1).
					Return(sqlc.User{}, &pq.Error{
						Code:    "23505",
						Message: "unique_violation",
					})
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, recorder.Code)
			},
		},
		{
			name: "ForeignKeyViolation",
			body: json.RawMessage(fmt.Sprintf(`{"username": "%s", "full_name": "%s", "email": "%s", "password": "%s"}`,
				user.Username, user.FullName, user.Email, "secret123")),
			buildStubs: func(store *mockdb.MockStore) {
				arg := sqlc.CreateUserParams{
					Username: user.Username,
					FullName: user.FullName,
					Email:    user.Email,
				}
				store.EXPECT().
					CreateUser(gomock.Any(), eqCreateUserParamsMatcher{
						arg:           arg,
						plainPassword: "secret123",
					}).
					Times(1).
					Return(sqlc.User{}, &pq.Error{
						Code:    "foreign_key_violation", // Эмулируем нарушение внешнего ключа
						Message: "foreign key violation error",
					})
			},
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
			url := "/users"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(tc.body))
			require.NoError(t, err)

			// Запускает сервер и обрабатывает запрос
			server.router.ServeHTTP(recorder, request)

			// Проверяет ответ
			tc.checkResponse(recorder)
		})
	}
}

func randomUser() sqlc.User {
	hashedPassword, err := util.HashPassword(util.RandomString(6))
	if err != nil {
		log.Fatal(err)
	}

	return sqlc.User{
		Username:       util.RandomOwner(),
		FullName:       util.RandomOwner(),
		Email:          util.RandomEmail(),
		HashedPassword: hashedPassword,
	}
}

func requireBodyMatchUser(t *testing.T, body *bytes.Buffer, user sqlc.User) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotUser sqlc.User
	err = json.Unmarshal(data, &gotUser)
	require.NoError(t, err)

	user.HashedPassword = ""
	require.Equal(t, user, gotUser)
}

type eqCreateUserParamsMatcher struct {
	arg           sqlc.CreateUserParams
	plainPassword string
}

func (e eqCreateUserParamsMatcher) Matches(x interface{}) bool {
	arg, ok := x.(sqlc.CreateUserParams)
	if !ok {
		return false
	}

	// Проверяем, что сгенерированный хэш соответствует исходному паролю
	err := util.CheckPassword(e.plainPassword, arg.HashedPassword)
	if err != nil {
		return false
	}

	// Присваиваем сгенерированный хэш для дальнейшего сравнения остальных полей.
	e.arg.HashedPassword = arg.HashedPassword
	return reflect.DeepEqual(e.arg, arg)
}

func (e eqCreateUserParamsMatcher) String() string {
	return fmt.Sprintf("matches arg %v and password %v", e.arg, e.plainPassword)
}
