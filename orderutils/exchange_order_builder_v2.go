package orderutils

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"

	"github.com/jeffWu90/go-clob-client-v2/signing"
	"github.com/jeffWu90/go-clob-client-v2/types"
)

// ExchangeOrderBuilderV2 builds, hashes and signs V2 orders against a specific
// CTF Exchange deployment (Exchange or NegRiskExchange) on a given chain.
type ExchangeOrderBuilderV2 struct {
	// ContractAddress is the EIP-712 verifyingContract (the Exchange V2 deployment).
	ContractAddress string
	// ChainID is the EIP-712 chain id used in the domain separator.
	ChainID types.Chain
	// Signer signs the EIP-712 digest. Must control the address matching OrderDataV2.Signer.
	Signer signing.Signer
	// SaltFn returns the per-order salt string. Defaults to GenerateOrderSalt;
	// tests may inject a deterministic generator.
	SaltFn func() string
}

// NewExchangeOrderBuilderV2 returns a builder with the default crypto/rand salt source.
func NewExchangeOrderBuilderV2(contractAddress string, chainID types.Chain, signer signing.Signer) *ExchangeOrderBuilderV2 {
	return &ExchangeOrderBuilderV2{
		ContractAddress: contractAddress,
		ChainID:         chainID,
		Signer:          signer,
		SaltFn:          GenerateOrderSalt,
	}
}

// BuildSignedOrder fills in defaults, signs, and returns the SignedOrderV2 ready
// to be POSTed to /order.
func (b *ExchangeOrderBuilderV2) BuildSignedOrder(d types.OrderDataV2) (*types.SignedOrderV2, error) {
	order, err := b.BuildOrder(d)
	if err != nil {
		return nil, err
	}
	td, err := b.BuildOrderTypedData(order)
	if err != nil {
		return nil, err
	}
	sig, err := b.BuildOrderSignature(td)
	if err != nil {
		return nil, err
	}
	return &types.SignedOrderV2{OrderV2: *order, Signature: sig}, nil
}

// BuildOrder constructs an unsigned OrderV2 from the user-facing OrderDataV2,
// filling in salt, timestamp, signer, metadata, builder, and expiration defaults.
func (b *ExchangeOrderBuilderV2) BuildOrder(d types.OrderDataV2) (*types.OrderV2, error) {
	if b.Signer == nil {
		return nil, fmt.Errorf("signer is required")
	}
	if d.Maker == "" {
		return nil, fmt.Errorf("maker is required")
	}

	signer := d.Signer
	if signer == "" {
		signer = d.Maker
	}
	if !strings.EqualFold(signer, b.Signer.Address()) {
		return nil, fmt.Errorf("signer %s does not match wallet %s", signer, b.Signer.Address())
	}

	saltFn := b.SaltFn
	if saltFn == nil {
		saltFn = GenerateOrderSalt
	}

	metadata := d.Metadata
	if metadata == "" {
		metadata = Bytes32Zero
	}
	builder := d.Builder
	if builder == "" {
		builder = Bytes32Zero
	}
	timestamp := d.Timestamp
	if timestamp == "" {
		timestamp = strconv.FormatInt(time.Now().UnixMilli(), 10)
	}
	expiration := d.Expiration
	if expiration == "" {
		expiration = "0"
	}

	return &types.OrderV2{
		Salt:          saltFn(),
		Maker:         d.Maker,
		Signer:        signer,
		TokenID:       d.TokenID,
		MakerAmount:   d.MakerAmount,
		TakerAmount:   d.TakerAmount,
		Side:          d.Side,
		SignatureType: d.SignatureType, // defaults to 0 (EOA)
		Timestamp:     timestamp,
		Metadata:      metadata,
		Builder:       builder,
		Expiration:    expiration,
	}, nil
}

// BuildOrderTypedData converts an OrderV2 into EIP-712 typed data ready for hashing/signing.
// Note: the on-chain Order struct does NOT include expiration; that field is wire-only.
// Side is encoded as uint8 (BUY=0, SELL=1).
func (b *ExchangeOrderBuilderV2) BuildOrderTypedData(order *types.OrderV2) (*apitypes.TypedData, error) {
	salt, err := parseBigInt("salt", order.Salt)
	if err != nil {
		return nil, err
	}
	tokenID, err := parseBigInt("tokenId", order.TokenID)
	if err != nil {
		return nil, err
	}
	makerAmt, err := parseBigInt("makerAmount", order.MakerAmount)
	if err != nil {
		return nil, err
	}
	takerAmt, err := parseBigInt("takerAmount", order.TakerAmount)
	if err != nil {
		return nil, err
	}
	timestamp, err := parseBigInt("timestamp", order.Timestamp)
	if err != nil {
		return nil, err
	}

	var sideU8 uint8
	switch order.Side {
	case types.SideBuy:
		sideU8 = 0
	case types.SideSell:
		sideU8 = 1
	default:
		return nil, fmt.Errorf("invalid side: %q", order.Side)
	}

	return &apitypes.TypedData{
		Types: apitypes.Types{
			"EIP712Domain": eip712DomainDef,
			"Order":        orderV2StructDef,
		},
		PrimaryType: "Order",
		Domain: apitypes.TypedDataDomain{
			Name:              CTFExchangeV2DomainName,
			Version:           CTFExchangeV2DomainVersion,
			ChainId:           (*math.HexOrDecimal256)(big.NewInt(int64(b.ChainID))),
			VerifyingContract: b.ContractAddress,
		},
		Message: apitypes.TypedDataMessage{
			"salt":          salt,
			"maker":         order.Maker,
			"signer":        order.Signer,
			"tokenId":       tokenID,
			"makerAmount":   makerAmt,
			"takerAmount":   takerAmt,
			"side":          new(big.Int).SetUint64(uint64(sideU8)),
			"signatureType": new(big.Int).SetUint64(uint64(order.SignatureType)),
			"timestamp":     timestamp,
			"metadata":      order.Metadata,
			"builder":       order.Builder,
		},
	}, nil
}

// BuildOrderSignature signs the typed-data EIP-712 digest using the configured signer.
func (b *ExchangeOrderBuilderV2) BuildOrderSignature(td *apitypes.TypedData) (string, error) {
	return b.Signer.SignTypedData(td)
}

// BuildOrderHash returns the 0x-prefixed hex EIP-712 digest of the typed data.
func (b *ExchangeOrderBuilderV2) BuildOrderHash(td *apitypes.TypedData) (string, error) {
	hash, _, err := apitypes.TypedDataAndHash(*td)
	if err != nil {
		return "", fmt.Errorf("hash typed data: %w", err)
	}
	return hexutil.Encode(hash), nil
}

func parseBigInt(field, s string) (*big.Int, error) {
	v, ok := new(big.Int).SetString(s, 10)
	if !ok {
		return nil, fmt.Errorf("invalid %s: %q", field, s)
	}
	return v, nil
}
