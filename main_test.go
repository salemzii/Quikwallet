package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

/* These test cases only work without authentication middleware set, so to run test;
make sure to remove the "private" group middleware wrapping the wallet's endpoints;

FOR EXample You Could Do:

	/*private := router.Group("/private")
	private.Use(auth.AuthRequired)
	{
		private.GET("/api/v1/wallets/:wallet_id/balance", getWalletBalance)
		private.POST("/api/v1/wallets/:wallet_id/credit", creditWallet)
		private.POST("/api/v1/wallets/:wallet_id/debit", debitWallet)
		router.POST("/api/v1/wallets/create", createWallet)
	}
*/
//router.GET("/api/v1/wallets/:wallet_id/balance", getWalletBalance)
//router.POST("/api/v1/wallets/:wallet_id/credit", creditWallet)
//router.POST("/api/v1/wallets/:wallet_id/debit", debitWallet)

func TestGetBalanceRoute(t *testing.T) {
	// The setupServer method, that we previously refactored
	// is injected into a test server
	ts := httptest.NewServer(setupServer())
	// Shut down the server and block until all requests have gone through
	defer ts.Close()

	// Make a request to our server with the {base url}/ping
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/wallets/1/balance", ts.URL))

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp.StatusCode != 200 {
		t.Fatalf("Expected status code 200, got %v", resp.StatusCode)
	}

	val, ok := resp.Header["Content-Type"]

	// Assert that the "content-type" header is actually set
	if !ok {
		t.Fatalf("Expected Content-Type header to be set")
	}

	// Assert that it was set as expected
	if val[0] != "application/json; charset=utf-8" {
		t.Fatalf("Expected \"application/json; charset=utf-8\", got %s", val[0])
	}
}

func TestCreditWallet(t *testing.T) {
	ts := httptest.NewServer(setupServer())
	defer ts.Close()

	var jsonStr = []byte(`{"balance":"10.00"}`)
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/wallets/1/credit", ts.URL), bytes.NewBuffer(jsonStr))
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("Expected status code 200, got %v", resp.StatusCode)
	}
}

func TestDebitWallet(t *testing.T) {
	ts := httptest.NewServer(setupServer())
	defer ts.Close()

	var jsonStr = []byte(`{"balance":"10.00"}`)
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/wallets/1/debit", ts.URL), bytes.NewBuffer(jsonStr))
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("Expected status code 200, got %v", resp.StatusCode)
	}
}
