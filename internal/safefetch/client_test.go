package safefetch

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

type dialerFunc func(context.Context, string, string) (net.Conn, error)

func (fn dialerFunc) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	return fn(ctx, network, address)
}

func serverDialer(server *httptest.Server, calls *atomic.Int64) Dialer {
	serverAddress := strings.TrimPrefix(server.URL, "http://")
	return dialerFunc(func(ctx context.Context, network, _ string) (net.Conn, error) {
		if calls != nil {
			calls.Add(1)
		}
		return (&net.Dialer{}).DialContext(ctx, network, serverAddress)
	})
}

func newTestClient(server *httptest.Server, resolver Resolver, options ClientOptions) *Client {
	options.Resolver = resolver
	options.Dialer = serverDialer(server, nil)
	client := NewClient(options)
	return client
}

func TestClientConditionalGETCarriesAndRefreshesCheckpoint(t *testing.T) {
	const (
		etag         = `"catalog-v2"`
		lastModified = "Sun, 13 Sep 2026 12:00:00 GMT"
	)
	var calls atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		calls.Add(1)
		if request.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", request.Method)
		}
		if request.UserAgent() == "" {
			t.Error("missing User-Agent")
		}
		writer.Header().Set("ETag", etag)
		writer.Header().Set("Last-Modified", lastModified)
		if request.Header.Get("If-None-Match") == etag && request.Header.Get("If-Modified-Since") == lastModified {
			writer.WriteHeader(http.StatusNotModified)
			return
		}
		_, _ = writer.Write([]byte(`{"success":true}`))
	}))
	defer server.Close()

	resolver := staticResolver(map[string][]string{"catalog.canada.ca": {"93.184.216.34"}})
	client := newTestClient(server, resolver, ClientOptions{})
	defer client.CloseIdleConnections()
	first, err := client.Get(context.Background(), "http://catalog.canada.ca/catalog", Checkpoint{})
	if err != nil {
		t.Fatal(err)
	}
	if first.NotModified || string(first.Body) != `{"success":true}` || first.ETag != etag || first.LastModified != lastModified {
		t.Fatalf("first response = %#v", first)
	}
	second, err := client.Get(context.Background(), "http://catalog.canada.ca/catalog", first.Checkpoint())
	if err != nil {
		t.Fatal(err)
	}
	if !second.NotModified || len(second.Body) != 0 || second.Checkpoint() != first.Checkpoint() {
		t.Fatalf("second response = %#v", second)
	}
	if calls.Load() != 2 {
		t.Fatalf("server calls = %d, want 2", calls.Load())
	}
}

func TestClientRevalidatesEveryRedirect(t *testing.T) {
	var finalCalls atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/safe-start":
			http.Redirect(writer, request, "http://redirect.canada.ca/final", http.StatusFound)
		case "/unsafe-start":
			http.Redirect(writer, request, "http://169.254.169.254/latest/meta-data", http.StatusFound)
		case "/final":
			finalCalls.Add(1)
			_, _ = writer.Write([]byte("ok"))
		default:
			http.NotFound(writer, request)
		}
	}))
	defer server.Close()

	resolver := staticResolver(map[string][]string{
		"catalog.canada.ca":  {"93.184.216.34"},
		"redirect.canada.ca": {"93.184.216.35"},
	})
	client := newTestClient(server, resolver, ClientOptions{})
	defer client.CloseIdleConnections()
	response, err := client.Get(context.Background(), "http://catalog.canada.ca/safe-start", Checkpoint{})
	if err != nil || string(response.Body) != "ok" {
		t.Fatalf("safe redirect response = %#v, error = %v", response, err)
	}
	if _, err := client.Get(context.Background(), "http://catalog.canada.ca/unsafe-start", Checkpoint{}); !errors.Is(err, ErrUnsafeDestination) {
		t.Fatalf("unsafe redirect error = %v, want ErrUnsafeDestination", err)
	}
	if finalCalls.Load() != 1 {
		t.Fatalf("final handler calls = %d, want 1", finalCalls.Load())
	}
}

func TestClientBlocksDNSRebindingBeforeDial(t *testing.T) {
	var resolutions atomic.Int64
	resolver := resolverFunc(func(_ context.Context, _ string, _ string) ([]net.IP, error) {
		if resolutions.Add(1) == 1 {
			return []net.IP{net.ParseIP("93.184.216.34")}, nil
		}
		return []net.IP{net.ParseIP("127.0.0.1")}, nil
	})
	var dialCalls atomic.Int64
	client := NewClient(ClientOptions{
		Resolver: resolver,
		Dialer: dialerFunc(func(context.Context, string, string) (net.Conn, error) {
			dialCalls.Add(1)
			return nil, errors.New("must not dial")
		}),
	})
	defer client.CloseIdleConnections()
	if _, err := client.Get(context.Background(), "https://catalog.canada.ca/data", Checkpoint{}); !errors.Is(err, ErrUnsafeDestination) {
		t.Fatalf("rebind error = %v, want ErrUnsafeDestination", err)
	}
	if dialCalls.Load() != 0 {
		t.Fatalf("unsafe dial attempted %d time(s)", dialCalls.Load())
	}
}

func TestClientEnforcesBodyLimitWithAndWithoutContentLength(t *testing.T) {
	for _, chunked := range []bool{false, true} {
		t.Run(fmt.Sprintf("chunked=%t", chunked), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
				if chunked {
					writer.(http.Flusher).Flush()
				}
				_, _ = writer.Write([]byte(strings.Repeat("x", 65)))
			}))
			defer server.Close()
			resolver := staticResolver(map[string][]string{"catalog.canada.ca": {"93.184.216.34"}})
			client := newTestClient(server, resolver, ClientOptions{MaxBodyBytes: 64})
			defer client.CloseIdleConnections()
			if _, err := client.Get(context.Background(), "http://catalog.canada.ca/large", Checkpoint{}); !errors.Is(err, ErrBodyTooLarge) {
				t.Fatalf("body-limit error = %v, want ErrBodyTooLarge", err)
			}
		})
	}
}

func TestClientReturnsTypedHTTPStatusWithRetryAfter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Retry-After", "120")
		writer.WriteHeader(http.StatusTooManyRequests)
		_, _ = writer.Write([]byte("slow down"))
	}))
	defer server.Close()
	resolver := staticResolver(map[string][]string{"catalog.canada.ca": {"93.184.216.34"}})
	client := newTestClient(server, resolver, ClientOptions{})
	defer client.CloseIdleConnections()
	response, err := client.Get(context.Background(), "http://catalog.canada.ca/data", Checkpoint{})
	var statusError *HTTPStatusError
	if !errors.As(err, &statusError) || statusError.StatusCode != http.StatusTooManyRequests || statusError.RetryAfter != "120" {
		t.Fatalf("status error = %#v (%v)", statusError, err)
	}
	if string(response.Body) != "slow down" {
		t.Fatalf("bounded error body = %q", response.Body)
	}
}

func TestClientRejectsUnsafeHeadersAndCheckpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("invalid request reached server")
	}))
	defer server.Close()
	resolver := staticResolver(map[string][]string{"catalog.canada.ca": {"93.184.216.34"}})
	client := newTestClient(server, resolver, ClientOptions{})
	defer client.CloseIdleConnections()
	_, err := client.Do(context.Background(), Request{
		URL:    "http://catalog.canada.ca/data",
		Header: http.Header{"Connection": {"close"}},
	})
	if err == nil || !strings.Contains(err.Error(), "forbidden") {
		t.Fatalf("forbidden header error = %v", err)
	}
	_, err = client.Get(context.Background(), "http://catalog.canada.ca/data", Checkpoint{ETag: "bad\r\nInjected: yes"})
	if err == nil || !strings.Contains(err.Error(), "checkpoint") {
		t.Fatalf("invalid checkpoint error = %v", err)
	}
}

func TestClientBoundsRedirectsAndRequestTime(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/slow" {
			select {
			case <-request.Context().Done():
			case <-time.After(time.Second):
			}
			return
		}
		next := "/redirect"
		if request.URL.Query().Get("n") == "" {
			next += "?n=1"
		} else {
			next += "?n=2"
		}
		http.Redirect(writer, request, next, http.StatusFound)
	}))
	defer server.Close()
	resolver := staticResolver(map[string][]string{"catalog.canada.ca": {"93.184.216.34"}})
	client := newTestClient(server, resolver, ClientOptions{MaxRedirects: 1, Timeout: 40 * time.Millisecond})
	defer client.CloseIdleConnections()
	if _, err := client.Get(context.Background(), "http://catalog.canada.ca/redirect", Checkpoint{}); err == nil || !strings.Contains(err.Error(), "stopped after 1 redirects") {
		t.Fatalf("redirect limit error = %v", err)
	}
	started := time.Now()
	if _, err := client.Get(context.Background(), "http://catalog.canada.ca/slow", Checkpoint{}); err == nil {
		t.Fatal("expected request timeout")
	}
	if elapsed := time.Since(started); elapsed > 500*time.Millisecond {
		t.Fatalf("timeout took %v", elapsed)
	}
}
