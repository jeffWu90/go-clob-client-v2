package httphelpers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	clob "github.com/jeffWu90/go-clob-client-v2"
)

// userAgent matches the value emitted by the TS clob-client.
const userAgent = "@polymarket/clob-client"

// transientRetryDelay matches the 30 ms backoff used by the TS retry path.
const transientRetryDelay = 30 * time.Millisecond

// Client wraps a standard library *http.Client with the request-building,
// header overloading, and error-translation conventions used across all
// CLOB API endpoints. A zero-value Client uses http.DefaultClient.
type Client struct {
	// HTTP is the underlying transport. nil falls back to http.DefaultClient.
	HTTP *http.Client
	// RetryOnError, when true, retries POST requests once after a 30 ms delay
	// when the first attempt failed with a transient error (network / 5xx).
	RetryOnError bool
}

// New returns a Client backed by the given http.Client. Pass nil to use the
// default client.
func New(httpClient *http.Client) *Client {
	return &Client{HTTP: httpClient}
}

// RequestOptions configures a single request.
type RequestOptions struct {
	// Headers overrides individual headers (User-Agent etc. are added automatically).
	Headers map[string]string
	// Body is marshaled to JSON and sent in the request body. Nil omits the body.
	Body any
	// Params populates the query string.
	Params url.Values
}

// Get issues a GET request and decodes the response JSON into out (which may
// be nil to discard the response).
func (c *Client) Get(ctx context.Context, rawURL string, opts RequestOptions, out any) error {
	return c.do(ctx, http.MethodGet, rawURL, opts, out, false)
}

// Post issues a POST request. Retries once on transient errors when RetryOnError is set.
func (c *Client) Post(ctx context.Context, rawURL string, opts RequestOptions, out any) error {
	return c.do(ctx, http.MethodPost, rawURL, opts, out, c.RetryOnError)
}

// Delete issues a DELETE request.
func (c *Client) Delete(ctx context.Context, rawURL string, opts RequestOptions, out any) error {
	return c.do(ctx, http.MethodDelete, rawURL, opts, out, false)
}

func (c *Client) do(ctx context.Context, method, rawURL string, opts RequestOptions, out any, retry bool) error {
	err := c.doOnce(ctx, method, rawURL, opts, out)
	if err == nil || !retry || !isTransientError(err) {
		return err
	}
	timer := time.NewTimer(transientRetryDelay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
	}
	return c.doOnce(ctx, method, rawURL, opts, out)
}

func (c *Client) doOnce(ctx context.Context, method, rawURL string, opts RequestOptions, out any) error {
	var bodyReader io.Reader
	if opts.Body != nil {
		buf, err := json.Marshal(opts.Body)
		if err != nil {
			return fmt.Errorf("marshal body: %w", err)
		}
		bodyReader = bytes.NewReader(buf)
	}

	if len(opts.Params) > 0 {
		sep := "?"
		if strings.Contains(rawURL, "?") {
			sep = "&"
		}
		rawURL = rawURL + sep + opts.Params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, rawURL, bodyReader)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	applyDefaultHeaders(req, method)
	for k, v := range opts.Headers {
		req.Header.Set(k, v)
	}

	httpClient := c.HTTP
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		if out == nil || len(bytes.TrimSpace(respBody)) == 0 {
			return nil
		}
		if err := json.Unmarshal(respBody, out); err != nil {
			return fmt.Errorf("decode response (status=%d body=%s): %w", resp.StatusCode, string(respBody), err)
		}
		return nil
	}

	// Non-2xx: build an ApiError. The CLOB server typically returns either a
	// {"error": "..."} body or a free-form string; preserve both shapes by
	// stashing the parsed value (or raw string) in ApiError.Data.
	return apiErrorFromResponse(resp.StatusCode, respBody)
}

// applyDefaultHeaders sets the Polymarket CLOB request defaults that match
// the TS client's overloadHeaders() helper.
func applyDefaultHeaders(req *http.Request, method string) {
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", userAgent)
	}
	if req.Header.Get("Accept") == "" {
		req.Header.Set("Accept", "*/*")
	}
	if req.Header.Get("Connection") == "" {
		req.Header.Set("Connection", "keep-alive")
	}
	if req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if method == http.MethodGet && req.Header.Get("Accept-Encoding") == "" {
		req.Header.Set("Accept-Encoding", "gzip")
	}
}

// apiErrorFromResponse extracts the canonical error message + data from a
// non-2xx CLOB response, accepting both {"error":"..."} and free-form bodies.
func apiErrorFromResponse(status int, body []byte) error {
	trimmed := bytes.TrimSpace(body)
	if len(trimmed) == 0 {
		return clob.NewApiError(http.StatusText(status), status, nil)
	}

	// Try to parse as a JSON object with an "error" key (canonical shape).
	var asMap map[string]any
	if err := json.Unmarshal(trimmed, &asMap); err == nil {
		if msg, ok := asMap["error"].(string); ok && msg != "" {
			return clob.NewApiError(msg, status, asMap)
		}
		// JSON object without an "error" key -> stash whole body, use status text.
		return clob.NewApiError(http.StatusText(status), status, asMap)
	}

	// Try JSON string body.
	var asStr string
	if err := json.Unmarshal(trimmed, &asStr); err == nil {
		return clob.NewApiError(asStr, status, asStr)
	}

	// Fallback: raw bytes as message.
	return clob.NewApiError(string(trimmed), status, string(trimmed))
}

// isTransientError reports whether the error is worth retrying. Transient
// covers connection-level failures and 5xx responses; client-side 4xx errors
// are not retried because the caller's request is malformed.
func isTransientError(err error) bool {
	var apiErr *clob.ApiError
	if errors.As(err, &apiErr) {
		return apiErr.Status >= 500 && apiErr.Status < 600
	}
	// Non-API error means transport failure.
	return true
}
