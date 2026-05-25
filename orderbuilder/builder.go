package orderbuilder

import (
	"github.com/jeffWu90/go-clob-client-v2/signing"
	"github.com/jeffWu90/go-clob-client-v2/types"
)

// OrderBuilder is the top-level convenience wrapper around CreateOrder /
// CreateMarketOrder. It binds a signer + chain + signature type + funder so
// callers don't need to thread them through every call.
type OrderBuilder struct {
	Signer        signing.Signer
	ChainID       types.Chain
	SignatureType types.SignatureTypeV2
	FunderAddress string // optional; defaults to Signer.Address() when empty
}

// NewOrderBuilder returns an OrderBuilder with SignatureTypeV2EOA as the default.
func NewOrderBuilder(signer signing.Signer, chainID types.Chain) *OrderBuilder {
	return &OrderBuilder{
		Signer:        signer,
		ChainID:       chainID,
		SignatureType: types.SignatureTypeV2EOA,
	}
}

// BuildOrder signs a limit order.
func (b *OrderBuilder) BuildOrder(userOrder types.UserOrderV2, options types.CreateOrderOptions) (*types.SignedOrderV2, error) {
	return CreateOrder(b.Signer, b.ChainID, b.SignatureType, b.FunderAddress, userOrder, options)
}

// BuildMarketOrder signs a market order (FOK / FAK).
func (b *OrderBuilder) BuildMarketOrder(userMarketOrder types.UserMarketOrderV2, options types.CreateOrderOptions) (*types.SignedOrderV2, error) {
	return CreateMarketOrder(b.Signer, b.ChainID, b.SignatureType, b.FunderAddress, userMarketOrder, options)
}
