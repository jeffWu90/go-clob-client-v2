package clob

import (
	"testing"

	"github.com/jeffWu90/go-clob-client-v2/types"
)

func TestDecimalPlaces(t *testing.T) {
	cases := []struct {
		in   float64
		want int
	}{
		{0, 0},
		{1, 0},
		{1.5, 1},
		{0.1, 1},
		{0.01, 2},
		{0.001, 3},
		{0.0001, 4},
		{123.456, 3},
	}
	for _, c := range cases {
		got := DecimalPlaces(c.in)
		if got != c.want {
			t.Errorf("DecimalPlaces(%v) = %d, want %d", c.in, got, c.want)
		}
	}
}

func TestRoundNormal(t *testing.T) {
	cases := []struct {
		in       float64
		decimals int
		want     float64
	}{
		{1.2345, 2, 1.23},
		{1.235, 2, 1.24}, // half rounds up
		{0.1 + 0.2, 2, 0.3},
		{1.0, 2, 1.0},
		{0.125, 2, 0.13},
		{0.5, 0, 1},
	}
	for _, c := range cases {
		got := RoundNormal(c.in, c.decimals)
		if got != c.want {
			t.Errorf("RoundNormal(%v, %d) = %v, want %v", c.in, c.decimals, got, c.want)
		}
	}
}

func TestRoundDown(t *testing.T) {
	cases := []struct {
		in       float64
		decimals int
		want     float64
	}{
		{1.2399, 2, 1.23},
		{1.230, 2, 1.23},
		{0.1, 2, 0.1},
		{1.999, 0, 1},
	}
	for _, c := range cases {
		got := RoundDown(c.in, c.decimals)
		if got != c.want {
			t.Errorf("RoundDown(%v, %d) = %v, want %v", c.in, c.decimals, got, c.want)
		}
	}
}

func TestRoundUp(t *testing.T) {
	cases := []struct {
		in       float64
		decimals int
		want     float64
	}{
		{1.231, 2, 1.24},
		{1.230, 2, 1.23},
		{0.1, 2, 0.1},
		{1.001, 0, 2},
	}
	for _, c := range cases {
		got := RoundUp(c.in, c.decimals)
		if got != c.want {
			t.Errorf("RoundUp(%v, %d) = %v, want %v", c.in, c.decimals, got, c.want)
		}
	}
}

func TestIsTickSizeSmaller(t *testing.T) {
	if !IsTickSizeSmaller(types.TickSize0001, types.TickSize01) {
		t.Errorf("0.0001 should be smaller than 0.01")
	}
	if IsTickSizeSmaller(types.TickSize01, types.TickSize0001) {
		t.Errorf("0.01 should not be smaller than 0.0001")
	}
	if IsTickSizeSmaller(types.TickSize01, types.TickSize01) {
		t.Errorf("0.01 should not be smaller than itself")
	}
}

func TestPriceValid(t *testing.T) {
	if !PriceValid(0.5, types.TickSize01) {
		t.Errorf("0.5 should be valid at tick 0.01")
	}
	if PriceValid(0.0, types.TickSize01) {
		t.Errorf("0.0 should not be valid at tick 0.01")
	}
	if PriceValid(1.0, types.TickSize01) {
		t.Errorf("1.0 should not be valid at tick 0.01")
	}
	if !PriceValid(0.01, types.TickSize01) {
		t.Errorf("0.01 should be valid at tick 0.01")
	}
	if !PriceValid(0.99, types.TickSize01) {
		t.Errorf("0.99 should be valid at tick 0.01")
	}
}

func TestGenerateOrderBookSummaryHash(t *testing.T) {
	ob := &types.OrderBookSummary{
		Market:  "0xabc",
		AssetID: "1",
		Bids:    []types.OrderSummary{{Price: "0.4", Size: "100"}},
		Asks:    []types.OrderSummary{{Price: "0.5", Size: "50"}},
	}
	h1 := GenerateOrderBookSummaryHash(ob)
	if h1 == "" || len(h1) != 40 {
		t.Errorf("expected 40-char sha1 hex, got %q", h1)
	}
	if ob.Hash != h1 {
		t.Errorf("hash field should be populated; got %q", ob.Hash)
	}
	// Recompute should be deterministic.
	h2 := GenerateOrderBookSummaryHash(ob)
	if h1 != h2 {
		t.Errorf("hash should be deterministic; got %q vs %q", h1, h2)
	}
}

func TestGetContractConfig(t *testing.T) {
	if cfg, err := GetContractConfig(types.ChainPolygon); err != nil {
		t.Errorf("polygon: unexpected error %v", err)
	} else if cfg.ExchangeV2 == "" {
		t.Errorf("polygon: ExchangeV2 should be populated")
	}
	if cfg, err := GetContractConfig(types.ChainAmoy); err != nil {
		t.Errorf("amoy: unexpected error %v", err)
	} else if cfg.ExchangeV2 == "" {
		t.Errorf("amoy: ExchangeV2 should be populated")
	}
	if _, err := GetContractConfig(types.Chain(1)); err == nil {
		t.Errorf("unknown chain: expected error")
	}
}
