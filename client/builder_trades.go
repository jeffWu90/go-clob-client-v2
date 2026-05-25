package client

import (
	"context"
	"fmt"
	"net/url"

	clob "github.com/jeffWu90/go-clob-client-v2"
	"github.com/jeffWu90/go-clob-client-v2/httphelpers"
	"github.com/jeffWu90/go-clob-client-v2/types"
)

// GetBuilderTrades returns trades attributed to a specific builder code.
// builder_code is required and cannot be the zero bytes32.
func (c *ClobClient) GetBuilderTrades(ctx context.Context, params types.BuilderTradeParams, nextCursor string) (*types.BuilderTradesResponse, error) {
	if params.BuilderCode == "" || params.BuilderCode == clob.Bytes32Zero {
		return nil, fmt.Errorf("builderCode is required and cannot be zero")
	}
	if nextCursor == "" {
		nextCursor = clob.InitialCursor
	}
	v := url.Values{
		"builder_code": []string{params.BuilderCode},
		"next_cursor":  []string{nextCursor},
	}
	if params.ID != "" {
		v.Set("id", params.ID)
	}
	if params.MakerAddress != "" {
		v.Set("maker_address", params.MakerAddress)
	}
	if params.Market != "" {
		v.Set("market", params.Market)
	}
	if params.AssetID != "" {
		v.Set("asset_id", params.AssetID)
	}
	if params.Before != "" {
		v.Set("before", params.Before)
	}
	if params.After != "" {
		v.Set("after", params.After)
	}

	var page struct {
		NextCursor string               `json:"next_cursor"`
		Limit      int                  `json:"limit"`
		Count      int                  `json:"count"`
		Data       []types.BuilderTrade `json:"data"`
	}
	if err := c.http.Get(ctx, c.url(clob.EndpointGetBuilderTrades), httphelpers.RequestOptions{Params: v}, &page); err != nil {
		return nil, err
	}
	return &types.BuilderTradesResponse{
		Trades:     page.Data,
		NextCursor: page.NextCursor,
		Limit:      page.Limit,
		Count:      page.Count,
	}, nil
}
