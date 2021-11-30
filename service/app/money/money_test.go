package money_test

import (
	"encoding/json"
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/nickbryan/go-template/service/app/money"
)

func TestFromPence(t *testing.T) {
	t.Parallel()

	tests := []struct {
		pence    int64
		expected money.Amount
	}{
		{pence: 0, expected: money.Amount(0)},
		{pence: 123, expected: money.Amount(123)},
		{pence: -123, expected: money.Amount(-123)},
	}

	for _, tc := range tests {
		assert.Equal(t, tc.expected, money.FromPence(tc.pence))
	}
}

func TestFromPounds(t *testing.T) {
	t.Parallel()

	tests := []struct {
		pounds   float64
		expected money.Amount
	}{
		{pounds: 0, expected: money.Amount(0)},
		{pounds: 0.00, expected: money.Amount(0)},
		{pounds: 0.000001, expected: money.Amount(0)},
		{pounds: -0.0001, expected: money.Amount(0)},
		{pounds: 0.01, expected: money.Amount(1)},
		{pounds: 0.12, expected: money.Amount(12)},
		{pounds: 1.23, expected: money.Amount(123)},
		{pounds: 101.23, expected: money.Amount(10123)},
		{pounds: -0, expected: money.Amount(0)},
		{pounds: -0.00, expected: money.Amount(0)},
		{pounds: -0.01, expected: money.Amount(-1)},
		{pounds: -0.12, expected: money.Amount(-12)},
		{pounds: -1.23, expected: money.Amount(-123)},
		{pounds: -101.23, expected: money.Amount(-10123)},
	}

	for _, tc := range tests {
		assert.Equal(t, tc.expected, money.FromPounds(tc.pounds))
	}
}

func TestMax(t *testing.T) {
	t.Parallel()

	tests := []struct {
		a, b     money.Amount
		expected money.Amount
	}{
		{a: money.Zero(), b: money.Zero(), expected: money.Amount(0)},
		{a: money.Zero(), b: 1, expected: money.Amount(1)},
		{a: 1, b: money.Zero(), expected: money.Amount(1)},
		{a: money.Zero(), b: -1, expected: money.Amount(0)},
		{a: -1, b: money.Zero(), expected: money.Amount(0)},
		{a: -3, b: 3, expected: money.Amount(3)},
		{a: 3, b: -3, expected: money.Amount(3)},
		{a: money.MinAmount, b: money.MaxAmount, expected: money.MaxAmount},
	}

	for _, tc := range tests {
		assert.Equal(t, tc.expected, money.Max(tc.a, tc.b))
	}
}

func TestMin(t *testing.T) {
	t.Parallel()

	tests := []struct {
		a, b     money.Amount
		expected money.Amount
	}{
		{a: money.Zero(), b: money.Zero(), expected: money.Amount(0)},
		{a: money.Zero(), b: 1, expected: money.Amount(0)},
		{a: 1, b: money.Zero(), expected: money.Amount(0)},
		{a: money.Zero(), b: -1, expected: money.Amount(-1)},
		{a: -1, b: money.Zero(), expected: money.Amount(-1)},
		{a: -3, b: 3, expected: money.Amount(-3)},
		{a: 3, b: -3, expected: money.Amount(-3)},
		{a: money.MinAmount, b: money.MaxAmount, expected: money.MinAmount},
	}

	for _, tc := range tests {
		assert.Equal(t, tc.expected, money.Min(tc.a, tc.b))
	}
}

func TestAmountString(t *testing.T) {
	t.Parallel()

	// Compile time assertion to check if we implement the fmt.Stringer interface.
	var _ fmt.Stringer = (*money.Amount)(nil)

	tests := []struct {
		amount   money.Amount
		expected string
	}{
		{amount: money.Zero(), expected: "£0.00"},
		{amount: 123 * money.Penny, expected: "£1.23"},
		{amount: 100 * money.Pound, expected: "£100.00"},
		{amount: -100 * money.Pound, expected: "-£100.00"},
		{amount: 1_000 * money.Pound, expected: "£1,000.00"},
		{amount: 10_000 * money.Pound, expected: "£10,000.00"},
		{amount: 100_000 * money.Pound, expected: "£100,000.00"},
		{amount: 1_000_000 * money.Pound, expected: "£1,000,000.00"},
		{amount: -1_123_456_789 * money.Pound, expected: "-£1,123,456,789.00"},
	}

	for _, tc := range tests {
		assert.Equal(t, tc.expected, tc.amount.String())
	}
}

func TestAmountPence(t *testing.T) {
	t.Parallel()

	tests := []struct {
		amount   money.Amount
		expected int64
	}{
		{amount: money.Zero(), expected: 0},
		{amount: -money.Zero(), expected: 0},
		{amount: 1 * money.Penny, expected: 1},
		{amount: -1 * money.Penny, expected: -1},
		{amount: 100 * money.Penny, expected: 100},
		{amount: -100 * money.Penny, expected: -100},
		{amount: 1 * money.Pound, expected: 100},
		{amount: -1 * money.Pound, expected: -100},
		{amount: 100 * money.Pound, expected: 10_000},
		{amount: -100 * money.Pound, expected: -10_000},
		{amount: money.MaxAmount, expected: math.MaxInt64},
		{amount: money.MinAmount, expected: math.MinInt64},
		{amount: 12345 * money.Penny, expected: 12345},
		{amount: -12345 * money.Penny, expected: -12345},
		{amount: (12 * money.Pound) + (155 * money.Penny), expected: 1355},
	}

	for _, tc := range tests {
		assert.Equal(t, tc.expected, tc.amount.Pence())
	}
}

func TestAmountPounds(t *testing.T) {
	t.Parallel()

	tests := []struct {
		amount   money.Amount
		expected float64
	}{
		{amount: money.Zero(), expected: 0.00},
		{amount: -money.Zero(), expected: 0.00},
		{amount: 1 * money.Penny, expected: 0.01},
		{amount: -1 * money.Penny, expected: -0.01},
		{amount: 100 * money.Penny, expected: 1.00},
		{amount: -100 * money.Penny, expected: -1.00},
		{amount: 1 * money.Pound, expected: 1.00},
		{amount: -1 * money.Pound, expected: -1.00},
		{amount: 100 * money.Pound, expected: 100.00},
		{amount: -100 * money.Pound, expected: -100.00},
		{amount: money.MaxAmount, expected: 92233720368547758.07},
		{amount: money.MinAmount, expected: -92233720368547758.08},
		{amount: 12399 * money.Penny, expected: 123.99},
		{amount: -12399 * money.Penny, expected: -123.99},
		{amount: 99 * money.Penny, expected: 0.99},
		{amount: -99 * money.Penny, expected: -0.99},
		{amount: (12 * money.Pound) + (155 * money.Penny), expected: 13.55},
	}

	for _, tc := range tests {
		assert.Equal(t, tc.expected, tc.amount.Pounds())
	}
}

func TestAmountCeil(t *testing.T) {
	t.Parallel()

	tests := []struct {
		amount   money.Amount
		expected money.Amount
	}{
		{amount: money.Zero(), expected: 0},
		{amount: 1 * money.Pound, expected: 100},
		{amount: -1 * money.Pound, expected: -100},
		{amount: money.FromPounds(0.01), expected: 100},
		{amount: money.FromPounds(1.50), expected: 200},
		{amount: money.FromPounds(-0.01), expected: 0},
		{amount: money.FromPounds(-1.50), expected: -100},
	}

	for _, tc := range tests {
		assert.Equal(t, tc.expected, tc.amount.Ceil())
	}
}

func TestAmountFloor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		amount   money.Amount
		expected money.Amount
	}{
		{amount: money.Zero(), expected: 0},
		{amount: 1 * money.Pound, expected: 100},
		{amount: -1 * money.Pound, expected: -100},
		{amount: money.FromPounds(0.01), expected: 0},
		{amount: money.FromPounds(1.50), expected: 100},
		{amount: money.FromPounds(-0.01), expected: -100},
		{amount: money.FromPounds(-1.50), expected: -200},
	}

	for _, tc := range tests {
		assert.Equal(t, tc.expected, tc.amount.Floor())
	}
}

func TestAmountRound(t *testing.T) {
	t.Parallel()

	tests := []struct {
		amount   money.Amount
		expected money.Amount
	}{
		{amount: money.Zero(), expected: 0},
		{amount: 1 * money.Pound, expected: 100},
		{amount: -1 * money.Pound, expected: -100},
		{amount: money.FromPounds(0.01), expected: 0},
		{amount: money.FromPounds(1.50), expected: 200},
		{amount: money.FromPounds(-0.01), expected: 0},
		{amount: money.FromPounds(-1.50), expected: -200},
		{amount: money.FromPounds(1.70), expected: 200},
		{amount: money.FromPounds(1.20), expected: 100},
	}

	for _, tc := range tests {
		assert.Equal(t, tc.expected, tc.amount.Round())
	}
}

func TestAmountTrunc(t *testing.T) {
	t.Parallel()

	tests := []struct {
		amount   money.Amount
		expected money.Amount
	}{
		{amount: money.Zero(), expected: 0},
		{amount: 1 * money.Pound, expected: 100},
		{amount: -1 * money.Pound, expected: -100},
		{amount: money.FromPounds(0.01), expected: 0},
		{amount: money.FromPounds(2.50), expected: 200},
		{amount: money.FromPounds(-0.01), expected: 0},
		{amount: money.FromPounds(-1.50), expected: -100},
		{amount: money.FromPounds(2.70), expected: 200},
		{amount: money.FromPounds(3.20), expected: 300},
	}

	for _, tc := range tests {
		assert.Equal(t, tc.expected, tc.amount.Trunc())
	}
}

func TestAmountMarshalJSON(t *testing.T) {
	t.Parallel()

	val := struct {
		A money.Amount
		B money.Amount
		C money.Amount
		D money.Amount
		E money.Amount
	}{A: money.Zero(), B: money.FromPence(1), C: money.FromPounds(1.50), D: -1, E: -150}

	expectedJSON := `{"A":0, "B":1, "C":150, "D":-1, "E":-150}`

	jsn, err := json.Marshal(&val)
	if err != nil {
		assert.Fail(t, "unable to marshal JSON: %v", err)
	}

	assert.JSONEq(t, expectedJSON, string(jsn))
}

func TestAmountUnmarshalJSON(t *testing.T) {
	t.Parallel()

	type s struct {
		A money.Amount
		B money.Amount
		C money.Amount
		D money.Amount
		E money.Amount
	}

	var val s

	jsn := `{"A":0, "B":1, "C":150, "D":-1, "E":-150}`

	err := json.Unmarshal([]byte(jsn), &val)
	if err != nil {
		assert.Fail(t, "unable to marshal JSON: %v", err)
	}

	expected := s{A: money.Zero(), B: money.FromPence(1), C: money.FromPounds(1.50), D: -1, E: -150}

	assert.Equal(t, expected, val)
}
