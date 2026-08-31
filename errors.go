package smsir

import (
	"fmt"
	"net/http"
)

// API status codes returned in the response envelope's status field.
const (
	StatusFailed                            = 0   // درخواست شما با خطا مواجه شده‌است
	StatusSuccess                           = 1   // عملیات با موفقیت انجام شد
	StatusInvalidAPIKey                     = 10  // کلید وب‌سرویس نامعتبر است
	StatusInactiveAPIKey                    = 11  // کلید وب‌سرویس غیرفعال است
	StatusAPIKeyIPRestricted                = 12  // کلید وب‌سرویس محدود به آی‌پی‌های تعریف‌شده است
	StatusInactiveAccount                   = 13  // حساب کاربری غیرفعال است
	StatusSuspendedAccount                  = 14  // حساب کاربری در حالت تعلیق قرار دارد
	StatusPlanUpgradeRequired               = 15  // به منظور استفاده از وب‌سرویس پلن خود را ارتقا دهید
	StatusInvalidParameter                  = 16  // مقدار ارسالی پارامتر نادرست است
	StatusTooManyRequests                   = 20  // تعداد درخواست بیشتر از حد مجاز است
	StatusInvalidLineNumber                 = 101 // شماره خط نامعتبر است
	StatusInsufficientCredit                = 102 // اعتبار کافی نیست
	StatusEmptyMessageText                  = 103 // درخواست دارای متن(های) خالی است
	StatusInvalidMobileNumber               = 104 // درخواست دارای موبایل(های) نادرست است
	StatusTooManyMobiles                    = 105 // تعداد موبایل‌ها بیشتر از حد مجاز (100) است
	StatusTooManyMessageTexts               = 106 // تعداد متن‌ها بیشتر از حد مجاز (100) است
	StatusEmptyMobileList                   = 107 // لیست موبایل‌ها خالی است
	StatusEmptyTextList                     = 108 // لیست متن‌ها خالی است
	StatusInvalidSendDateTime               = 109 // زمان ارسال نامعتبر است
	StatusMobileTextCountMismatch           = 110 // تعداد شماره موبایل‌ها و متن‌ها برابر نیست
	StatusSendNotFound                      = 111 // با این شناسه، ارسالی ثبت نشده است
	StatusNoRecordToDelete                  = 112 // رکوردی برای حذف یافت نشد
	StatusTemplateNotFound                  = 113 // قالب یافت نشد
	StatusParameterValueTooLong             = 114 // طول مقدار پارامتر بیش از حد مجاز (25 کاراکتر) است
	StatusMobileInBlacklist                 = 115 // شماره موبایل(ها) در لیست سیاه سامانه است
	StatusMissingParameterName              = 116 // نام یک یا چند پارامتر مقداردهی نشده است
	StatusTextNotApproved                   = 117 // متن ارسال‌شده مورد تایید نیست
	StatusTooManyMessages                   = 118 // تعداد پیام‌ها بیشتر از حد مجاز است
	StatusCustomTemplatePlanUpgradeRequired = 119 // برای قالب شخصی‌سازی‌شده پلن خود را ارتقا دهید
	StatusLineActivationRequired            = 123 // خط ارسال‌کننده نیاز به فعال‌سازی دارد
	StatusOnlyOTPTemplateAllowed            = 124 // فقط امکان ارسال پیامک OTP وجود دارد و قالب شما OTP شناسایی نشده است
)

// APIError represents an error returned by the SMS.ir API.
type APIError struct {
	// HTTPStatus is the HTTP status code of the response (e.g. 400, 401, 429).
	HTTPStatus int
	// Status is the API status code from the response envelope. It is -1 when
	// the response body was not a valid envelope.
	Status int
	// Message is the server-provided description (usually in Persian), or a
	// truncated copy of the raw body when the envelope could not be decoded.
	Message string
}

// Error implements the error interface.
func (e *APIError) Error() string {
	return fmt.Sprintf("smsir: api error (http %d, status %d): %s", e.HTTPStatus, e.Status, e.Message)
}

// IsAuthError reports whether the error relates to authentication or account
// state (invalid/inactive/IP-restricted key, inactive or suspended account,
// or an HTTP 401 response).
func (e *APIError) IsAuthError() bool {
	switch e.Status {
	case StatusInvalidAPIKey, StatusInactiveAPIKey, StatusAPIKeyIPRestricted,
		StatusInactiveAccount, StatusSuspendedAccount:
		return true
	}
	return e.HTTPStatus == http.StatusUnauthorized
}

// IsRateLimited reports whether the request was rejected due to rate limiting
// (API status 20 or HTTP 429).
func (e *APIError) IsRateLimited() bool {
	return e.Status == StatusTooManyRequests || e.HTTPStatus == http.StatusTooManyRequests
}
