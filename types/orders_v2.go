package types

// OrderDataV2 is the user-facing input bundle that builders turn into a signed V2 order.
// Optional fields default to a zero / sensible value if empty (see order builder).
type OrderDataV2 struct {
	Maker         string          // source of funds
	TokenID       string          // CTF ERC1155 token id
	MakerAmount   string          // max amount of maker token to spend (raw, 6 decimals)
	TakerAmount   string          // min amount of taker token to receive (raw, 6 decimals)
	Side          Side            // BUY / SELL
	Signer        string          // signer address; defaults to Maker if empty
	SignatureType SignatureTypeV2 // defaults to EOA
	Timestamp     string          // unix ms, set by builder if empty
	Metadata      string          // bytes32 hex, defaults to zero
	Builder       string          // bytes32 builder code, defaults to zero
	Expiration    string          // unix seconds, "0" = no expiration
}

// OrderV2 is the canonical EIP-712 V2 order struct used for hashing/signing.
// All numeric fields are stringified to avoid loss of precision over JSON.
type OrderV2 struct {
	Salt          string          `json:"salt"`
	Maker         string          `json:"maker"`
	Signer        string          `json:"signer"`
	TokenID       string          `json:"tokenId"`
	MakerAmount   string          `json:"makerAmount"`
	TakerAmount   string          `json:"takerAmount"`
	Side          Side            `json:"side"`
	SignatureType SignatureTypeV2 `json:"signatureType"`
	Timestamp     string          `json:"timestamp"`
	Metadata      string          `json:"metadata"`
	Builder       string          `json:"builder"`
	Expiration    string          `json:"expiration"`
}

// SignedOrderV2 is an OrderV2 paired with its signature.
type SignedOrderV2 struct {
	OrderV2
	Signature string `json:"signature"`
}

// UserOrderV2 is the high-level limit-order input exposed to library users.
type UserOrderV2 struct {
	TokenID     string  // CTF token id
	Price       float64 // order price
	Size        float64 // size in CTF shares
	Side        Side
	Metadata    string // bytes32 hex (optional)
	BuilderCode string // bytes32 hex (optional)
	Expiration  int64  // unix seconds, 0 = no expiration
}

// UserMarketOrderV2 is the high-level market-order input exposed to library users.
type UserMarketOrderV2 struct {
	TokenID         string
	Price           float64   // optional; if 0 the market price is used
	Amount          float64   // BUY: USDC to spend, SELL: shares to sell
	Side            Side
	OrderType       OrderType // FOK or FAK
	UserUSDCBalance float64   // optional; affects fee deduction strategy
	Metadata        string    // bytes32 hex (optional)
	BuilderCode     string    // bytes32 hex (optional)
}

// NewOrderV2Body is the JSON body POSTed when placing a V2 order.
type NewOrderV2Body struct {
	DeferExec bool         `json:"deferExec"`
	PostOnly  bool         `json:"postOnly"`
	Order     OrderV2Wire  `json:"order"`
	Owner     string       `json:"owner"`
	OrderType OrderType    `json:"orderType"`
}

// OrderV2Wire is the wire-format V2 order (salt as int, taker preserved for
// backward compatibility, signature included).
type OrderV2Wire struct {
	Salt          int             `json:"salt"`
	Maker         string          `json:"maker"`
	Signer        string          `json:"signer"`
	Taker         string          `json:"taker"`
	TokenID       string          `json:"tokenId"`
	MakerAmount   string          `json:"makerAmount"`
	TakerAmount   string          `json:"takerAmount"`
	Side          Side            `json:"side"`
	SignatureType SignatureTypeV2 `json:"signatureType"`
	Timestamp     string          `json:"timestamp"`
	Expiration    string          `json:"expiration"`
	Metadata      string          `json:"metadata"`
	Builder       string          `json:"builder"`
	Signature     string          `json:"signature"`
}
