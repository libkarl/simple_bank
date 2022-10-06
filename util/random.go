package util

import (
	"math/rand"
	"strings"
	"time"
)

const alphabet = "abcdefghijklmnopqrstuvwxyz"

func init() {
	rand.Seed(time.Now().UnixNano())
}

// Random Int generates a random integer between min and max
func RandomInt(min, max int64) int64 {
	return min + rand.Int63n(max-min+1)
}

// RandomString generates a ranodm string
func RandomString(n int) string {
	var sb strings.Builder
	k := len(alphabet)

	for i := 0; i < n; i++ {
		c := alphabet[rand.Intn(k)]
		sb.WriteByte(c)
	}

	return sb.String()
}

// Random Owner generates random owner name
func RandomOwner() string {
	return RandomString(6)
}

// Random Money generate random amount of money
func RandomMoney() int64 {
	return RandomInt(0, 1000)
}

// Random Currency generates random currency

func RandomCurrency() string {
	currencies := []string{"USD", "EUR", "CAD"}
	n := len(currencies)
	// vrátí typ měny podle náhodně vybraného indexu
	// z intervalu 0 až n -> konec intervalu jsem získal z len() funkce
	return currencies[rand.Intn(n)]
}
