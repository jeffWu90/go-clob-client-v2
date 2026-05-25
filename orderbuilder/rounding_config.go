package orderbuilder

import "github.com/jeffWu90/go-clob-client-v2/types"

// RoundingConfig maps each canonical tick size to the per-field decimal cap
// used by the order-amount calculations. price = decimals kept on the unit
// price; size = decimals kept on the order size; amount = decimals kept on
// the derived USDC amount before being shifted into base units.
var RoundingConfig = map[types.TickSize]types.RoundConfig{
	types.TickSize1: {
		Price:  1,
		Size:   2,
		Amount: 3,
	},
	types.TickSize01: {
		Price:  2,
		Size:   2,
		Amount: 4,
	},
	types.TickSize001: {
		Price:  3,
		Size:   2,
		Amount: 5,
	},
	types.TickSize0001: {
		Price:  4,
		Size:   2,
		Amount: 6,
	},
}
