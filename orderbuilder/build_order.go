package orderbuilder

import (
	"github.com/jeffWu90/go-clob-client-v2/orderutils"
	"github.com/jeffWu90/go-clob-client-v2/signing"
	"github.com/jeffWu90/go-clob-client-v2/types"
)

// BuildOrder signs an OrderDataV2 against a specific V2 Exchange deployment
// and returns the SignedOrderV2 ready to be POSTed.
func BuildOrder(
	signer signing.Signer,
	exchangeAddress string,
	chainID types.Chain,
	orderData types.OrderDataV2,
) (*types.SignedOrderV2, error) {
	b := orderutils.NewExchangeOrderBuilderV2(exchangeAddress, chainID, signer)
	return b.BuildSignedOrder(orderData)
}
