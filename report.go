package smsir

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// MessageReport describes a sent message and its delivery status.
type MessageReport struct {
	// MessageID is the unique ID of the message.
	MessageID int64 `json:"messageId"`
	// Mobile is the recipient mobile number (returned as a number by the API).
	Mobile int64 `json:"mobile"`
	// MessageText is the SMS text.
	MessageText string `json:"messageText"`
	// SendDateTime is when the message was sent.
	SendDateTime UnixTime `json:"sendDateTime"`
	// LineNumber is the sending line number.
	LineNumber int64 `json:"lineNumber"`
	// Cost is the credit consumed.
	Cost float64 `json:"cost"`
	// DeliveryState is the delivery status; nil until a delivery report
	// arrives.
	DeliveryState *DeliveryState `json:"deliveryState"`
	// DeliveryDateTime is when the delivery report arrived; nil until then.
	DeliveryDateTime *UnixTime `json:"deliveryDateTime"`
}

// PackSummary summarizes one send pack in today's pack report.
type PackSummary struct {
	// PackID is the unique GUID of the send pack.
	PackID string `json:"packId"`
	// RecipientCount is the number of recipients in the pack.
	RecipientCount int `json:"recipientCount"`
	// CreationDateTime is when the pack was created.
	CreationDateTime UnixTime `json:"creationDateTime"`
}

// ArchiveReportParams filters Client.ArchiveReport. All fields are optional.
type ArchiveReportParams struct {
	PageParams
	// FromDate filters messages sent at or after this time.
	FromDate *time.Time
	// ToDate filters messages sent at or before this time.
	ToDate *time.Time
}

// MessageReport fetches the details and delivery status of one sent message
// by its unique ID (GET /v1/send/{messageId}).
func (c *Client) MessageReport(ctx context.Context, messageID int64) (*MessageReport, error) {
	var out MessageReport
	if err := c.do(ctx, http.MethodGet, "/v1/send/"+strconv.FormatInt(messageID, 10), nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// TodayPacks lists the send packs created today (GET /v1/send/pack). p may be
// nil to use the server defaults.
func (c *Client) TodayPacks(ctx context.Context, p *PageParams) ([]PackSummary, error) {
	q := url.Values{}
	p.apply(q)
	var out []PackSummary
	if err := c.do(ctx, http.MethodGet, "/v1/send/pack", q, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// PackReport lists every message of one send pack with its delivery status
// (GET /v1/send/pack/{packId}).
func (c *Client) PackReport(ctx context.Context, packID string) ([]MessageReport, error) {
	var out []MessageReport
	if err := c.do(ctx, http.MethodGet, "/v1/send/pack/"+url.PathEscape(packID), nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// TodayReport lists the messages sent today (GET /v1/send/live). p may be nil
// to use the server defaults.
func (c *Client) TodayReport(ctx context.Context, p *PageParams) ([]MessageReport, error) {
	q := url.Values{}
	p.apply(q)
	var out []MessageReport
	if err := c.do(ctx, http.MethodGet, "/v1/send/live", q, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ArchiveReport lists messages sent in the past, up to the end of yesterday
// (GET /v1/send/archive). p may be nil to use the server defaults.
func (c *Client) ArchiveReport(ctx context.Context, p *ArchiveReportParams) ([]MessageReport, error) {
	q := url.Values{}
	if p != nil {
		p.PageParams.apply(q)
		if p.FromDate != nil {
			q.Set("fromDate", strconv.FormatInt(p.FromDate.Unix(), 10))
		}
		if p.ToDate != nil {
			q.Set("toDate", strconv.FormatInt(p.ToDate.Unix(), 10))
		}
	}
	var out []MessageReport
	if err := c.do(ctx, http.MethodGet, "/v1/send/archive", q, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}
