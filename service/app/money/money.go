package money

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
)

// An Amount represents a unit value of money in GBP. It is a type
// alias for int64 that allows us to attache domain specific helper methods
// for working with money. A single Amount is the same as a single Penny.
type Amount int64

const (
	// MaxAmount represents the maximum possible value for an Amount.
	MaxAmount Amount = math.MaxInt64

	// MinAmount represents the minimum possible value for an Amount.
	MinAmount Amount = math.MinInt64
)

// Common Amounts.
const (
	Penny Amount = 1
	Pound        = 100 * Penny
)

// FromPence creates an Amount from an integer representation of pennies.
func FromPence(pence int64) Amount {
	return Amount(pence)
}

// FromPounds creates an Amount from an float representation of pounds.
func FromPounds(pounds float64) Amount {
	const penniesInPound float64 = 100

	return Amount(pounds * penniesInPound)
}

// Zero creates an Amount of zero pennies.
func Zero() Amount {
	return Amount(0)
}

// Max returns the max value of the given Amounts.
func Max(a, b Amount) Amount {
	if a > b {
		return a
	}

	return b
}

// Min returns the min value of the given Amounts.
func Min(a, b Amount) Amount {
	if a < b {
		return a
	}

	return b
}

// String allows us to satisfy the fmt.Stringer interface for printing human readable values eg. -£3.50.
func (a Amount) String() string {
	end := "00"
	neg := ""

	amt := a
	if amt < 0 {
		amt = -amt
		neg = "-"
	}

	endNum := amt % Pound
	if endNum > 0 {
		end = strconv.Itoa(int(endNum))
	}

	start := strconv.Itoa(int((amt - endNum) / Pound))

	return fmt.Sprintf("%s£%s.%s", neg, fmtCommas(start), end)
}

func fmtCommas(s string) string {
	var buffer bytes.Buffer

	s = reverse(s)

	n := 3 // every 3 chars
	n1 := n - 1
	l1 := len(s) - 1

	for i, r := range s {
		buffer.WriteRune(r)

		if i%n == n1 && i != l1 {
			buffer.WriteRune(',')
		}
	}

	return reverse(buffer.String())
}

func reverse(s string) (result string) {
	for _, v := range s {
		result = string(v) + result
	}

	return
}

// Pence returns the integer representation of the Amount.
func (a Amount) Pence() int64 {
	return int64(a)
}

// Pounds returns the float representation of the Amount.
func (a Amount) Pounds() float64 {
	pounds := a / Pound
	pence := a % Pound

	return float64(pounds) + float64(pence)/100
}

// Ceil to the nearest pound.
func (a Amount) Ceil() Amount {
	return FromPounds(math.Ceil(a.Pounds()))
}

// Floor to the nearest pound.
func (a Amount) Floor() Amount {
	return FromPounds(math.Floor(a.Pounds()))
}

// Round to the nearest pound.
func (a Amount) Round() Amount {
	return FromPounds(math.Round(a.Pounds()))
}

// Trunc removes the remainder value of the pound representation of the Amount.
func (a Amount) Trunc() Amount {
	return FromPounds(math.Trunc(a.Pounds()))
}

// MarshalJSON encodes the wrapped int64 into JSON.
func (a Amount) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.Pence())
}

// UnmarshalJSON decodes JSON into the wrapped int64.
func (a *Amount) UnmarshalJSON(b []byte) error {
	var v int64

	err := json.Unmarshal(b, &v)
	if err != nil {
		return fmt.Errorf("unable to unmarshal to int64: %w", err)
	}

	*a = FromPence(v)

	return nil
}
