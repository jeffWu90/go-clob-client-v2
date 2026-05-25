package orderbuilder

import (
	"fmt"

	clob "github.com/jeffWu90/go-clob-client-v2"
	"github.com/jeffWu90/go-clob-client-v2/signing"
	"github.com/jeffWu90/go-clob-client-v2/types"
)

// CreateOrder is the high-level entry point for limit orders. It resolves the
// maker (defaulting to the signer address when funderAddress is empty), looks
// up the right Exchange V2 contract from the chain config, builds the order
// args, and signs the order.
func CreateOrder(
	signer signing.Signer,
	chainID types.Chain,
	signatureType types.SignatureTypeV2,
	funderAddress string,
	userOrder types.UserOrderV2,
	options types.CreateOrderOptions,
) (*types.SignedOrderV2, error) {
	if signer == nil {
		return nil, fmt.Errorf("signer is required")
	}
	signerAddr := signer.Address()

	maker := funderAddress
	if maker == "" {
		maker = signerAddr
	}

	cc, err := clob.GetContractConfig(chainID)
	if err != nil {
		return nil, err
	}
	rc, ok := RoundingConfig[options.TickSize]
	if !ok {
		return nil, fmt.Errorf("unsupported tick size: %s", options.TickSize)
	}

	orderData := BuildOrderCreationArgs(signerAddr, maker, signatureType, userOrder, rc)

	exchange := cc.ExchangeV2
	if options.NegRisk {
		exchange = cc.NegRiskExchangeV2
	}
	return BuildOrder(signer, exchange, chainID, orderData)
}

// CreateMarketOrder is the high-level entry point for market orders (FOK/FAK).
// See CreateOrder for the funderAddress / contract-selection semantics.
func CreateMarketOrder(
	signer signing.Signer,
	chainID types.Chain,
	signatureType types.SignatureTypeV2,
	funderAddress string,
	userMarketOrder types.UserMarketOrderV2,
	options types.CreateOrderOptions,
) (*types.SignedOrderV2, error) {
	if signer == nil {
		return nil, fmt.Errorf("signer is required")
	}
	signerAddr := signer.Address()

	maker := funderAddress
	if maker == "" {
		maker = signerAddr
	}

	cc, err := clob.GetContractConfig(chainID)
	if err != nil {
		return nil, err
	}
	rc, ok := RoundingConfig[options.TickSize]
	if !ok {
		return nil, fmt.Errorf("unsupported tick size: %s", options.TickSize)
	}

	orderData := BuildMarketOrderCreationArgs(signerAddr, maker, signatureType, userMarketOrder, rc)

	exchange := cc.ExchangeV2
	if options.NegRisk {
		exchange = cc.NegRiskExchangeV2
	}
	return BuildOrder(signer, exchange, chainID, orderData)
}
