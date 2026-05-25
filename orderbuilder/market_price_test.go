package orderbuilder

import (
	"errors"
	"testing"

	"github.com/jeffWu90/go-clob-client-v2/types"
)

func summaries(in [][2]string) []types.OrderSummary {
	out := make([]types.OrderSummary, len(in))
	for i, p := range in {
		out[i] = types.OrderSummary{Price: p[0], Size: p[1]}
	}
	return out
}

func TestCalculateBuyMarketPrice_FOK(t *testing.T) {
	// empty book
	if _, err := CalculateBuyMarketPrice(nil, 100, types.OrderTypeFOK); !errors.Is(err, ErrNoMatch) {
		t.Errorf("empty book: expected ErrNoMatch, got %v", err)
	}
	// not enough liquidity
	if _, err := CalculateBuyMarketPrice(summaries([][2]string{{"0.5", "100"}, {"0.4", "100"}}), 100, types.OrderTypeFOK); !errors.Is(err, ErrNoMatch) {
		t.Errorf("insufficient liquidity: expected ErrNoMatch, got %v", err)
	}
	// ok cases (mirroring TS test vectors)
	cases := []struct {
		positions [][2]string
		amount    float64
		want      float64
	}{
		{[][2]string{{"0.5", "100"}, {"0.4", "100"}, {"0.3", "100"}}, 100, 0.5},
		{[][2]string{{"0.5", "100"}, {"0.4", "200"}, {"0.3", "100"}}, 100, 0.4},
		{[][2]string{{"0.5", "120"}, {"0.4", "100"}, {"0.3", "100"}}, 100, 0.5},
		{[][2]string{{"0.5", "200"}, {"0.4", "100"}, {"0.3", "100"}}, 100, 0.5},
	}
	for _, c := range cases {
		got, err := CalculateBuyMarketPrice(summaries(c.positions), c.amount, types.OrderTypeFOK)
		if err != nil {
			t.Errorf("positions=%v amount=%v: %v", c.positions, c.amount, err)
		}
		if got != c.want {
			t.Errorf("positions=%v amount=%v: got %v, want %v", c.positions, c.amount, got, c.want)
		}
	}
}

func TestCalculateBuyMarketPrice_FAK(t *testing.T) {
	if _, err := CalculateBuyMarketPrice(nil, 100, types.OrderTypeFAK); !errors.Is(err, ErrNoMatch) {
		t.Errorf("empty book: expected ErrNoMatch, got %v", err)
	}
	// not enough: returns top of book
	p, err := CalculateBuyMarketPrice(summaries([][2]string{{"0.5", "100"}, {"0.4", "100"}}), 100, types.OrderTypeFAK)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if p != 0.5 {
		t.Errorf("not enough FAK: got %v, want 0.5", p)
	}
	p, err = CalculateBuyMarketPrice(summaries([][2]string{{"0.6", "100"}, {"0.55", "100"}, {"0.5", "100"}}), 200, types.OrderTypeFAK)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if p != 0.6 {
		t.Errorf("not enough FAK: got %v, want 0.6", p)
	}
}

func TestCalculateSellMarketPrice_FOK(t *testing.T) {
	if _, err := CalculateSellMarketPrice(nil, 100, types.OrderTypeFOK); !errors.Is(err, ErrNoMatch) {
		t.Errorf("empty book: expected ErrNoMatch, got %v", err)
	}
	if _, err := CalculateSellMarketPrice(summaries([][2]string{{"0.4", "10"}, {"0.5", "10"}}), 100, types.OrderTypeFOK); !errors.Is(err, ErrNoMatch) {
		t.Errorf("insufficient liquidity: expected ErrNoMatch, got %v", err)
	}
	cases := []struct {
		positions [][2]string
		amount    float64
		want      float64
	}{
		{[][2]string{{"0.3", "100"}, {"0.4", "100"}, {"0.5", "100"}}, 100, 0.5},
		{[][2]string{{"0.3", "100"}, {"0.4", "100"}, {"0.5", "100"}}, 300, 0.3},
		{[][2]string{{"0.3", "100"}, {"0.4", "200"}, {"0.5", "100"}}, 300, 0.4},
		{[][2]string{{"0.3", "334"}, {"0.4", "100"}, {"0.5", "1000"}}, 600, 0.5},
	}
	for _, c := range cases {
		got, err := CalculateSellMarketPrice(summaries(c.positions), c.amount, types.OrderTypeFOK)
		if err != nil {
			t.Errorf("positions=%v amount=%v: %v", c.positions, c.amount, err)
		}
		if got != c.want {
			t.Errorf("positions=%v amount=%v: got %v, want %v", c.positions, c.amount, got, c.want)
		}
	}
}

func TestCalculateSellMarketPrice_FAK(t *testing.T) {
	if _, err := CalculateSellMarketPrice(nil, 100, types.OrderTypeFAK); !errors.Is(err, ErrNoMatch) {
		t.Errorf("empty book: expected ErrNoMatch, got %v", err)
	}
	p, err := CalculateSellMarketPrice(summaries([][2]string{{"0.4", "10"}, {"0.5", "10"}}), 100, types.OrderTypeFAK)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if p != 0.4 {
		t.Errorf("not enough FAK: got %v, want 0.4", p)
	}
}
