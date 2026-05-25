package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	clob "github.com/jeffWu90/go-clob-client-v2"
	"github.com/jeffWu90/go-clob-client-v2/httphelpers"
	"github.com/jeffWu90/go-clob-client-v2/types"
)

// GetOk hits the /ok health probe. Returns true when the server is healthy.
func (c *ClobClient) GetOk(ctx context.Context) (bool, error) {
	var out any
	if err := c.http.Get(ctx, c.url(clob.EndpointOK), httphelpers.RequestOptions{}, &out); err != nil {
		return false, err
	}
	return true, nil
}

// GetServerTime returns the server's current unix-seconds time.
func (c *ClobClient) GetServerTime(ctx context.Context) (int64, error) {
	var raw json.Number
	if err := c.http.Get(ctx, c.url(clob.EndpointTime), httphelpers.RequestOptions{}, &raw); err != nil {
		return 0, err
	}
	return raw.Int64()
}

// GetMarkets pages through /markets, returning a single page at the supplied
// cursor. Pass clob.InitialCursor for the first page; the response includes
// the cursor for the next page (clob.EndCursor when exhausted).
func (c *ClobClient) GetMarkets(ctx context.Context, nextCursor string) (*types.PaginationPayload, error) {
	if nextCursor == "" {
		nextCursor = clob.InitialCursor
	}
	var out types.PaginationPayload
	err := c.http.Get(ctx, c.url(clob.EndpointGetMarkets), httphelpers.RequestOptions{
		Params: url.Values{"next_cursor": []string{nextCursor}},
	}, &out)
	return &out, err
}

// GetSamplingMarkets pages through /sampling-markets.
func (c *ClobClient) GetSamplingMarkets(ctx context.Context, nextCursor string) (*types.PaginationPayload, error) {
	if nextCursor == "" {
		nextCursor = clob.InitialCursor
	}
	var out types.PaginationPayload
	err := c.http.Get(ctx, c.url(clob.EndpointGetSamplingMarkets), httphelpers.RequestOptions{
		Params: url.Values{"next_cursor": []string{nextCursor}},
	}, &out)
	return &out, err
}

// GetSimplifiedMarkets pages through /simplified-markets.
func (c *ClobClient) GetSimplifiedMarkets(ctx context.Context, nextCursor string) (*types.PaginationPayload, error) {
	if nextCursor == "" {
		nextCursor = clob.InitialCursor
	}
	var out types.PaginationPayload
	err := c.http.Get(ctx, c.url(clob.EndpointGetSimplifiedMarkets), httphelpers.RequestOptions{
		Params: url.Values{"next_cursor": []string{nextCursor}},
	}, &out)
	return &out, err
}

// GetSamplingSimplifiedMarkets pages through /sampling-simplified-markets.
func (c *ClobClient) GetSamplingSimplifiedMarkets(ctx context.Context, nextCursor string) (*types.PaginationPayload, error) {
	if nextCursor == "" {
		nextCursor = clob.InitialCursor
	}
	var out types.PaginationPayload
	err := c.http.Get(ctx, c.url(clob.EndpointGetSamplingSimplifiedMarkets), httphelpers.RequestOptions{
		Params: url.Values{"next_cursor": []string{nextCursor}},
	}, &out)
	return &out, err
}

// GetMarket fetches one market by condition id. The shape is dynamic; the
// raw map is returned so callers can pick out the fields they need.
func (c *ClobClient) GetMarket(ctx context.Context, conditionID string) (map[string]any, error) {
	var out map[string]any
	err := c.http.Get(ctx, c.url(clob.EndpointGetMarket+conditionID), httphelpers.RequestOptions{}, &out)
	return out, err
}

// GetClobMarketInfo fetches /clob-markets/{conditionID}. Side effects:
// populates the internal tickSize / negRisk / feeInfo / tokenConditionMap
// caches for every token id in the returned market.
func (c *ClobClient) GetClobMarketInfo(ctx context.Context, conditionID string) (*types.MarketDetails, error) {
	var out types.MarketDetails
	if err := c.http.Get(ctx, c.url(clob.EndpointGetClobMarket+conditionID), httphelpers.RequestOptions{}, &out); err != nil {
		return nil, err
	}
	if out.Tokens == [2]*types.ClobToken{} {
		return nil, fmt.Errorf("failed to fetch market info for condition id %s", conditionID)
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	for _, tok := range out.Tokens {
		if tok == nil {
			continue
		}
		c.tokenConditionMap[tok.TokenID] = conditionID
		c.tickSizes[tok.TokenID] = types.TickSize(strconv.FormatFloat(out.MinTickSize, 'f', -1, 64))
		c.negRisk[tok.TokenID] = out.NegRisk
		var rate, exp float64
		if out.FeeDetails != nil {
			if out.FeeDetails.Rate != nil {
				rate = *out.FeeDetails.Rate
			}
			if out.FeeDetails.Exponent != nil {
				exp = *out.FeeDetails.Exponent
			}
		}
		c.feeInfos[tok.TokenID] = types.FeeInfo{Rate: rate, Exponent: exp}
	}
	return &out, nil
}

// GetOrderBook returns the orderbook snapshot for a token.
func (c *ClobClient) GetOrderBook(ctx context.Context, tokenID string) (*types.OrderBookSummary, error) {
	var out types.OrderBookSummary
	err := c.http.Get(ctx, c.url(clob.EndpointGetOrderBook), httphelpers.RequestOptions{
		Params: url.Values{"token_id": []string{tokenID}},
	}, &out)
	return &out, err
}

// GetOrderBooks batch-fetches multiple orderbooks via POST /books.
func (c *ClobClient) GetOrderBooks(ctx context.Context, params []types.BookParams) ([]types.OrderBookSummary, error) {
	var out []types.OrderBookSummary
	err := c.http.Post(ctx, c.url(clob.EndpointGetOrderBooks), httphelpers.RequestOptions{Body: params}, &out)
	return out, err
}

// GetOrderBookHash returns the SHA-1 hash of the orderbook (also sets it on the
// passed object). Useful for client-side change detection.
func (c *ClobClient) GetOrderBookHash(ob *types.OrderBookSummary) string {
	return clob.GenerateOrderBookSummaryHash(ob)
}

// GetTickSize returns the minimum tick size for the token, using the local
// cache when warm.
func (c *ClobClient) GetTickSize(ctx context.Context, tokenID string) (types.TickSize, error) {
	c.mu.RLock()
	if ts, ok := c.tickSizes[tokenID]; ok {
		c.mu.RUnlock()
		return ts, nil
	}
	conditionID, hasCondition := c.tokenConditionMap[tokenID]
	c.mu.RUnlock()

	if hasCondition {
		if _, err := c.GetClobMarketInfo(ctx, conditionID); err != nil {
			return "", err
		}
		c.mu.RLock()
		ts := c.tickSizes[tokenID]
		c.mu.RUnlock()
		return ts, nil
	}

	var raw struct {
		MinimumTickSize json.Number `json:"minimum_tick_size"`
	}
	if err := c.http.Get(ctx, c.url(clob.EndpointGetTickSize), httphelpers.RequestOptions{
		Params: url.Values{"token_id": []string{tokenID}},
	}, &raw); err != nil {
		return "", err
	}
	ts := types.TickSize(raw.MinimumTickSize.String())
	c.mu.Lock()
	c.tickSizes[tokenID] = ts
	c.mu.Unlock()
	return ts, nil
}

// GetNegRisk returns whether the token belongs to a negative-risk market.
func (c *ClobClient) GetNegRisk(ctx context.Context, tokenID string) (bool, error) {
	c.mu.RLock()
	if v, ok := c.negRisk[tokenID]; ok {
		c.mu.RUnlock()
		return v, nil
	}
	conditionID, hasCondition := c.tokenConditionMap[tokenID]
	c.mu.RUnlock()

	if hasCondition {
		if _, err := c.GetClobMarketInfo(ctx, conditionID); err != nil {
			return false, err
		}
		c.mu.RLock()
		v := c.negRisk[tokenID]
		c.mu.RUnlock()
		return v, nil
	}

	var raw struct {
		NegRisk bool `json:"neg_risk"`
	}
	if err := c.http.Get(ctx, c.url(clob.EndpointGetNegRisk), httphelpers.RequestOptions{
		Params: url.Values{"token_id": []string{tokenID}},
	}, &raw); err != nil {
		return false, err
	}
	c.mu.Lock()
	c.negRisk[tokenID] = raw.NegRisk
	c.mu.Unlock()
	return raw.NegRisk, nil
}

// GetFeeRateBps returns the per-market base fee rate (V1 compat — V2 uses
// FeeInfo instead, but the endpoint is still exposed for completeness).
func (c *ClobClient) GetFeeRateBps(ctx context.Context, tokenID string) (float64, error) {
	c.mu.RLock()
	if v, ok := c.feeRates[tokenID]; ok {
		c.mu.RUnlock()
		return v, nil
	}
	c.mu.RUnlock()

	var raw struct {
		BaseFee json.Number `json:"base_fee"`
	}
	if err := c.http.Get(ctx, c.url(clob.EndpointGetFeeRate), httphelpers.RequestOptions{
		Params: url.Values{"token_id": []string{tokenID}},
	}, &raw); err != nil {
		return 0, err
	}
	v, err := raw.BaseFee.Float64()
	if err != nil {
		return 0, err
	}
	c.mu.Lock()
	c.feeRates[tokenID] = v
	c.mu.Unlock()
	return v, nil
}

// GetFeeExponent returns the per-market fee exponent (V2 fee model).
func (c *ClobClient) GetFeeExponent(ctx context.Context, tokenID string) (float64, error) {
	if err := c.ensureMarketInfoCached(ctx, tokenID); err != nil {
		return 0, err
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.feeInfos[tokenID].Exponent, nil
}

// GetMidpoint returns the midpoint price for a token.
func (c *ClobClient) GetMidpoint(ctx context.Context, tokenID string) (map[string]any, error) {
	var out map[string]any
	err := c.http.Get(ctx, c.url(clob.EndpointGetMidpoint), httphelpers.RequestOptions{
		Params: url.Values{"token_id": []string{tokenID}},
	}, &out)
	return out, err
}

// GetMidpoints batch-fetches midpoint prices.
func (c *ClobClient) GetMidpoints(ctx context.Context, params []types.BookParams) (map[string]any, error) {
	var out map[string]any
	err := c.http.Post(ctx, c.url(clob.EndpointGetMidpoints), httphelpers.RequestOptions{Body: params}, &out)
	return out, err
}

// GetPrice returns the BUY/SELL price for a single token.
func (c *ClobClient) GetPrice(ctx context.Context, tokenID string, side types.Side) (map[string]any, error) {
	var out map[string]any
	err := c.http.Get(ctx, c.url(clob.EndpointGetPrice), httphelpers.RequestOptions{
		Params: url.Values{"token_id": []string{tokenID}, "side": []string{string(side)}},
	}, &out)
	return out, err
}

// GetPrices batch-fetches prices.
func (c *ClobClient) GetPrices(ctx context.Context, params []types.BookParams) (map[string]any, error) {
	var out map[string]any
	err := c.http.Post(ctx, c.url(clob.EndpointGetPrices), httphelpers.RequestOptions{Body: params}, &out)
	return out, err
}

// GetSpread returns the spread for a single token.
func (c *ClobClient) GetSpread(ctx context.Context, tokenID string) (map[string]any, error) {
	var out map[string]any
	err := c.http.Get(ctx, c.url(clob.EndpointGetSpread), httphelpers.RequestOptions{
		Params: url.Values{"token_id": []string{tokenID}},
	}, &out)
	return out, err
}

// GetSpreads batch-fetches spreads.
func (c *ClobClient) GetSpreads(ctx context.Context, params []types.BookParams) (map[string]any, error) {
	var out map[string]any
	err := c.http.Post(ctx, c.url(clob.EndpointGetSpreads), httphelpers.RequestOptions{Body: params}, &out)
	return out, err
}

// GetLastTradePrice returns the last trade price for a token.
func (c *ClobClient) GetLastTradePrice(ctx context.Context, tokenID string) (map[string]any, error) {
	var out map[string]any
	err := c.http.Get(ctx, c.url(clob.EndpointGetLastTradePrice), httphelpers.RequestOptions{
		Params: url.Values{"token_id": []string{tokenID}},
	}, &out)
	return out, err
}

// GetLastTradesPrices batch-fetches last trade prices.
func (c *ClobClient) GetLastTradesPrices(ctx context.Context, params []types.BookParams) (map[string]any, error) {
	var out map[string]any
	err := c.http.Post(ctx, c.url(clob.EndpointGetLastTradesPrices), httphelpers.RequestOptions{Body: params}, &out)
	return out, err
}

// GetPricesHistory returns the historical price series. Requires either an
// Interval, or both StartTs and EndTs.
func (c *ClobClient) GetPricesHistory(ctx context.Context, p types.PriceHistoryFilterParams) ([]types.MarketPrice, error) {
	if p.Interval == "" && (p.StartTs == 0 || p.EndTs == 0) {
		return nil, fmt.Errorf("GetPricesHistory requires either interval or both startTs and endTs")
	}
	v := url.Values{}
	if p.Market != "" {
		v.Set("market", p.Market)
	}
	if p.StartTs != 0 {
		v.Set("startTs", strconv.FormatInt(p.StartTs, 10))
	}
	if p.EndTs != 0 {
		v.Set("endTs", strconv.FormatInt(p.EndTs, 10))
	}
	if p.Fidelity != 0 {
		v.Set("fidelity", strconv.Itoa(p.Fidelity))
	}
	if p.Interval != "" {
		v.Set("interval", string(p.Interval))
	}
	var out []types.MarketPrice
	err := c.http.Get(ctx, c.url(clob.EndpointGetPricesHistory), httphelpers.RequestOptions{Params: v}, &out)
	return out, err
}

// GetMarketTradesEvents returns the live-activity trade events for a market.
func (c *ClobClient) GetMarketTradesEvents(ctx context.Context, conditionID string) ([]types.MarketTradeEvent, error) {
	var out []types.MarketTradeEvent
	err := c.http.Get(ctx, c.url(clob.EndpointGetMarketTradesEvents+conditionID), httphelpers.RequestOptions{}, &out)
	return out, err
}

// ensureMarketInfoCached fetches the market info if the token's fee info
// hasn't been cached yet. Used by methods that need fees / decimals before
// signing a market order.
func (c *ClobClient) ensureMarketInfoCached(ctx context.Context, tokenID string) error {
	c.mu.RLock()
	if _, ok := c.feeInfos[tokenID]; ok {
		c.mu.RUnlock()
		return nil
	}
	conditionID, hasCondition := c.tokenConditionMap[tokenID]
	c.mu.RUnlock()

	if !hasCondition {
		var raw struct {
			ConditionID string `json:"condition_id"`
		}
		if err := c.http.Get(ctx, c.url(clob.EndpointGetMarketByToken+tokenID), httphelpers.RequestOptions{}, &raw); err != nil {
			return err
		}
		if raw.ConditionID == "" {
			return fmt.Errorf("failed to resolve condition id for token %s", tokenID)
		}
		c.mu.Lock()
		c.tokenConditionMap[tokenID] = raw.ConditionID
		conditionID = raw.ConditionID
		c.mu.Unlock()
	}

	_, err := c.GetClobMarketInfo(ctx, conditionID)
	return err
}
