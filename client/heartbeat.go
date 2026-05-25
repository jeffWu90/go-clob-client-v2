package client

import (
	"context"
	"encoding/json"
	"net/http"

	clob "github.com/jeffWu90/go-clob-client-v2"
	"github.com/jeffWu90/go-clob-client-v2/httphelpers"
)

// HeartbeatResponse is the body returned by POST /v1/heartbeats.
type HeartbeatResponse struct {
	HeartbeatID string `json:"heartbeat_id"`
	ErrorMsg    string `json:"error_msg,omitempty"`
}

// PostHeartbeat publishes a heartbeat for the account, optionally tagging it
// with a caller-supplied id.
func (c *ClobClient) PostHeartbeat(ctx context.Context, heartbeatID string) (*HeartbeatResponse, error) {
	body := map[string]string{"heartbeat_id": heartbeatID}
	bodyJSON, _ := json.Marshal(body)
	hdrs, err := c.buildL2Headers(ctx, http.MethodPost, clob.EndpointHeartbeat, string(bodyJSON))
	if err != nil {
		return nil, err
	}
	var out HeartbeatResponse
	err = c.http.Post(ctx, c.url(clob.EndpointHeartbeat), httphelpers.RequestOptions{Headers: hdrs, Body: body}, &out)
	return &out, err
}
