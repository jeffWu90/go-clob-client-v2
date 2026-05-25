package signing

import (
	"strings"
	"testing"
)

const (
	testPrivateKey  = "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"
	testWalletAddr  = "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"
)

func TestNewPrivateKeySigner_DerivesAddress(t *testing.T) {
	cases := []string{
		testPrivateKey,
		"0x" + testPrivateKey,
	}
	for _, hex := range cases {
		s, err := NewPrivateKeySigner(hex)
		if err != nil {
			t.Fatalf("NewPrivateKeySigner(%q): %v", hex, err)
		}
		if !strings.EqualFold(s.Address(), testWalletAddr) {
			t.Errorf("address mismatch:\n  got  %q\n  want %q", s.Address(), testWalletAddr)
		}
	}
}

func TestNewPrivateKeySigner_RejectsBadHex(t *testing.T) {
	if _, err := NewPrivateKeySigner("not-hex"); err == nil {
		t.Errorf("expected error for invalid private key")
	}
}
