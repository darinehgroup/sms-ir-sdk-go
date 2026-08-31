package smsir

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestMessageReport_Success(t *testing.T) {
	var got *http.Request
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{
		  "status": 1,
		  "message": "موفق",
		  "data": {
		    "messageId": 89545112,
		    "mobile": 9121234677,
		    "messageText": "سلام",
		    "sendDateTime": 1628683626,
		    "lineNumber": 30004505000017,
		    "cost": 1.0,
		    "deliveryState": 1,
		    "deliveryDateTime": 1628683629
		  }
		}`)
	}))
	defer ts.Close()

	c := newTestClient(t, ts)
	res, err := c.MessageReport(context.Background(), 89545112)
	if err != nil {
		t.Fatalf("MessageReport: %v", err)
	}
	if got.Method != http.MethodGet {
		t.Errorf("method = %q", got.Method)
	}
	if got.URL.Path != "/v1/send/89545112" {
		t.Errorf("path = %q", got.URL.Path)
	}
	assertHeader(t, got, "X-API-KEY", testAPIKey)

	if res.MessageID != 89545112 {
		t.Errorf("MessageID = %d", res.MessageID)
	}
	if res.Mobile != 9121234677 {
		t.Errorf("Mobile = %d", res.Mobile)
	}
	if res.MessageText != "سلام" {
		t.Errorf("MessageText = %q", res.MessageText)
	}
	if res.SendDateTime.Time().Unix() != 1628683626 {
		t.Errorf("SendDateTime = %v", res.SendDateTime)
	}
	if res.LineNumber != 30004505000017 {
		t.Errorf("LineNumber = %d", res.LineNumber)
	}
	if res.Cost != 1.0 {
		t.Errorf("Cost = %v", res.Cost)
	}
	if res.DeliveryState == nil || *res.DeliveryState != DeliveryStateDelivered {
		t.Errorf("DeliveryState = %v", res.DeliveryState)
	}
	if res.DeliveryDateTime == nil || res.DeliveryDateTime.Time().Unix() != 1628683629 {
		t.Errorf("DeliveryDateTime = %v", res.DeliveryDateTime)
	}
}

func TestMessageReport_NullableDelivery(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{
		  "status": 1,
		  "data": {
		    "messageId": 1, "mobile": 9, "messageText": "x",
		    "sendDateTime": 1628683626, "lineNumber": 3000, "cost": 1.0,
		    "deliveryState": null, "deliveryDateTime": null
		  }
		}`)
	}))
	defer ts.Close()

	c := newTestClient(t, ts)
	res, err := c.MessageReport(context.Background(), 1)
	if err != nil {
		t.Fatalf("MessageReport: %v", err)
	}
	if res.DeliveryState != nil {
		t.Errorf("DeliveryState should be nil, got %v", res.DeliveryState)
	}
	if res.DeliveryDateTime != nil {
		t.Errorf("DeliveryDateTime should be nil, got %v", res.DeliveryDateTime)
	}
}

func TestMessageReport_NotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, `{"status":111,"message":"با این شناسه، ارسالی ثبت نشده است","data":null}`)
	}))
	defer ts.Close()

	c := newTestClient(t, ts)
	_, err := c.MessageReport(context.Background(), 99999)
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("err = %v, want *APIError", err)
	}
	if apiErr.Status != StatusSendNotFound {
		t.Errorf("Status = %d", apiErr.Status)
	}
}

func TestTodayPacks_Success(t *testing.T) {
	var got *http.Request
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"status":1,"data":[
		  {"packId":"p1","recipientCount":3,"creationDateTime":1628683626},
		  {"packId":"p2","recipientCount":5,"creationDateTime":1628683627}
		]}`)
	}))
	defer ts.Close()

	c := newTestClient(t, ts)
	res, err := c.TodayPacks(context.Background(), &PageParams{PageSize: 10, PageNumber: 2})
	if err != nil {
		t.Fatalf("TodayPacks: %v", err)
	}
	if got.URL.Path != "/v1/send/pack" {
		t.Errorf("path = %q", got.URL.Path)
	}
	q := got.URL.Query()
	if q.Get("pageSize") != "10" {
		t.Errorf("pageSize = %q", q.Get("pageSize"))
	}
	if q.Get("pageNumber") != "2" {
		t.Errorf("pageNumber = %q", q.Get("pageNumber"))
	}
	if len(res) != 2 {
		t.Fatalf("len = %d", len(res))
	}
	if res[0].PackID != "p1" || res[0].RecipientCount != 3 {
		t.Errorf("res[0] = %+v", res[0])
	}
	if res[1].CreationDateTime.Time().Unix() != 1628683627 {
		t.Errorf("res[1].CreationDateTime = %v", res[1].CreationDateTime)
	}
}

func TestTodayPacks_NilParamsNoQuery(t *testing.T) {
	var got *http.Request
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"status":1,"data":[]}`)
	}))
	defer ts.Close()

	c := newTestClient(t, ts)
	_, err := c.TodayPacks(context.Background(), nil)
	if err != nil {
		t.Fatalf("TodayPacks: %v", err)
	}
	if got.URL.RawQuery != "" {
		t.Errorf("RawQuery should be empty, got %q", got.URL.RawQuery)
	}
}

func TestPackReport_Success(t *testing.T) {
	var got *http.Request
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"status":1,"data":[
		  {"messageId":1,"mobile":9,"messageText":"a","sendDateTime":1628683626,"lineNumber":3000,"cost":1.0,"deliveryState":1,"deliveryDateTime":1628683627},
		  {"messageId":2,"mobile":8,"messageText":"b","sendDateTime":1628683626,"lineNumber":3000,"cost":1.0,"deliveryState":null,"deliveryDateTime":null}
		]}`)
	}))
	defer ts.Close()

	c := newTestClient(t, ts)
	res, err := c.PackReport(context.Background(), "pack-guid")
	if err != nil {
		t.Fatalf("PackReport: %v", err)
	}
	if got.URL.Path != "/v1/send/pack/pack-guid" {
		t.Errorf("path = %q", got.URL.Path)
	}
	if len(res) != 2 {
		t.Fatalf("len = %d", len(res))
	}
	if res[0].DeliveryState == nil || *res[0].DeliveryState != DeliveryStateDelivered {
		t.Errorf("res[0].DeliveryState = %v", res[0].DeliveryState)
	}
	if res[1].DeliveryState != nil {
		t.Errorf("res[1].DeliveryState should be nil, got %v", res[1].DeliveryState)
	}
}

func TestPackReport_EscapesPackID(t *testing.T) {
	var got *http.Request
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"status":1,"data":[]}`)
	}))
	defer ts.Close()

	c := newTestClient(t, ts)
	_, err := c.PackReport(context.Background(), "a b")
	if err != nil {
		t.Fatalf("PackReport: %v", err)
	}
	if got.URL.EscapedPath() != "/v1/send/pack/a%20b" {
		t.Errorf("escaped path = %q", got.URL.EscapedPath())
	}
}

func TestTodayReport_Success(t *testing.T) {
	var got *http.Request
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"status":1,"data":[
		  {"messageId":1,"mobile":9,"messageText":"a","sendDateTime":1628683626,"lineNumber":3000,"cost":1.0,"deliveryState":1,"deliveryDateTime":1628683627}
		]}`)
	}))
	defer ts.Close()

	c := newTestClient(t, ts)
	res, err := c.TodayReport(context.Background(), &PageParams{PageSize: 50})
	if err != nil {
		t.Fatalf("TodayReport: %v", err)
	}
	if got.URL.Path != "/v1/send/live" {
		t.Errorf("path = %q", got.URL.Path)
	}
	if got.URL.Query().Get("pageSize") != "50" {
		t.Errorf("pageSize = %q", got.URL.Query().Get("pageSize"))
	}
	if len(res) != 1 {
		t.Fatalf("len = %d", len(res))
	}
	if res[0].MessageID != 1 {
		t.Errorf("MessageID = %d", res[0].MessageID)
	}
}

func TestArchiveReport_Success(t *testing.T) {
	var got *http.Request
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"status":1,"data":[
		  {"messageId":1,"mobile":9,"messageText":"a","sendDateTime":1628683626,"lineNumber":3000,"cost":1.0,"deliveryState":6,"deliveryDateTime":1628683627}
		]}`)
	}))
	defer ts.Close()

	c := newTestClient(t, ts)
	from := time.Unix(1600000000, 0).UTC()
	to := time.Unix(1700000000, 0).UTC()
	res, err := c.ArchiveReport(context.Background(), &ArchiveReportParams{
		PageParams: PageParams{PageSize: 20, PageNumber: 3},
		FromDate:   &from,
		ToDate:     &to,
	})
	if err != nil {
		t.Fatalf("ArchiveReport: %v", err)
	}
	if got.URL.Path != "/v1/send/archive" {
		t.Errorf("path = %q", got.URL.Path)
	}
	q := got.URL.Query()
	if q.Get("fromDate") != "1600000000" {
		t.Errorf("fromDate = %q", q.Get("fromDate"))
	}
	if q.Get("toDate") != "1700000000" {
		t.Errorf("toDate = %q", q.Get("toDate"))
	}
	if q.Get("pageSize") != "20" {
		t.Errorf("pageSize = %q", q.Get("pageSize"))
	}
	if q.Get("pageNumber") != "3" {
		t.Errorf("pageNumber = %q", q.Get("pageNumber"))
	}
	if len(res) != 1 {
		t.Fatalf("len = %d", len(res))
	}
	if res[0].DeliveryState == nil || *res[0].DeliveryState != DeliveryStateFailed {
		t.Errorf("DeliveryState = %v", res[0].DeliveryState)
	}
}

func TestArchiveReport_NilParamsNoQuery(t *testing.T) {
	var got *http.Request
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"status":1,"data":[]}`)
	}))
	defer ts.Close()

	c := newTestClient(t, ts)
	_, err := c.ArchiveReport(context.Background(), nil)
	if err != nil {
		t.Fatalf("ArchiveReport: %v", err)
	}
	if got.URL.RawQuery != "" {
		t.Errorf("RawQuery should be empty, got %q", got.URL.RawQuery)
	}
}
