package smsir

import (
	"context"
	"net/http"
)

// Credit returns the account's current credit balance (GET /v1/credit).
func (c *Client) Credit(ctx context.Context) (float64, error) {
	var out float64
	if err := c.do(ctx, http.MethodGet, "/v1/credit", nil, nil, &out); err != nil {
		return 0, err
	}
	return out, nil
}

// Lines returns the line numbers available for sending (GET /v1/line).
func (c *Client) Lines(ctx context.Context) ([]int64, error) {
	var out []int64
	if err := c.do(ctx, http.MethodGet, "/v1/line", nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}
