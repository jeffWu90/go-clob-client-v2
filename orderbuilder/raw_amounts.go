package orderbuilder

import (
	clob "github.com/jeffWu90/go-clob-client-v2"
	"github.com/jeffWu90/go-clob-client-v2/types"
)

// RawAmounts holds the maker/taker amounts (as floats) produced by the
// pre-encoding rounding step. They are still in human units; the calling code
// shifts them to base units via parseUnits.
type RawAmounts struct {
	Side        types.Side
	RawMakerAmt float64
	RawTakerAmt float64
}

// GetOrderRawAmounts computes the raw maker/taker amounts for a LIMIT order.
//
// BUY: taker amount is the size (in shares), maker amount is size * price (USDC).
// SELL: maker amount is the size (in shares), taker amount is size * price (USDC).
//
// The derived amount is rounded down to roundConfig.amount fractional digits.
// To absorb float-multiplication noise, the rounding goes via an intermediate
// roundUp at +4 digits before the final truncation (matches the TS implementation).
func GetOrderRawAmounts(side types.Side, size, price float64, rc types.RoundConfig) RawAmounts {
	rawPrice := clob.RoundNormal(price, rc.Price)

	if side == types.SideBuy {
		rawTaker := clob.RoundDown(size, rc.Size)
		rawMaker := rawTaker * rawPrice
		if clob.DecimalPlaces(rawMaker) > rc.Amount {
			rawMaker = clob.RoundUp(rawMaker, rc.Amount+4)
			if clob.DecimalPlaces(rawMaker) > rc.Amount {
				rawMaker = clob.RoundDown(rawMaker, rc.Amount)
			}
		}
		return RawAmounts{Side: types.SideBuy, RawMakerAmt: rawMaker, RawTakerAmt: rawTaker}
	}

	// SELL
	rawMaker := clob.RoundDown(size, rc.Size)
	rawTaker := rawMaker * rawPrice
	if clob.DecimalPlaces(rawTaker) > rc.Amount {
		rawTaker = clob.RoundUp(rawTaker, rc.Amount+4)
		if clob.DecimalPlaces(rawTaker) > rc.Amount {
			rawTaker = clob.RoundDown(rawTaker, rc.Amount)
		}
	}
	return RawAmounts{Side: types.SideSell, RawMakerAmt: rawMaker, RawTakerAmt: rawTaker}
}

// GetMarketOrderRawAmounts computes raw maker/taker amounts for a MARKET order.
//
// BUY market: amount is USDC to spend; maker = amount (USDC), taker = amount / price (shares).
// SELL market: amount is shares to sell; maker = amount (shares), taker = amount * price (USDC).
//
// Note the price is rounded DOWN here (not normal), which keeps market buys/sells
// from over-committing in the presence of float noise.
func GetMarketOrderRawAmounts(side types.Side, amount, price float64, rc types.RoundConfig) RawAmounts {
	rawPrice := clob.RoundDown(price, rc.Price)

	if side == types.SideBuy {
		rawMaker := clob.RoundDown(amount, rc.Size)
		rawTaker := rawMaker / rawPrice
		if clob.DecimalPlaces(rawTaker) > rc.Amount {
			rawTaker = clob.RoundUp(rawTaker, rc.Amount+4)
			if clob.DecimalPlaces(rawTaker) > rc.Amount {
				rawTaker = clob.RoundDown(rawTaker, rc.Amount)
			}
		}
		return RawAmounts{Side: types.SideBuy, RawMakerAmt: rawMaker, RawTakerAmt: rawTaker}
	}

	// SELL market
	rawMaker := clob.RoundDown(amount, rc.Size)
	rawTaker := rawMaker * rawPrice
	if clob.DecimalPlaces(rawTaker) > rc.Amount {
		rawTaker = clob.RoundUp(rawTaker, rc.Amount+4)
		if clob.DecimalPlaces(rawTaker) > rc.Amount {
			rawTaker = clob.RoundDown(rawTaker, rc.Amount)
		}
	}
	return RawAmounts{Side: types.SideSell, RawMakerAmt: rawMaker, RawTakerAmt: rawTaker}
}
