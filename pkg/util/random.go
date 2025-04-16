package util

import "math/rand"

func RandomCurrency() string {
	currencies := []string{EUR, USD, RUB}
	n := len(currencies)
	return currencies[rand.Intn(n)]
}
