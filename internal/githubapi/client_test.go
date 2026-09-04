package githubapi

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestClientRetriesRateLimit(t *testing.T) {
	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if attempts.Add(1) == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		writeJSON(t, w, []byte("[]"))
	}))
	t.Cleanup(server.Close)

	got, err := New(server.URL, "", server.Client()).ListOwnedPublic(t.Context(), "rekurt")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 || attempts.Load() != 2 {
		t.Fatalf("repos = %d, attempts = %d; want 0, 2", len(got), attempts.Load())
	}
}

func TestClientReusesETagBody(t *testing.T) {
	fixture := mustFixture(t, "repos-page-2.json")
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch requests.Add(1) {
		case 1:
			if got := r.Header.Get("If-None-Match"); got != "" {
				t.Fatalf("first If-None-Match = %q", got)
			}
			w.Header().Set("ETag", `"catalog-v1"`)
			writeJSON(t, w, fixture)
		case 2:
			if got := r.Header.Get("If-None-Match"); got != `"catalog-v1"` {
				t.Fatalf("second If-None-Match = %q", got)
			}
			w.WriteHeader(http.StatusNotModified)
		default:
			t.Fatalf("unexpected request %d", requests.Load())
		}
	}))
	t.Cleanup(server.Close)

	client := New(server.URL, "", server.Client())
	first, err := client.ListOwnedPublic(t.Context(), "rekurt")
	if err != nil {
		t.Fatal(err)
	}
	second, err := client.ListOwnedPublic(t.Context(), "rekurt")
	if err != nil {
		t.Fatal(err)
	}
	if len(first) != 1 || len(second) != 1 || second[0].NameWithOwner != "rekurt/beta" {
		t.Fatalf("first = %#v, second = %#v", first, second)
	}
}
