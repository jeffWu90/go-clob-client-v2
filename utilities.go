package clob

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"math"
	"strconv"
	"strings"

	"github.com/jeffWu90/go-clob-client-v2/types"
)

// numberEpsilon mirrors JavaScript's Number.EPSILON (2^-52). It is added before
// rounding to nudge tie cases like 0.1+0.2 back to the expected decimal value.
const numberEpsilon = 2.220446049250313e-16

// RoundNormal rounds n to the given number of decimal places, matching
// JavaScript's Math.round(n + EPSILON) semantics (half rounds towards +infinity).
func RoundNormal(n float64, decimals int) float64 {
	if DecimalPlaces(n) <= decimals {
		return n
	}
	factor := math.Pow(10, float64(decimals))
	// JS Math.round rounds half towards +infinity; math.Floor(x+0.5) reproduces that
	// for the non-negative numbers we care about (prices, sizes, amounts).
	return math.Floor((n+numberEpsilon)*factor+0.5) / factor
}

// RoundDown rounds n down (towards -infinity) to the given number of decimal places.
func RoundDown(n float64, decimals int) float64 {
	if DecimalPlaces(n) <= decimals {
		return n
	}
	factor := math.Pow(10, float64(decimals))
	return math.Floor(n*factor) / factor
}

// RoundUp rounds n up (towards +infinity) to the given number of decimal places.
func RoundUp(n float64, decimals int) float64 {
	if DecimalPlaces(n) <= decimals {
		return n
	}
	factor := math.Pow(10, float64(decimals))
	return math.Ceil(n*factor) / factor
}

// DecimalPlaces returns the number of decimal places in n. Integers return 0.
func DecimalPlaces(n float64) int {
	if n == math.Trunc(n) && !math.IsInf(n, 0) {
		return 0
	}
	s := strconv.FormatFloat(n, 'f', -1, 64)
	dot := strings.IndexByte(s, '.')
	if dot < 0 {
		return 0
	}
	return len(s) - dot - 1
}

// GenerateOrderBookSummaryHash computes the SHA-1 hash of the orderbook JSON
// (with the hash field cleared) and writes the hex digest back into ob.Hash.
func GenerateOrderBookSummaryHash(ob *types.OrderBookSummary) string {
	ob.Hash = ""
	b, _ := json.Marshal(ob)
	sum := sha1.Sum(b)
	ob.Hash = hex.EncodeToString(sum[:])
	return ob.Hash
}

// IsTickSizeSmaller reports whether tick a is strictly smaller than tick b.
func IsTickSizeSmaller(a, b types.TickSize) bool {
	pa, _ := strconv.ParseFloat(string(a), 64)
	pb, _ := strconv.ParseFloat(string(b), 64)
	return pa < pb
}

// PriceValid checks whether a price lies within [tickSize, 1 - tickSize].
func PriceValid(price float64, tick types.TickSize) bool {
	t, err := strconv.ParseFloat(string(tick), 64)
	if err != nil {
		return false
	}
	return price >= t && price <= 1-t
}
