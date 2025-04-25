package util

import "time"

type Maker interface {
	// CreateToken creates a new token for the given username and duration
	CreateToken(username string, duration time.Duration) (string, *Payload, error)

	// VerifyToken verifies the token and returns the payload
	VerifyToken(token string) (*Payload, error)
}
