package util

// Constants with supported currencies
const (
	USD = "USD"
	EUR = "EUR"
	CAD = "CAD"
)

// It returns true if the currency inside the input is supported 
func IsSupportedCurrency(currency string) bool{
	switch currency {
	case USD, EUR, CAD:
		return true
	}
	return false
}
