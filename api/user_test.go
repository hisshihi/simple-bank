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
	"testing"

	"github.com/golang/mock/gomock"
	mockdb "github.com/hisshihi/simple-bank-go/db/mock"
	"github.com/hisshihi/simple-bank-go/db/sqlc"
	"github.com/hisshihi/simple-bank-go/util"
	"github.com/stretchr/testify/require"
)

// TODO: Добавить тесты api для полного покрытия кода
func TestGetUserAPI(t *testing.T) {
	user := randomUser()
	
	testCases := []struct {
		name string
		username string
		buildStubs func(store *mockdb.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			username: user.Username,
			buildStubs: func(store *mockdb.MockStore){
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(user.Username)).
					Times(1).
					Return(user, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder){
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchUser(t, recorder.Body, user)
			},
		},
		{
			name: "NotFound",
			username: user.Username,
			buildStubs: func(store *mockdb.MockStore){
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(user.Username)).
					Times(1).
					Return(sqlc.User{}, sql.ErrNoRows)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder){
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "BadRequest",
			username: "hi%23ss",
			buildStubs: func(store *mockdb.MockStore){
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder){
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InternalError",
			username: user.Username,
			buildStubs: func(store *mockdb.MockStore){
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(user.Username)).
					Times(1).
					Return(sqlc.User{}, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder){
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
			server := NewServer(store)
			recorder := httptest.NewRecorder() // Записывает ответы сервера

			// Указываем параметр равный каждому тестовому случаю
			url := fmt.Sprintf("/users/%s", tc.username)
			request, err := http.NewRequest(http.MethodGet, url, nil)
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
		Username: util.RandomOwner(),
		FullName: util.RandomOwner(),
		Email:    util.RandomEmail(),
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