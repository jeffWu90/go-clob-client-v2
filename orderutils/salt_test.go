package orderutils

import "testing"

func TestGenerateOrderSalt_NotEmpty(t *testing.T) {
	s := GenerateOrderSalt()
	if s == "" {
		t.Errorf("expected non-empty salt")
	}
}

func TestGenerateOrderSalt_Unique(t *testing.T) {
	// Mirrors the TS test which loops 100x and demands all unique.
	const n = 100
	seen := make(map[string]struct{}, n)
	for i := 0; i < n; i++ {
		s := GenerateOrderSalt()
		if _, dup := seen[s]; dup {
			t.Fatalf("duplicate salt at iteration %d: %s", i, s)
		}
		seen[s] = struct{}{}
	}
}
