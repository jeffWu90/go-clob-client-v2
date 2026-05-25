package types

// SignatureTypeV2 enumerates the signature schemes the V2 CTF Exchange accepts.
type SignatureTypeV2 uint8

const (
	// SignatureTypeV2EOA is an ECDSA EIP-712 signature from an externally owned account.
	SignatureTypeV2EOA SignatureTypeV2 = 0
	// SignatureTypeV2PolyProxy is an EIP-712 signature from an EOA that owns a Polymarket proxy wallet.
	SignatureTypeV2PolyProxy SignatureTypeV2 = 1
	// SignatureTypeV2PolyGnosisSafe is an EIP-712 signature from an EOA that owns a Polymarket Gnosis Safe.
	SignatureTypeV2PolyGnosisSafe SignatureTypeV2 = 2
	// SignatureTypeV2Poly1271 is an EIP-1271 signature from a smart-contract wallet.
	SignatureTypeV2Poly1271 SignatureTypeV2 = 3
)
