package signing

import (
	"testing"

	"github.com/jeffWu90/go-clob-client-v2/types"
)

func TestBuildClobEip712Signature(t *testing.T) {
	// Test vector ported verbatim from TS clob-client-v2 tests/signing/eip712.test.ts.
	// Same input must yield the same signature across ethers / viem / go-ethereum,
	// because the EIP-712 digest is encoding-independent.
	signer, err := NewPrivateKeySigner(testPrivateKey)
	if err != nil {
		t.Fatalf("NewPrivateKeySigner: %v", err)
	}

	sig, err := BuildClobEip712Signature(signer, types.ChainAmoy, 10000000, 23)
	if err != nil {
		t.Fatalf("BuildClobEip712Signature: %v", err)
	}

	want := "0xf62319a987514da40e57e2f4d7529f7bac38f0355bd88bb5adbb3768d80de6c1682518e0af677d5260366425f4361e7b70c25ae232aff0ab2331e2b164a1aedc1b"
	if sig != want {
		t.Errorf("signature mismatch:\n  got  %q\n  want %q", sig, want)
	}
}

func TestBuildClobEip712Signature_NilSigner(t *testing.T) {
	if _, err := BuildClobEip712Signature(nil, types.ChainAmoy, 10000000, 23); err == nil {
		t.Errorf("expected error when signer is nil")
	}
}
