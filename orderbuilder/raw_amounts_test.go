package orderbuilder

import (
	"testing"

	clob "github.com/jeffWu90/go-clob-client-v2"
	"github.com/jeffWu90/go-clob-client-v2/types"
)

var (
	testSizes        = []float64{0.01, 0.1, 1, 10, 100}
	testPrices01     = []float64{0.1, 0.3, 0.5, 0.7, 0.9, 1}
	testPrices001    = []float64{0.01, 0.1, 0.25, 0.5, 0.75, 0.99}
	testPrices0001   = []float64{0.001, 0.01, 0.1, 0.5, 0.999}
	testPrices00001  = []float64{0.0001, 0.001, 0.01, 0.1, 0.5, 0.9999}
)

// limitInvariantsBuy mirrors the TS getOrderRawAmounts BUY assertions: maker
// decimal cap, taker decimal cap, and the maker/taker ratio meeting the price.
func limitInvariantsBuy(t *testing.T, tick types.TickSize, makerCap, takerCap, ratioPrec int, prices []float64, sizes []float64) {
	t.Helper()
	rc := RoundingConfig[tick]
	for _, size := range sizes {
		for _, price := range prices {
			amts := GetOrderRawAmounts(types.SideBuy, size, price, rc)
			if d := clob.DecimalPlaces(amts.RawMakerAmt); d > makerCap {
				t.Errorf("BUY %s size=%v price=%v: makerAmt %v has %d decimals (cap %d)", tick, size, price, amts.RawMakerAmt, d, makerCap)
			}
			if d := clob.DecimalPlaces(amts.RawTakerAmt); d > takerCap {
				t.Errorf("BUY %s size=%v price=%v: takerAmt %v has %d decimals (cap %d)", tick, size, price, amts.RawTakerAmt, d, takerCap)
			}
			if amts.RawTakerAmt == 0 {
				continue
			}
			ratio := clob.RoundNormal(amts.RawMakerAmt/amts.RawTakerAmt, ratioPrec)
			want := clob.RoundNormal(price, ratioPrec)
			if ratio < want {
				t.Errorf("BUY %s size=%v price=%v: ratio %v < price %v", tick, size, price, ratio, want)
			}
		}
	}
}

func limitInvariantsSell(t *testing.T, tick types.TickSize, makerCap, takerCap, ratioPrec int, prices []float64, sizes []float64) {
	t.Helper()
	rc := RoundingConfig[tick]
	for _, size := range sizes {
		for _, price := range prices {
			amts := GetOrderRawAmounts(types.SideSell, size, price, rc)
			if d := clob.DecimalPlaces(amts.RawMakerAmt); d > makerCap {
				t.Errorf("SELL %s size=%v price=%v: makerAmt %v has %d decimals (cap %d)", tick, size, price, amts.RawMakerAmt, d, makerCap)
			}
			if d := clob.DecimalPlaces(amts.RawTakerAmt); d > takerCap {
				t.Errorf("SELL %s size=%v price=%v: takerAmt %v has %d decimals (cap %d)", tick, size, price, amts.RawTakerAmt, d, takerCap)
			}
			if amts.RawMakerAmt == 0 {
				continue
			}
			ratio := clob.RoundNormal(amts.RawTakerAmt/amts.RawMakerAmt, ratioPrec)
			want := clob.RoundNormal(price, ratioPrec)
			if ratio > want {
				t.Errorf("SELL %s size=%v price=%v: ratio %v > price %v", tick, size, price, ratio, want)
			}
		}
	}
}

func TestGetOrderRawAmounts_BuyDecimalCaps(t *testing.T) {
	limitInvariantsBuy(t, types.TickSize1, 3, 2, 2, testPrices01, testSizes)
	limitInvariantsBuy(t, types.TickSize01, 4, 2, 4, testPrices001, testSizes)
	limitInvariantsBuy(t, types.TickSize001, 5, 2, 6, testPrices0001, testSizes[:4])
	limitInvariantsBuy(t, types.TickSize0001, 6, 2, 8, testPrices00001, testSizes[:3])
}

func TestGetOrderRawAmounts_SellDecimalCaps(t *testing.T) {
	limitInvariantsSell(t, types.TickSize1, 2, 3, 2, testPrices01, testSizes)
	limitInvariantsSell(t, types.TickSize01, 2, 4, 4, testPrices001, testSizes)
	limitInvariantsSell(t, types.TickSize001, 2, 5, 6, testPrices0001, testSizes[:4])
	limitInvariantsSell(t, types.TickSize0001, 2, 6, 8, testPrices00001, testSizes[:3])
}

func marketInvariantsBuy(t *testing.T, tick types.TickSize, makerCap, takerCap, ratioPrec int, prices []float64, sizes []float64) {
	t.Helper()
	rc := RoundingConfig[tick]
	for _, size := range sizes {
		for _, price := range prices {
			amts := GetMarketOrderRawAmounts(types.SideBuy, size, price, rc)
			if d := clob.DecimalPlaces(amts.RawMakerAmt); d > makerCap {
				t.Errorf("mBUY %s size=%v price=%v: makerAmt %v has %d decimals (cap %d)", tick, size, price, amts.RawMakerAmt, d, makerCap)
			}
			if d := clob.DecimalPlaces(amts.RawTakerAmt); d > takerCap {
				t.Errorf("mBUY %s size=%v price=%v: takerAmt %v has %d decimals (cap %d)", tick, size, price, amts.RawTakerAmt, d, takerCap)
			}
			if amts.RawTakerAmt == 0 {
				continue
			}
			ratio := clob.RoundNormal(amts.RawMakerAmt/amts.RawTakerAmt, ratioPrec)
			want := clob.RoundNormal(price, ratioPrec)
			if ratio < want {
				t.Errorf("mBUY %s size=%v price=%v: ratio %v < price %v", tick, size, price, ratio, want)
			}
		}
	}
}

func marketInvariantsSell(t *testing.T, tick types.TickSize, makerCap, takerCap, ratioPrec int, prices []float64, sizes []float64) {
	t.Helper()
	rc := RoundingConfig[tick]
	for _, size := range sizes {
		for _, price := range prices {
			amts := GetMarketOrderRawAmounts(types.SideSell, size, price, rc)
			if d := clob.DecimalPlaces(amts.RawMakerAmt); d > makerCap {
				t.Errorf("mSELL %s size=%v price=%v: makerAmt %v has %d decimals (cap %d)", tick, size, price, amts.RawMakerAmt, d, makerCap)
			}
			if d := clob.DecimalPlaces(amts.RawTakerAmt); d > takerCap {
				t.Errorf("mSELL %s size=%v price=%v: takerAmt %v has %d decimals (cap %d)", tick, size, price, amts.RawTakerAmt, d, takerCap)
			}
			if amts.RawMakerAmt == 0 {
				continue
			}
			ratio := clob.RoundNormal(amts.RawTakerAmt/amts.RawMakerAmt, ratioPrec)
			want := clob.RoundNormal(price, ratioPrec)
			if ratio > want {
				t.Errorf("mSELL %s size=%v price=%v: ratio %v > price %v", tick, size, price, ratio, want)
			}
		}
	}
}

func TestGetMarketOrderRawAmounts_BuyDecimalCaps(t *testing.T) {
	marketInvariantsBuy(t, types.TickSize1, 2, 3, 2, testPrices01, testSizes)
	marketInvariantsBuy(t, types.TickSize01, 2, 4, 4, testPrices001, testSizes)
	marketInvariantsBuy(t, types.TickSize001, 2, 5, 6, testPrices0001, testSizes[:4])
	marketInvariantsBuy(t, types.TickSize0001, 2, 6, 8, testPrices00001, testSizes[:3])
}

func TestGetMarketOrderRawAmounts_SellDecimalCaps(t *testing.T) {
	marketInvariantsSell(t, types.TickSize1, 2, 3, 2, testPrices01, testSizes)
	marketInvariantsSell(t, types.TickSize01, 2, 4, 4, testPrices001, testSizes)
	marketInvariantsSell(t, types.TickSize001, 2, 5, 6, testPrices0001, testSizes[:4])
	marketInvariantsSell(t, types.TickSize0001, 2, 6, 8, testPrices00001, testSizes[:3])
}
