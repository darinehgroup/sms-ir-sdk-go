package smsir

import (
	"context"
	"fmt"
	"time"
)

// These Example functions are rendered by godoc. They do not make network
// calls; their purpose is to show idiomatic SDK usage. The Example* naming
// convention lets `go test -run Example` exercise them, so each one compiles
// as part of the test suite.

// ExampleNew shows how to construct a client.
func ExampleNew() {
	// Production key (or a sandbox key — only the key type differs).
	client := New("YOUR_API_KEY")
	_ = client // use client.SendVerify, client.Credit, ...
}

// ExampleClient_SendVerify shows sending a one-time verification code.
func ExampleClient_SendVerify() {
	ctx := context.Background()
	client := New("YOUR_API_KEY")

	res, err := client.SendVerify(ctx, &SendVerifyRequest{
		Mobile:     "9120000000",
		TemplateID: 123456,
		Parameters: []VerifyParameter{{Name: "Code", Value: "12345"}},
	})
	if err != nil {
		fmt.Println("send failed:", err)
		return
	}
	fmt.Println("message id:", res.MessageID)
}

// ExampleClient_SendBulk shows sending the same text to several recipients.
func ExampleClient_SendBulk() {
	ctx := context.Background()
	client := New("YOUR_API_KEY")

	res, err := client.SendBulk(ctx, &SendBulkRequest{
		LineNumber:  30004505000017,
		MessageText: "سلام از SMS.ir",
		Mobiles:     []string{"9120000000", "9120000001"},
	})
	if err != nil {
		fmt.Println("send failed:", err)
		return
	}
	fmt.Println("pack id:", res.PackID, "recipients:", len(res.MessageIDs))
}

// ExampleClient_SendBulk_scheduled shows scheduling a bulk send two hours
// from now using UnixTime.
func ExampleClient_SendBulk_scheduled() {
	ctx := context.Background()
	client := New("YOUR_API_KEY")

	when := time.Now().Add(2 * time.Hour)
	_, err := client.SendBulk(ctx, &SendBulkRequest{
		LineNumber:   30004505000017,
		MessageText:  "یادآوری قرار ساعت ۱۴",
		Mobiles:      []string{"9120000000"},
		SendDateTime: NewUnixTime(when),
	})
	if err != nil {
		fmt.Println("schedule failed:", err)
	}
}

// ExampleClient_Credit shows reading the account's current balance.
func ExampleClient_Credit() {
	ctx := context.Background()
	client := New("YOUR_API_KEY")

	credit, err := client.Credit(ctx)
	if err != nil {
		fmt.Println("credit failed:", err)
		return
	}
	fmt.Printf("remaining credit: %.2f\n", credit)
}
