package client

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	clob "github.com/jeffWu90/go-clob-client-v2"
	"github.com/jeffWu90/go-clob-client-v2/httphelpers"
	"github.com/jeffWu90/go-clob-client-v2/types"
)

// GetEarningsForUserForDay returns the per-asset earnings for a specific date,
// fully paginated.
func (c *ClobClient) GetEarningsForUserForDay(ctx context.Context, date string) ([]types.UserEarning, error) {
	hdrs, err := c.buildL2Headers(ctx, http.MethodGet, clob.EndpointGetEarningsForUserForDay, "")
	if err != nil {
		return nil, err
	}
	nextCursor := clob.InitialCursor
	var results []types.UserEarning
	for nextCursor != clob.EndCursor {
		v := url.Values{
			"date":           []string{date},
			"signature_type": []string{strconv.Itoa(int(c.orderBuilder.SignatureType))},
			"next_cursor":    []string{nextCursor},
		}
		var page struct {
			NextCursor string              `json:"next_cursor"`
			Data       []types.UserEarning `json:"data"`
		}
		if err := c.http.Get(ctx, c.url(clob.EndpointGetEarningsForUserForDay), httphelpers.RequestOptions{Headers: hdrs, Params: v}, &page); err != nil {
			return results, err
		}
		results = append(results, page.Data...)
		nextCursor = page.NextCursor
	}
	return results, nil
}

// GetTotalEarningsForUserForDay returns the aggregated per-asset earnings (not paginated).
func (c *ClobClient) GetTotalEarningsForUserForDay(ctx context.Context, date string) ([]types.TotalUserEarning, error) {
	hdrs, err := c.buildL2Headers(ctx, http.MethodGet, clob.EndpointGetTotalEarningsForUserForDay, "")
	if err != nil {
		return nil, err
	}
	v := url.Values{
		"date":           []string{date},
		"signature_type": []string{strconv.Itoa(int(c.orderBuilder.SignatureType))},
	}
	var out []types.TotalUserEarning
	err = c.http.Get(ctx, c.url(clob.EndpointGetTotalEarningsForUserForDay), httphelpers.RequestOptions{Headers: hdrs, Params: v}, &out)
	return out, err
}

// GetUserEarningsAndMarketsConfig pages through /rewards/user/markets,
// returning earnings + reward config per market.
func (c *ClobClient) GetUserEarningsAndMarketsConfig(ctx context.Context, date, orderBy, position string, noCompetition bool) ([]types.UserRewardsEarning, error) {
	hdrs, err := c.buildL2Headers(ctx, http.MethodGet, clob.EndpointGetRewardsEarningsPercentages, "")
	if err != nil {
		return nil, err
	}
	nextCursor := clob.InitialCursor
	var results []types.UserRewardsEarning
	for nextCursor != clob.EndCursor {
		v := url.Values{
			"date":           []string{date},
			"signature_type": []string{strconv.Itoa(int(c.orderBuilder.SignatureType))},
			"next_cursor":    []string{nextCursor},
			"no_competition": []string{strconv.FormatBool(noCompetition)},
		}
		if orderBy != "" {
			v.Set("order_by", orderBy)
		}
		if position != "" {
			v.Set("position", position)
		}
		var page struct {
			NextCursor string                     `json:"next_cursor"`
			Data       []types.UserRewardsEarning `json:"data"`
		}
		if err := c.http.Get(ctx, c.url(clob.EndpointGetRewardsEarningsPercentages), httphelpers.RequestOptions{Headers: hdrs, Params: v}, &page); err != nil {
			return results, err
		}
		results = append(results, page.Data...)
		nextCursor = page.NextCursor
	}
	return results, nil
}

// GetRewardPercentages returns the liquidity reward percentages per market.
func (c *ClobClient) GetRewardPercentages(ctx context.Context) (types.RewardsPercentages, error) {
	hdrs, err := c.buildL2Headers(ctx, http.MethodGet, clob.EndpointGetLiquidityRewardPercentages, "")
	if err != nil {
		return nil, err
	}
	v := url.Values{"signature_type": []string{strconv.Itoa(int(c.orderBuilder.SignatureType))}}
	var out types.RewardsPercentages
	err = c.http.Get(ctx, c.url(clob.EndpointGetLiquidityRewardPercentages), httphelpers.RequestOptions{Headers: hdrs, Params: v}, &out)
	return out, err
}

// GetCurrentRewards pages through /rewards/markets/current.
func (c *ClobClient) GetCurrentRewards(ctx context.Context) ([]types.MarketReward, error) {
	nextCursor := clob.InitialCursor
	var results []types.MarketReward
	for nextCursor != clob.EndCursor {
		v := url.Values{"next_cursor": []string{nextCursor}}
		var page struct {
			NextCursor string               `json:"next_cursor"`
			Data       []types.MarketReward `json:"data"`
		}
		if err := c.http.Get(ctx, c.url(clob.EndpointGetRewardsMarketsCurrent), httphelpers.RequestOptions{Params: v}, &page); err != nil {
			return results, err
		}
		results = append(results, page.Data...)
		nextCursor = page.NextCursor
	}
	return results, nil
}

// GetRawRewardsForMarket pages through /rewards/markets/{conditionID}.
func (c *ClobClient) GetRawRewardsForMarket(ctx context.Context, conditionID string) ([]types.MarketReward, error) {
	nextCursor := clob.InitialCursor
	var results []types.MarketReward
	endpoint := clob.EndpointGetRewardsMarkets + conditionID
	for nextCursor != clob.EndCursor {
		v := url.Values{"next_cursor": []string{nextCursor}}
		var page struct {
			NextCursor string               `json:"next_cursor"`
			Data       []types.MarketReward `json:"data"`
		}
		if err := c.http.Get(ctx, c.url(endpoint), httphelpers.RequestOptions{Params: v}, &page); err != nil {
			return results, err
		}
		results = append(results, page.Data...)
		nextCursor = page.NextCursor
	}
	return results, nil
}
