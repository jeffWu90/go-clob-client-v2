package signing

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
)

// BuildPolyHmacSignature builds the canonical Polymarket CLOB HMAC-SHA256
// signature used in L2 (API-key) headers.
//
// The signed payload is the concatenation: timestamp + method + requestPath + body
// (body is appended only when non-empty). The secret is base64-decoded before
// being used as the HMAC key. The resulting digest is encoded as standard base64
// with '+' / '/' rewritten to URL-safe '-' / '_'; padding ('=') is preserved.
func BuildPolyHmacSignature(secret string, timestamp int64, method, requestPath, body string) (string, error) {
	keyBytes, err := decodeBase64Flexible(secret)
	if err != nil {
		return "", fmt.Errorf("decode secret: %w", err)
	}

	message := strconv.FormatInt(timestamp, 10) + method + requestPath + body

	mac := hmac.New(sha256.New, keyBytes)
	mac.Write([]byte(message))
	sig := mac.Sum(nil)

	encoded := base64.StdEncoding.EncodeToString(sig)
	encoded = strings.ReplaceAll(encoded, "+", "-")
	encoded = strings.ReplaceAll(encoded, "/", "_")
	return encoded, nil
}

// decodeBase64Flexible decodes a base64 string accepting either standard or
// URL-safe alphabets and tolerating missing padding.
func decodeBase64Flexible(s string) ([]byte, error) {
	s = strings.ReplaceAll(s, "-", "+")
	s = strings.ReplaceAll(s, "_", "/")
	if pad := len(s) % 4; pad != 0 {
		s += strings.Repeat("=", 4-pad)
	}
	return base64.StdEncoding.DecodeString(s)
}
