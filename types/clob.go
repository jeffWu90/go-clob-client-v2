package types

// ApiKeyCreds is the L2 (HMAC) credentials triplet returned by the auth endpoints.
type ApiKeyCreds struct {
	Key        string `json:"apiKey"`
	Secret     string `json:"secret"`
	Passphrase string `json:"passphrase"`
}

// BuilderConfig configures the builder code attached to orders placed by the client.
type BuilderConfig struct {
	BuilderCode string
}

// L2HeaderArgs are the inputs needed to build an HMAC L2 auth header.
type L2HeaderArgs struct {
	Method      string
	RequestPath string
	Body        string
}

// L1PolyHeader is the L1 (EIP-712 wallet signature) request-header set.
type L1PolyHeader struct {
	PolyAddress   string `json:"POLY_ADDRESS"`
	PolySignature string `json:"POLY_SIGNATURE"`
	PolyTimestamp string `json:"POLY_TIMESTAMP"`
	PolyNonce     string `json:"POLY_NONCE"`
}

// L2PolyHeader is the L2 (HMAC API-key) request-header set.
type L2PolyHeader struct {
	PolyAddress    string `json:"POLY_ADDRESS"`
	PolySignature  string `json:"POLY_SIGNATURE"`
	PolyTimestamp  string `json:"POLY_TIMESTAMP"`
	PolyApiKey     string `json:"POLY_API_KEY"`
	PolyPassphrase string `json:"POLY_PASSPHRASE"`
}

// OrderPayload is the {orderID} response shape from order endpoints.
type OrderPayload struct {
	OrderID string `json:"orderID"`
}

// OrderResponse is the full response from POST /order.
type OrderResponse struct {
	Success            bool     `json:"success"`
	ErrorMsg           string   `json:"errorMsg"`
	OrderID            string   `json:"orderID"`
	TransactionsHashes []string `json:"transactionsHashes"`
	Status             string   `json:"status"`
	TakingAmount       string   `json:"takingAmount"`
	MakingAmount       string   `json:"makingAmount"`
}

// OpenOrder describes an open or recently filled order.
type OpenOrder struct {
	ID               string   `json:"id"`
	Status           string   `json:"status"`
	Owner            string   `json:"owner"`
	MakerAddress     string   `json:"maker_address"`
	Market           string   `json:"market"`
	AssetID          string   `json:"asset_id"`
	Side             string   `json:"side"`
	OriginalSize     string   `json:"original_size"`
	SizeMatched      string   `json:"size_matched"`
	Price            string   `json:"price"`
	AssociateTrades  []string `json:"associate_trades"`
	Outcome          string   `json:"outcome"`
	CreatedAt        int64    `json:"created_at"`
	Expiration       string   `json:"expiration"`
	OrderType        string   `json:"order_type"`
}

// MakerOrder is the maker side of a Trade.
type MakerOrder struct {
	OrderID       string `json:"order_id"`
	Owner         string `json:"owner"`
	MakerAddress  string `json:"maker_address"`
	MatchedAmount string `json:"matched_amount"`
	Price         string `json:"price"`
	FeeRateBps    string `json:"fee_rate_bps"`
	AssetID       string `json:"asset_id"`
	Outcome       string `json:"outcome"`
	Side          Side   `json:"side"`
}

// Trade is a CLOB trade event.
type Trade struct {
	ID              string       `json:"id"`
	TakerOrderID    string       `json:"taker_order_id"`
	Market          string       `json:"market"`
	AssetID         string       `json:"asset_id"`
	Side            Side         `json:"side"`
	Size            string       `json:"size"`
	FeeRateBps      string       `json:"fee_rate_bps"`
	Price           string       `json:"price"`
	Status          string       `json:"status"`
	MatchTime       string       `json:"match_time"`
	LastUpdate      string       `json:"last_update"`
	Outcome         string       `json:"outcome"`
	BucketIndex     int          `json:"bucket_index"`
	Owner           string       `json:"owner"`
	MakerAddress    string       `json:"maker_address"`
	MakerOrders     []MakerOrder `json:"maker_orders"`
	TransactionHash string       `json:"transaction_hash"`
	TraderSide      string       `json:"trader_side"` // "TAKER" or "MAKER"
}

// ApiKeysResponse is the response of GET /auth/api-keys.
type ApiKeysResponse struct {
	ApiKeys []ApiKeyCreds `json:"apiKeys"`
}

// BanStatus is the response of GET /auth/ban-status/closed-only.
type BanStatus struct {
	ClosedOnly bool `json:"closed_only"`
}

// TradeParams are filters accepted by GET /data/trades.
type TradeParams struct {
	ID           string
	MakerAddress string
	Market       string
	AssetID      string
	Before       string
	After        string
}

// BuilderTradeParams extends TradeParams with a builder-code filter.
type BuilderTradeParams struct {
	TradeParams
	BuilderCode string
}

// OpenOrderParams are filters accepted by GET /data/orders.
type OpenOrderParams struct {
	ID      string
	Market  string
	AssetID string
}

// MarketPrice is a single (timestamp, price) sample.
type MarketPrice struct {
	T int64   `json:"t"`
	P float64 `json:"p"`
}

// PriceHistoryInterval enumerates the canonical history buckets.
type PriceHistoryInterval string

const (
	PriceHistoryIntervalMax     PriceHistoryInterval = "max"
	PriceHistoryIntervalOneWeek PriceHistoryInterval = "1w"
	PriceHistoryIntervalOneDay  PriceHistoryInterval = "1d"
	PriceHistoryIntervalSixHour PriceHistoryInterval = "6h"
	PriceHistoryIntervalOneHour PriceHistoryInterval = "1h"
)

// PriceHistoryFilterParams are filters accepted by GET /prices-history.
type PriceHistoryFilterParams struct {
	Market   string
	StartTs  int64
	EndTs    int64
	Fidelity int
	Interval PriceHistoryInterval
}

// DropNotificationParams is the body of DELETE /notifications.
type DropNotificationParams struct {
	IDs []string `json:"ids"`
}

// Notification is a single notification record.
type Notification struct {
	Type    int            `json:"type"`
	Owner   string         `json:"owner"`
	Payload map[string]any `json:"payload"`
}

// OrderMarketCancelParams are filters for DELETE /cancel-market-orders.
type OrderMarketCancelParams struct {
	Market  string
	AssetID string
}

// OrderBookSummary is the canonical orderbook snapshot.
type OrderBookSummary struct {
	Market         string         `json:"market"`
	AssetID        string         `json:"asset_id"`
	Timestamp      string         `json:"timestamp"`
	Bids           []OrderSummary `json:"bids"`
	Asks           []OrderSummary `json:"asks"`
	MinOrderSize   string         `json:"min_order_size"`
	TickSize       string         `json:"tick_size"`
	NegRisk        bool           `json:"neg_risk"`
	Hash           string         `json:"hash"`
	LastTradePrice string         `json:"last_trade_price"`
}

// OrderSummary is a single price-level entry in the orderbook.
type OrderSummary struct {
	Price string `json:"price"`
	Size  string `json:"size"`
}

// AssetType identifies the token kind for balance / allowance queries.
type AssetType string

const (
	AssetTypeCollateral  AssetType = "COLLATERAL"
	AssetTypeConditional AssetType = "CONDITIONAL"
)

// BalanceAllowanceParams is the input for GET /balance-allowance.
type BalanceAllowanceParams struct {
	AssetType AssetType
	TokenID   string
}

// BalanceAllowanceResponse is the result of GET /balance-allowance.
type BalanceAllowanceResponse struct {
	Balance   string `json:"balance"`
	Allowance string `json:"allowance"`
}

// OrderScoringParams is the input for GET /order-scoring.
type OrderScoringParams struct {
	OrderID string
}

// OrderScoring is a single-order scoring response.
type OrderScoring struct {
	Scoring bool `json:"scoring"`
}

// OrdersScoringParams is the input for GET /orders-scoring.
type OrdersScoringParams struct {
	OrderIDs []string
}

// OrdersScoring maps each order id to its scoring flag.
type OrdersScoring map[string]bool

// TickSize is one of the canonical tick-size buckets.
type TickSize string

const (
	TickSize1     TickSize = "0.1"
	TickSize01    TickSize = "0.01"
	TickSize001   TickSize = "0.001"
	TickSize0001  TickSize = "0.0001"
)

// CreateOrderOptions configures order-builder behaviour.
type CreateOrderOptions struct {
	TickSize TickSize
	NegRisk  bool
}

// RoundConfig describes how many decimal places to round price/size/amount to
// at a given tick size.
type RoundConfig struct {
	Price  int
	Size   int
	Amount int
}

// TickSizes maps token id -> tick size.
type TickSizes map[string]TickSize

// FeeRates maps token id -> fee rate (in bps).
type FeeRates map[string]float64

// NegRisk maps token id -> negative-risk flag.
type NegRisk map[string]bool

// FeeInfo describes the V2 platform fee for a token.
type FeeInfo struct {
	Rate     float64 `json:"rate"`
	Exponent float64 `json:"exponent"`
}

// FeeInfos maps token id -> FeeInfo.
type FeeInfos map[string]FeeInfo

// BuilderFeeRates maps a builder code to its maker/taker fee rates.
type BuilderFeeRates map[string]struct {
	Maker float64 `json:"maker"`
	Taker float64 `json:"taker"`
}

// TokenConditionMap maps token id -> condition id.
type TokenConditionMap map[string]string

// FeeDetails is the platform fee descriptor from /clob-markets.
type FeeDetails struct {
	Rate      *float64 `json:"r,omitempty"`
	Exponent  *float64 `json:"e,omitempty"`
	TakerOnly bool     `json:"to"`
}

// ClobToken is one outcome token in MarketDetails.
type ClobToken struct {
	TokenID string `json:"t"`
	Outcome string `json:"o"`
}

// MarketDetails is the response shape of GET /clob-markets/{conditionID}.
type MarketDetails struct {
	ConditionID string        `json:"c"`
	Tokens      [2]*ClobToken `json:"t"`
	MinTickSize float64       `json:"mts"`
	NegRisk     bool          `json:"nr"`
	FeeDetails  *FeeDetails   `json:"fd,omitempty"`
	V1MakerFee  *float64      `json:"mbf,omitempty"`
	V1TakerFee  *float64      `json:"tbf,omitempty"`
}

// PaginationPayload is a generic paginated response envelope.
type PaginationPayload struct {
	Limit      int   `json:"limit"`
	Count      int   `json:"count"`
	NextCursor string `json:"next_cursor"`
	Data       []any  `json:"data"`
}

// BookParams is one entry in the batch /books request.
type BookParams struct {
	TokenID string `json:"token_id"`
	Side    Side   `json:"side"`
}

// UserEarning is a single per-asset earnings row.
type UserEarning struct {
	Date         string  `json:"date"`
	ConditionID  string  `json:"condition_id"`
	AssetAddress string  `json:"asset_address"`
	MakerAddress string  `json:"maker_address"`
	Earnings     float64 `json:"earnings"`
	AssetRate    float64 `json:"asset_rate"`
}

// TotalUserEarning is the aggregated earnings row.
type TotalUserEarning struct {
	Date         string  `json:"date"`
	AssetAddress string  `json:"asset_address"`
	MakerAddress string  `json:"maker_address"`
	Earnings     float64 `json:"earnings"`
	AssetRate    float64 `json:"asset_rate"`
}

// RewardsPercentages maps market -> earning percentage.
type RewardsPercentages map[string]float64

// Token is an outcome token used by reward responses.
type Token struct {
	TokenID string  `json:"token_id"`
	Outcome string  `json:"outcome"`
	Price   float64 `json:"price"`
}

// RewardsConfig is the per-asset reward configuration.
type RewardsConfig struct {
	AssetAddress  string  `json:"asset_address"`
	StartDate     string  `json:"start_date"`
	EndDate       string  `json:"end_date"`
	RatePerDay    float64 `json:"rate_per_day"`
	TotalRewards  float64 `json:"total_rewards"`
}

// MarketReward is the per-market reward shape.
type MarketReward struct {
	ConditionID       string          `json:"condition_id"`
	Question          string          `json:"question"`
	MarketSlug        string          `json:"market_slug"`
	EventSlug         string          `json:"event_slug"`
	Image             string          `json:"image"`
	RewardsMaxSpread  float64         `json:"rewards_max_spread"`
	RewardsMinSize    float64         `json:"rewards_min_size"`
	Tokens            []Token         `json:"tokens"`
	RewardsConfig     []RewardsConfig `json:"rewards_config"`
}

// Earning is a single asset's earnings entry inside UserRewardsEarning.
type Earning struct {
	AssetAddress string  `json:"asset_address"`
	Earnings     float64 `json:"earnings"`
	AssetRate    float64 `json:"asset_rate"`
}

// UserRewardsEarning is the per-market user reward record.
type UserRewardsEarning struct {
	ConditionID            string          `json:"condition_id"`
	Question               string          `json:"question"`
	MarketSlug             string          `json:"market_slug"`
	EventSlug              string          `json:"event_slug"`
	Image                  string          `json:"image"`
	RewardsMaxSpread       float64         `json:"rewards_max_spread"`
	RewardsMinSize         float64         `json:"rewards_min_size"`
	MarketCompetitiveness  float64         `json:"market_competitiveness"`
	Tokens                 []Token         `json:"tokens"`
	RewardsConfig          []RewardsConfig `json:"rewards_config"`
	MakerAddress           string          `json:"maker_address"`
	EarningPercentage      float64         `json:"earning_percentage"`
	Earnings               []Earning       `json:"earnings"`
}

// BuilderTrade is a single trade attributed to a builder code.
type BuilderTrade struct {
	ID              string  `json:"id"`
	TradeType       string  `json:"tradeType"`
	TakerOrderHash  string  `json:"takerOrderHash"`
	Builder         string  `json:"builder"`
	Market          string  `json:"market"`
	AssetID         string  `json:"assetId"`
	Side            string  `json:"side"`
	Size            string  `json:"size"`
	SizeUsdc        string  `json:"sizeUsdc"`
	Price           string  `json:"price"`
	Status          string  `json:"status"`
	Outcome         string  `json:"outcome"`
	OutcomeIndex    int     `json:"outcomeIndex"`
	Owner           string  `json:"owner"`
	Maker           string  `json:"maker"`
	TransactionHash string  `json:"transactionHash"`
	MatchTime       string  `json:"matchTime"`
	BucketIndex     int     `json:"bucketIndex"`
	Fee             string  `json:"fee"`
	FeeUsdc         string  `json:"feeUsdc"`
	ErrMsg          *string `json:"err_msg,omitempty"`
	CreatedAt       *string `json:"createdAt"`
	UpdatedAt       *string `json:"updatedAt"`
}

// ReadonlyApiKeyResponse is the result of POST /auth/readonly-api-key.
type ReadonlyApiKeyResponse struct {
	ApiKey string `json:"apiKey"`
}

// MarketTradeEvent is one live-activity event.
type MarketTradeEvent struct {
	EventType string `json:"event_type"`
	Market    struct {
		ConditionID string `json:"condition_id"`
		AssetID     string `json:"asset_id"`
		Question    string `json:"question"`
		Icon        string `json:"icon"`
		Slug        string `json:"slug"`
	} `json:"market"`
	User struct {
		Address                 string `json:"address"`
		Username                string `json:"username"`
		ProfilePicture          string `json:"profile_picture"`
		OptimizedProfilePicture string `json:"optimized_profile_picture"`
		Pseudonym               string `json:"pseudonym"`
	} `json:"user"`
	Side            Side   `json:"side"`
	Size            string `json:"size"`
	FeeRateBps      string `json:"fee_rate_bps"`
	Price           string `json:"price"`
	Outcome         string `json:"outcome"`
	OutcomeIndex    int    `json:"outcome_index"`
	TransactionHash string `json:"transaction_hash"`
	Timestamp       string `json:"timestamp"`
}

// BuilderApiKey is the response shape of POST /auth/builder-api-key.
type BuilderApiKey struct {
	Key        string `json:"key"`
	Secret     string `json:"secret"`
	Passphrase string `json:"passphrase"`
}

// BuilderApiKeyResponse is the lifecycle response (list / revoke) for builder keys.
type BuilderApiKeyResponse struct {
	Key       string `json:"key"`
	CreatedAt string `json:"createdAt,omitempty"`
	RevokedAt string `json:"revokedAt,omitempty"`
}

// ClobErrorResponseBody is the canonical {error: "..."} body returned on failures.
type ClobErrorResponseBody struct {
	Error string `json:"error"`
}

// TradesPaginatedResponse is the result of paginated GET /data/trades.
type TradesPaginatedResponse struct {
	Trades     []Trade `json:"trades"`
	NextCursor string  `json:"next_cursor"`
	Limit      int     `json:"limit"`
	Count      int     `json:"count"`
}

// BuilderTradesResponse is the result of GET /builder/trades.
type BuilderTradesResponse struct {
	Trades     []BuilderTrade `json:"trades"`
	NextCursor string         `json:"next_cursor"`
	Limit      int            `json:"limit"`
	Count      int            `json:"count"`
}
