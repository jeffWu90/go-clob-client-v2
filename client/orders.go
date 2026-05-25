package client

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	clob "github.com/jeffWu90/go-clob-client-v2"
	"github.com/jeffWu90/go-clob-client-v2/httphelpers"
	"github.com/jeffWu90/go-clob-client-v2/orderbuilder"
	"github.com/jeffWu90/go-clob-client-v2/types"
)

// CreateOrder signs a limit order. If options.TickSize is empty the minimum
// tick size for the token is fetched. If options.NegRisk is left unset the
// market's negRisk flag is fetched. The order's BuilderCode is filled from
// the client's BuilderConfig when blank.
func (c *ClobClient) CreateOrder(ctx context.Context, userOrder types.UserOrderV2, options types.CreateOrderOptions) (*types.SignedOrderV2, error) {
	if err := c.canL1Auth(); err != nil {
		return nil, err
	}
	if c.builderConfig != nil && userOrder.BuilderCode == "" {
		userOrder.BuilderCode = c.builderConfig.BuilderCode
	}

	tick, err := c.resolveTickSize(ctx, userOrder.TokenID, options.TickSize)
	if err != nil {
		return nil, err
	}
	if !clob.PriceValid(userOrder.Price, tick) {
		t, _ := strconv.ParseFloat(string(tick), 64)
		return nil, fmt.Errorf("invalid price (%v), min: %v - max: %v", userOrder.Price, t, 1-t)
	}

	negRisk := options.NegRisk
	if !negRisk {
		nr, err := c.GetNegRisk(ctx, userOrder.TokenID)
		if err != nil {
			return nil, err
		}
		negRisk = nr
	}

	return c.orderBuilder.BuildOrder(userOrder, types.CreateOrderOptions{TickSize: tick, NegRisk: negRisk})
}

// CreateMarketOrder signs a market order. If userMarketOrder.Price is 0, the
// price is resolved from the orderbook via CalculateMarketPrice. When the
// caller provides a UserUSDCBalance on a BUY, the order amount is adjusted to
// stay within the balance after platform + builder fees.
func (c *ClobClient) CreateMarketOrder(ctx context.Context, userMarketOrder types.UserMarketOrderV2, options types.CreateOrderOptions) (*types.SignedOrderV2, error) {
	if err := c.canL1Auth(); err != nil {
		return nil, err
	}
	if err := c.ensureMarketInfoCached(ctx, userMarketOrder.TokenID); err != nil {
		return nil, err
	}

	tick, err := c.resolveTickSize(ctx, userMarketOrder.TokenID, options.TickSize)
	if err != nil {
		return nil, err
	}
	if userMarketOrder.Price == 0 {
		ot := userMarketOrder.OrderType
		if ot == "" {
			ot = types.OrderTypeFOK
		}
		price, err := c.CalculateMarketPrice(ctx, userMarketOrder.TokenID, userMarketOrder.Side, userMarketOrder.Amount, ot)
		if err != nil {
			return nil, err
		}
		userMarketOrder.Price = price
	}
	if !clob.PriceValid(userMarketOrder.Price, tick) {
		t, _ := strconv.ParseFloat(string(tick), 64)
		return nil, fmt.Errorf("invalid price (%v), min: %v - max: %v", userMarketOrder.Price, t, 1-t)
	}

	if c.builderConfig != nil && userMarketOrder.BuilderCode == "" {
		userMarketOrder.BuilderCode = c.builderConfig.BuilderCode
	}
	if err := c.ensureBuilderFeeRateCached(ctx, userMarketOrder.BuilderCode); err != nil {
		return nil, err
	}

	if userMarketOrder.Side == types.SideBuy && userMarketOrder.UserUSDCBalance > 0 {
		var takerFee float64
		if c.isBuilderOrder(userMarketOrder.BuilderCode) {
			c.mu.RLock()
			if rates, ok := c.builderFeeRates[userMarketOrder.BuilderCode]; ok {
				takerFee = rates.Taker
			}
			c.mu.RUnlock()
		}
		c.mu.RLock()
		fi := c.feeInfos[userMarketOrder.TokenID]
		c.mu.RUnlock()
		userMarketOrder.Amount = adjustBuyAmountForFees(userMarketOrder.Amount, userMarketOrder.Price, userMarketOrder.UserUSDCBalance, fi.Rate, fi.Exponent, takerFee)
	}

	negRisk := options.NegRisk
	if !negRisk {
		nr, err := c.GetNegRisk(ctx, userMarketOrder.TokenID)
		if err != nil {
			return nil, err
		}
		negRisk = nr
	}
	return c.orderBuilder.BuildMarketOrder(userMarketOrder, types.CreateOrderOptions{TickSize: tick, NegRisk: negRisk})
}

// CreateAndPostOrder signs and immediately POSTs a limit order. orderType
// defaults to GTC; the caller picks FOK/FAK/GTD as needed.
func (c *ClobClient) CreateAndPostOrder(ctx context.Context, userOrder types.UserOrderV2, options types.CreateOrderOptions, orderType types.OrderType, postOnly, deferExec bool) (*types.OrderResponse, error) {
	if orderType == "" {
		orderType = types.OrderTypeGTC
	}
	signed, err := c.CreateOrder(ctx, userOrder, options)
	if err != nil {
		return nil, err
	}
	return c.PostOrder(ctx, signed, orderType, postOnly, deferExec)
}

// CreateAndPostMarketOrder signs and immediately POSTs a market order. Defaults to FOK.
func (c *ClobClient) CreateAndPostMarketOrder(ctx context.Context, userMarketOrder types.UserMarketOrderV2, options types.CreateOrderOptions, orderType types.OrderType, deferExec bool) (*types.OrderResponse, error) {
	if orderType == "" {
		orderType = types.OrderTypeFOK
	}
	signed, err := c.CreateMarketOrder(ctx, userMarketOrder, options)
	if err != nil {
		return nil, err
	}
	return c.PostOrder(ctx, signed, orderType, false, deferExec)
}

// PostOrder POSTs a single signed order to /order. postOnly is rejected for
// FOK / FAK (those order types must be allowed to take liquidity).
func (c *ClobClient) PostOrder(ctx context.Context, order *types.SignedOrderV2, orderType types.OrderType, postOnly, deferExec bool) (*types.OrderResponse, error) {
	if err := c.canL2Auth(); err != nil {
		return nil, err
	}
	if postOnly && (orderType == types.OrderTypeFOK || orderType == types.OrderTypeFAK) {
		return nil, fmt.Errorf("postOnly is not supported for FOK/FAK orders")
	}

	owner := ""
	if c.creds != nil {
		owner = c.creds.Key
	}
	payload := orderbuilder.OrderToWireV2(order, owner, orderType, postOnly, deferExec)
	body, _ := json.Marshal(payload)

	hdrs, err := c.buildL2Headers(ctx, http.MethodPost, clob.EndpointPostOrder, string(body))
	if err != nil {
		return nil, err
	}
	var out types.OrderResponse
	err = c.http.Post(ctx, c.url(clob.EndpointPostOrder), httphelpers.RequestOptions{Headers: hdrs, Body: payload}, &out)
	return &out, err
}

// PostOrdersArg pairs an order with its OrderType for batch submission.
type PostOrdersArg struct {
	Order     *types.SignedOrderV2
	OrderType types.OrderType
}

// PostOrders submits multiple signed orders in a single request to /orders.
func (c *ClobClient) PostOrders(ctx context.Context, args []PostOrdersArg, postOnly, deferExec bool) (any, error) {
	if err := c.canL2Auth(); err != nil {
		return nil, err
	}
	if postOnly {
		for _, a := range args {
			if a.OrderType == types.OrderTypeFOK || a.OrderType == types.OrderTypeFAK {
				return nil, fmt.Errorf("postOnly is not supported for FOK/FAK orders")
			}
		}
	}
	owner := ""
	if c.creds != nil {
		owner = c.creds.Key
	}
	payloads := make([]*types.NewOrderV2Body, len(args))
	for i, a := range args {
		payloads[i] = orderbuilder.OrderToWireV2(a.Order, owner, a.OrderType, postOnly, deferExec)
	}
	body, _ := json.Marshal(payloads)

	hdrs, err := c.buildL2Headers(ctx, http.MethodPost, clob.EndpointPostOrders, string(body))
	if err != nil {
		return nil, err
	}
	var out any
	err = c.http.Post(ctx, c.url(clob.EndpointPostOrders), httphelpers.RequestOptions{Headers: hdrs, Body: payloads}, &out)
	return out, err
}

// CancelOrder cancels a single open order by id.
func (c *ClobClient) CancelOrder(ctx context.Context, payload types.OrderPayload) (any, error) {
	body, _ := json.Marshal(payload)
	hdrs, err := c.buildL2Headers(ctx, http.MethodDelete, clob.EndpointCancelOrder, string(body))
	if err != nil {
		return nil, err
	}
	var out any
	err = c.http.Delete(ctx, c.url(clob.EndpointCancelOrder), httphelpers.RequestOptions{Headers: hdrs, Body: payload}, &out)
	return out, err
}

// CancelOrders cancels multiple open orders by hash.
func (c *ClobClient) CancelOrders(ctx context.Context, orderHashes []string) (any, error) {
	body, _ := json.Marshal(orderHashes)
	hdrs, err := c.buildL2Headers(ctx, http.MethodDelete, clob.EndpointCancelOrders, string(body))
	if err != nil {
		return nil, err
	}
	var out any
	err = c.http.Delete(ctx, c.url(clob.EndpointCancelOrders), httphelpers.RequestOptions{Headers: hdrs, Body: orderHashes}, &out)
	return out, err
}

// CancelAll cancels every open order on the account.
func (c *ClobClient) CancelAll(ctx context.Context) (any, error) {
	hdrs, err := c.buildL2Headers(ctx, http.MethodDelete, clob.EndpointCancelAll, "")
	if err != nil {
		return nil, err
	}
	var out any
	err = c.http.Delete(ctx, c.url(clob.EndpointCancelAll), httphelpers.RequestOptions{Headers: hdrs}, &out)
	return out, err
}

// CancelMarketOrders cancels all orders in a market.
func (c *ClobClient) CancelMarketOrders(ctx context.Context, payload types.OrderMarketCancelParams) (any, error) {
	body, _ := json.Marshal(payload)
	hdrs, err := c.buildL2Headers(ctx, http.MethodDelete, clob.EndpointCancelMarketOrders, string(body))
	if err != nil {
		return nil, err
	}
	var out any
	err = c.http.Delete(ctx, c.url(clob.EndpointCancelMarketOrders), httphelpers.RequestOptions{Headers: hdrs, Body: payload}, &out)
	return out, err
}

// GetOrder fetches a single order by id.
func (c *ClobClient) GetOrder(ctx context.Context, orderID string) (*types.OpenOrder, error) {
	endpoint := clob.EndpointGetOrder + orderID
	hdrs, err := c.buildL2Headers(ctx, http.MethodGet, endpoint, "")
	if err != nil {
		return nil, err
	}
	var out types.OpenOrder
	err = c.http.Get(ctx, c.url(endpoint), httphelpers.RequestOptions{Headers: hdrs}, &out)
	return &out, err
}

// GetOpenOrders pages through /data/orders until the EndCursor.
func (c *ClobClient) GetOpenOrders(ctx context.Context, params *types.OpenOrderParams, onlyFirstPage bool, nextCursor string) ([]types.OpenOrder, error) {
	hdrs, err := c.buildL2Headers(ctx, http.MethodGet, clob.EndpointGetOpenOrders, "")
	if err != nil {
		return nil, err
	}
	if nextCursor == "" {
		nextCursor = clob.InitialCursor
	}
	var results []types.OpenOrder
	for nextCursor != clob.EndCursor && (nextCursor == clob.InitialCursor || !onlyFirstPage) {
		v := openOrderParamsToValues(params)
		v.Set("next_cursor", nextCursor)
		var page struct {
			NextCursor string            `json:"next_cursor"`
			Data       []types.OpenOrder `json:"data"`
		}
		if err := c.http.Get(ctx, c.url(clob.EndpointGetOpenOrders), httphelpers.RequestOptions{Headers: hdrs, Params: v}, &page); err != nil {
			return results, err
		}
		results = append(results, page.Data...)
		nextCursor = page.NextCursor
	}
	return results, nil
}

// GetPreMigrationOrders pages through /data/pre-migration-orders (V1→V2 cutover artifacts).
func (c *ClobClient) GetPreMigrationOrders(ctx context.Context, onlyFirstPage bool, nextCursor string) ([]types.OpenOrder, error) {
	hdrs, err := c.buildL2Headers(ctx, http.MethodGet, clob.EndpointGetPreMigrationOrders, "")
	if err != nil {
		return nil, err
	}
	if nextCursor == "" {
		nextCursor = clob.InitialCursor
	}
	var results []types.OpenOrder
	for nextCursor != clob.EndCursor && (nextCursor == clob.InitialCursor || !onlyFirstPage) {
		v := url.Values{"next_cursor": []string{nextCursor}}
		var page struct {
			NextCursor string            `json:"next_cursor"`
			Data       []types.OpenOrder `json:"data"`
		}
		if err := c.http.Get(ctx, c.url(clob.EndpointGetPreMigrationOrders), httphelpers.RequestOptions{Headers: hdrs, Params: v}, &page); err != nil {
			return results, err
		}
		results = append(results, page.Data...)
		nextCursor = page.NextCursor
	}
	return results, nil
}

// GetTrades pages through /data/trades.
func (c *ClobClient) GetTrades(ctx context.Context, params *types.TradeParams, onlyFirstPage bool, nextCursor string) ([]types.Trade, error) {
	hdrs, err := c.buildL2Headers(ctx, http.MethodGet, clob.EndpointGetTrades, "")
	if err != nil {
		return nil, err
	}
	if nextCursor == "" {
		nextCursor = clob.InitialCursor
	}
	var results []types.Trade
	for nextCursor != clob.EndCursor && (nextCursor == clob.InitialCursor || !onlyFirstPage) {
		v := tradeParamsToValues(params)
		v.Set("next_cursor", nextCursor)
		var page struct {
			NextCursor string        `json:"next_cursor"`
			Data       []types.Trade `json:"data"`
		}
		if err := c.http.Get(ctx, c.url(clob.EndpointGetTrades), httphelpers.RequestOptions{Headers: hdrs, Params: v}, &page); err != nil {
			return results, err
		}
		results = append(results, page.Data...)
		nextCursor = page.NextCursor
	}
	return results, nil
}

// GetTradesPaginated returns a single page of trades (no auto-paging).
func (c *ClobClient) GetTradesPaginated(ctx context.Context, params *types.TradeParams, nextCursor string) (*types.TradesPaginatedResponse, error) {
	hdrs, err := c.buildL2Headers(ctx, http.MethodGet, clob.EndpointGetTrades, "")
	if err != nil {
		return nil, err
	}
	if nextCursor == "" {
		nextCursor = clob.InitialCursor
	}
	v := tradeParamsToValues(params)
	v.Set("next_cursor", nextCursor)

	var page struct {
		NextCursor string        `json:"next_cursor"`
		Limit      int           `json:"limit"`
		Count      int           `json:"count"`
		Data       []types.Trade `json:"data"`
	}
	if err := c.http.Get(ctx, c.url(clob.EndpointGetTrades), httphelpers.RequestOptions{Headers: hdrs, Params: v}, &page); err != nil {
		return nil, err
	}
	return &types.TradesPaginatedResponse{
		Trades:     page.Data,
		NextCursor: page.NextCursor,
		Limit:      page.Limit,
		Count:      page.Count,
	}, nil
}

// IsOrderScoring returns whether a specific order is in scoring range.
func (c *ClobClient) IsOrderScoring(ctx context.Context, params *types.OrderScoringParams) (*types.OrderScoring, error) {
	hdrs, err := c.buildL2Headers(ctx, http.MethodGet, clob.EndpointIsOrderScoring, "")
	if err != nil {
		return nil, err
	}
	v := url.Values{}
	if params != nil && params.OrderID != "" {
		v.Set("order_id", params.OrderID)
	}
	var out types.OrderScoring
	err = c.http.Get(ctx, c.url(clob.EndpointIsOrderScoring), httphelpers.RequestOptions{Headers: hdrs, Params: v}, &out)
	return &out, err
}

// AreOrdersScoring batch-checks scoring for a list of order ids.
func (c *ClobClient) AreOrdersScoring(ctx context.Context, params *types.OrdersScoringParams) (types.OrdersScoring, error) {
	body, _ := json.Marshal(params.OrderIDs)
	hdrs, err := c.buildL2Headers(ctx, http.MethodPost, clob.EndpointAreOrdersScoring, string(body))
	if err != nil {
		return nil, err
	}
	var out types.OrdersScoring
	err = c.http.Post(ctx, c.url(clob.EndpointAreOrdersScoring), httphelpers.RequestOptions{Headers: hdrs, Body: string(body)}, &out)
	return out, err
}

// CalculateMarketPrice resolves the worst-fill price from the live orderbook
// (asks for BUY, bids for SELL). orderType controls FOK / FAK semantics.
func (c *ClobClient) CalculateMarketPrice(ctx context.Context, tokenID string, side types.Side, amount float64, orderType types.OrderType) (float64, error) {
	book, err := c.GetOrderBook(ctx, tokenID)
	if err != nil {
		return 0, err
	}
	if side == types.SideBuy {
		if len(book.Asks) == 0 {
			return 0, orderbuilder.ErrNoMatch
		}
		return orderbuilder.CalculateBuyMarketPrice(book.Asks, amount, orderType)
	}
	if len(book.Bids) == 0 {
		return 0, orderbuilder.ErrNoMatch
	}
	return orderbuilder.CalculateSellMarketPrice(book.Bids, amount, orderType)
}

// resolveTickSize fills in a default tick when the caller didn't pick one,
// and validates that a user-supplied tick isn't finer than the market's minimum.
func (c *ClobClient) resolveTickSize(ctx context.Context, tokenID string, tick types.TickSize) (types.TickSize, error) {
	minTick, err := c.GetTickSize(ctx, tokenID)
	if err != nil {
		return "", err
	}
	if tick == "" {
		return minTick, nil
	}
	if clob.IsTickSizeSmaller(tick, minTick) {
		return "", fmt.Errorf("invalid tick size (%s), minimum for the market is %s", tick, minTick)
	}
	return tick, nil
}

// ensureBuilderFeeRateCached fetches /fees/builder-fees/{code} when the rate
// for the builder isn't cached yet. A zero / empty code skips the lookup.
func (c *ClobClient) ensureBuilderFeeRateCached(ctx context.Context, builderCode string) error {
	if !c.isBuilderOrder(builderCode) {
		return nil
	}
	c.mu.RLock()
	_, ok := c.builderFeeRates[builderCode]
	c.mu.RUnlock()
	if ok {
		return nil
	}

	var raw struct {
		BuilderMakerFeeRateBps json.Number `json:"builder_maker_fee_rate_bps"`
		BuilderTakerFeeRateBps json.Number `json:"builder_taker_fee_rate_bps"`
	}
	if err := c.http.Get(ctx, c.url(clob.EndpointGetBuilderFees+builderCode), httphelpers.RequestOptions{}, &raw); err != nil {
		return err
	}
	maker, _ := raw.BuilderMakerFeeRateBps.Float64()
	taker, _ := raw.BuilderTakerFeeRateBps.Float64()

	c.mu.Lock()
	c.builderFeeRates[builderCode] = struct {
		Maker float64 `json:"maker"`
		Taker float64 `json:"taker"`
	}{Maker: maker / clob.BuilderFeesBps, Taker: taker / clob.BuilderFeesBps}
	c.mu.Unlock()
	return nil
}

func (c *ClobClient) isBuilderOrder(builderCode string) bool {
	return builderCode != "" && builderCode != clob.Bytes32Zero && !strings.EqualFold(builderCode, clob.Bytes32Zero)
}

// adjustBuyAmountForFees mirrors the TS helper: if the user's USDC balance is
// less than the order + platform + builder fee, shrink the order so the total
// fits. Returns the (possibly reduced) order amount in USDC.
func adjustBuyAmountForFees(amount, price, userUSDCBalance, feeRate, feeExponent, builderTakerFeeRate float64) float64 {
	platformFeeRate := feeRate * math.Pow(price*(1-price), feeExponent)
	platformFee := (amount / price) * platformFeeRate
	totalCost := amount + platformFee + amount*builderTakerFeeRate
	if userUSDCBalance <= totalCost {
		return userUSDCBalance / (1 + platformFeeRate/price + builderTakerFeeRate)
	}
	return amount
}

func openOrderParamsToValues(p *types.OpenOrderParams) url.Values {
	v := url.Values{}
	if p == nil {
		return v
	}
	if p.ID != "" {
		v.Set("id", p.ID)
	}
	if p.Market != "" {
		v.Set("market", p.Market)
	}
	if p.AssetID != "" {
		v.Set("asset_id", p.AssetID)
	}
	return v
}

func tradeParamsToValues(p *types.TradeParams) url.Values {
	v := url.Values{}
	if p == nil {
		return v
	}
	if p.ID != "" {
		v.Set("id", p.ID)
	}
	if p.MakerAddress != "" {
		v.Set("maker_address", p.MakerAddress)
	}
	if p.Market != "" {
		v.Set("market", p.Market)
	}
	if p.AssetID != "" {
		v.Set("asset_id", p.AssetID)
	}
	if p.Before != "" {
		v.Set("before", p.Before)
	}
	if p.After != "" {
		v.Set("after", p.After)
	}
	return v
}
