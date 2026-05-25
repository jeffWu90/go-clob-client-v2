package httphelpers

import (
	"net/url"
	"strings"

	"github.com/jeffWu90/go-clob-client-v2/types"
)

// ParseOrdersScoringParams converts OrdersScoringParams to a query string.
// Matches the TS helper of the same name (comma-joined order ids).
func ParseOrdersScoringParams(p *types.OrdersScoringParams) url.Values {
	v := url.Values{}
	if p != nil && len(p.OrderIDs) > 0 {
		v.Set("order_ids", strings.Join(p.OrderIDs, ","))
	}
	return v
}

// ParseDropNotificationParams converts DropNotificationParams to a query string.
func ParseDropNotificationParams(p *types.DropNotificationParams) url.Values {
	v := url.Values{}
	if p != nil && len(p.IDs) > 0 {
		v.Set("ids", strings.Join(p.IDs, ","))
	}
	return v
}
