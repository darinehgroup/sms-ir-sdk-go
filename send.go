package smsir

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"
)

// errNilRequest is returned when a required request struct is nil.
var errNilRequest = errors.New("smsir: nil request")

// SendBulkRequest is the request body for Client.SendBulk.
type SendBulkRequest struct {
	// LineNumber is the sending line number. Required.
	LineNumber int64 `json:"lineNumber"`
	// MessageText is the SMS text sent to every recipient. Required.
	MessageText string `json:"messageText"`
	// Mobiles is the list of recipient mobile numbers (max 100). Required.
	Mobiles []string `json:"mobiles"`
	// SendDateTime optionally schedules the send; nil sends immediately.
	// A valid schedule is between 1 hour and 365 days in the future.
	SendDateTime *UnixTime `json:"sendDateTime,omitempty"`
}

// SendPackResult is the result of a bulk or like-to-like send.
type SendPackResult struct {
	// PackID is the unique GUID of the send pack.
	PackID string `json:"packId"`
	// MessageIDs holds the unique ID of each individual message.
	MessageIDs []int64 `json:"messageIds"`
	// Cost is the credit consumed by the whole pack.
	Cost float64 `json:"cost"`
}

// SendLikeToLikeRequest is the request body for Client.SendLikeToLike. The
// lengths of MessageTexts and Mobiles must be equal; element i of
// MessageTexts is sent to element i of Mobiles.
type SendLikeToLikeRequest struct {
	// LineNumber is the sending line number. Required.
	LineNumber int64 `json:"lineNumber"`
	// MessageTexts holds one SMS text per recipient (max 100). Required.
	MessageTexts []string `json:"messageTexts"`
	// Mobiles is the list of recipient mobile numbers (max 100). Required.
	Mobiles []string `json:"mobiles"`
	// SendDateTime optionally schedules the send; nil sends immediately.
	// A valid schedule is between 1 hour and 365 days in the future.
	SendDateTime *UnixTime `json:"sendDateTime,omitempty"`
}

// SendVerifyRequest is the request body for Client.SendVerify.
type SendVerifyRequest struct {
	// Mobile is the recipient mobile number. Required.
	Mobile string `json:"mobile"`
	// TemplateID identifies a template defined in the panel (fast-send
	// section). Required. In the sandbox environment only template 123456
	// ("کد تایید شما: #CODE#") is available.
	TemplateID int `json:"templateId"`
	// Parameters supplies the values substituted into the template. Required.
	Parameters []VerifyParameter `json:"parameters"`
}

// VerifyParameter is a single template substitution for Client.SendVerify.
type VerifyParameter struct {
	// Name is the template key without the surrounding # characters.
	Name string `json:"name"`
	// Value replaces the key in the template text (max 25 characters).
	Value string `json:"value"`
}

// SendSingleResult is the result of sending a single message (verify or
// URL-based send).
type SendSingleResult struct {
	// MessageID is the unique ID of the message.
	MessageID int64 `json:"messageId"`
	// Cost is the credit consumed by the send.
	Cost float64 `json:"cost"`
}

// CancelScheduledResult is the result of Client.CancelScheduledSend.
type CancelScheduledResult struct {
	// ReturnedCreditCount is the amount of credit refunded.
	ReturnedCreditCount float64 `json:"returnedCreditCount"`
	// SmsCount is the number of cancelled messages.
	SmsCount int `json:"smsCount"`
}

// SendViaURLParams holds the credentials and message for Client.SendViaURL.
type SendViaURLParams struct {
	// Username is the panel account username.
	Username string
	// Password is the private API key from the developers panel.
	Password string
	// Line is the sending line number.
	Line int64
	// Mobile is the recipient mobile number.
	Mobile string
	// Text is the SMS text.
	Text string
}

// SendBulk sends one message text to up to 100 mobile numbers
// (POST /v1/send/bulk). Set SendDateTime to schedule the send for a future
// time.
func (c *Client) SendBulk(ctx context.Context, req *SendBulkRequest) (*SendPackResult, error) {
	if req == nil {
		return nil, errNilRequest
	}
	var out SendPackResult
	if err := c.do(ctx, http.MethodPost, "/v1/send/bulk", nil, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// SendLikeToLike sends a distinct message text to each mobile number
// (POST /v1/send/likeToLike). Set SendDateTime to schedule the send for a
// future time.
func (c *Client) SendLikeToLike(ctx context.Context, req *SendLikeToLikeRequest) (*SendPackResult, error) {
	if req == nil {
		return nil, errNilRequest
	}
	var out SendPackResult
	if err := c.do(ctx, http.MethodPost, "/v1/send/likeToLike", nil, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// SendVerify sends a high-priority templated message such as a verification
// code (POST /v1/send/verify). Verify messages are sent through service lines
// and are delivered even to numbers that block promotional SMS.
func (c *Client) SendVerify(ctx context.Context, req *SendVerifyRequest) (*SendSingleResult, error) {
	if req == nil {
		return nil, errNilRequest
	}
	var out SendSingleResult
	if err := c.do(ctx, http.MethodPost, "/v1/send/verify", nil, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// SendViaURL sends a single message through the query-string endpoint
// (GET /v1/send). Unlike every other method it authenticates with the
// username/password pair in p instead of the client's X-API-KEY header.
func (c *Client) SendViaURL(ctx context.Context, p *SendViaURLParams) (*SendSingleResult, error) {
	if p == nil {
		return nil, errNilRequest
	}
	q := url.Values{}
	q.Set("username", p.Username)
	q.Set("password", p.Password)
	q.Set("line", strconv.FormatInt(p.Line, 10))
	q.Set("mobile", p.Mobile)
	q.Set("text", p.Text)
	var out SendSingleResult
	if err := c.doRequest(ctx, http.MethodGet, "/v1/send", q, nil, &out, false); err != nil {
		return nil, err
	}
	return &out, nil
}

// CancelScheduledSend cancels a scheduled bulk or like-to-like send
// (DELETE /v1/send/scheduled/{packId}). Cancellation is allowed until 3
// minutes before the scheduled send time.
func (c *Client) CancelScheduledSend(ctx context.Context, packID string) (*CancelScheduledResult, error) {
	var out CancelScheduledResult
	if err := c.do(ctx, http.MethodDelete, "/v1/send/scheduled/"+url.PathEscape(packID), nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
