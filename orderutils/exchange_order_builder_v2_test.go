package orderutils

import (
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"

	"github.com/jeffWu90/go-clob-client-v2/signing"
	"github.com/jeffWu90/go-clob-client-v2/types"
)

const (
	testPrivateKey   = "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"
	testWalletAddr   = "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"
	testExchangeAddr = "0xE111180000d2663C0091e4f400237545B87B996B"
)

func mustSigner(t *testing.T) signing.Signer {
	t.Helper()
	s, err := signing.NewPrivateKeySigner(testPrivateKey)
	if err != nil {
		t.Fatalf("signer: %v", err)
	}
	return s
}

func newDeterministicBuilder(t *testing.T, salt string) *ExchangeOrderBuilderV2 {
	b := NewExchangeOrderBuilderV2(testExchangeAddr, types.ChainPolygon, mustSigner(t))
	b.SaltFn = func() string { return salt }
	return b
}

func sampleOrderData() types.OrderDataV2 {
	return types.OrderDataV2{
		Maker:       testWalletAddr,
		TokenID:     "1234",
		MakerAmount: "1000000",
		TakerAmount: "500000",
		Side:        types.SideBuy,
		Timestamp:   "1735000000000",
	}
}

func TestBuildOrder_FillsDefaults(t *testing.T) {
	b := newDeterministicBuilder(t, "479249096354")

	o, err := b.BuildOrder(sampleOrderData())
	if err != nil {
		t.Fatalf("BuildOrder: %v", err)
	}

	if o.Salt != "479249096354" {
		t.Errorf("salt = %q, want injected value", o.Salt)
	}
	if !strings.EqualFold(o.Signer, testWalletAddr) {
		t.Errorf("signer should default to maker: %q", o.Signer)
	}
	if o.SignatureType != types.SignatureTypeV2EOA {
		t.Errorf("signatureType should default to EOA, got %d", o.SignatureType)
	}
	if o.Metadata != Bytes32Zero {
		t.Errorf("metadata should default to bytes32 zero, got %q", o.Metadata)
	}
	if o.Builder != Bytes32Zero {
		t.Errorf("builder should default to bytes32 zero, got %q", o.Builder)
	}
	if o.Expiration != "0" {
		t.Errorf("expiration should default to %q, got %q", "0", o.Expiration)
	}
}

func TestBuildOrder_RejectsMismatchedSigner(t *testing.T) {
	b := newDeterministicBuilder(t, "1")

	d := sampleOrderData()
	d.Signer = "0x0000000000000000000000000000000000000001"
	if _, err := b.BuildOrder(d); err == nil {
		t.Errorf("expected error when signer != wallet address")
	}
}

func TestBuildOrder_RejectsMissingMaker(t *testing.T) {
	b := newDeterministicBuilder(t, "1")
	d := sampleOrderData()
	d.Maker = ""
	if _, err := b.BuildOrder(d); err == nil {
		t.Errorf("expected error when maker is empty")
	}
}

func TestBuildOrderTypedData_Shape(t *testing.T) {
	b := newDeterministicBuilder(t, "479249096354")
	o, err := b.BuildOrder(sampleOrderData())
	if err != nil {
		t.Fatalf("BuildOrder: %v", err)
	}
	td, err := b.BuildOrderTypedData(o)
	if err != nil {
		t.Fatalf("BuildOrderTypedData: %v", err)
	}

	if td.PrimaryType != "Order" {
		t.Errorf("primaryType = %q", td.PrimaryType)
	}
	if td.Domain.Name != "Polymarket CTF Exchange" {
		t.Errorf("domain.name = %q", td.Domain.Name)
	}
	if td.Domain.Version != "2" {
		t.Errorf("domain.version = %q", td.Domain.Version)
	}
	if td.Domain.VerifyingContract != testExchangeAddr {
		t.Errorf("domain.verifyingContract = %q", td.Domain.VerifyingContract)
	}
	if len(td.Types["Order"]) != 11 {
		t.Errorf("Order struct should have 11 fields, got %d", len(td.Types["Order"]))
	}
}

func TestBuildOrderHash_DeterministicAndSaltSensitive(t *testing.T) {
	b1 := newDeterministicBuilder(t, "1")
	b2 := newDeterministicBuilder(t, "2")

	mk := func(b *ExchangeOrderBuilderV2) string {
		o, err := b.BuildOrder(sampleOrderData())
		if err != nil {
			t.Fatalf("BuildOrder: %v", err)
		}
		td, err := b.BuildOrderTypedData(o)
		if err != nil {
			t.Fatalf("BuildOrderTypedData: %v", err)
		}
		h, err := b.BuildOrderHash(td)
		if err != nil {
			t.Fatalf("BuildOrderHash: %v", err)
		}
		return h
	}

	h1a := mk(b1)
	h1b := mk(b1)
	h2 := mk(b2)

	if h1a != h1b {
		t.Errorf("hash should be deterministic for same inputs: %s vs %s", h1a, h1b)
	}
	if h1a == h2 {
		t.Errorf("hash should change with salt, but both produced %s", h1a)
	}
}

func TestBuildSignedOrder_RecoversToSigner(t *testing.T) {
	b := newDeterministicBuilder(t, "42")
	so, err := b.BuildSignedOrder(sampleOrderData())
	if err != nil {
		t.Fatalf("BuildSignedOrder: %v", err)
	}
	if so.Signature == "" {
		t.Fatalf("expected non-empty signature")
	}

	// Recompute the digest and recover the signer address from the signature.
	td, err := b.BuildOrderTypedData(&so.OrderV2)
	if err != nil {
		t.Fatalf("BuildOrderTypedData: %v", err)
	}
	digest, _, err := apitypes.TypedDataAndHash(*td)
	if err != nil {
		t.Fatalf("TypedDataAndHash: %v", err)
	}

	sigBytes, err := hexutil.Decode(so.Signature)
	if err != nil {
		t.Fatalf("decode signature: %v", err)
	}
	if len(sigBytes) != 65 {
		t.Fatalf("expected 65-byte signature, got %d", len(sigBytes))
	}
	// crypto.Ecrecover expects V in {0,1}; we encoded V in {27,28}.
	sigBytes[64] -= 27

	pubKey, err := crypto.SigToPub(digest, sigBytes)
	if err != nil {
		t.Fatalf("SigToPub: %v", err)
	}
	recovered := crypto.PubkeyToAddress(*pubKey)
	want := common.HexToAddress(testWalletAddr)
	if recovered != want {
		t.Errorf("recovered %s, want %s", recovered.Hex(), want.Hex())
	}
}
