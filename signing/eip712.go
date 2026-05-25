package signing

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"

	"github.com/jeffWu90/go-clob-client-v2/types"
)

// BuildClobEip712Signature signs the canonical Polymarket ClobAuth typed
// payload used to authenticate L1 (wallet) requests.
//
// The address is taken from the signer, lowercased before being placed in the
// EIP-712 message (matching what the CLOB server expects in POLY_ADDRESS).
// The signature is independent of the address letter casing because EIP-712
// encodes addresses as 20 raw bytes.
func BuildClobEip712Signature(s Signer, chainID types.Chain, timestamp, nonce int64) (string, error) {
	if s == nil {
		return "", fmt.Errorf("signer is required")
	}

	address := strings.ToLower(s.Address())

	td := &apitypes.TypedData{
		Types: apitypes.Types{
			"EIP712Domain": {
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
			},
			"ClobAuth": {
				{Name: "address", Type: "address"},
				{Name: "timestamp", Type: "string"},
				{Name: "nonce", Type: "uint256"},
				{Name: "message", Type: "string"},
			},
		},
		PrimaryType: "ClobAuth",
		Domain: apitypes.TypedDataDomain{
			Name:    ClobDomainName,
			Version: ClobVersion,
			ChainId: (*math.HexOrDecimal256)(big.NewInt(int64(chainID))),
		},
		Message: apitypes.TypedDataMessage{
			"address":   address,
			"timestamp": strconv.FormatInt(timestamp, 10),
			"nonce":     strconv.FormatInt(nonce, 10),
			"message":   MsgToSign,
		},
	}
	return s.SignTypedData(td)
}
