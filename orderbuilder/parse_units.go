package orderbuilder

import (
	"strconv"
	"strings"
)

// parseUnits shifts amount by 10^decimals and returns the result as a decimal
// integer string, matching viem.parseUnits for non-negative amounts whose
// fractional part fits within `decimals` digits.
//
// The implementation deliberately avoids math/big.Float to keep behaviour
// predictable: callers always feed in values that have already been rounded
// to `decimals` (or fewer) fractional digits via roundUp/roundDown.
func parseUnits(amount float64, decimals int) string {
	s := strconv.FormatFloat(amount, 'f', decimals, 64)
	s = strings.Replace(s, ".", "", 1)
	s = strings.TrimLeft(s, "0")
	if s == "" {
		return "0"
	}
	return s
}
