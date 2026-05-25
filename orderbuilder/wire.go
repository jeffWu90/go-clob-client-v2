package orderbuilder

import (
	"strconv"

	"github.com/jeffWu90/go-clob-client-v2/types"
)

// ZeroAddress is the canonical 20-byte zero address. V2 orders preserve a
// `taker` field on the wire (set to zero) for backward compatibility, even
// though it is not part of the EIP-712 digest.
const ZeroAddress = "0x0000000000000000000000000000000000000000"

// OrderToWireV2 converts a SignedOrderV2 into the JSON body POSTed to /order
// or /orders. The owner field is the L2 API key (POLY_API_KEY value).
func OrderToWireV2(so *types.SignedOrderV2, owner string, orderType types.OrderType, postOnly, deferExec bool) *types.NewOrderV2Body {
	salt, _ := strconv.Atoi(so.Salt) // V2 salt fits in int64 by construction; ignore parse failure here.
	return &types.NewOrderV2Body{
		DeferExec: deferExec,
		PostOnly:  postOnly,
		Order: types.OrderV2Wire{
			Salt:          salt,
			Maker:         so.Maker,
			Signer:        so.Signer,
			Taker:         ZeroAddress,
			TokenID:       so.TokenID,
			MakerAmount:   so.MakerAmount,
			TakerAmount:   so.TakerAmount,
			Side:          so.Side,
			SignatureType: so.SignatureType,
			Timestamp:     so.Timestamp,
			Expiration:    so.Expiration,
			Metadata:      so.Metadata,
			Builder:       so.Builder,
			Signature:     so.Signature,
		},
		Owner:     owner,
		OrderType: orderType,
	}
}
