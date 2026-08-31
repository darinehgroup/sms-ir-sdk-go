package smsir

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSendBulk_Success(t *testing.T) {
	var got *http.Request
	var gotBody []byte
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r
		gotBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"status":1,"message":"موفق","data":{"packId":"2b99e63c-9bf8-4a21-9bfe-3f72dc1b46f1","messageIds":[86522023,86522024],"cost":2.0}}`)
	}))
	defer ts.Close()

	c := newTestClient(t, ts)
	res, err := c.SendBulk(context.Background(), &SendBulkRequest{
		LineNumber:  30004505000017,
		MessageText: "سلام",
		Mobiles:     []string{"9120000000", "9120000001"},
	})
	if err != nil {
		t.Fatalf("SendBulk: %v", err)
	}
	if got.Method != http.MethodPost {
		t.Errorf("method = %q", got.Method)
	}
	if got.URL.Path != "/v1/send/bulk" {
		t.Errorf("path = %q", got.URL.Path)
	}
	assertHeader(t, got, "X-API-KEY", testAPIKey)
	assertHeader(t, got, "Content-Type", "application/json")

	var body map[string]any
	if err := json.Unmarshal(gotBody, &body); err != nil {
		t.Fatalf("body not json: %v", err)
	}
	if v, ok := body["sendDateTime"]; ok {
		t.Errorf("sendDateTime should be omitted when nil, got %v", v)
	}
	if body["lineNumber"].(float64) != 30004505000017 {
		t.Errorf("lineNumber = %v", body["lineNumber"])
	}

	if res.PackID != "2b99e63c-9bf8-4a21-9bfe-3f72dc1b46f1" {
		t.Errorf("PackID = %q", res.PackID)
	}
	if len(res.MessageIDs) != 2 || res.MessageIDs[0] != 86522023 {
		t.Errorf("MessageIDs = %v", res.MessageIDs)
	}
	if res.Cost != 2.0 {
		t.Errorf("Cost = %v", res.Cost)
	}
}

func TestSendBulk_ScheduledIncludesUnixTime(t *testing.T) {
	var gotBody []byte
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"status":1,"data":{"packId":"p","messageIds":[1],"cost":1.0}}`)
	}))
	defer ts.Close()

	when := time.Now().Add(2 * time.Hour).UTC().Truncate(time.Second)
	c := newTestClient(t, ts)
	_, err := c.SendBulk(context.Background(), &SendBulkRequest{
		LineNumber:   30004505000017,
		MessageText:  "x",
		Mobiles:      []string{"9120000000"},
		SendDateTime: NewUnixTime(when),
	})
	if err != nil {
		t.Fatalf("SendBulk: %v", err)
	}
	var body map[string]any
	if err := json.Unmarshal(gotBody, &body); err != nil {
		t.Fatalf("body not json: %v", err)
	}
	v, ok := body["sendDateTime"]
	if !ok {
		t.Fatal("sendDateTime missing")
	}
	if int64(v.(float64)) != when.Unix() {
		t.Errorf("sendDateTime = %v, want %d", v, when.Unix())
	}
}

func TestSendBulk_NilRequest(t *testing.T) {
	c := New(testAPIKey)
	_, err := c.SendBulk(context.Background(), nil)
	if !errors.Is(err, errNilRequest) {
		t.Errorf("err = %v, want errNilRequest", err)
	}
}

func TestSendBulk_APIError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, `{"status":102,"message":"اعتبار کافی نیست","data":null}`)
	}))
	defer ts.Close()

	c := newTestClient(t, ts)
	_, err := c.SendBulk(context.Background(), &SendBulkRequest{
		LineNumber: 1, MessageText: "x", Mobiles: []string{"9"},
	})
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("err = %v, want *APIError", err)
	}
	if apiErr.Status != StatusInsufficientCredit {
		t.Errorf("Status = %d", apiErr.Status)
	}
}

func TestSendLikeToLike_Success(t *testing.T) {
	var gotBody []byte
	var gotPath, gotMethod string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"status":1,"data":{"packId":"p","messageIds":[10,11],"cost":2.0}}`)
	}))
	defer ts.Close()

	c := newTestClient(t, ts)
	res, err := c.SendLikeToLike(context.Background(), &SendLikeToLikeRequest{
		LineNumber:   30004505000017,
		MessageTexts: []string{"a", "b"},
		Mobiles:      []string{"9120000000", "9120000001"},
	})
	if err != nil {
		t.Fatalf("SendLikeToLike: %v", err)
	}
	if gotPath != "/v1/send/likeToLike" {
		t.Errorf("path = %q", gotPath)
	}
	if gotMethod != http.MethodPost {
		t.Errorf("method = %q", gotMethod)
	}
	var body map[string]any
	_ = json.Unmarshal(gotBody, &body)
	if _, ok := body["sendDateTime"]; ok {
		t.Error("sendDateTime should be omitted when nil")
	}
	if len(res.MessageIDs) != 2 || res.MessageIDs[1] != 11 {
		t.Errorf("MessageIDs = %v", res.MessageIDs)
	}
}

func TestSendLikeToLike_NilRequest(t *testing.T) {
	c := New(testAPIKey)
	_, err := c.SendLikeToLike(context.Background(), nil)
	if !errors.Is(err, errNilRequest) {
		t.Errorf("err = %v", err)
	}
}

func TestSendVerify_Success(t *testing.T) {
	var gotPath, gotMethod string
	var gotBody []byte
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"status":1,"message":"موفق","data":{"messageId":89545112,"cost":1.0}}`)
	}))
	defer ts.Close()

	c := newTestClient(t, ts)
	res, err := c.SendVerify(context.Background(), &SendVerifyRequest{
		Mobile:     "9190000004",
		TemplateID: 123456,
		Parameters: []VerifyParameter{{Name: "Code", Value: "12345"}},
	})
	if err != nil {
		t.Fatalf("SendVerify: %v", err)
	}
	if gotPath != "/v1/send/verify" || gotMethod != http.MethodPost {
		t.Errorf("method/path = %q %q", gotMethod, gotPath)
	}
	var body map[string]any
	_ = json.Unmarshal(gotBody, &body)
	if body["mobile"] != "9190000004" {
		t.Errorf("mobile = %v", body["mobile"])
	}
	if int(body["templateId"].(float64)) != 123456 {
		t.Errorf("templateId = %v", body["templateId"])
	}
	params := body["parameters"].([]any)
	if len(params) != 1 {
		t.Fatalf("parameters len = %d", len(params))
	}
	p0 := params[0].(map[string]any)
	if p0["name"] != "Code" || p0["value"] != "12345" {
		t.Errorf("parameter[0] = %v", p0)
	}
	if res.MessageID != 89545112 || res.Cost != 1.0 {
		t.Errorf("result = %+v", res)
	}
}

func TestSendVerify_NilRequest(t *testing.T) {
	c := New(testAPIKey)
	_, err := c.SendVerify(context.Background(), nil)
	if !errors.Is(err, errNilRequest) {
		t.Errorf("err = %v", err)
	}
}

func TestSendViaURL_Success(t *testing.T) {
	var got *http.Request
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"status":1,"message":"موفق","data":{"messageId":89545112,"cost":1.0}}`)
	}))
	defer ts.Close()

	c := newTestClient(t, ts)
	res, err := c.SendViaURL(context.Background(), &SendViaURLParams{
		Username: "user",
		Password: "secret",
		Line:     30004505000017,
		Mobile:   "9120000000",
		Text:     "hello world",
	})
	if err != nil {
		t.Fatalf("SendViaURL: %v", err)
	}
	if got.Method != http.MethodGet {
		t.Errorf("method = %q", got.Method)
	}
	if got.URL.Path != "/v1/send" {
		t.Errorf("path = %q", got.URL.Path)
	}
	// No X-API-KEY on this endpoint.
	if v := got.Header.Get("X-API-KEY"); v != "" {
		t.Errorf("X-API-KEY should be empty, got %q", v)
	}
	q := got.URL.Query()
	if q.Get("username") != "user" {
		t.Errorf("username = %q", q.Get("username"))
	}
	if q.Get("password") != "secret" {
		t.Errorf("password = %q", q.Get("password"))
	}
	if q.Get("line") != "30004505000017" {
		t.Errorf("line = %q", q.Get("line"))
	}
	if q.Get("mobile") != "9120000000" {
		t.Errorf("mobile = %q", q.Get("mobile"))
	}
	if q.Get("text") != "hello world" {
		t.Errorf("text = %q", q.Get("text"))
	}
	if res.MessageID != 89545112 || res.Cost != 1.0 {
		t.Errorf("result = %+v", res)
	}
}

func TestSendViaURL_NilRequest(t *testing.T) {
	c := New(testAPIKey)
	_, err := c.SendViaURL(context.Background(), nil)
	if !errors.Is(err, errNilRequest) {
		t.Errorf("err = %v", err)
	}
}

func TestCancelScheduledSend_Success(t *testing.T) {
	var gotPath, gotMethod string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"status":1,"message":"موفق","data":{"returnedCreditCount":10.0,"smsCount":5}}`)
	}))
	defer ts.Close()

	c := newTestClient(t, ts)
	res, err := c.CancelScheduledSend(context.Background(), "2b99e63c-9bf8-4a21-9bfe-3f72dc1b46f1")
	if err != nil {
		t.Fatalf("CancelScheduledSend: %v", err)
	}
	if gotMethod != http.MethodDelete {
		t.Errorf("method = %q", gotMethod)
	}
	wantPath := "/v1/send/scheduled/2b99e63c-9bf8-4a21-9bfe-3f72dc1b46f1"
	if gotPath != wantPath {
		t.Errorf("path = %q, want %q", gotPath, wantPath)
	}
	if res.ReturnedCreditCount != 10.0 || res.SmsCount != 5 {
		t.Errorf("result = %+v", res)
	}
}

func TestCancelScheduledSend_EscapesPackID(t *testing.T) {
	var gotPath string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.EscapedPath()
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"status":1,"data":{"returnedCreditCount":0,"smsCount":0}}`)
	}))
	defer ts.Close()

	c := newTestClient(t, ts)
	// A pack ID with a path-unsafe character must be escaped.
	_, err := c.CancelScheduledSend(context.Background(), "a b/c")
	if err != nil {
		t.Fatalf("CancelScheduledSend: %v", err)
	}
	if gotPath != "/v1/send/scheduled/a%20b%2Fc" {
		t.Errorf("escaped path = %q", gotPath)
	}
}

func TestCancelScheduledSend_NotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, `{"status":111,"message":"با این شناسه، ارسالی ثبت نشده است","data":null}`)
	}))
	defer ts.Close()

	c := newTestClient(t, ts)
	_, err := c.CancelScheduledSend(context.Background(), "missing")
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("err = %v, want *APIError", err)
	}
	if apiErr.Status != StatusSendNotFound {
		t.Errorf("Status = %d", apiErr.Status)
	}
}
