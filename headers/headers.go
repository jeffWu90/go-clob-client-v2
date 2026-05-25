package headers

import (
	"strconv"

	"github.com/jeffWu90/go-clob-client-v2/signing"
	"github.com/jeffWu90/go-clob-client-v2/types"
)

// L1AuthHeaders carries the four POLY_* headers required for L1 (wallet) auth.
// Keys are the literal HTTP header names expected by the CLOB server.
type L1AuthHeaders = map[string]string

// L2AuthHeaders carries the five POLY_* headers required for L2 (API key) auth.
type L2AuthHeaders = map[string]string

// CreateL1Headers signs a ClobAuth EIP-712 message with the given signer and
// returns the four headers the server expects on L1 endpoints.
//
// If timestamp is zero a fresh time.Now().Unix() is used; otherwise the value
// is treated as a unix-seconds timestamp (typically taken from /time).
// nonce defaults to 0 — the CLOB uses it as the API-key index so multiple
// keys can be derived from the same wallet.
func CreateL1Headers(s signing.Signer, chainID types.Chain, nonce, timestamp int64) (L1AuthHeaders, error) {
	ts := timestamp
	if ts == 0 {
		ts = nowUnix()
	}
	sig, err := signing.BuildClobEip712Signature(s, chainID, ts, nonce)
	if err != nil {
		return nil, err
	}
	return L1AuthHeaders{
		"POLY_ADDRESS":   s.Address(),
		"POLY_SIGNATURE": sig,
		"POLY_TIMESTAMP": strconv.FormatInt(ts, 10),
		"POLY_NONCE":     strconv.FormatInt(nonce, 10),
	}, nil
}

// CreateL2Headers signs an HMAC payload with the API secret and returns the
// five headers the server expects on L2 endpoints. The signer is needed only
// to populate POLY_ADDRESS (the HMAC itself doesn't touch the wallet).
func CreateL2Headers(s signing.Signer, creds types.ApiKeyCreds, args types.L2HeaderArgs, timestamp int64) (L2AuthHeaders, error) {
	ts := timestamp
	if ts == 0 {
		ts = nowUnix()
	}
	sig, err := signing.BuildPolyHmacSignature(creds.Secret, ts, args.Method, args.RequestPath, args.Body)
	if err != nil {
		return nil, err
	}
	return L2AuthHeaders{
		"POLY_ADDRESS":    s.Address(),
		"POLY_SIGNATURE":  sig,
		"POLY_TIMESTAMP":  strconv.FormatInt(ts, 10),
		"POLY_API_KEY":    creds.Key,
		"POLY_PASSPHRASE": creds.Passphrase,
	}, nil
}
