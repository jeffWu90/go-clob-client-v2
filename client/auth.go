package client

import (
	"context"
	"encoding/json"
	"net/http"

	clob "github.com/jeffWu90/go-clob-client-v2"
	"github.com/jeffWu90/go-clob-client-v2/headers"
	"github.com/jeffWu90/go-clob-client-v2/httphelpers"
	"github.com/jeffWu90/go-clob-client-v2/types"
)

// apiKeyRaw is the wire shape of an API-key response.
type apiKeyRaw struct {
	ApiKey     string `json:"apiKey"`
	Secret     string `json:"secret"`
	Passphrase string `json:"passphrase"`
}

// CreateApiKey mints a fresh API-key credential via L1 (wallet) auth.
// nonce defaults to 0 — it acts as the API-key index within the wallet.
func (c *ClobClient) CreateApiKey(ctx context.Context, nonce int64) (*types.ApiKeyCreds, error) {
	if err := c.canL1Auth(); err != nil {
		return nil, err
	}
	ts, err := c.timestampForHeaders(ctx)
	if err != nil {
		return nil, err
	}
	hdrs, err := headers.CreateL1Headers(c.signer, c.chainID, nonce, ts)
	if err != nil {
		return nil, err
	}
	var raw apiKeyRaw
	if err := c.http.Post(ctx, c.url(clob.EndpointCreateApiKey), httphelpers.RequestOptions{Headers: hdrs}, &raw); err != nil {
		return nil, err
	}
	return rawToCreds(raw), nil
}

// DeriveApiKey returns the existing API credentials for the wallet (the same
// nonce-derived key that CreateApiKey would mint).
func (c *ClobClient) DeriveApiKey(ctx context.Context, nonce int64) (*types.ApiKeyCreds, error) {
	if err := c.canL1Auth(); err != nil {
		return nil, err
	}
	ts, err := c.timestampForHeaders(ctx)
	if err != nil {
		return nil, err
	}
	hdrs, err := headers.CreateL1Headers(c.signer, c.chainID, nonce, ts)
	if err != nil {
		return nil, err
	}
	var raw apiKeyRaw
	if err := c.http.Get(ctx, c.url(clob.EndpointDeriveApiKey), httphelpers.RequestOptions{Headers: hdrs}, &raw); err != nil {
		return nil, err
	}
	return rawToCreds(raw), nil
}

// CreateOrDeriveApiKey tries CreateApiKey first; if the result is missing a
// key (i.e. the credentials already existed) it falls back to DeriveApiKey.
func (c *ClobClient) CreateOrDeriveApiKey(ctx context.Context, nonce int64) (*types.ApiKeyCreds, error) {
	creds, err := c.CreateApiKey(ctx, nonce)
	if err == nil && creds.Key != "" {
		return creds, nil
	}
	return c.DeriveApiKey(ctx, nonce)
}

// GetApiKeys lists the API keys for the wallet.
func (c *ClobClient) GetApiKeys(ctx context.Context) (*types.ApiKeysResponse, error) {
	hdrs, err := c.buildL2Headers(ctx, http.MethodGet, clob.EndpointGetApiKeys, "")
	if err != nil {
		return nil, err
	}
	var out types.ApiKeysResponse
	err = c.http.Get(ctx, c.url(clob.EndpointGetApiKeys), httphelpers.RequestOptions{Headers: hdrs}, &out)
	return &out, err
}

// DeleteApiKey removes the API key associated with the current credentials.
func (c *ClobClient) DeleteApiKey(ctx context.Context) error {
	hdrs, err := c.buildL2Headers(ctx, http.MethodDelete, clob.EndpointDeleteApiKey, "")
	if err != nil {
		return err
	}
	return c.http.Delete(ctx, c.url(clob.EndpointDeleteApiKey), httphelpers.RequestOptions{Headers: hdrs}, nil)
}

// GetClosedOnlyMode reports whether the account is currently in closed-only
// (only allowed to close existing positions).
func (c *ClobClient) GetClosedOnlyMode(ctx context.Context) (*types.BanStatus, error) {
	hdrs, err := c.buildL2Headers(ctx, http.MethodGet, clob.EndpointClosedOnly, "")
	if err != nil {
		return nil, err
	}
	var out types.BanStatus
	err = c.http.Get(ctx, c.url(clob.EndpointClosedOnly), httphelpers.RequestOptions{Headers: hdrs}, &out)
	return &out, err
}

// CreateReadonlyApiKey mints a read-only API key.
func (c *ClobClient) CreateReadonlyApiKey(ctx context.Context) (*types.ReadonlyApiKeyResponse, error) {
	hdrs, err := c.buildL2Headers(ctx, http.MethodPost, clob.EndpointCreateReadonlyApiKey, "")
	if err != nil {
		return nil, err
	}
	var out types.ReadonlyApiKeyResponse
	err = c.http.Post(ctx, c.url(clob.EndpointCreateReadonlyApiKey), httphelpers.RequestOptions{Headers: hdrs}, &out)
	return &out, err
}

// GetReadonlyApiKeys lists the readonly API keys for the wallet.
func (c *ClobClient) GetReadonlyApiKeys(ctx context.Context) ([]string, error) {
	hdrs, err := c.buildL2Headers(ctx, http.MethodGet, clob.EndpointGetReadonlyApiKeys, "")
	if err != nil {
		return nil, err
	}
	var out []string
	err = c.http.Get(ctx, c.url(clob.EndpointGetReadonlyApiKeys), httphelpers.RequestOptions{Headers: hdrs}, &out)
	return out, err
}

// DeleteReadonlyApiKey deletes a single readonly API key by id.
func (c *ClobClient) DeleteReadonlyApiKey(ctx context.Context, key string) (bool, error) {
	payload := map[string]string{"key": key}
	body, _ := json.Marshal(payload)
	hdrs, err := c.buildL2Headers(ctx, http.MethodDelete, clob.EndpointDeleteReadonlyApiKey, string(body))
	if err != nil {
		return false, err
	}
	var out bool
	err = c.http.Delete(ctx, c.url(clob.EndpointDeleteReadonlyApiKey), httphelpers.RequestOptions{
		Headers: hdrs,
		Body:    payload,
	}, &out)
	return out, err
}

// CreateBuilderApiKey mints a fresh builder-fees credential.
func (c *ClobClient) CreateBuilderApiKey(ctx context.Context) (*types.BuilderApiKey, error) {
	hdrs, err := c.buildL2Headers(ctx, http.MethodPost, clob.EndpointCreateBuilderApiKey, "")
	if err != nil {
		return nil, err
	}
	var out types.BuilderApiKey
	err = c.http.Post(ctx, c.url(clob.EndpointCreateBuilderApiKey), httphelpers.RequestOptions{Headers: hdrs}, &out)
	return &out, err
}

// GetBuilderApiKeys lists the wallet's builder API keys (active + revoked).
func (c *ClobClient) GetBuilderApiKeys(ctx context.Context) ([]types.BuilderApiKeyResponse, error) {
	hdrs, err := c.buildL2Headers(ctx, http.MethodGet, clob.EndpointGetBuilderApiKeys, "")
	if err != nil {
		return nil, err
	}
	var out []types.BuilderApiKeyResponse
	err = c.http.Get(ctx, c.url(clob.EndpointGetBuilderApiKeys), httphelpers.RequestOptions{Headers: hdrs}, &out)
	return out, err
}

// RevokeBuilderApiKey revokes the current builder API key.
func (c *ClobClient) RevokeBuilderApiKey(ctx context.Context) error {
	hdrs, err := c.buildL2Headers(ctx, http.MethodDelete, clob.EndpointRevokeBuilderApiKey, "")
	if err != nil {
		return err
	}
	return c.http.Delete(ctx, c.url(clob.EndpointRevokeBuilderApiKey), httphelpers.RequestOptions{Headers: hdrs}, nil)
}

// buildL2Headers is an internal helper that constructs L2 HMAC headers for
// the given method + path + (optional) JSON body, using the timestamp source
// (local vs server clock) configured on the client.
func (c *ClobClient) buildL2Headers(ctx context.Context, method, path, body string) (map[string]string, error) {
	if err := c.canL2Auth(); err != nil {
		return nil, err
	}
	ts, err := c.timestampForHeaders(ctx)
	if err != nil {
		return nil, err
	}
	return headers.CreateL2Headers(c.signer, *c.creds, types.L2HeaderArgs{
		Method:      method,
		RequestPath: path,
		Body:        body,
	}, ts)
}

func rawToCreds(raw apiKeyRaw) *types.ApiKeyCreds {
	return &types.ApiKeyCreds{Key: raw.ApiKey, Secret: raw.Secret, Passphrase: raw.Passphrase}
}
