package signing

import (
	"crypto/ecdsa"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

// Signer is the abstraction the CLOB client uses for L1 (EIP-712) signing.
// Implementations cover plain EOAs as well as remote signers (HSM, hardware
// wallets, etc.).
type Signer interface {
	// Address returns the EIP-55 checksummed address controlled by this signer.
	Address() string

	// SignTypedData hashes the given typed data per EIP-712 and produces a
	// 0x-prefixed 65-byte signature with the recovery byte V in {27, 28}.
	SignTypedData(td *apitypes.TypedData) (string, error)
}

// PrivateKeySigner is a Signer backed by a raw secp256k1 private key.
type PrivateKeySigner struct {
	key     *ecdsa.PrivateKey
	address string
}

// NewPrivateKeySigner parses a hex-encoded private key (with or without the
// 0x prefix) and returns a Signer that signs in-process. The returned address
// is EIP-55 checksummed.
func NewPrivateKeySigner(hexKey string) (*PrivateKeySigner, error) {
	hexKey = strings.TrimPrefix(hexKey, "0x")
	key, err := crypto.HexToECDSA(hexKey)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}
	addr := crypto.PubkeyToAddress(key.PublicKey).Hex()
	return &PrivateKeySigner{key: key, address: addr}, nil
}

// Address returns the EIP-55 checksummed address of the underlying key.
func (s *PrivateKeySigner) Address() string {
	return s.address
}

// SignTypedData signs the EIP-712 digest of td and returns a hex signature.
func (s *PrivateKeySigner) SignTypedData(td *apitypes.TypedData) (string, error) {
	hash, _, err := apitypes.TypedDataAndHash(*td)
	if err != nil {
		return "", fmt.Errorf("hash typed data: %w", err)
	}
	sig, err := crypto.Sign(hash, s.key)
	if err != nil {
		return "", fmt.Errorf("sign typed data: %w", err)
	}
	// go-ethereum's crypto.Sign returns V in {0,1}; the Ethereum wire format
	// (and what viem/ethers emit) expects V in {27,28}.
	sig[64] += 27
	return hexutil.Encode(sig), nil
}
