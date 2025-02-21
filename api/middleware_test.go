package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hisshihi/simple-bank-go/token"
	"github.com/stretchr/testify/require"
)

func addAuthorization(
	t *testing.T,
	request *http.Request,
	tokenMaker token.Maker,
	authorizationType string,
	username string,
	duration time.Duration,
) {
	token, err := tokenMaker.CreateToken(username, duration)
	require.NoError(t, err)

	authorizationHeader := fmt.Sprintf("%s %s", authorizationType, token)
	request.Header.Set(authorizationHeaderKey, authorizationHeader)
}

// Функция Test_authMiddleware выполняет тестирование middleware для проверки авторизации.
func Test_authMiddleware(t *testing.T) {
	// Определяем набор тестовых случаев. Каждый случай описывается структурой, содержащей:
	// - name: описание (имя) тестового случая.
	// - setupAuth: функция, которая настраивает авторизацию (например, устанавливает необходимые заголовки) в HTTP-запросе.
	// - checkResponse: функция, которая проверяет корректность ответа, полученного от сервера.
	testCases := []struct {
		name          string // Описание тестового случая.
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)    // Функция настройки авторизации запроса.
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder) // Функция проверки HTTP-ответа.
	}{
		{
			name: "OK",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, "user", time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "NoAuthorization",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "UnsupportedAuthorization",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, "unsupported", "user", time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "InvalidAuthorizationFormat",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, "", "user", time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "ExpiredToken",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, "user", -time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
	}

	// Проходимся по каждому тестовому случаю с помощью цикла.
	for i := range testCases {
		// Извлекаем текущий тестовый случай.
		tc := testCases[i]

		// Запускаем под-тест с именем, заданным в поле tc.name.
		t.Run(tc.name, func(t *testing.T) {
			// Создаем новый тестовый сервер. Функция newTestServer инициализирует сервер с маршрутизатором и tokenMaker.
			server := newTestServer(t, nil)

			// Определяем путь маршрута, по которому будет производиться проверка авторизации.
			authPath := "/auth"

			// Регистрируем маршрут GET для пути authPath на маршрутизаторе сервера.
			// При обработке запроса сначала вызывается authMiddleware, который проверяет валидность токена,
			// а затем, если авторизация успешна, вызывается функция-обработчик, возвращающая HTTP статус 200 с пустым JSON.
			server.router.GET(
				authPath,
				// authMiddleware: middleware, ответственное за проверку действительности токена.
				authMiddleware(server.tokenMaker),
				// Обработчик, отправляющий пустой JSON-ответ с HTTP статусом 200 (OK).
				func(ctx *gin.Context) {
					ctx.JSON(http.StatusOK, gin.H{})
				},
			)

			// Создаем новый объект ResponseRecorder для перехвата HTTP ответа сервера.
			recorder := httptest.NewRecorder()

			// Формируем новый HTTP GET запрос по пути authPath. Тело запроса отсутствует (nil).
			request, err := http.NewRequest(http.MethodGet, authPath, nil)
			// Проверяем, что не возникла ошибка при создании запроса.
			require.NoError(t, err)

			// Вызываем функцию setupAuth тестового случая, чтобы настроить авторизацию в запросе.
			// Эта функция может, например, установить нужные заголовки с токеном.
			tc.setupAuth(t, request, server.tokenMaker)

			// Передаем запрос в маршрутизатор сервера, который обрабатывает его с учетом установленного middleware,
			// а результат записывается в recorder.
			server.router.ServeHTTP(recorder, request)

			// Вызываем функцию checkResponse тестового случая для проверки корректности полученного ответа.
			tc.checkResponse(t, recorder)
		})
	}
}
