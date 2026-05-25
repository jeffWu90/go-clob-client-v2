package client

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	clob "github.com/jeffWu90/go-clob-client-v2"
	"github.com/jeffWu90/go-clob-client-v2/httphelpers"
	"github.com/jeffWu90/go-clob-client-v2/types"
)

// GetNotifications returns the notifications pending for the account.
func (c *ClobClient) GetNotifications(ctx context.Context) ([]types.Notification, error) {
	hdrs, err := c.buildL2Headers(ctx, http.MethodGet, clob.EndpointGetNotifications, "")
	if err != nil {
		return nil, err
	}
	v := url.Values{"signature_type": []string{strconv.Itoa(int(c.orderBuilder.SignatureType))}}
	var out []types.Notification
	err = c.http.Get(ctx, c.url(clob.EndpointGetNotifications), httphelpers.RequestOptions{Headers: hdrs, Params: v}, &out)
	return out, err
}

// DropNotifications acknowledges (deletes) the listed notifications.
func (c *ClobClient) DropNotifications(ctx context.Context, params *types.DropNotificationParams) error {
	hdrs, err := c.buildL2Headers(ctx, http.MethodDelete, clob.EndpointDropNotifications, "")
	if err != nil {
		return err
	}
	return c.http.Delete(ctx, c.url(clob.EndpointDropNotifications), httphelpers.RequestOptions{
		Headers: hdrs,
		Params:  httphelpers.ParseDropNotificationParams(params),
	}, nil)
}
