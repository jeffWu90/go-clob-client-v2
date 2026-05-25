// Package client provides ClobClient, the top-level Polymarket CLOB V2 API
// surface. A client holds optional L1 (wallet) and L2 (API-key) credentials
// and exposes ~50 HTTP endpoints for market data, orders, balances, rewards,
// notifications, and builder operations.
package client

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"

	clob "github.com/jeffWu90/go-clob-client-v2"
	"github.com/jeffWu90/go-clob-client-v2/httphelpers"
	"github.com/jeffWu90/go-clob-client-v2/orderbuilder"
	"github.com/jeffWu90/go-clob-client-v2/signing"
	"github.com/jeffWu90/go-clob-client-v2/types"
)

// Options configures a ClobClient.
type Options struct {
	// Host is the CLOB base URL, e.g. "https://clob.polymarket.com".
	// Trailing slashes are trimmed.
	Host string
	// Chain is the chain id used in EIP-712 domain separators.
	Chain types.Chain
	// Signer signs L1 (wallet) authenticated requests and orders. Optional —
	// pass nil for read-only / public-endpoint usage.
	Signer signing.Signer
	// Creds are the L2 (API key) credentials. Optional — required only for
	// authenticated endpoints (order placement, account data, etc).
	Creds *types.ApiKeyCreds
	// SignatureType defaults to SignatureTypeV2EOA when zero. Set to
	// POLY_PROXY / POLY_GNOSIS_SAFE / POLY_1271 for non-EOA flows.
	SignatureType types.SignatureTypeV2
	// FunderAddress is the proxy wallet that holds the user's funds. Defaults
	// to the signer address for plain EOAs.
	FunderAddress string
	// UseServerTime, when true, fetches /time before every signed request so
	// timestamps come from the server clock. Helps with clock-skew issues.
	UseServerTime bool
	// BuilderConfig opts orders into a builder code automatically.
	BuilderConfig *types.BuilderConfig
	// RetryOnError, when true, retries transient POST failures once.
	RetryOnError bool
	// HTTPClient overrides the underlying transport. Nil falls back to a
	// preconfigured http.Client with a 30s timeout.
	HTTPClient *http.Client
}

// ClobClient is the main API entrypoint.
type ClobClient struct {
	host    string
	chainID types.Chain

	signer signing.Signer
	creds  *types.ApiKeyCreds

	orderBuilder  *orderbuilder.OrderBuilder
	builderConfig *types.BuilderConfig

	http          *httphelpers.Client
	useServerTime bool

	mu                sync.RWMutex
	tickSizes         types.TickSizes
	negRisk           types.NegRisk
	feeInfos          types.FeeInfos
	feeRates          types.FeeRates
	builderFeeRates   types.BuilderFeeRates
	tokenConditionMap types.TokenConditionMap
}

// New constructs a ClobClient from the supplied Options.
func New(opts Options) (*ClobClient, error) {
	host := strings.TrimRight(opts.Host, "/")
	if host == "" {
		host = "https://clob.polymarket.com"
	}

	httpClient := opts.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	hh := httphelpers.New(httpClient)
	hh.RetryOnError = opts.RetryOnError

	var ob *orderbuilder.OrderBuilder
	if opts.Signer != nil {
		ob = orderbuilder.NewOrderBuilder(opts.Signer, opts.Chain)
		if opts.SignatureType != 0 {
			ob.SignatureType = opts.SignatureType
		}
		ob.FunderAddress = opts.FunderAddress
	}

	return &ClobClient{
		host:              host,
		chainID:           opts.Chain,
		signer:            opts.Signer,
		creds:             opts.Creds,
		orderBuilder:      ob,
		builderConfig:     opts.BuilderConfig,
		http:              hh,
		useServerTime:     opts.UseServerTime,
		tickSizes:         types.TickSizes{},
		negRisk:           types.NegRisk{},
		feeInfos:          types.FeeInfos{},
		feeRates:          types.FeeRates{},
		builderFeeRates:   types.BuilderFeeRates{},
		tokenConditionMap: types.TokenConditionMap{},
	}, nil
}

// Host returns the (trimmed) CLOB base URL.
func (c *ClobClient) Host() string { return c.host }

// ChainID returns the configured chain id.
func (c *ClobClient) ChainID() types.Chain { return c.chainID }

// OrderBuilder returns the underlying OrderBuilder so callers can build orders
// directly without going through the post helpers.
func (c *ClobClient) OrderBuilder() *orderbuilder.OrderBuilder { return c.orderBuilder }

func (c *ClobClient) canL1Auth() error {
	if c.signer == nil {
		return clob.ErrL1AuthUnavailable
	}
	return nil
}

func (c *ClobClient) canL2Auth() error {
	if c.signer == nil {
		return clob.ErrL1AuthUnavailable
	}
	if c.creds == nil {
		return clob.ErrL2AuthUnavailable
	}
	return nil
}

// timestampForHeaders returns the unix-seconds timestamp to embed in auth
// headers, fetching the server clock when UseServerTime is enabled.
func (c *ClobClient) timestampForHeaders(ctx context.Context) (int64, error) {
	if !c.useServerTime {
		return time.Now().Unix(), nil
	}
	return c.GetServerTime(ctx)
}

// url assembles a fully-qualified URL by concatenating host + path.
func (c *ClobClient) url(path string) string {
	return c.host + path
}
