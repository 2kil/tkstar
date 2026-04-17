package network

import "testing"

func TestNetParseCurlComdWithHeadersAndData(t *testing.T) {
	cmd := `curl "https://example.com/api?q=1" -X POST -H "Content-Type: application/json" -H "X-Test: abc" --data-raw "{\"name\":\"tkstar\"}"`

	method, urlStr, headers, body, err := NetParseCurlComd(cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if method != "POST" {
		t.Fatalf("method = %q, want POST", method)
	}
	if urlStr != "https://example.com/api?q=1" {
		t.Fatalf("url = %q", urlStr)
	}
	if got := headers.Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q", got)
	}
	if got := headers.Get("X-Test"); got != "abc" {
		t.Fatalf("X-Test = %q", got)
	}
	if got := string(body); got != `{"name":"tkstar"}` {
		t.Fatalf("body = %q", got)
	}
}

func TestNetParseCurlComdSupportsOptionsBeforeURL(t *testing.T) {
	cmd := `curl -H 'Accept: application/json' --data 'a=1' https://example.com/login`

	method, urlStr, headers, body, err := NetParseCurlComd(cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if method != "POST" {
		t.Fatalf("method = %q, want POST", method)
	}
	if urlStr != "https://example.com/login" {
		t.Fatalf("url = %q", urlStr)
	}
	if got := headers.Get("Accept"); got != "application/json" {
		t.Fatalf("Accept = %q", got)
	}
	if got := string(body); got != "a=1" {
		t.Fatalf("body = %q", got)
	}
}

func TestNetParseCurlComdRejectsBrokenQuotes(t *testing.T) {
	_, _, _, _, err := NetParseCurlComd(`curl "https://example.com`)
	if err == nil {
		t.Fatal("expected error for broken quotes")
	}
}

func TestNetParseCurlComdSkipsKnownOptionValues(t *testing.T) {
	cmd := `curl --proxy http://127.0.0.1:7890 https://example.com/api`

	method, urlStr, _, _, err := NetParseCurlComd(cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if method != "GET" {
		t.Fatalf("method = %q, want GET", method)
	}
	if urlStr != "https://example.com/api" {
		t.Fatalf("url = %q, want https://example.com/api", urlStr)
	}
}

func TestNetParseCurlComdSupportsExplicitURLFlag(t *testing.T) {
	cmd := `curl --url https://example.com/api -H "Accept: application/json"`

	_, urlStr, headers, _, err := NetParseCurlComd(cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if urlStr != "https://example.com/api" {
		t.Fatalf("url = %q", urlStr)
	}
	if got := headers.Get("Accept"); got != "application/json" {
		t.Fatalf("Accept = %q", got)
	}
}
