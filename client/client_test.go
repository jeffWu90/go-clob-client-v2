package client

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	clob "github.com/jeffWu90/go-clob-client-v2"
	"github.com/jeffWu90/go-clob-client-v2/signing"
	"github.com/jeffWu90/go-clob-client-v2/types"
)

const (
	testPrivateKey = "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"
	testWalletAddr = "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"
)

func newTestSigner(t *testing.T) signing.Signer {
	t.Helper()
	s, err := signing.NewPrivateKeySigner(testPrivateKey)
	if err != nil {
		t.Fatalf("signer: %v", err)
	}
	return s
}

func newTestClient(t *testing.T, host string, withCreds bool) *ClobClient {
	t.Helper()
	opts := Options{
		Host:   host,
		Chain:  types.ChainPolygon,
		Signer: newTestSigner(t),
	}
	if withCreds {
		opts.Creds = &types.ApiKeyCreds{Key: "k", Secret: "c2VjcmV0", Passphrase: "p"}
	}
	c, err := New(opts)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return c
}

// -----------------------------------------------------------------------------
// Public endpoints (no auth)
// -----------------------------------------------------------------------------

func TestGetServerTime(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != clob.EndpointTime {
			t.Errorf("path = %s, want %s", r.URL.Path, clob.EndpointTime)
		}
		w.Write([]byte("1700000000"))
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL, false)
	ts, err := c.GetServerTime(context.Background())
	if err != nil {
		t.Fatalf("GetServerTime: %v", err)
	}
	if ts != 1700000000 {
		t.Errorf("got %d, want 1700000000", ts)
	}
}

func TestGetOrderBook(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != clob.EndpointGetOrderBook {
			t.Errorf("path = %s", r.URL.Path)
		}
		if tok := r.URL.Query().Get("token_id"); tok != "abc" {
			t.Errorf("token_id = %s", tok)
		}
		w.Write([]byte(`{"market":"m","asset_id":"abc","bids":[{"price":"0.4","size":"100"}],"asks":[{"price":"0.5","size":"50"}],"tick_size":"0.01"}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL, false)
	ob, err := c.GetOrderBook(context.Background(), "abc")
	if err != nil {
		t.Fatalf("GetOrderBook: %v", err)
	}
	if ob.AssetID != "abc" {
		t.Errorf("asset_id = %s", ob.AssetID)
	}
	if len(ob.Bids) != 1 || ob.Bids[0].Price != "0.4" {
		t.Errorf("bids = %+v", ob.Bids)
	}
}

// -----------------------------------------------------------------------------
// Error handling
// -----------------------------------------------------------------------------

func TestApiErrorMapping(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"No orderbook exists for the requested token id"}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL, false)
	_, err := c.GetOrderBook(context.Background(), "missing")
	if err == nil {
		t.Fatalf("expected error")
	}
	var apiErr *clob.ApiError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *clob.ApiError, got %T", err)
	}
	if apiErr.Status != http.StatusNotFound {
		t.Errorf("status = %d", apiErr.Status)
	}
	if !strings.Contains(apiErr.Message, "No orderbook exists") {
		t.Errorf("message = %q", apiErr.Message)
	}
}

// -----------------------------------------------------------------------------
// L1 auth — header construction
// -----------------------------------------------------------------------------

func TestCreateApiKey_SendsL1Headers(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != clob.EndpointCreateApiKey {
			t.Errorf("%s %s", r.Method, r.URL.Path)
		}
		// L1 headers are POLY_ADDRESS / POLY_SIGNATURE / POLY_TIMESTAMP / POLY_NONCE.
		for _, h := range []string{"POLY_ADDRESS", "POLY_SIGNATURE", "POLY_TIMESTAMP", "POLY_NONCE"} {
			if r.Header.Get(h) == "" {
				t.Errorf("missing header %s", h)
			}
		}
		if addr := r.Header.Get("POLY_ADDRESS"); !strings.EqualFold(addr, testWalletAddr) {
			t.Errorf("POLY_ADDRESS = %s", addr)
		}
		w.Write([]byte(`{"apiKey":"newkey","secret":"newsecret","passphrase":"newphrase"}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL, false)
	creds, err := c.CreateApiKey(context.Background(), 0)
	if err != nil {
		t.Fatalf("CreateApiKey: %v", err)
	}
	if creds.Key != "newkey" || creds.Secret != "newsecret" || creds.Passphrase != "newphrase" {
		t.Errorf("creds = %+v", creds)
	}
}

// -----------------------------------------------------------------------------
// L2 auth — header construction + order POST shape
// -----------------------------------------------------------------------------

func TestPostOrder_SendsL2HeadersAndV2WireShape(t *testing.T) {
	var seenBody types.NewOrderV2Body
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != clob.EndpointPostOrder {
			t.Errorf("%s %s", r.Method, r.URL.Path)
		}
		// L2 headers expected: 5 of them.
		for _, h := range []string{"POLY_ADDRESS", "POLY_SIGNATURE", "POLY_TIMESTAMP", "POLY_API_KEY", "POLY_PASSPHRASE"} {
			if r.Header.Get(h) == "" {
				t.Errorf("missing header %s", h)
			}
		}
		if k := r.Header.Get("POLY_API_KEY"); k != "k" {
			t.Errorf("POLY_API_KEY = %s", k)
		}

		body, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(body, &seenBody); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.Write([]byte(`{"success":true,"orderID":"oid-123","status":"matched"}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL, true)

	signed := &types.SignedOrderV2{
		OrderV2: types.OrderV2{
			Salt:          "42",
			Maker:         testWalletAddr,
			Signer:        testWalletAddr,
			TokenID:       "1234",
			MakerAmount:   "40000000",
			TakerAmount:   "100000000",
			Side:          types.SideBuy,
			SignatureType: types.SignatureTypeV2EOA,
			Timestamp:     "1700000000000",
			Metadata:      clob.Bytes32Zero,
			Builder:       clob.Bytes32Zero,
			Expiration:    "0",
		},
		Signature: "0xabc",
	}

	resp, err := c.PostOrder(context.Background(), signed, types.OrderTypeGTC, false, false)
	if err != nil {
		t.Fatalf("PostOrder: %v", err)
	}
	if !resp.Success || resp.OrderID != "oid-123" {
		t.Errorf("resp = %+v", resp)
	}

	// Wire shape: salt is a JSON number, taker is the zero address, side is "BUY",
	// owner is the API key.
	if seenBody.Order.Salt != 42 {
		t.Errorf("salt on the wire = %d, want 42", seenBody.Order.Salt)
	}
	if seenBody.Order.Taker != "0x0000000000000000000000000000000000000000" {
		t.Errorf("taker = %s", seenBody.Order.Taker)
	}
	if seenBody.Order.Side != types.SideBuy {
		t.Errorf("side = %s", seenBody.Order.Side)
	}
	if seenBody.Owner != "k" {
		t.Errorf("owner = %s, want \"k\" (api key)", seenBody.Owner)
	}
	if seenBody.OrderType != types.OrderTypeGTC {
		t.Errorf("orderType = %s", seenBody.OrderType)
	}
}

func TestPostOrder_RejectsPostOnlyForFOK(t *testing.T) {
	c := newTestClient(t, "http://unused", true)
	_, err := c.PostOrder(context.Background(), &types.SignedOrderV2{}, types.OrderTypeFOK, true, false)
	if err == nil {
		t.Errorf("expected error when postOnly is true on FOK")
	}
}

// -----------------------------------------------------------------------------
// Retry behavior
// -----------------------------------------------------------------------------

func TestRetryOnError_OnceForTransient(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := hits.Add(1)
		if h == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"transient"}`))
			return
		}
		w.Write([]byte(`{"success":true,"orderID":"ok","status":"matched"}`))
	}))
	defer srv.Close()

	c, err := New(Options{
		Host:         srv.URL,
		Chain:        types.ChainPolygon,
		Signer:       newTestSigner(t),
		Creds:        &types.ApiKeyCreds{Key: "k", Secret: "c2VjcmV0", Passphrase: "p"},
		RetryOnError: true,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	resp, err := c.PostOrder(context.Background(), &types.SignedOrderV2{
		OrderV2: types.OrderV2{Salt: "1", Side: types.SideBuy, SignatureType: types.SignatureTypeV2EOA, Timestamp: "1"},
	}, types.OrderTypeGTC, false, false)
	if err != nil {
		t.Fatalf("PostOrder: %v", err)
	}
	if resp.OrderID != "ok" {
		t.Errorf("orderID = %s", resp.OrderID)
	}
	if got := hits.Load(); got != 2 {
		t.Errorf("expected 2 hits (1 fail + 1 retry), got %d", got)
	}
}

// -----------------------------------------------------------------------------
// Auth gating
// -----------------------------------------------------------------------------

func TestL1AuthGating(t *testing.T) {
	c, _ := New(Options{Host: "http://unused", Chain: types.ChainPolygon})
	if _, err := c.CreateApiKey(context.Background(), 0); !errors.Is(err, clob.ErrL1AuthUnavailable) {
		t.Errorf("expected ErrL1AuthUnavailable, got %v", err)
	}
}

func TestL2AuthGating(t *testing.T) {
	c, _ := New(Options{Host: "http://unused", Chain: types.ChainPolygon, Signer: newTestSigner(t)})
	if _, err := c.PostOrder(context.Background(), &types.SignedOrderV2{}, types.OrderTypeGTC, false, false); !errors.Is(err, clob.ErrL2AuthUnavailable) {
		t.Errorf("expected ErrL2AuthUnavailable, got %v", err)
	}
}

// -----------------------------------------------------------------------------
// Cancel
// -----------------------------------------------------------------------------

func TestCancelOrder_SendsDeleteWithBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != clob.EndpointCancelOrder {
			t.Errorf("%s %s", r.Method, r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		var payload types.OrderPayload
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if payload.OrderID != "oid-1" {
			t.Errorf("orderID = %s", payload.OrderID)
		}
		w.Write([]byte(`{"canceled":["oid-1"]}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL, true)
	if _, err := c.CancelOrder(context.Background(), types.OrderPayload{OrderID: "oid-1"}); err != nil {
		t.Fatalf("CancelOrder: %v", err)
	}
}

// -----------------------------------------------------------------------------
// Tick size caching
// -----------------------------------------------------------------------------

func TestGetTickSize_CachesResult(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.Write([]byte(`{"minimum_tick_size":0.01}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL, false)
	for i := 0; i < 3; i++ {
		ts, err := c.GetTickSize(context.Background(), "tok-1")
		if err != nil {
			t.Fatalf("GetTickSize: %v", err)
		}
		if ts != types.TickSize01 {
			t.Errorf("tick = %s, want 0.01", ts)
		}
	}
	if got := hits.Load(); got != 1 {
		t.Errorf("expected single network hit due to cache, got %d", got)
	}
}
