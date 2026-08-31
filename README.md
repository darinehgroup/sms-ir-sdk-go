# sms-ir-sdk-go

A small, dependency-free Go client for the [SMS.ir](https://sms.ir) web service.
It covers sending (bulk, like-to-like, verify, URL-based, scheduled cancellation),
delivery reports, received messages, and account settings (credit, lines).

Also available in Persian: [README-fa.md](README-fa.md)

- **Zero external dependencies** — the standard library only.
- **Context-aware** — every method takes a `context.Context`.
- **Immutable client** — safe for concurrent use after construction.
- **Typed errors** — failures surface as `*smsir.APIError` with the API status code.

Design and implementation reference (Persian): [DESIGN.md](DESIGN.md).

[![Go Reference](https://pkg.go.dev/badge/github.com/darinehgroup/sms-ir-sdk-go.svg)](https://pkg.go.dev/github.com/darinehgroup/sms-ir-sdk-go)
[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

---

## Installation

```bash
go get github.com/darinehgroup/sms-ir-sdk-go
```

The package is imported as `smsir`:

```go
import "github.com/darinehgroup/sms-ir-sdk-go"
```

---

## Quick start

Create a client with your API key (generated in the SMS.ir developers panel),
then call any method. Every method takes a `context.Context` as its first
argument and returns a typed result or an error.

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/darinehgroup/sms-ir-sdk-go"
)

func main() {
    ctx := context.Background()
    client := smsir.New(os.Getenv("SMSIR_API_KEY"))

    result, err := client.SendVerify(ctx, &smsir.SendVerifyRequest{
        Mobile:     "9120000000",
        TemplateID: 123456,
        Parameters: []smsir.VerifyParameter{{Name: "Code", Value: "12345"}},
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("message id:", result.MessageID, "cost:", result.Cost)
}
```

---

## Creating a client

```go
client := smsir.New("YOUR_API_KEY")
```

Optional configuration via functional options:

```go
client := smsir.New(
    "YOUR_API_KEY",
    smsir.WithHTTPClient(&http.Client{Timeout: 10 * time.Second}),
    smsir.WithUserAgent("my-app/1.0"),
    smsir.WithBaseURL("https://api.sms.ir"), // override only if needed
)
```

| Option            | Purpose                                                        |
|-------------------|----------------------------------------------------------------|
| `WithBaseURL`     | Override the base URL (`https://api.sms.ir` by default).      |
| `WithHTTPClient`  | Provide a custom `*http.Client` (proxies, timeouts, etc.).    |
| `WithUserAgent`   | Override the `User-Agent` header (`smsir-go/<version>`).      |

The default HTTP client uses a 30-second timeout.

---

## Sending messages

### Verify (OTP / templated)

Send a high-priority templated message such as a verification code. Verify
messages go through service lines and are delivered even to numbers that have
blocked promotional SMS. Templates are defined in the panel's fast-send section.

```go
res, err := client.SendVerify(ctx, &smsir.SendVerifyRequest{
    Mobile:     "9120000000",
    TemplateID: 123456,
    Parameters: []smsir.VerifyParameter{{Name: "Code", Value: "12345"}},
})
// res.MessageID, res.Cost
```

### Bulk (same text to many)

One message text to up to 100 recipients.

```go
res, err := client.SendBulk(ctx, &smsir.SendBulkRequest{
    LineNumber:  30004505000017,
    MessageText: "Hello from SMS.ir",
    Mobiles:     []string{"9120000000", "9120000001"},
})
// res.PackID, res.MessageIDs, res.Cost
```

### Like-to-like (different text per recipient)

Each recipient gets a different text. `MessageTexts` and `Mobiles` must have
equal length.

```go
res, err := client.SendLikeToLike(ctx, &smsir.SendLikeToLikeRequest{
    LineNumber:   30004505000017,
    MessageTexts: []string{"Hello A", "Hello B"},
    Mobiles:      []string{"9120000000", "9120000001"},
})
```

### Scheduled send

Set `SendDateTime` to schedule. A valid schedule is between **1 hour** and
**365 days** in the future.

```go
when := time.Now().Add(2 * time.Hour)
res, err := client.SendBulk(ctx, &smsir.SendBulkRequest{
    LineNumber:   30004505000017,
    MessageText:  "Reminder at 14:00",
    Mobiles:      []string{"9120000000"},
    SendDateTime: smsir.NewUnixTime(when),
})
```

### Cancel a scheduled send

Allowed until **3 minutes** before the scheduled send time.

```go
res, err := client.CancelScheduledSend(ctx, "pack-guid")
// res.ReturnedCreditCount, res.SmsCount
```

### Send via URL (legacy)

The legacy query-string endpoint authenticates with `username`/`password` in the
query string instead of the `X-API-KEY` header. `Password` is the private API
key from the developers panel.

```go
res, err := client.SendViaURL(ctx, &smsir.SendViaURLParams{
    Username: "panel-user",
    Password: "panel-private-key",
    Line:     30004505000017,
    Mobile:   "9120000000",
    Text:     "hello",
})
```

---

## Delivery reports

```go
// One message by ID.
rep, err := client.MessageReport(ctx, 89545112)

// Today's send packs.
packs, err := client.TodayPacks(ctx, &smsir.PageParams{PageSize: 50, PageNumber: 1})

// All messages in one pack.
msgs, err := client.PackReport(ctx, "pack-guid")

// Today's messages.
msgs, err := client.TodayReport(ctx, &smsir.PageParams{PageSize: 100})

// Archived messages (up to end of yesterday), optional date range.
from := time.Now().AddDate(0, -1, 0)
to := time.Now()
msgs, err := client.ArchiveReport(ctx, &smsir.ArchiveReportParams{
    PageParams: smsir.PageParams{PageSize: 100},
    FromDate:   &from,
    ToDate:     &to,
})
```

`MessageReport.DeliveryState` is a pointer that is `nil` until a delivery report
arrives. Once delivered it holds one of the `smsir.DeliveryState*` constants.

---

## Received messages

```go
// ⚠️ Destructive read: each message is returned only once, then marked read.
latest, err := client.LatestReceived(ctx, 50)

// Today's received messages (read and unread). p may be nil.
today, err := client.TodayReceived(ctx, &smsir.TodayReceivedParams{
    PageParams:   smsir.PageParams{PageSize: 50},
    SortByNewest: true,
})

// Archived received messages (up to end of yesterday).
archive, err := client.ArchiveReceived(ctx, &smsir.ArchiveReceivedParams{
    PageParams: smsir.PageParams{PageSize: 50},
    Mobile:     "9120000000",
})
```

> **⚠️ `LatestReceived` is a destructive read.** Each received message is
> returned by this endpoint only once. After being fetched it is marked as read
> and can no longer be retrieved through `LatestReceived`. Use `TodayReceived`
> or `ArchiveReceived` to re-read messages.

Note: `TodayReceived` does **not** populate `ReceiveReturnID` (it stays `0`),
while `LatestReceived` and `ArchiveReceived` do.

---

## Account settings

```go
credit, err := client.Credit(ctx)     // current balance (float64)
lines, err := client.Lines(ctx)       // available sending line numbers
```

---

## API reference (method → endpoint)

| Method                  | HTTP                                  | Returns                       |
|-------------------------|---------------------------------------|-------------------------------|
| `SendBulk`              | `POST /v1/send/bulk`                  | `*SendPackResult`             |
| `SendLikeToLike`        | `POST /v1/send/likeToLike`            | `*SendPackResult`             |
| `SendVerify`            | `POST /v1/send/verify`                | `*SendSingleResult`           |
| `SendViaURL`            | `GET /v1/send` (query-string auth)    | `*SendSingleResult`           |
| `CancelScheduledSend`   | `DELETE /v1/send/scheduled/{packId}`  | `*CancelScheduledResult`      |
| `MessageReport`         | `GET /v1/send/{messageId}`            | `*MessageReport`              |
| `TodayPacks`            | `GET /v1/send/pack`                   | `[]PackSummary`               |
| `PackReport`            | `GET /v1/send/pack/{packId}`          | `[]MessageReport`             |
| `TodayReport`           | `GET /v1/send/live`                   | `[]MessageReport`             |
| `ArchiveReport`         | `GET /v1/send/archive`                | `[]MessageReport`             |
| `LatestReceived`        | `GET /v1/receive/latest`              | `[]ReceivedMessage`           |
| `TodayReceived`         | `GET /v1/receive/live`                | `[]ReceivedMessage`           |
| `ArchiveReceived`       | `GET /v1/receive/archive`             | `[]ReceivedMessage`           |
| `Credit`                | `GET /v1/credit`                      | `float64`                     |
| `Lines`                 | `GET /v1/line`                        | `[]int64`                     |

Full API documentation is available on
[pkg.go.dev](https://pkg.go.dev/github.com/darinehgroup/sms-ir-sdk-go).

---

## Error handling

Failures returned by the SMS.ir API are typed as `*smsir.APIError`. Use
`errors.As` to inspect them, and the helper methods `IsAuthError()` and
`IsRateLimited()` to branch on common cases.

```go
res, err := client.SendBulk(ctx, req)
if err != nil {
    var apiErr *smsir.APIError
    if errors.As(err, &apiErr) {
        fmt.Println("status:", apiErr.Status)   // API status code
        fmt.Println("http:", apiErr.HTTPStatus) // HTTP status code
        fmt.Println("message:", apiErr.Message) // server message (Persian)

        if apiErr.IsAuthError() {
            // refresh key, alert ops, etc.
        }
        if apiErr.IsRateLimited() {
            // back off and retry
        }
    } else {
        // network error, context cancellation, JSON error, etc.
        fmt.Println("transport/decode error:", err)
    }
    return
}
```

`APIError.Error()` format:
`smsir: api error (http 400, status 102): اعتبار کافی نیست`.

### API status codes

| Code | Constant                            | Meaning                                       |
|-----:|-------------------------------------|-----------------------------------------------|
| 0    | `StatusFailed`                      | Request failed                                |
| 1    | `StatusSuccess`                     | Success                                       |
| 10   | `StatusInvalidAPIKey`               | Invalid API key                               |
| 11   | `StatusInactiveAPIKey`              | Inactive API key                              |
| 12   | `StatusAPIKeyIPRestricted`          | API key restricted to specific IPs            |
| 13   | `StatusInactiveAccount`             | Inactive account                              |
| 14   | `StatusSuspendedAccount`            | Suspended account                             |
| 15   | `StatusPlanUpgradeRequired`         | Plan upgrade required to use the web service  |
| 16   | `StatusInvalidParameter`            | Invalid parameter value                       |
| 20   | `StatusTooManyRequests`             | Too many requests (rate limited)              |
| 101  | `StatusInvalidLineNumber`           | Invalid line number                           |
| 102  | `StatusInsufficientCredit`          | Insufficient credit                           |
| 103  | `StatusEmptyMessageText`            | Empty message text(s)                         |
| 104  | `StatusInvalidMobileNumber`         | Invalid mobile number(s)                      |
| 105  | `StatusTooManyMobiles`              | More than 100 mobiles                         |
| 106  | `StatusTooManyMessageTexts`         | More than 100 message texts                   |
| 107  | `StatusEmptyMobileList`             | Empty mobile list                             |
| 108  | `StatusEmptyTextList`               | Empty text list                               |
| 109  | `StatusInvalidSendDateTime`         | Invalid send date/time                        |
| 110  | `StatusMobileTextCountMismatch`     | Mobile and text counts differ                 |
| 111  | `StatusSendNotFound`                | No send found with this ID                    |
| 112  | `StatusNoRecordToDelete`            | No record to delete                           |
| 113  | `StatusTemplateNotFound`            | Template not found                            |
| 114  | `StatusParameterValueTooLong`       | Parameter value longer than 25 chars          |
| 115  | `StatusMobileInBlacklist`           | Mobile number is blacklisted                  |
| 116  | `StatusMissingParameterName`        | A parameter name was not provided             |
| 117  | `StatusTextNotApproved`             | Text not approved                             |
| 118  | `StatusTooManyMessages`             | Too many messages                             |
| 119  | `StatusCustomTemplatePlanUpgradeRequired` | Plan upgrade required for custom template |
| 123  | `StatusLineActivationRequired`      | Sender line needs activation                  |
| 124  | `StatusOnlyOTPTemplateAllowed`      | Only OTP templates allowed; template not OTP  |

---

## Time values

The SMS.ir API expresses every timestamp as **Unix seconds (UTC)**. This SDK
exposes them as `smsir.UnixTime` (a wrapper around `time.Time`):

```go
// Convert to time.Time.
t := report.SendDateTime.Time()

// Build a value for an optional request field.
sendAt := smsir.NewUnixTime(time.Now().Add(2 * time.Hour))
```

---

## Sandbox

Sandbox shares the same base URL (`https://api.sms.ir`) — there is no separate
URL. Only the **API key type** differs: generate a sandbox key in the panel and
pass it to `smsir.New` exactly like a production key.

In sandbox:

- No real SMS is sent, no credit is deducted, and no report is recorded.
- Input validation and errors behave like production.
- Only **one** verify template is active: **template `123456`** with the text
  `کد تایید شما: #CODE#`, parameter name `Code`.

```go
// Sandbox: same client, sandbox key.
client := smsir.New("YOUR_SANDBOX_KEY")
_, _ = client.SendVerify(ctx, &smsir.SendVerifyRequest{
    Mobile:     "9120000000",
    TemplateID: 123456,
    Parameters: []smsir.VerifyParameter{{Name: "Code", Value: "12345"}},
})
```

---

## License

[MIT](LICENSE) © Darineh Group

