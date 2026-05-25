package orderbuilder

import (
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"

	"github.com/jeffWu90/go-clob-client-v2/orderutils"
	"github.com/jeffWu90/go-clob-client-v2/signing"
	"github.com/jeffWu90/go-clob-client-v2/types"
)

const (
	testPrivateKey = "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"
	testWalletAddr = "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"
)

func newSigner(t *testing.T) signing.Signer {
	t.Helper()
	s, err := signing.NewPrivateKeySigner(testPrivateKey)
	if err != nil {
		t.Fatalf("signer: %v", err)
	}
	return s
}

func TestCreateOrder_LimitBuy_RecoversToSigner(t *testing.T) {
	signer := newSigner(t)
	so, err := CreateOrder(
		signer,
		types.ChainPolygon,
		types.SignatureTypeV2EOA,
		"", // maker defaults to signer
		types.UserOrderV2{
			TokenID: "1234",
			Price:   0.4,
			Size:    100,
			Side:    types.SideBuy,
		},
		types.CreateOrderOptions{TickSize: types.TickSize01},
	)
	if err != nil {
		t.Fatalf("CreateOrder: %v", err)
	}

	// 100 shares @ 0.40 => maker = 40 USDC, taker = 100 shares.
	if so.MakerAmount != "40000000" {
		t.Errorf("makerAmount = %s, want 40000000", so.MakerAmount)
	}
	if so.TakerAmount != "100000000" {
		t.Errorf("takerAmount = %s, want 100000000", so.TakerAmount)
	}
	if so.Side != types.SideBuy {
		t.Errorf("side = %s, want BUY", so.Side)
	}
	if !strings.EqualFold(so.Maker, testWalletAddr) {
		t.Errorf("maker = %s, want %s", so.Maker, testWalletAddr)
	}
	if !strings.EqualFold(so.Signer, testWalletAddr) {
		t.Errorf("signer = %s, want %s", so.Signer, testWalletAddr)
	}

	requireSignatureRecovers(t, signer, so)
}

func TestCreateOrder_LimitSell_RecoversToSigner(t *testing.T) {
	signer := newSigner(t)
	so, err := CreateOrder(
		signer,
		types.ChainPolygon,
		types.SignatureTypeV2EOA,
		"",
		types.UserOrderV2{
			TokenID: "9999",
			Price:   0.6,
			Size:    50,
			Side:    types.SideSell,
		},
		types.CreateOrderOptions{TickSize: types.TickSize01},
	)
	if err != nil {
		t.Fatalf("CreateOrder: %v", err)
	}
	if so.MakerAmount != "50000000" {
		t.Errorf("makerAmount = %s, want 50000000", so.MakerAmount)
	}
	if so.TakerAmount != "30000000" {
		t.Errorf("takerAmount = %s, want 30000000", so.TakerAmount)
	}
	requireSignatureRecovers(t, signer, so)
}

func TestCreateMarketOrder_Buy_RecoversToSigner(t *testing.T) {
	signer := newSigner(t)
	so, err := CreateMarketOrder(
		signer,
		types.ChainPolygon,
		types.SignatureTypeV2EOA,
		"",
		types.UserMarketOrderV2{
			TokenID:   "1234",
			Price:     0.5,
			Amount:    100, // 100 USDC
			Side:      types.SideBuy,
			OrderType: types.OrderTypeFOK,
		},
		types.CreateOrderOptions{TickSize: types.TickSize01},
	)
	if err != nil {
		t.Fatalf("CreateMarketOrder: %v", err)
	}
	// 100 USDC at price 0.5 => 200 shares
	if so.MakerAmount != "100000000" {
		t.Errorf("makerAmount = %s, want 100000000", so.MakerAmount)
	}
	if so.TakerAmount != "200000000" {
		t.Errorf("takerAmount = %s, want 200000000", so.TakerAmount)
	}
	requireSignatureRecovers(t, signer, so)
}

func TestCreateOrder_RejectsUnknownTickSize(t *testing.T) {
	signer := newSigner(t)
	_, err := CreateOrder(
		signer,
		types.ChainPolygon,
		types.SignatureTypeV2EOA,
		"",
		types.UserOrderV2{TokenID: "1", Price: 0.5, Size: 1, Side: types.SideBuy},
		types.CreateOrderOptions{TickSize: types.TickSize("0.5")},
	)
	if err == nil {
		t.Errorf("expected error for unknown tick size")
	}
}

// requireSignatureRecovers signs the order's typed data, recovers the public
// key from the signature, and confirms the recovered address matches the signer.
func requireSignatureRecovers(t *testing.T, s signing.Signer, so *types.SignedOrderV2) {
	t.Helper()

	// Build the typed data using the same domain the signer used: V2 Exchange on Polygon.
	b := orderutils.NewExchangeOrderBuilderV2("0xE111180000d2663C0091e4f400237545B87B996B", types.ChainPolygon, s)
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
		t.Fatalf("expected 65-byte sig, got %d", len(sigBytes))
	}
	sigBytes[64] -= 27

	pub, err := crypto.SigToPub(digest, sigBytes)
	if err != nil {
		t.Fatalf("SigToPub: %v", err)
	}
	got := crypto.PubkeyToAddress(*pub)
	want := common.HexToAddress(s.Address())
	if got != want {
		t.Errorf("recovered %s, want %s", got.Hex(), want.Hex())
	}
}
