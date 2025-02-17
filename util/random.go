package util

import (
	"math/rand"
	"strings"
	"time"
)

/*
Для версий Go ниже 1.20, можно использовать следующий код:

rand.Seed(time.Now().UnixNano())

*/

func init() {
	rand.Seed(time.Now().UnixNano())
}

const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// RandomInt возвращает случайное число между min и max
func RandomInt(min, max int64) int64 {
	return min + rand.Int63n(max-min+1)
}

// RandomString возвращает случайную строку длиной n
func RandomString(n int) string {
	var sb strings.Builder
	k := len(alphabet)

	for range n {
		c := alphabet[rand.Intn(k)]
		sb.WriteByte(c)
	}

	return sb.String()
}

// RandomOwner возвращает случайную строку длиной 6
func RandomOwner() string {
	return RandomString(6)
}

// RandomMoney возвращает случайное число между 0 и 10000
func RandomMoney() int64 {
	return RandomInt(0, 10000)
}

// RandomCurrency возвращает случайное значение из списка валют
func RandomCurrency() string {
	currencies := []string{EUR, USD, CAD, RUB}
	n := len(currencies)
	return currencies[rand.Intn(n)]
}
