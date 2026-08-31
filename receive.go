package smsir

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// ReceivedMessage describes an SMS received on one of the account's lines.
type ReceivedMessage struct {
	// ReceiveReturnID is the unique ID of the received message. It is
	// returned by LatestReceived and ArchiveReceived; TodayReceived does not
	// include it, so it stays zero there.
	ReceiveReturnID int64 `json:"receiveReturnId"`
	// MessageText is the SMS text.
	MessageText string `json:"messageText"`
	// Number is the receiving line number.
	Number int64 `json:"number"`
	// Mobile is the sender's mobile number (returned as a number by the API).
	Mobile int64 `json:"mobile"`
	// ReceivedDateTime is when the message was received.
	ReceivedDateTime UnixTime `json:"receivedDateTime"`
}

// TodayReceivedParams filters Client.TodayReceived. All fields are optional.
type TodayReceivedParams struct {
	PageParams
	// SortByNewest sorts by receive time descending; the server default is
	// ascending.
	SortByNewest bool
	// Mobile filters by the sender's mobile number.
	Mobile string
}

// ArchiveReceivedParams filters Client.ArchiveReceived. All fields are
// optional.
type ArchiveReceivedParams struct {
	PageParams
	// FromDate filters messages received at or after this time.
	FromDate *time.Time
	// ToDate filters messages received at or before this time.
	ToDate *time.Time
	// Mobile filters by the sender's mobile number.
	Mobile string
}

// LatestReceived fetches the newest unread received messages
// (GET /v1/receive/latest). count is capped at 100 by the server; values <= 0
// omit the parameter so the server default (100) applies.
//
// This is a destructive read: each received message is returned by this
// endpoint only once. After being fetched it is marked as read and can no
// longer be retrieved through this method (use TodayReceived or
// ArchiveReceived instead).
func (c *Client) LatestReceived(ctx context.Context, count int) ([]ReceivedMessage, error) {
	q := url.Values{}
	if count > 0 {
		q.Set("count", strconv.Itoa(count))
	}
	var out []ReceivedMessage
	if err := c.do(ctx, http.MethodGet, "/v1/receive/latest", q, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// TodayReceived lists the messages received today, read or unread
// (GET /v1/receive/live). During the first hours of the day it also returns
// yesterday's messages. The returned items do not include ReceiveReturnID.
// p may be nil to use the server defaults.
func (c *Client) TodayReceived(ctx context.Context, p *TodayReceivedParams) ([]ReceivedMessage, error) {
	q := url.Values{}
	if p != nil {
		p.PageParams.apply(q)
		if p.SortByNewest {
			q.Set("sortByNewest", "true")
		}
		if p.Mobile != "" {
			q.Set("mobile", p.Mobile)
		}
	}
	var out []ReceivedMessage
	if err := c.do(ctx, http.MethodGet, "/v1/receive/live", q, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ArchiveReceived lists messages received in the past, up to the end of
// yesterday (GET /v1/receive/archive). p may be nil to use the server
// defaults.
func (c *Client) ArchiveReceived(ctx context.Context, p *ArchiveReceivedParams) ([]ReceivedMessage, error) {
	q := url.Values{}
	if p != nil {
		p.PageParams.apply(q)
		if p.FromDate != nil {
			q.Set("fromDate", strconv.FormatInt(p.FromDate.Unix(), 10))
		}
		if p.ToDate != nil {
			q.Set("toDate", strconv.FormatInt(p.ToDate.Unix(), 10))
		}
		if p.Mobile != "" {
			q.Set("mobile", p.Mobile)
		}
	}
	var out []ReceivedMessage
	if err := c.do(ctx, http.MethodGet, "/v1/receive/archive", q, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}
