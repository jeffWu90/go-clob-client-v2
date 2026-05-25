package clob

// CLOB API endpoint paths. Keep in sync with the upstream server routes.
const (
	// Health
	EndpointOK        = "/ok"
	EndpointHeartbeat = "/v1/heartbeats"

	// Server time
	EndpointTime = "/time"

	// API key lifecycle
	EndpointCreateApiKey = "/auth/api-key"
	EndpointGetApiKeys   = "/auth/api-keys"
	EndpointDeleteApiKey = "/auth/api-key"
	EndpointDeriveApiKey = "/auth/derive-api-key"
	EndpointClosedOnly   = "/auth/ban-status/closed-only"

	// Markets
	EndpointGetSamplingSimplifiedMarkets = "/sampling-simplified-markets"
	EndpointGetSamplingMarkets           = "/sampling-markets"
	EndpointGetSimplifiedMarkets         = "/simplified-markets"
	EndpointGetMarkets                   = "/markets"
	EndpointGetMarket                    = "/markets/"
	EndpointGetMarketByToken             = "/markets-by-token/"
	EndpointGetClobMarket                = "/clob-markets/"
	EndpointGetOrderBook                 = "/book"
	EndpointGetOrderBooks                = "/books"
	EndpointGetMidpoint                  = "/midpoint"
	EndpointGetMidpoints                 = "/midpoints"
	EndpointGetPrice                     = "/price"
	EndpointGetPrices                    = "/prices"
	EndpointGetSpread                    = "/spread"
	EndpointGetSpreads                   = "/spreads"
	EndpointGetLastTradePrice            = "/last-trade-price"
	EndpointGetLastTradesPrices          = "/last-trades-prices"
	EndpointGetTickSize                  = "/tick-size"
	EndpointGetNegRisk                   = "/neg-risk"
	EndpointGetFeeRate                   = "/fee-rate"

	// Orders
	EndpointPostOrder            = "/order"
	EndpointPostOrders           = "/orders"
	EndpointCancelOrder          = "/order"
	EndpointCancelOrders         = "/orders"
	EndpointGetOrder             = "/data/order/"
	EndpointCancelAll            = "/cancel-all"
	EndpointCancelMarketOrders   = "/cancel-market-orders"
	EndpointGetOpenOrders        = "/data/orders"
	EndpointGetPreMigrationOrders = "/data/pre-migration-orders"
	EndpointGetTrades            = "/data/trades"
	EndpointIsOrderScoring       = "/order-scoring"
	EndpointAreOrdersScoring     = "/orders-scoring"

	// Prices
	EndpointGetPricesHistory = "/prices-history"

	// Notifications
	EndpointGetNotifications  = "/notifications"
	EndpointDropNotifications = "/notifications"

	// Balance
	EndpointGetBalanceAllowance    = "/balance-allowance"
	EndpointUpdateBalanceAllowance = "/balance-allowance/update"

	// Rewards
	EndpointGetEarningsForUserForDay       = "/rewards/user"
	EndpointGetTotalEarningsForUserForDay  = "/rewards/user/total"
	EndpointGetLiquidityRewardPercentages  = "/rewards/user/percentages"
	EndpointGetRewardsMarketsCurrent       = "/rewards/markets/current"
	EndpointGetRewardsMarkets              = "/rewards/markets/"
	EndpointGetRewardsEarningsPercentages  = "/rewards/user/markets"

	// Readonly API keys
	EndpointCreateReadonlyApiKey = "/auth/readonly-api-key"
	EndpointGetReadonlyApiKeys   = "/auth/readonly-api-keys"
	EndpointDeleteReadonlyApiKey = "/auth/readonly-api-key"

	// Builder API keys
	EndpointCreateBuilderApiKey = "/auth/builder-api-key"
	EndpointGetBuilderApiKeys   = "/auth/builder-api-key"
	EndpointRevokeBuilderApiKey = "/auth/builder-api-key"

	// Live activity
	EndpointGetMarketTradesEvents = "/markets/live-activity/"

	// Builder
	EndpointGetBuilderTrades = "/builder/trades"

	// Fees
	EndpointGetBuilderFees = "/fees/builder-fees/"
)
