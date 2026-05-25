package orderbuilder

import (
	"testing"

	"github.com/jeffWu90/go-clob-client-v2/types"
)

func TestOrderBuilder_BuildOrder(t *testing.T) {
	signer := newSigner(t)
	b := NewOrderBuilder(signer, types.ChainPolygon)

	so, err := b.BuildOrder(
		types.UserOrderV2{TokenID: "1234", Price: 0.5, Size: 10, Side: types.SideBuy},
		types.CreateOrderOptions{TickSize: types.TickSize01},
	)
	if err != nil {
		t.Fatalf("BuildOrder: %v", err)
	}
	if so.Signature == "" {
		t.Errorf("expected non-empty signature")
	}
	if so.Maker == "" {
		t.Errorf("expected maker to be populated (defaults to signer)")
	}
}

func TestOrderBuilder_BuildMarketOrder(t *testing.T) {
	signer := newSigner(t)
	b := NewOrderBuilder(signer, types.ChainPolygon)

	so, err := b.BuildMarketOrder(
		types.UserMarketOrderV2{TokenID: "9999", Price: 0.5, Amount: 50, Side: types.SideBuy, OrderType: types.OrderTypeFAK},
		types.CreateOrderOptions{TickSize: types.TickSize01},
	)
	if err != nil {
		t.Fatalf("BuildMarketOrder: %v", err)
	}
	if so.Signature == "" {
		t.Errorf("expected non-empty signature")
	}
}
