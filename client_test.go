package smsir

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

const testAPIKey = "test-api-key"

// newTestClient builds a Client pointed at ts and returns it.
func newTestClient(t *testing.T, ts *httptest.Server) *Client {
	t.Helper()
	return New(testAPIKey, WithBaseURL(ts.URL), WithHTTPClient(ts.Client()))
}

// assertHeader fails the test if the request did not carry header h with value v.
func assertHeader(t *testing.T, r *http.Request, h, v string) {
	t.Helper()
	if got := r.Header.Get(h); got != v {
		t.Errorf("header %q = %q, want %q", h, got, v)
	}
}

func TestNew_Defaults(t *testing.T) {
	c := New("k")
	if c.apiKey != "k" {
		t.Errorf("apiKey = %q", c.apiKey)
	}
	if c.baseURL != DefaultBaseURL {
		t.Errorf("baseURL = %q, want %q", c.baseURL, DefaultBaseURL)
	}
	if c.httpClient == nil || c.httpClient.Timeout != 30*time.Second {
		t.Errorf("httpClient timeout = %v, want 30s", c.httpClient.Timeout)
	}
	if c.userAgent != "smsir-go/"+Version {
		t.Errorf("userAgent = %q", c.userAgent)
	}
}

func TestWithBaseURL_TrimsTrailingSlash(t *testing.T) {
	c := New("k", WithBaseURL("https://example.com/"))
	if c.baseURL != "https://example.com" {
		t.Errorf("baseURL = %q", c.baseURL)
	}
}

func TestWithUserAgent(t *testing.T) {
	c := New("k", WithUserAgent("custom/1"))
	if c.userAgent != "custom/1" {
		t.Errorf("userAgent = %q", c.userAgent)
	}
}

func TestDo_SuccessDecodesData(t *testing.T) {
	var gotPath, gotMethod string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"status":1,"message":"موفق","data":{"packId":"abc","messageIds":[1,2],"cost":2.0}}`)
	}))
	defer ts.Close()

	c := newTestClient(t, ts)
	var out SendPackResult
	if err := c.do(context.Background(), http.MethodPost, "/v1/send/bulk", nil, struct{ X int }{1}, &out); err != nil {
		t.Fatalf("do: %v", err)
	}
	if gotPath != "/v1/send/bulk" {
		t.Errorf("path = %q", gotPath)
	}
	if gotMethod != http.MethodPost {
		t.Errorf("method = %q", gotMethod)
	}
	if out.PackID != "abc" || len(out.MessageIDs) != 2 || out.Cost != 2.0 {
		t.Errorf("decoded = %+v", out)
	}
}

func TestDo_SetsHeadersForPostBody(t *testing.T) {
	var got *http.Request
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"status":1,"data":null}`)
	}))
	defer ts.Close()

	c := newTestClient(t, ts)
	_ = c.do(context.Background(), http.MethodPost, "/v1/send/bulk", nil, struct{ X int }{1}, nil)

	if got == nil {
		t.Fatal("no request captured")
	}
	assertHeader(t, got, "X-API-KEY", testAPIKey)
	assertHeader(t, got, "Accept", "application/json")
	assertHeader(t, got, "Content-Type", "application/json")
	if ua := got.Header.Get("User-Agent"); !strings.HasPrefix(ua, "smsir-go/") {
		t.Errorf("User-Agent = %q", ua)
	}
}

func TestDo_SetsHeadersForGetNoBody(t *testing.T) {
	var got *http.Request
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"status":1,"data":null}`)
	}))
	defer ts.Close()

	c := newTestClient(t, ts)
	_ = c.do(context.Background(), http.MethodGet, "/v1/credit", nil, nil, nil)

	if got == nil {
		t.Fatal("no request captured")
	}
	assertHeader(t, got, "X-API-KEY", testAPIKey)
	assertHeader(t, got, "Accept", "application/json")
	if ct := got.Header.Get("Content-Type"); ct != "" {
		t.Errorf("GET should not set Content-Type, got %q", ct)
	}
}

func TestDo_APIErrorOnStatusNotOne(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, `{"status":102,"message":"اعتبار کافی نمیباشد","data":null}`)
	}))
	defer ts.Close()

	c := newTestClient(t, ts)
	err := c.do(context.Background(), http.MethodGet, "/v1/credit", nil, nil, new(float64))

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("err = %v, want *APIError", err)
	}
	if apiErr.Status != StatusInsufficientCredit {
		t.Errorf("Status = %d, want %d", apiErr.Status, StatusInsufficientCredit)
	}
	if apiErr.HTTPStatus != http.StatusBadRequest {
		t.Errorf("HTTPStatus = %d", apiErr.HTTPStatus)
	}
	if apiErr.Message != "اعتبار کافی نمیباشد" {
		t.Errorf("Message = %q", apiErr.Message)
	}
	wantStr := "smsir: api error (http 400, status 102): اعتبار کافی نمیباشد"
	if apiErr.Error() != wantStr {
		t.Errorf("Error() = %q, want %q", apiErr.Error(), wantStr)
	}
}

func TestDo_NonJSONBody(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		io.WriteString(w, "<html>502 bad gateway</html>")
	}))
	defer ts.Close()

	c := newTestClient(t, ts)
	err := c.do(context.Background(), http.MethodGet, "/v1/credit", nil, nil, new(float64))

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("err = %v, want *APIError", err)
	}
	if apiErr.HTTPStatus != http.StatusBadGateway {
		t.Errorf("HTTPStatus = %d", apiErr.HTTPStatus)
	}
	if apiErr.Status != -1 {
		t.Errorf("Status = %d, want -1", apiErr.Status)
	}
	if !strings.Contains(apiErr.Message, "502") {
		t.Errorf("Message = %q", apiErr.Message)
	}
}

func TestDo_HTTP500EmptyBody(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	c := newTestClient(t, ts)
	err := c.do(context.Background(), http.MethodGet, "/v1/credit", nil, nil, new(float64))

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("err = %v, want *APIError", err)
	}
	if apiErr.HTTPStatus != http.StatusInternalServerError {
		t.Errorf("HTTPStatus = %d", apiErr.HTTPStatus)
	}
	if apiErr.Status != -1 {
		t.Errorf("Status = %d, want -1", apiErr.Status)
	}
}

func TestDo_ContextCanceled(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Block long enough for the caller to cancel.
		time.Sleep(50 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"status":1,"data":null}`)
	}))
	defer ts.Close()

	c := newTestClient(t, ts)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer cancel()

	err := c.do(ctx, http.MethodGet, "/v1/credit", nil, nil, new(float64))
	if err == nil {
		t.Fatal("err = nil, want a context/timeout error")
	}
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		t.Fatalf("err should not be *APIError, got %v", err)
	}
}

func TestDo_TruncatesRawBody(t *testing.T) {
	long := strings.Repeat("x", 1200)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		io.WriteString(w, long)
	}))
	defer ts.Close()

	c := newTestClient(t, ts)
	err := c.do(context.Background(), http.MethodGet, "/v1/credit", nil, nil, new(float64))

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("err = %v, want *APIError", err)
	}
	if len(apiErr.Message) > 516 { // 512 + "...": trimming may drop trailing whitespace
		t.Errorf("Message not truncated, len = %d", len(apiErr.Message))
	}
}

func TestAPIError_IsAuthError(t *testing.T) {
	cases := []struct {
		e    *APIError
		want bool
	}{
		{&APIError{Status: StatusInvalidAPIKey}, true},
		{&APIError{Status: StatusInactiveAPIKey}, true},
		{&APIError{Status: StatusAPIKeyIPRestricted}, true},
		{&APIError{Status: StatusInactiveAccount}, true},
		{&APIError{Status: StatusSuspendedAccount}, true},
		{&APIError{HTTPStatus: http.StatusUnauthorized}, true},
		{&APIError{Status: StatusInsufficientCredit}, false},
		{&APIError{Status: StatusTooManyRequests}, false},
	}
	for _, tc := range cases {
		if got := tc.e.IsAuthError(); got != tc.want {
			t.Errorf("IsAuthError(%+v) = %v, want %v", tc.e, got, tc.want)
		}
	}
}

func TestAPIError_IsRateLimited(t *testing.T) {
	cases := []struct {
		e    *APIError
		want bool
	}{
		{&APIError{Status: StatusTooManyRequests}, true},
		{&APIError{HTTPStatus: http.StatusTooManyRequests}, true},
		{&APIError{Status: StatusInsufficientCredit}, false},
		{&APIError{HTTPStatus: http.StatusBadRequest}, false},
	}
	for _, tc := range cases {
		if got := tc.e.IsRateLimited(); got != tc.want {
			t.Errorf("IsRateLimited(%+v) = %v, want %v", tc.e, got, tc.want)
		}
	}
}

func TestUnixTime_MarshalUnmarshal(t *testing.T) {
	tm := time.Unix(1628683626, 0).UTC()
	u := UnixTime(tm)

	b, err := u.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON: %v", err)
	}
	if string(b) != "1628683626" {
		t.Errorf("MarshalJSON = %s, want 1628683626", b)
	}

	var got UnixTime
	if err := got.UnmarshalJSON([]byte("1628683626")); err != nil {
		t.Fatalf("UnmarshalJSON: %v", err)
	}
	if !got.Time().Equal(tm) {
		t.Errorf("round trip = %v, want %v", got.Time(), tm)
	}
}

func TestUnixTime_UnmarshalNull(t *testing.T) {
	var u UnixTime
	if err := u.UnmarshalJSON([]byte("null")); err != nil {
		t.Fatalf("UnmarshalJSON(null): %v", err)
	}
}

func TestUnixTime_UnmarshalInvalid(t *testing.T) {
	var u UnixTime
	if err := u.UnmarshalJSON([]byte("\"oops\"")); err == nil {
		t.Fatal("UnmarshalJSON expected error, got nil")
	}
}

func TestUnixTime_String(t *testing.T) {
	u := UnixTime(time.Unix(1628683626, 0).UTC())
	if got := u.String(); got != "2021-08-11T12:07:06Z" {
		t.Errorf("String() = %q", got)
	}
}

func TestPageParams_Apply_NilSafe(t *testing.T) {
	// nil receiver must not panic and must not set anything.
	var p *PageParams
	q := url.Values{}
	p.apply(q)
	if len(q) != 0 {
		t.Errorf("nil PageParams should add nothing, got %v", q)
	}
}

func TestPageParams_Apply_NonZeroFields(t *testing.T) {
	p := &PageParams{PageSize: 50, PageNumber: 2}
	q := url.Values{}
	p.apply(q)
	if got := q.Get("pageSize"); got != "50" {
		t.Errorf("pageSize = %q", got)
	}
	if got := q.Get("pageNumber"); got != "2" {
		t.Errorf("pageNumber = %q", got)
	}
}

func TestPageParams_Apply_ZeroFieldsOmitted(t *testing.T) {
	p := &PageParams{}
	q := url.Values{}
	p.apply(q)
	if len(q) != 0 {
		t.Errorf("zero PageParams should add nothing, got %v", q)
	}
}
