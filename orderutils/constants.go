package orderutils

import "github.com/ethereum/go-ethereum/signer/core/apitypes"

// CTF Exchange V2 EIP-712 domain constants. These are constant across all
// networks; only chainId and verifyingContract change per deployment.
const (
	CTFExchangeV2DomainName    = "Polymarket CTF Exchange"
	CTFExchangeV2DomainVersion = "2"
)

// Bytes32Zero is the canonical zero bytes32 hex string used as default for the
// V2 metadata and builder fields.
const Bytes32Zero = "0x0000000000000000000000000000000000000000000000000000000000000000"

// orderV2StructDef is the V2 Order EIP-712 struct definition in canonical order.
// NOTE: V2 excludes taker, expiration, nonce, and feeRateBps from the digest
// (those were V1 only) — expiration is still POSTed on the wire but does not
// participate in the signed hash.
var orderV2StructDef = []apitypes.Type{
	{Name: "salt", Type: "uint256"},
	{Name: "maker", Type: "address"},
	{Name: "signer", Type: "address"},
	{Name: "tokenId", Type: "uint256"},
	{Name: "makerAmount", Type: "uint256"},
	{Name: "takerAmount", Type: "uint256"},
	{Name: "side", Type: "uint8"},
	{Name: "signatureType", Type: "uint8"},
	{Name: "timestamp", Type: "uint256"},
	{Name: "metadata", Type: "bytes32"},
	{Name: "builder", Type: "bytes32"},
}

// eip712DomainDef is the standard EIP-712 domain struct for V2 orders.
var eip712DomainDef = []apitypes.Type{
	{Name: "name", Type: "string"},
	{Name: "version", Type: "string"},
	{Name: "chainId", Type: "uint256"},
	{Name: "verifyingContract", Type: "address"},
}
