package token

import (
	"fmt"
	"time"

	"golang.org/x/crypto/chacha20poly1305"

	"github.com/o1egl/paseto"
)

// PASETOMaker реализует интерфейс Maker с использованием библиотеки пакета PASETO
type PASETOMaker struct {
	paseto       *paseto.V2
	symmetricKey []byte
}

// NewPASETOMaker создаёт новый PASETOMaker
func NewPASETOMaker(synmetricKey string) (Maker, error) {
	if len(synmetricKey) != chacha20poly1305.KeySize {
		return nil, fmt.Errorf("invalid key size: must be exactly %d characters", chacha20poly1305.KeySize)
	}

	// Создаём новый PASETOMaker
	maker := &PASETOMaker{
		paseto:       paseto.NewV2(),
		symmetricKey: []byte(synmetricKey),
	}

	return maker, nil
}

// CreateToken создает новый токен для определенного пользователя и срока действия
func (maker *PASETOMaker) CreateToken(username string, duration time.Duration) (string, error) {
	payload, err := NewPayload(username, duration)
	if err != nil {
		return "", err
	}

	return maker.paseto.Encrypt(maker.symmetricKey, payload, nil)
}

// VerifyToken проверяет, действителен ли токен
func (maker *PASETOMaker) VerifyToken(token string) (*Payload, error) {
	payload := &Payload{}

	err := maker.paseto.Decrypt(token, maker.symmetricKey, payload, nil)
	if err != nil {
		return nil, ErrInvalidToken
	}

	err = payload.Valid()
	if err != nil {
		return nil, err
	}

	return payload, nil
}
