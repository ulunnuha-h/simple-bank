package util

var ValidCurrency = []string{"USD", "EUR", "IDR"}

func IsValidCurrency(currency string) bool {
	for _, val := range ValidCurrency {
		if currency == val { return true }
	}
	return false
}