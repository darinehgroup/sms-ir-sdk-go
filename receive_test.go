package smsir

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestLatestReceived_Success(t *testing.T) {
	var got *http.Request
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{
		  "status": 1,
		  "message": "موفق",
		  "data": [
		    {
		      "receiveReturnId": 123456789,
		      "messageText": "سلام",
		      "number": 30004505000017,
		      "mobile": 9121234002,
		      "receivedDateTime": 1628683625
		    }
		  ]
		}`)
	}))
	defer ts.Close()

	c := newTestClient(t, ts)
	res, err := c.LatestReceived(context.Background(), 50)
	if err != nil {
		t.Fatalf("LatestReceived: %v", err)
	}
	if got.Method != http.MethodGet {
		t.Errorf("method = %q", got.Method)
	}
	if got.URL.Path != "/v1/receive/latest" {
		t.Errorf("path = %q", got.URL.Path)
	}
	if got.URL.Query().Get("count") != "50" {
		t.Errorf("count = %q", got.URL.Query().Get("count"))
	}
	if len(res) != 1 {
		t.Fatalf("len = %d", len(res))
	}
	m := res[0]
	if m.ReceiveReturnID != 123456789 {
		t.Errorf("ReceiveReturnID = %d", m.ReceiveReturnID)
	}
	if m.MessageText != "سلام" {
		t.Errorf("MessageText = %q", m.MessageText)
	}
	if m.Number != 30004505000017 {
		t.Errorf("Number = %d", m.Number)
	}
	if m.Mobile != 9121234002 {
		t.Errorf("Mobile = %d", m.Mobile)
	}
	if m.ReceivedDateTime.Time().Unix() != 1628683625 {
		t.Errorf("ReceivedDateTime = %v", m.ReceivedDateTime)
	}
}

func TestLatestReceived_NonPositiveCountOmitted(t *testing.T) {
	var got *http.Request
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"status":1,"data":[]}`)
	}))
	defer ts.Close()

	c := newTestClient(t, ts)
	_, err := c.LatestReceived(context.Background(), 0)
	if err != nil {
		t.Fatalf("LatestReceived: %v", err)
	}
	if got.URL.Query().Get("count") != "" {
		t.Errorf("count should be omitted, got %q", got.URL.Query().Get("count"))
	}
}

func TestTodayReceived_Success(t *testing.T) {
	var got *http.Request
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r
		w.Header().Set("Content-Type", "application/json")
		// live endpoint omits receiveReturnId
		io.WriteString(w, `{"status":1,"data":[
		  {"messageText":"hi","number":30004505000017,"mobile":9121234002,"receivedDateTime":1628683625}
		]}`)
	}))
	defer ts.Close()

	c := newTestClient(t, ts)
	res, err := c.TodayReceived(context.Background(), &TodayReceivedParams{
		PageParams:   PageParams{PageSize: 10, PageNumber: 1},
		SortByNewest: true,
		Mobile:       "9121234002",
	})
	if err != nil {
		t.Fatalf("TodayReceived: %v", err)
	}
	if got.URL.Path != "/v1/receive/live" {
		t.Errorf("path = %q", got.URL.Path)
	}
	q := got.URL.Query()
	if q.Get("pageSize") != "10" {
		t.Errorf("pageSize = %q", q.Get("pageSize"))
	}
	if q.Get("pageNumber") != "1" {
		t.Errorf("pageNumber = %q", q.Get("pageNumber"))
	}
	if q.Get("sortByNewest") != "true" {
		t.Errorf("sortByNewest = %q", q.Get("sortByNewest"))
	}
	if q.Get("mobile") != "9121234002" {
		t.Errorf("mobile = %q", q.Get("mobile"))
	}
	if len(res) != 1 {
		t.Fatalf("len = %d", len(res))
	}
	if res[0].ReceiveReturnID != 0 {
		t.Errorf("ReceiveReturnID should be 0 for live endpoint, got %d", res[0].ReceiveReturnID)
	}
}

func TestTodayReceived_SortByNewestFalseOmitted(t *testing.T) {
	var got *http.Request
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"status":1,"data":[]}`)
	}))
	defer ts.Close()

	c := newTestClient(t, ts)
	_, err := c.TodayReceived(context.Background(), &TodayReceivedParams{SortByNewest: false})
	if err != nil {
		t.Fatalf("TodayReceived: %v", err)
	}
	if q := got.URL.Query().Get("sortByNewest"); q != "" {
		t.Errorf("sortByNewest should be omitted when false, got %q", q)
	}
}

func TestTodayReceived_NilParamsNoQuery(t *testing.T) {
	var got *http.Request
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"status":1,"data":[]}`)
	}))
	defer ts.Close()

	c := newTestClient(t, ts)
	_, err := c.TodayReceived(context.Background(), nil)
	if err != nil {
		t.Fatalf("TodayReceived: %v", err)
	}
	if got.URL.RawQuery != "" {
		t.Errorf("RawQuery should be empty, got %q", got.URL.RawQuery)
	}
}

func TestArchiveReceived_Success(t *testing.T) {
	var got *http.Request
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"status":1,"data":[
		  {"receiveReturnId":777,"messageText":"x","number":3000,"mobile":912,"receivedDateTime":1628683625}
		]}`)
	}))
	defer ts.Close()

	c := newTestClient(t, ts)
	from := time.Unix(1600000000, 0).UTC()
	to := time.Unix(1700000000, 0).UTC()
	res, err := c.ArchiveReceived(context.Background(), &ArchiveReceivedParams{
		PageParams: PageParams{PageSize: 50},
		FromDate:   &from,
		ToDate:     &to,
		Mobile:     "912",
	})
	if err != nil {
		t.Fatalf("ArchiveReceived: %v", err)
	}
	if got.URL.Path != "/v1/receive/archive" {
		t.Errorf("path = %q", got.URL.Path)
	}
	q := got.URL.Query()
	if q.Get("fromDate") != "1600000000" {
		t.Errorf("fromDate = %q", q.Get("fromDate"))
	}
	if q.Get("toDate") != "1700000000" {
		t.Errorf("toDate = %q", q.Get("toDate"))
	}
	if q.Get("pageSize") != "50" {
		t.Errorf("pageSize = %q", q.Get("pageSize"))
	}
	if q.Get("mobile") != "912" {
		t.Errorf("mobile = %q", q.Get("mobile"))
	}
	if len(res) != 1 {
		t.Fatalf("len = %d", len(res))
	}
	if res[0].ReceiveReturnID != 777 {
		t.Errorf("ReceiveReturnID = %d", res[0].ReceiveReturnID)
	}
}

func TestArchiveReceived_NilParamsNoQuery(t *testing.T) {
	var got *http.Request
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"status":1,"data":[]}`)
	}))
	defer ts.Close()

	c := newTestClient(t, ts)
	_, err := c.ArchiveReceived(context.Background(), nil)
	if err != nil {
		t.Fatalf("ArchiveReceived: %v", err)
	}
	if got.URL.RawQuery != "" {
		t.Errorf("RawQuery should be empty, got %q", got.URL.RawQuery)
	}
}
