package orderbuilder

import (
	"errors"
	"strconv"

	"github.com/jeffWu90/go-clob-client-v2/types"
)

// ErrNoMatch is returned by the market-price helpers when the orderbook is
// empty, or when there is not enough depth to fully fill a FOK order.
var ErrNoMatch = errors.New("no match")

// CalculateBuyMarketPrice returns the worst price at which a market BUY worth
// `amountToMatch` (USDC) can fully fill. The `positions` slice is the asks
// side of the orderbook, ordered from highest price (index 0) to lowest
// (index len-1). FOK requires the full amount to be matchable, otherwise
// ErrNoMatch is returned. FAK returns the top-of-book price if the depth is
// insufficient.
func CalculateBuyMarketPrice(positions []types.OrderSummary, amountToMatch float64, ot types.OrderType) (float64, error) {
	if len(positions) == 0 {
		return 0, ErrNoMatch
	}
	var sum float64
	for i := len(positions) - 1; i >= 0; i-- {
		p := positions[i]
		size, err := strconv.ParseFloat(p.Size, 64)
		if err != nil {
			return 0, err
		}
		price, err := strconv.ParseFloat(p.Price, 64)
		if err != nil {
			return 0, err
		}
		sum += size * price
		if sum >= amountToMatch {
			return price, nil
		}
	}
	if ot == types.OrderTypeFOK {
		return 0, ErrNoMatch
	}
	// FAK: fall back to top-of-book.
	top, err := strconv.ParseFloat(positions[0].Price, 64)
	if err != nil {
		return 0, err
	}
	return top, nil
}

// CalculateSellMarketPrice returns the worst price at which a market SELL of
// `amountToMatch` shares can fully fill. The `positions` slice is the bids
// side, ordered from lowest price (index 0) to highest (index len-1). FOK
// throws ErrNoMatch on insufficient depth; FAK falls back to top-of-book.
func CalculateSellMarketPrice(positions []types.OrderSummary, amountToMatch float64, ot types.OrderType) (float64, error) {
	if len(positions) == 0 {
		return 0, ErrNoMatch
	}
	var sum float64
	for i := len(positions) - 1; i >= 0; i-- {
		p := positions[i]
		size, err := strconv.ParseFloat(p.Size, 64)
		if err != nil {
			return 0, err
		}
		price, err := strconv.ParseFloat(p.Price, 64)
		if err != nil {
			return 0, err
		}
		sum += size
		if sum >= amountToMatch {
			return price, nil
		}
	}
	if ot == types.OrderTypeFOK {
		return 0, ErrNoMatch
	}
	top, err := strconv.ParseFloat(positions[0].Price, 64)
	if err != nil {
		return 0, err
	}
	return top, nil
}
