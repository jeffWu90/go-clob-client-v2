package orderbuilder

import (
	"strconv"
	"time"

	clob "github.com/jeffWu90/go-clob-client-v2"
	"github.com/jeffWu90/go-clob-client-v2/orderutils"
	"github.com/jeffWu90/go-clob-client-v2/types"
)

// BuildOrderCreationArgs turns a user-facing UserOrderV2 into the OrderDataV2
// that the order builder accepts: rounded amounts shifted to USDC base units,
// timestamp set to now, default metadata/builder/expiration applied.
func BuildOrderCreationArgs(
	signer, maker string,
	signatureType types.SignatureTypeV2,
	userOrder types.UserOrderV2,
	rc types.RoundConfig,
) types.OrderDataV2 {
	amts := GetOrderRawAmounts(userOrder.Side, userOrder.Size, userOrder.Price, rc)

	makerAmount := parseUnits(amts.RawMakerAmt, clob.CollateralTokenDecimals)
	takerAmount := parseUnits(amts.RawTakerAmt, clob.CollateralTokenDecimals)

	metadata := userOrder.Metadata
	if metadata == "" {
		metadata = orderutils.Bytes32Zero
	}
	builder := userOrder.BuilderCode
	if builder == "" {
		builder = orderutils.Bytes32Zero
	}
	expiration := "0"
	if userOrder.Expiration != 0 {
		expiration = strconv.FormatInt(userOrder.Expiration, 10)
	}

	return types.OrderDataV2{
		Maker:         maker,
		Signer:        signer,
		TokenID:       userOrder.TokenID,
		MakerAmount:   makerAmount,
		TakerAmount:   takerAmount,
		Side:          amts.Side,
		SignatureType: signatureType,
		Timestamp:     strconv.FormatInt(time.Now().UnixMilli(), 10),
		Metadata:      metadata,
		Builder:       builder,
		Expiration:    expiration,
	}
}

// BuildMarketOrderCreationArgs turns a UserMarketOrderV2 into an OrderDataV2.
// If userOrder.Price is 0, a default price of 1.0 is used so the math still
// works; the actual market price should be resolved upstream via
// CalculateBuyMarketPrice / CalculateSellMarketPrice.
//
// Note: V2 market orders do not POST an expiration field on the wire; the
// builder still treats the OrderDataV2.Expiration field as "0" by default.
func BuildMarketOrderCreationArgs(
	signer, maker string,
	signatureType types.SignatureTypeV2,
	userMarketOrder types.UserMarketOrderV2,
	rc types.RoundConfig,
) types.OrderDataV2 {
	price := userMarketOrder.Price
	if price == 0 {
		price = 1
	}
	amts := GetMarketOrderRawAmounts(userMarketOrder.Side, userMarketOrder.Amount, price, rc)

	makerAmount := parseUnits(amts.RawMakerAmt, clob.CollateralTokenDecimals)
	takerAmount := parseUnits(amts.RawTakerAmt, clob.CollateralTokenDecimals)

	metadata := userMarketOrder.Metadata
	if metadata == "" {
		metadata = orderutils.Bytes32Zero
	}
	builder := userMarketOrder.BuilderCode
	if builder == "" {
		builder = orderutils.Bytes32Zero
	}

	return types.OrderDataV2{
		Maker:         maker,
		Signer:        signer,
		TokenID:       userMarketOrder.TokenID,
		MakerAmount:   makerAmount,
		TakerAmount:   takerAmount,
		Side:          amts.Side,
		SignatureType: signatureType,
		Timestamp:     strconv.FormatInt(time.Now().UnixMilli(), 10),
		Metadata:      metadata,
		Builder:       builder,
		// V2 market orders don't ship an expiration in the typed-data digest;
		// the wire field defaults to "0".
		Expiration: "0",
	}
}
