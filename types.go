package smsir

import (
	"fmt"
	"net/url"
	"strconv"
	"time"
)

// UnixTime wraps time.Time and (de)serializes as Unix seconds. The SMS.ir API
// expresses every timestamp as Unix time in UTC.
type UnixTime time.Time

// NewUnixTime returns a *UnixTime for t, convenient for optional request
// fields such as SendDateTime.
func NewUnixTime(t time.Time) *UnixTime {
	u := UnixTime(t)
	return &u
}

// Time returns the wrapped time.Time.
func (u UnixTime) Time() time.Time { return time.Time(u) }

// String implements fmt.Stringer using RFC 3339 in UTC.
func (u UnixTime) String() string { return time.Time(u).UTC().Format(time.RFC3339) }

// MarshalJSON encodes the time as an integer number of Unix seconds.
func (u UnixTime) MarshalJSON() ([]byte, error) {
	return strconv.AppendInt(nil, time.Time(u).Unix(), 10), nil
}

// UnmarshalJSON decodes an integer number of Unix seconds. JSON null leaves
// the value unchanged.
func (u *UnixTime) UnmarshalJSON(b []byte) error {
	s := string(b)
	if s == "null" {
		return nil
	}
	sec, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return fmt.Errorf("smsir: invalid unix time %q: %w", s, err)
	}
	*u = UnixTime(time.Unix(sec, 0).UTC())
	return nil
}

// DeliveryState is the delivery status of a sent message.
type DeliveryState byte

// Delivery status codes reported by the API.
const (
	DeliveryStateDelivered            DeliveryState = 1 // رسیده
	DeliveryStateUndeliveredToHandset DeliveryState = 2 // نرسیده به گوشی
	DeliveryStateReachedTelecom       DeliveryState = 3 // رسیده به مخابرات
	DeliveryStateNotReachedTelecom    DeliveryState = 4 // نرسیده به مخابرات
	DeliveryStateReachedOperator      DeliveryState = 5 // رسیده به اپراتور
	DeliveryStateFailed               DeliveryState = 6 // ناموفق
	DeliveryStateBlacklisted          DeliveryState = 7 // لیست سیاه
	DeliveryStateUnknown              DeliveryState = 8 // نامشخص
)

// String implements fmt.Stringer with a readable English name.
func (d DeliveryState) String() string {
	switch d {
	case DeliveryStateDelivered:
		return "Delivered"
	case DeliveryStateUndeliveredToHandset:
		return "UndeliveredToHandset"
	case DeliveryStateReachedTelecom:
		return "ReachedTelecom"
	case DeliveryStateNotReachedTelecom:
		return "NotReachedTelecom"
	case DeliveryStateReachedOperator:
		return "ReachedOperator"
	case DeliveryStateFailed:
		return "Failed"
	case DeliveryStateBlacklisted:
		return "Blacklisted"
	case DeliveryStateUnknown:
		return "Unknown"
	default:
		return "DeliveryState(" + strconv.Itoa(int(d)) + ")"
	}
}

// PageParams controls pagination for list endpoints. Zero values are omitted
// from the query string so the server defaults apply (PageSize 100,
// PageNumber 1).
type PageParams struct {
	// PageSize is the number of items per page (1..100).
	PageSize int
	// PageNumber is the requested page, starting at 1.
	PageNumber int
}

// apply adds the non-zero pagination fields to q. It is safe on a nil
// receiver.
func (p *PageParams) apply(q url.Values) {
	if p == nil {
		return
	}
	if p.PageSize > 0 {
		q.Set("pageSize", strconv.Itoa(p.PageSize))
	}
	if p.PageNumber > 0 {
		q.Set("pageNumber", strconv.Itoa(p.PageNumber))
	}
}
