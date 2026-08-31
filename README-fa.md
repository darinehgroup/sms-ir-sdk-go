<div dir="rtl">

# sms-ir-sdk-go

کلاینت Go برای وب‌سرویس [SMS.ir](https://sms.ir) — **بدون هیچ وابستگی خارجی**، فقط با کتابخانه استاندارد Go. ارسال (گروهی، مانند‌به‌مانند، verify، لغو زمان‌بندی‌شده)، گزارش تحویل، پیامک‌های دریافتی و تنظیمات حساب (اعتبار، خطوط).

نسخهٔ انگلیسی این سند (مرجع کامل متدها و جدول کدهای خطا): [README.md](README.md)

[![Go Reference](https://pkg.go.dev/badge/github.com/darinehgroup/sms-ir-sdk-go.svg)](https://pkg.go.dev/github.com/darinehgroup/sms-ir-sdk-go)
[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

## ویژگی‌ها

- **صفر وابستگی خارجی** — فقط کتابخانه استاندارد
- **Context-aware** — همه متدها آرگومان اول `context.Context` می‌گیرند
- **کلاینت تغییرناپذیر** — پس از ساخت، برای استفاده همزمان امن است
- **خطاهای typed** — خطاها به‌صورت `*smsir.APIError` با کد وضعیت API برمی‌گردند

## نصب

```bash
go get github.com/darinehgroup/sms-ir-sdk-go
```

نام پکیج هنگام import برابر `smsir` است:

```go
import "github.com/darinehgroup/sms-ir-sdk-go"
```

## شروع سریع

کلاینت را با کلید API (از پنل توسعه‌دهندگان SMS.ir) بسازید:

```go
client := smsir.New(os.Getenv("SMSIR_API_KEY"))

res, err := client.SendVerify(ctx, &smsir.SendVerifyRequest{
    Mobile:     "9120000000",
    TemplateID: 123456,
    Parameters: []smsir.VerifyParameter{{Name: "Code", Value: "12345"}},
})
```

## نکات کلیدی

- **محیط Sandbox:** URL جداگانه ندارد؛ فقط کلید Sandbox را به `smsir.New` بدهید. در Sandbox فقط قالب `123456` با پارامتر `Code` فعال است و پیامک واقعی ارسال نمی‌شود.
- **ارورها:** خطاهای سرور به‌صورت `*smsir.APIError` برمی‌گردند. با `errors.As` آن‌ها را بگیرید و با `IsAuthError()` / `IsRateLimited()` شاخه‌بندی کنید. کد وضعیت در فیلد `Status` (جدول کامل در نسخهٔ انگلیسی) و کد HTTP در `HTTPStatus` قرار دارد.
- **زمان‌ها:** همه زمان‌ها به‌صورت Unix ثانیه (UTC) هستند و در SDK با `smsir.UnixTime` نمایش داده می‌شوند.
- **`LatestReceived` مصرف‌شونده است:** هر پیامک دریافتی فقط یک بار از این متد خوانده می‌شود و سپس «خوانده‌شده» می‌گردد؛ برای خواندن مجدد از `TodayReceived` یا `ArchiveReceived` استفاده کنید.
- **ارسال زمان‌بندی‌شده:** با `SendDateTime` از ۱ ساعت تا ۳۶۵ روز بعد قابل زمان‌بندی است و تا ۳ دقیقه مانده به ارسال با `CancelScheduledSend` قابل لغو.

## لایسنس

[MIT](LICENSE) © Darineh Group

مرجع طراحی و پیاده‌سازی: [DESIGN.md](DESIGN.md)

</div>
