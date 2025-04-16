package util

const (
	USD = "USD"
	EUR = "EUR"
	RUB = "RUB"
)

func IsSupportedCurrnecy(currency string) bool {
	switch currency {
	case USD, EUR, RUB:
		return true
	}
	return false
}
