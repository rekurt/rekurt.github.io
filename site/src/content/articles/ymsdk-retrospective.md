# Testing message retries without sending a message

A messaging client receives a temporary server error. Retrying may be reasonable, but a second attempt raises a separate question: does it still identify the same logical message?

A local test can inspect that behavior without a real account, token or recipient. This walkthrough uses [ymsdk](https://github.com/rekurt/ymsdk) v0.2.0, released on September 2, 2026. The original check ran September 18; the same version-pinned example was rerun October 15.

## Separate the retry decision from the request identity

The check gives the SDK a fake HTTP transport. The transport reads each request body, records the first non-empty `payload_id`, and requires the second request to carry the same value. It returns HTTP 502 on the first call, then a successful response with message ID 42.

Two attempts are configured explicitly. The fake transport never invokes a network client: its `Do` method constructs both responses locally. This tests client behavior, not a live messaging service.

## A runnable example

In a fresh Go module, pin the dependency:

```sh
go mod init example.com/ymsdk-retry-example
go get github.com/rekurt/ymsdk@v0.2.0
```

Save this as `main.go` and run `go run .`:

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/rekurt/ymsdk/client/ym"
	"github.com/rekurt/ymsdk/client/ym/messages"
	"github.com/rekurt/ymsdk/client/ym/ymerrors"
	"io"
	"net/http"
	"strings"
	"time"
)

// Fake transport: no network request or real message is sent.
type fakeDoer struct {
	calls   int
	payload string
}

func (f *fakeDoer) Do(r *http.Request) (*http.Response, error) {
	f.calls++
	b, e := io.ReadAll(r.Body)
	if e != nil {
		return nil, e
	}
	var body map[string]any
	if e = json.Unmarshal(b, &body); e != nil {
		return nil, e
	}
	id, ok := body["payload_id"].(string)
	if !ok || id == "" {
		panic("missing payload_id")
	}
	if f.calls == 1 {
		f.payload = id
	} else if id != f.payload {
		panic("retry changed payload_id")
	}
	code := http.StatusOK
	text := `{"ok":true,"message_id":42}`
	if f.calls == 1 {
		code = http.StatusBadGateway
		text = `{"ok":false,"description":"test transient error"}`
	}
	return &http.Response{StatusCode: code, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(text)), Request: r}, nil
}
func main() {
	f := &fakeDoer{}
	cl := ym.NewClientWithHTTP(ym.Config{Token: "fake-test-token", BaseURL: "https://example.invalid", ErrorHandling: ymerrors.ErrorHandlingConfig{RetryStrategy: ymerrors.RetryStrategy{MaxAttempts: 2, InitialBackoff: time.Millisecond, MaxBackoff: time.Millisecond, DisableJitter: true}}}, f)
	m, e := messages.NewService(cl).SendToChat(context.Background(), "test-chat", "example", nil)
	if e != nil {
		panic(e)
	}
	if f.calls != 2 || m.ID != 42 {
		panic("unexpected result")
	}
	fmt.Println("calls=2; message_id=42; payload_id preserved; network requests=0")
}
```

The successful run prints:

```text
calls=2; message_id=42; payload_id preserved; network requests=0
```

The assertions matter as much as the output. A missing identifier, a changed identifier, an unexpected attempt count or a different returned message ID fails the example. The token is a fake test string and the configured address uses `example.invalid`; the injected transport returns responses without contacting it.

## What the test actually proves

For this configured send operation and this simulated 502 response, the client makes two attempts and preserves `payload_id`. That is the observation. It does not establish exactly-once delivery, server-side deduplication or the behavior of an ambiguous timeout after a server has accepted a message.

Those are different properties. A stable identifier can support a deduplication mechanism, but this local test does not observe the remote service implementing one.

Retries also need to be considered operation by operation. The SDK documents that uploads and share operations do not have the same deduplication key. Do not transfer this send-message result to every endpoint. Retries are configured here explicitly; the example should not be read as a claim that they are automatically enabled by default.

## Where to take it next

A broader client test suite can cover cancellation, exhaustion of attempts and responses that should not be retried. A separate integration check, with authorization and an appropriate test destination, would be needed to establish current live API compatibility. Neither a network request nor a real message was part of this walkthrough.

The benefit of the fake transport is a narrow, repeatable assertion about request construction under a controlled failure. It lets that property be checked before introducing accounts, recipients or external availability into the experiment.

[Source and documentation](https://github.com/rekurt/ymsdk) · [v0.2.0 release](https://github.com/rekurt/ymsdk/releases/tag/v0.2.0)

Disclosure: this article was drafted with AI assistance. The fake-transport example was executed locally; no live Yandex API check is claimed.
