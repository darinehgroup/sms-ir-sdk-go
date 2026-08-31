# طراحی SDK وب‌سرویس SMS.ir برای Go

این سند، طراحی کامل و خودکفای SDK رسمی‌نمای Go برای وب‌سرویس [SMS.ir](https://sms.ir) است. تمام اطلاعات لازم برای پیاده‌سازی (endpoint ها، مدل‌های ورودی/خروجی، کدهای خطا، ساختار پکیج، امضای متدها و استراتژی تست) در همین سند آمده و به هیچ منبع دیگری نیاز نیست.

> **نقش این فایل:** سند طراحی *و* مرجع پیاده‌سازی — کد موجود در همین ریپو دقیقاً بر اساس همین سند نوشته شده و ساختار فایل‌ها (بخش ۱) با ساختار واقعی ریپو یکی است. برای «چطور از SDK استفاده کنم» [README.md](README.md) را ببینید؛ برای «چرا این‌طور پیاده شده و قراردادهای سرور چیستند» همین سند.

---

## ۱. مشخصات کلی پروژه

| مورد | مقدار |
|---|---|
| نام ماژول | `github.com/darinehgroup/sms-ir-sdk-go` |
| نام پکیج | `smsir` (پکیج ریشه‌ی ریپازیتوری) |
| حداقل نسخه Go | 1.22 |
| وابستگی خارجی | **هیچ** — فقط کتابخانه استاندارد |
| Base URL پیش‌فرض | `https://api.sms.ir` |
| فرمت داده | JSON (ارسال و دریافت) |
| لایسنس | MIT |

### قواعد کدنویسی
- تمام identifier ها و doc comment ها به انگلیسی و مطابق قواعد استاندارد Go (`gofmt`, `go vet` بدون خطا).
- تمام متدهای عمومی `context.Context` را به‌عنوان پارامتر اول می‌گیرند.
- هیچ متغیر global قابل‌تغییری وجود ندارد؛ `Client` بعد از ساخت thread-safe و immutable است.
- SDK اعتبارسنجی سمت‌کلاینتِ قواعد سرور را **تکرار نمی‌کند** (مثلاً سقف ۱۰۰ شماره)؛ خطاهای سرور عیناً به‌صورت `*APIError` برگردانده می‌شوند. فقط چک‌های بدیهی (nil بودن request) انجام می‌شود.

### ساختار فایل‌ها

```
sms-ir-sdk-go/
├── go.mod              // module github.com/darinehgroup/sms-ir-sdk-go — go 1.22
├── client.go           // Client, Option, New, متد داخلی do/get/post/delete
├── errors.go           // APIError + ثابت‌های کد وضعیت API
├── types.go            // UnixTime, DeliveryState, PageParams و مدل‌های مشترک
├── send.go             // SendBulk, SendLikeToLike, SendVerify, SendViaURL, CancelScheduledSend
├── report.go           // MessageReport, TodayPacks, PackReport, TodayReport, ArchiveReport
├── receive.go          // LatestReceived, TodayReceived, ArchiveReceived
├── settings.go         // Credit, Lines
├── client_test.go      // تست‌های Client و envelope و خطاها
├── send_test.go
├── report_test.go
├── receive_test.go
├── settings_test.go
├── examples_test.go    // Example functions برای godoc
├── README.md
└── LICENSE
```

---

## ۲. قراردادهای وب‌سرویس (مرجع کامل)

### ۲.۱. احراز هویت
- کلید API در هدر `X-API-KEY` هر درخواست ارسال می‌شود (به‌جز endpoint «ارسال از طریق URL» که username/password را به‌صورت query string می‌گیرد).
- هدرهای ثابت هر درخواست:
  - `X-API-KEY: <apiKey>`
  - `Accept: application/json`
  - `Content-Type: application/json` (فقط وقتی بدنه دارد)

### ۲.۲. محیط Sandbox
- Sandbox هیچ URL جداگانه‌ای ندارد؛ همان `https://api.sms.ir` است و فقط **نوع کلید API** فرق می‌کند (کلید نوع Sandbox از پنل ساخته می‌شود).
- در Sandbox پیامک واقعی ارسال نمی‌شود، اعتباری کسر نمی‌شود، گزارشی ثبت نمی‌شود و داده‌های بازگشتی شبیه‌سازی‌شده‌اند؛ ولی اعتبارسنجی ورودی‌ها و خطاها مثل محیط اصلی است.
- در Sandbox فقط یک قالب Verify فعال است: شناسه `123456` با متن `کد تایید شما: #CODE#`.
- در SDK نیاز به هیچ سوییچ خاصی نیست؛ فقط در README توضیح داده شود که کاربر کلید Sandbox را به همان `New()` بدهد.

### ۲.۳. ساختار پاسخ یکپارچه (Envelope)
تمام endpoint ها این ساختار را برمی‌گردانند:

```json
{
  "status": 1,
  "message": "موفق",
  "data": <payload>
}
```

- `status == 1` یعنی موفق؛ هر مقدار دیگر یعنی خطا (جدول بخش ۵).
- `data` بسته به endpoint می‌تواند object، آرایه یا عدد ساده باشد.

### ۲.۴. کدهای HTTP

| HTTP Status | معنی |
|---|---|
| 200 | موفق |
| 400 | خطای منطقی |
| 401 | خطای احراز هویت |
| 429 | تعداد درخواست بیش از حد مجاز |
| 500 | خطای غیرمنتظره سرور |

### ۲.۵. زمان
تمام مقادیر زمانی، **Unix Time بر حسب ثانیه و UTC** هستند.

---

## ۳. سطح API عمومی SDK

### ۳.۱. ساخت کلاینت (`client.go`)

```go
package smsir

// New creates an SMS.ir API client. apiKey may be a production or sandbox key.
func New(apiKey string, opts ...Option) *Client

type Option func(*Client)

// WithBaseURL overrides the default base URL (https://api.sms.ir).
// Trailing slashes must be trimmed.
func WithBaseURL(u string) Option

// WithHTTPClient sets a custom *http.Client (e.g. for proxies or custom timeouts).
func WithHTTPClient(hc *http.Client) Option

// WithUserAgent overrides the default User-Agent header
// ("smsir-go/<version>" — version is a const in client.go).
func WithUserAgent(ua string) Option
```

- `http.Client` پیش‌فرض: `&http.Client{Timeout: 30 * time.Second}`.
- متد داخلی `do(ctx, method, path string, query url.Values, body any, out any) error` مسئول ساخت درخواست، ست‌کردن هدرها، اجرای درخواست، decode کردن envelope و تبدیل خطاست. همه‌ی متدهای عمومی از آن استفاده می‌کنند.
- منطق `do`:
  1. بدنه (در صورت وجود) با `encoding/json` سریالایز می‌شود.
  2. پاسخ (بدون توجه به HTTP status) به `apiEnvelope` decode می‌شود:
     ```go
     type apiEnvelope struct {
         Status  int             `json:"status"`
         Message string          `json:"message"`
         Data    json.RawMessage `json:"data"`
     }
     ```
  3. اگر decode شکست خورد یا HTTP status غیر 200 بود و بدنه envelope معتبر نداشت → `*APIError{HTTPStatus: <code>, Status: -1, Message: "<متن خطا/بدنه خام کوتاه‌شده>"}`.
  4. اگر `Status != 1` → `*APIError{HTTPStatus: <code>, Status: envelope.Status, Message: envelope.Message}`.
  5. در حالت موفق، `Data` روی `out` (در صورت غیر nil بودن) decode می‌شود.

### ۳.۲. متدهای Client — فهرست کامل

```go
// ارسال‌ها
func (c *Client) SendBulk(ctx context.Context, req *SendBulkRequest) (*SendPackResult, error)
func (c *Client) SendLikeToLike(ctx context.Context, req *SendLikeToLikeRequest) (*SendPackResult, error)
func (c *Client) SendVerify(ctx context.Context, req *SendVerifyRequest) (*SendSingleResult, error)
func (c *Client) SendViaURL(ctx context.Context, p *SendViaURLParams) (*SendSingleResult, error)
func (c *Client) CancelScheduledSend(ctx context.Context, packID string) (*CancelScheduledResult, error)

// گزارش‌های ارسال
func (c *Client) MessageReport(ctx context.Context, messageID int64) (*MessageReport, error)
func (c *Client) TodayPacks(ctx context.Context, p *PageParams) ([]PackSummary, error)
func (c *Client) PackReport(ctx context.Context, packID string) ([]MessageReport, error)
func (c *Client) TodayReport(ctx context.Context, p *PageParams) ([]MessageReport, error)
func (c *Client) ArchiveReport(ctx context.Context, p *ArchiveReportParams) ([]MessageReport, error)

// پیامک‌های دریافتی
func (c *Client) LatestReceived(ctx context.Context, count int) ([]ReceivedMessage, error)
func (c *Client) TodayReceived(ctx context.Context, p *TodayReceivedParams) ([]ReceivedMessage, error)
func (c *Client) ArchiveReceived(ctx context.Context, p *ArchiveReceivedParams) ([]ReceivedMessage, error)

// تنظیمات
func (c *Client) Credit(ctx context.Context) (float64, error)
func (c *Client) Lines(ctx context.Context) ([]int64, error)
```

- پارامترهای pointer ای که «اختیاری» هستند (`p *PageParams` و مشابه) می‌توانند `nil` باشند؛ در این حالت هیچ query string ای فرستاده نمی‌شود و پیش‌فرض سرور اعمال می‌شود.
- پارامترهای request ای که «اجباری» هستند (`req` در متدهای Send) اگر `nil` باشند، خطای `errors.New("smsir: nil request")` برگردانده می‌شود.

### ۳.۳. تایپ‌های مشترک (`types.go`)

```go
// UnixTime wraps time.Time and (de)serializes as Unix seconds (UTC).
type UnixTime time.Time

func NewUnixTime(t time.Time) *UnixTime      // helper برای مقداردهی فیلدهای اختیاری
func (u UnixTime) Time() time.Time
func (u UnixTime) MarshalJSON() ([]byte, error)   // → عدد صحیح ثانیه‌ی Unix
func (u *UnixTime) UnmarshalJSON(b []byte) error  // ← عدد صحیح؛ null را هم بی‌خطا رد می‌کند

// DeliveryState is the delivery status of a sent message.
type DeliveryState byte

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

func (d DeliveryState) String() string // نام انگلیسی خوانا، مثلا "Delivered"

// PageParams controls pagination for list endpoints.
type PageParams struct {
    PageSize   int // 1..100; صفر یعنی ارسال نشود (پیش‌فرض سرور: 100)
    PageNumber int // از 1؛ صفر یعنی ارسال نشود (پیش‌فرض سرور: 1)
}
```

قاعده‌ی ساخت query string برای پارامترهای اختیاری: **فقط فیلدهای غیر صفر/غیر خالی/غیر nil ارسال شوند** (به‌جز `SortByNewest` که چون bool است، فقط وقتی `true` است ارسال شود).

### ۳.۴. مدیریت خطا (`errors.go`)

```go
// APIError represents an error returned by the SMS.ir API.
type APIError struct {
    HTTPStatus int    // کد HTTP پاسخ (مثلاً 400، 401، 429)
    Status     int    // کد وضعیت API از envelope؛ -1 اگر بدنه‌ی معتبر نداشت
    Message    string // متن پیام سرور (فارسی)
}

func (e *APIError) Error() string
// فرمت: "smsir: api error (http 400, status 102): اعتبار کافی نمیباشد"

// Helper ها:
func (e *APIError) IsAuthError() bool    // status ∈ {10, 11, 12, 13, 14} یا HTTP 401
func (e *APIError) IsRateLimited() bool  // status == 20 یا HTTP 429
```

ثابت‌های کد وضعیت API (تمام مقادیر جدول بخش ۵ — دقیقاً با همین نام‌ها):

```go
const (
    StatusFailed                            = 0
    StatusSuccess                           = 1
    StatusInvalidAPIKey                     = 10
    StatusInactiveAPIKey                    = 11
    StatusAPIKeyIPRestricted                = 12
    StatusInactiveAccount                   = 13
    StatusSuspendedAccount                  = 14
    StatusPlanUpgradeRequired               = 15
    StatusInvalidParameter                  = 16
    StatusTooManyRequests                   = 20
    StatusInvalidLineNumber                 = 101
    StatusInsufficientCredit                = 102
    StatusEmptyMessageText                  = 103
    StatusInvalidMobileNumber               = 104
    StatusTooManyMobiles                    = 105
    StatusTooManyMessageTexts               = 106
    StatusEmptyMobileList                   = 107
    StatusEmptyTextList                     = 108
    StatusInvalidSendDateTime               = 109
    StatusMobileTextCountMismatch           = 110
    StatusSendNotFound                      = 111
    StatusNoRecordToDelete                  = 112
    StatusTemplateNotFound                  = 113
    StatusParameterValueTooLong             = 114
    StatusMobileInBlacklist                 = 115
    StatusMissingParameterName              = 116
    StatusTextNotApproved                   = 117
    StatusTooManyMessages                   = 118
    StatusCustomTemplatePlanUpgradeRequired = 119
    StatusLineActivationRequired            = 123
    StatusOnlyOTPTemplateAllowed            = 124
)
```

---

## ۴. مرجع کامل Endpoint ها و مدل‌ها

> نکته درباره‌ی تایپ‌ها: در پاسخ‌های سرور، شماره موبایل و شماره خط به‌صورت **عدد** برمی‌گردند → `int64`. در بدنه‌ی درخواست‌های ارسال، شماره موبایل‌ها **رشته** هستند (فرمت‌های `912xxxx677`، `0912...`، `00912...`، `+9891...` همه پذیرفته می‌شوند) → `[]string`. مقادیر پولی (`cost`, `credit`) → `float64`. شناسه پیامک → `int64`. شناسه Pack یک GUID رشته‌ای است → `string`.

### ۴.۱. ارسال گروهی — `POST /v1/send/bulk`

یک متن به حداکثر ۱۰۰ شماره. با `SendDateTime` می‌توان زمانبندی کرد.

قیود سرور: زمان گذشته نامعتبر؛ زمان معتبر از ۱ ساعت آینده تا ۳۶۵ روز آینده؛ حداکثر ۱۰۰ شماره.

```go
type SendBulkRequest struct {
    LineNumber   int64     `json:"lineNumber"`             // اجباری — شماره خط ارسالی
    MessageText  string    `json:"messageText"`            // اجباری — متن پیامک
    Mobiles      []string  `json:"mobiles"`                // اجباری — حداکثر 100 شماره
    SendDateTime *UnixTime `json:"sendDateTime,omitempty"` // اختیاری — خالی = ارسال فوری
}

type SendPackResult struct {
    PackID     string  `json:"packId"`     // Guid — شناسه یکتای مجموعه ارسال
    MessageIDs []int64 `json:"messageIds"` // شناسه یکتای هر پیامک
    Cost       float64 `json:"cost"`       // اعتبار مصرفی مجموعه
}
```

نمونه پاسخ سرور:

```json
{
  "status": 1,
  "message": "موفق",
  "data": {
    "packId": "2b99e63c-9bf8-4a21-9bfe-3f72dc1b46f1",
    "messageIds": [86522023, 86522024],
    "cost": 2.0
  }
}
```

### ۴.۲. ارسال نظیر به نظیر — `POST /v1/send/likeToLike`

هر شماره یک متن متفاوت. تعداد `MessageTexts` و `Mobiles` باید برابر باشد. همان قیود زمانبندی و سقف ۱۰۰ مورد بخش ۴.۱.

```go
type SendLikeToLikeRequest struct {
    LineNumber   int64     `json:"lineNumber"`
    MessageTexts []string  `json:"messageTexts"`
    Mobiles      []string  `json:"mobiles"`
    SendDateTime *UnixTime `json:"sendDateTime,omitempty"`
}
```

خروجی: همان `SendPackResult` بخش ۴.۱.

### ۴.۳. لغو ارسال زمانبندی‌شده — `DELETE /v1/send/scheduled/{packId}`

حداکثر تا ۳ دقیقه مانده به زمان ارسال قابل لغو است. `packId` در مسیر URL قرار می‌گیرد (با `url.PathEscape`).

```go
type CancelScheduledResult struct {
    ReturnedCreditCount float64 `json:"returnedCreditCount"` // اعتبار بازگشتی
    SmsCount            int     `json:"smsCount"`            // تعداد پیامک‌های لغوشده
}
```

نمونه پاسخ:

```json
{"status":1,"message":"موفق","data":{"returnedCreditCount":10.0,"smsCount":5}}
```

### ۴.۴. ارسال Verify — `POST /v1/send/verify`

ارسال پیامک قالب‌دار با اولویت بالا از خطوط خدماتی (کد تایید، فاکتور و …). قالب‌ها در پنل (بخش ارسال سریع) تعریف می‌شوند. حتی به شماره‌هایی که پیامک تبلیغاتی را مسدود کرده‌اند می‌رسد.

```go
type SendVerifyRequest struct {
    Mobile     string            `json:"mobile"`     // اجباری
    TemplateID int               `json:"templateId"` // اجباری — شناسه قالب پنل
    Parameters []VerifyParameter `json:"parameters"` // اجباری
}

type VerifyParameter struct {
    Name  string `json:"name"`  // کلید قالب، بدون # ابتدا/انتها
    Value string `json:"value"` // حداکثر 25 کاراکتر
}

type SendSingleResult struct {
    MessageID int64   `json:"messageId"`
    Cost      float64 `json:"cost"`
}
```

نمونه درخواست و پاسخ:

```json
{
  "mobile": "919xxxx904",
  "templateId": 123456,
  "parameters": [{"name": "Code", "value": "12345"}]
}
```

```json
{"status":1,"message":"موفق","data":{"messageId":89545112,"cost":1.0}}
```

### ۴.۵. ارسال از طریق URL — `GET /v1/send`

Endpoint قدیمی/ساده؛ احراز هویت با username/password در query string انجام می‌شود (**نه** هدر `X-API-KEY`؛ password همان کلید خصوصی پنل برنامه‌نویسان است). متد `SendViaURL` از baseURL و httpClient کلاینت استفاده می‌کند ولی هدر API-key نمی‌فرستد.

Query string: `username`, `password`, `line`, `mobile`, `text` — همگی اجباری و URL-encode شده.

```go
type SendViaURLParams struct {
    Username string // نام کاربری پنل
    Password string // کلید خصوصی
    Line     int64  // شماره خط
    Mobile   string // شماره موبایل مقصد
    Text     string // متن پیامک
}
```

خروجی: همان `SendSingleResult` بخش ۴.۴.

### ۴.۶. گزارش یک پیامک — `GET /v1/send/{messageId}`

وضعیت و اطلاعات یک پیامک ارسال‌شده با شناسه‌ی آن.

```go
type MessageReport struct {
    MessageID        int64          `json:"messageId"`
    Mobile           int64          `json:"mobile"`
    MessageText      string         `json:"messageText"`
    SendDateTime     UnixTime       `json:"sendDateTime"`
    LineNumber       int64          `json:"lineNumber"`
    Cost             float64        `json:"cost"`
    DeliveryState    *DeliveryState `json:"deliveryState"`    // nullable — تا قبل از دلیوری null
    DeliveryDateTime *UnixTime      `json:"deliveryDateTime"` // nullable
}
```

نمونه پاسخ:

```json
{
  "status": 1,
  "message": "موفق",
  "data": {
    "messageId": 89545112,
    "mobile": 9121234677,
    "messageText": "...",
    "sendDateTime": 1628683626,
    "lineNumber": 30004505000017,
    "cost": 1.0,
    "deliveryState": 1,
    "deliveryDateTime": 1628683629
  }
}
```

### ۴.۷. گزارش مجموعه ارسال‌های روز — `GET /v1/send/pack`

لیست Pack های روز جاری. Query اختیاری: `pageSize` (حداکثر و پیش‌فرض 100)، `pageNumber` (پیش‌فرض 1).

```go
type PackSummary struct {
    PackID           string   `json:"packId"`
    RecipientCount   int      `json:"recipientCount"`
    CreationDateTime UnixTime `json:"creationDateTime"`
}
```

`data` آرایه‌ای از این مدل است.

### ۴.۸. گزارش یک مجموعه ارسال — `GET /v1/send/pack/{packId}`

تمام پیامک‌های یک Pack با وضعیت‌هایشان. `data` آرایه‌ای از `MessageReport` (بخش ۴.۶) است.

### ۴.۹. گزارش ارسال‌های روز — `GET /v1/send/live`

ارسال‌های روز جاری. Query اختیاری: `pageSize`, `pageNumber` (مثل ۴.۷). خروجی: آرایه‌ی `MessageReport`.

### ۴.۱۰. گزارش ارسال‌های آرشیو — `GET /v1/send/archive`

ارسال‌های گذشته (تا انتهای روز قبل). Query همگی اختیاری: `fromDate`, `toDate` (هر دو Unix seconds)، `pageSize`, `pageNumber`.

```go
type ArchiveReportParams struct {
    PageParams
    FromDate *time.Time // در query به ثانیه‌ی Unix تبدیل می‌شود
    ToDate   *time.Time
}
```

خروجی: آرایه‌ی `MessageReport`.

### ۴.۱۱. تازه‌ترین پیامک‌های دریافتی — `GET /v1/receive/latest`

⚠️ **مصرف‌شونده (destructive read):** هر پیامک دریافتی فقط یک بار از این endpoint قابل خواندن است؛ بعد از خواندن، «خوانده‌شده» می‌شود و دیگر از این متد برنمی‌گردد. این نکته باید در doc comment متد `LatestReceived` صریحاً ذکر شود.

Query اختیاری: `count` (حداکثر و پیش‌فرض 100). مقدار `count <= 0` یعنی ارسال نشود.

```go
type ReceivedMessage struct {
    // ReceiveReturnID در endpoint های latest و archive برمی‌گردد؛
    // در live وجود ندارد و صفر می‌ماند.
    ReceiveReturnID  int64    `json:"receiveReturnId"`
    MessageText      string   `json:"messageText"`
    Number           int64    `json:"number"` // شماره خط دریافت‌کننده
    Mobile           int64    `json:"mobile"` // شماره موبایل ارسال‌کننده
    ReceivedDateTime UnixTime `json:"receivedDateTime"`
}
```

نمونه پاسخ:

```json
{
  "status": 1,
  "message": "موفق",
  "data": [
    {
      "receiveReturnId": 123456789,
      "messageText": "...",
      "number": 30004505000017,
      "mobile": 9121234002,
      "receivedDateTime": 1628683625
    }
  ]
}
```

### ۴.۱۲. پیامک‌های دریافتی روز — `GET /v1/receive/live`

دریافتی‌های روز جاری (خوانده‌شده و نشده). در ساعات ابتدایی روز، دریافتی‌های روز قبل هم برمی‌گردد. مدل بازگشتی این endpoint فیلد `receiveReturnId` **ندارد**.

Query اختیاری: `pageSize` (حداکثر/پیش‌فرض 100)، `pageNumber` (پیش‌فرض 1)، `sortByNewest` (bool، پیش‌فرض false = صعودی)، `mobile` (فیلتر شماره‌ی ارسال‌کننده).

```go
type TodayReceivedParams struct {
    PageParams
    SortByNewest bool
    Mobile       string
}
```

خروجی: آرایه‌ی `ReceivedMessage` (با `ReceiveReturnID == 0`).

### ۴.۱۳. پیامک‌های دریافتی آرشیو — `GET /v1/receive/archive`

دریافتی‌های گذشته (تا انتهای روز قبل). Query اختیاری: `fromDate`, `toDate` (Unix seconds)، `pageSize`, `pageNumber`, `mobile`.

```go
type ArchiveReceivedParams struct {
    PageParams
    FromDate *time.Time
    ToDate   *time.Time
    Mobile   string
}
```

خروجی: آرایه‌ی `ReceivedMessage`.

### ۴.۱۴. اعتبار فعلی — `GET /v1/credit`

`data` یک عدد اعشاری ساده است:

```json
{"status":1,"message":"موفق","data":165.3}
```

امضا: `Credit(ctx) (float64, error)`.

### ۴.۱۵. لیست خطوط — `GET /v1/line`

`data` آرایه‌ای از شماره خط (عدد):

```json
{"status":1,"message":"موفق","data":[10002155613464, 30004505000017]}
```

امضا: `Lines(ctx) ([]int64, error)`.

---

## ۵. جدول کامل کدهای وضعیت API

این جدول باید عیناً به ثابت‌های بخش ۳.۴ نگاشت شود و در README هم بیاید.

| کد | توضیح |
|---|---|
| 0 | درخواست شما با خطا مواجه شده‌است |
| 1 | عملیات با موفقیت انجام شد |
| 10 | کلید وب‌سرویس نامعتبر است |
| 11 | کلید وب‌سرویس غیرفعال است |
| 12 | کلید وب‌سرویس محدود به آی‌پی‌های تعریف‌شده است |
| 13 | حساب کاربری غیرفعال است |
| 14 | حساب کاربری در حالت تعلیق قرار دارد |
| 15 | به منظور استفاده از وب‌سرویس پلن خود را ارتقا دهید |
| 16 | مقدار ارسالی پارامتر نادرست است |
| 20 | تعداد درخواست بیشتر از حد مجاز است |
| 101 | شماره خط نامعتبر است |
| 102 | اعتبار کافی نیست |
| 103 | درخواست دارای متن(های) خالی است |
| 104 | درخواست دارای موبایل(های) نادرست است |
| 105 | تعداد موبایل‌ها بیشتر از حد مجاز (100) است |
| 106 | تعداد متن‌ها بیشتر از حد مجاز (100) است |
| 107 | لیست موبایل‌ها خالی است |
| 108 | لیست متن‌ها خالی است |
| 109 | زمان ارسال نامعتبر است |
| 110 | تعداد شماره موبایل‌ها و متن‌ها برابر نیست |
| 111 | با این شناسه، ارسالی ثبت نشده است |
| 112 | رکوردی برای حذف یافت نشد |
| 113 | قالب یافت نشد |
| 114 | طول مقدار پارامتر بیش از حد مجاز (25 کاراکتر) است |
| 115 | شماره موبایل(ها) در لیست سیاه سامانه است |
| 116 | نام یک یا چند پارامتر مقداردهی نشده است |
| 117 | متن ارسال‌شده مورد تایید نیست |
| 118 | تعداد پیام‌ها بیشتر از حد مجاز است |
| 119 | برای قالب شخصی‌سازی‌شده پلن خود را ارتقا دهید |
| 123 | خط ارسال‌کننده نیاز به فعال‌سازی دارد |
| 124 | فقط امکان ارسال پیامک OTP وجود دارد و قالب شما OTP شناسایی نشده است |

---

## ۶. تست‌ها

- تمام تست‌ها با `net/http/httptest` نوشته می‌شوند؛ **هیچ درخواست واقعی به شبکه زده نمی‌شود.**
- برای هر endpoint حداقل این موارد تست شود:
  1. صحت method و path و query string ساخته‌شده.
  2. ارسال هدرهای `X-API-KEY` و `Accept` (و `Content-Type` برای POST ها).
  3. سریالایز صحیح بدنه‌ی درخواست (شامل حذف `sendDateTime` وقتی nil است، و مقدار Unix صحیح وقتی ست شده).
  4. decode صحیح پاسخ موفق با نمونه‌های JSON همین سند (شامل فیلدهای nullable: `deliveryState`/`deliveryDateTime` هم null و هم مقداردار).
  5. نگاشت پاسخ خطا (مثلاً `{"status":102,...}` با HTTP 400) به `*APIError` با فیلدهای درست، و بررسی `errors.As`.
- تست‌های مشترک در `client_test.go`: پاسخ non-JSON، HTTP 500 با بدنه خالی، لغو context، رفتار `WithBaseURL` با اسلش انتهایی، helper های `IsAuthError`/`IsRateLimited`، مارشال/آن‌مارشال `UnixTime`.
- `examples_test.go` شامل `Example` function های قابل‌نمایش در godoc برای `SendBulk`، `SendVerify` و `Credit` (با کلاینت ساختگی یا فقط کد نمایشی که کامپایل شود).
- هدف: `go test ./...` سبز، بدون flag اضافه.

---

## ۷. README.md

README (به انگلیسی، با یک بخش خلاصه فارسی) شامل:

1. معرفی کوتاه و badge وضعیت (اختیاری).
2. نصب: `go get github.com/darinehgroup/sms-ir-sdk-go`.
3. Quick start — ارسال Verify:

```go
client := smsir.New(os.Getenv("SMSIR_API_KEY"))

result, err := client.SendVerify(ctx, &smsir.SendVerifyRequest{
    Mobile:     "9120000000",
    TemplateID: 123456,
    Parameters: []smsir.VerifyParameter{{Name: "Code", Value: "12345"}},
})
```

4. نمونه‌ی ارسال گروهی و ارسال زمانبندی‌شده (با `smsir.NewUnixTime(time.Now().Add(2*time.Hour))`).
5. جدول نگاشت تمام متدهای SDK به endpoint ها.
6. بخش مدیریت خطا (نمونه `errors.As` با `*smsir.APIError` و جدول کدهای وضعیت).
7. بخش Sandbox (خلاصه‌ی بخش ۲.۲ همین سند، شامل قالب `123456` با پارامتر `Code`).
8. توضیح نکته‌ی مصرف‌شونده بودن `LatestReceived`.

---

## ۸. معیارهای پذیرش (چک‌لیست نهایی)

- [ ] `go build ./...` و `go vet ./...` و `go test ./...` بدون خطا.
- [ ] `go.mod` فقط شامل ماژول خود پروژه (بدون require خارجی).
- [ ] هر ۱۵ متد بخش ۳.۲ پیاده‌سازی و تست شده‌اند.
- [ ] تمام تایپ‌های بخش ۴ با همان نام‌ها و JSON tag های ذکرشده وجود دارند.
- [ ] تمام ثابت‌های کد وضعیت (بخش ۳.۴) و دلیوری (بخش ۳.۳) تعریف شده‌اند.
- [ ] تمام identifier های عمومی doc comment دارند.
- [ ] README مطابق بخش ۷ نوشته شده است.
- [ ] فایل LICENSE با متن MIT موجود است.
