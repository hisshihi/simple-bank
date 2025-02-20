package token

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const minSecretKeySize = 32

type JWTMaker struct {
	secretKey string
}

func NewJWTMaker(secretKey string) (Maker, error) {
	if len(secretKey) < minSecretKeySize {
		return nil, fmt.Errorf("invalid key size: must be at least %d characters", minSecretKeySize)
	}

	return &JWTMaker{secretKey}, nil
}

// CreateToken создает новый токен для определенного пользователя и срока действия
func (maker *JWTMaker) CreateToken(username string, duration time.Duration) (string, error) {
	payload, err := NewPayload(username, duration)
	if err != nil {
		return "", err
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	return jwtToken.SignedString([]byte(maker.secretKey))
}

// VerifyToken проверяет, действителен ли токен, выполняя последовательную проверку его подписи и данных.
func (maker *JWTMaker) VerifyToken(token string) (*Payload, error) {
	// Определяем функцию keyFunc, которая предоставляет ключ для проверки подписи токена.
	// Эта функция вызывается библиотекой jwt для извлечения секретного ключа.
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		// Проверяем, что метод подписания токена соответствует HMAC.
		// Это гарантирует, что токен подписан с использованием ожидаемого алгоритма.
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			// Если метод подписи не является HMAC, возвращаем ошибку "недопустимый токен".
			return nil, ErrInvalidToken
		}
		// Если проверка прошла успешно, возвращаем секретный ключ в виде байтового среза.
		return []byte(maker.secretKey), nil
	}

	// Разбираем токен и выполняем его валидацию с помощью функции jwt.ParseWithClaims.
	// Здесь:
	// - первый аргумент — это строковое представление токена;
	// - второй аргумент — указатель на пустой объект Payload, куда будут записаны данные токена;
	// - третий аргумент — функция keyFunc, возвращающая ключ подписи.
	jwtToken, err := jwt.ParseWithClaims(token, &Payload{}, keyFunc)
	if err != nil {
		// Если возникла ошибка при разборе токена, проверяем, вызвана ли она истечением срока действия.
		if errors.Is(err, jwt.ErrTokenExpired) {
			// Если токен просрочен, возвращаем ошибку "токен истёк".
			return nil, ErrExpiredToken
		}
		// Для всех остальных ошибок возвращаем ошибку "токен недействителен".
		return nil, ErrInvalidToken
	}

	// После успешного разбора токена, выполняем приведение типа токеновых данных к *Payload.
	payload, ok := jwtToken.Claims.(*Payload)
	if !ok {
		// Если приведение типа не удалось, это означает, что структура токена не соответствует ожидаемой.
		return nil, ErrInvalidToken
	}

	// Если все этапы проверки пройдены, возвращаем полученный payload, содержащий полезные данные токена.
	return payload, nil
}
