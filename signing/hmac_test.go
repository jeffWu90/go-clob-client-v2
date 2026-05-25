package signing

import "testing"

func TestBuildPolyHmacSignature(t *testing.T) {
	// Test vector ported verbatim from TS clob-client-v2 tests/signing/hmac.test.ts.
	sig, err := BuildPolyHmacSignature(
		"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
		1000000,
		"test-sign",
		"/orders",
		`{"hash": "0x123"}`,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "ZwAdJKvoYRlEKDkNMwd5BuwNNtg93kNaR_oU2HrfVvc="
	if sig != want {
		t.Errorf("signature mismatch:\n  got  %q\n  want %q", sig, want)
	}
}

func TestBuildPolyHmacSignature_EmptyBody(t *testing.T) {
	// Empty body should be a no-op append, not a literal "" appended.
	sig, err := BuildPolyHmacSignature(
		"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
		1000000,
		"GET",
		"/markets",
		"",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sig == "" {
		t.Errorf("expected non-empty signature")
	}
}
