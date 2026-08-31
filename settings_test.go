package smsir

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCredit_Success(t *testing.T) {
	var got *http.Request
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"status":1,"message":"موفق","data":165.3}`)
	}))
	defer ts.Close()

	c := newTestClient(t, ts)
	credit, err := c.Credit(context.Background())
	if err != nil {
		t.Fatalf("Credit: %v", err)
	}
	if got.Method != http.MethodGet {
		t.Errorf("method = %q", got.Method)
	}
	if got.URL.Path != "/v1/credit" {
		t.Errorf("path = %q", got.URL.Path)
	}
	assertHeader(t, got, "X-API-KEY", testAPIKey)
	if credit != 165.3 {
		t.Errorf("credit = %v, want 165.3", credit)
	}
}

func TestCredit_APIError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		io.WriteString(w, `{"status":10,"message":"کلید وب‌سرویس نامعتبر است","data":null}`)
	}))
	defer ts.Close()

	c := newTestClient(t, ts)
	_, err := c.Credit(context.Background())
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("err = %v, want *APIError", err)
	}
	if apiErr.Status != StatusInvalidAPIKey {
		t.Errorf("Status = %d", apiErr.Status)
	}
	if !apiErr.IsAuthError() {
		t.Error("IsAuthError should be true")
	}
}

func TestLines_Success(t *testing.T) {
	var got *http.Request
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"status":1,"message":"موفق","data":[10002155613464, 30004505000017]}`)
	}))
	defer ts.Close()

	c := newTestClient(t, ts)
	lines, err := c.Lines(context.Background())
	if err != nil {
		t.Fatalf("Lines: %v", err)
	}
	if got.Method != http.MethodGet {
		t.Errorf("method = %q", got.Method)
	}
	if got.URL.Path != "/v1/line" {
		t.Errorf("path = %q", got.URL.Path)
	}
	if len(lines) != 2 {
		t.Fatalf("len = %d", len(lines))
	}
	if lines[0] != 10002155613464 || lines[1] != 30004505000017 {
		t.Errorf("lines = %v", lines)
	}
}
