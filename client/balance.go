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

// GetBalanceAllowance fetches the USDC / CTF balance + on-chain allowance for
// the configured wallet.
func (c *ClobClient) GetBalanceAllowance(ctx context.Context, params *types.BalanceAllowanceParams) (*types.BalanceAllowanceResponse, error) {
	hdrs, err := c.buildL2Headers(ctx, http.MethodGet, clob.EndpointGetBalanceAllowance, "")
	if err != nil {
		return nil, err
	}
	v := balanceAllowanceParamsToValues(params)
	v.Set("signature_type", strconv.Itoa(int(c.orderBuilder.SignatureType)))

	var out types.BalanceAllowanceResponse
	err = c.http.Get(ctx, c.url(clob.EndpointGetBalanceAllowance), httphelpers.RequestOptions{Headers: hdrs, Params: v}, &out)
	return &out, err
}

// UpdateBalanceAllowance triggers an on-chain allowance refresh for the wallet.
// Implemented as a GET request to match the upstream API contract.
func (c *ClobClient) UpdateBalanceAllowance(ctx context.Context, params *types.BalanceAllowanceParams) error {
	hdrs, err := c.buildL2Headers(ctx, http.MethodGet, clob.EndpointUpdateBalanceAllowance, "")
	if err != nil {
		return err
	}
	v := balanceAllowanceParamsToValues(params)
	v.Set("signature_type", strconv.Itoa(int(c.orderBuilder.SignatureType)))
	return c.http.Get(ctx, c.url(clob.EndpointUpdateBalanceAllowance), httphelpers.RequestOptions{Headers: hdrs, Params: v}, nil)
}

func balanceAllowanceParamsToValues(p *types.BalanceAllowanceParams) url.Values {
	v := url.Values{}
	if p == nil {
		return v
	}
	if p.AssetType != "" {
		v.Set("asset_type", string(p.AssetType))
	}
	if p.TokenID != "" {
		v.Set("token_id", p.TokenID)
	}
	return v
}
