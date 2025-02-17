package util

const (
	USD = "USD"
	EUR = "EUR"
	CAD = "CAD"
	RUB = "RUB"
)

// IsSupportedCurrency проверяет поддерживаемую валюту
func IsSupportedCurrency(currency string) bool {
	switch currency {
	case USD, EUR, CAD, RUB:
		return true
	}
	return false
}