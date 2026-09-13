package safefetch

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultMaxBodyBytes int64 = 8 << 20
	DefaultTimeout            = 30 * time.Second
	DefaultMaxRedirects       = 3
)

var ErrBodyTooLarge = errors.New("public HTTP response body exceeds limit")

// Dialer is injectable so tests can verify that rejected addresses never
// reach the network and can route public test names to local fixtures.
type Dialer interface {
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}

// Checkpoint contains standard HTTP validators used for conditional GETs.
type Checkpoint struct {
	ETag         string `json:"etag,omitempty"`
	LastModified string `json:"last_modified,omitempty"`
}

// Request describes a GET request. No method field is intentionally exposed:
// the discovery fetcher cannot be used to execute unsafe API operations.
type Request struct {
	URL        string
	Checkpoint Checkpoint
	Header     http.Header
}

// Response is a bounded response plus the validators needed for the next
// checkpoint. A 304 is represented by NotModified rather than as an error.
type Response struct {
	FinalURL     string
	StatusCode   int
	Header       http.Header
	Body         []byte
	NotModified  bool
	ETag         string
	LastModified string
	RetrievedAt  time.Time
}

// Checkpoint returns validators suitable for a subsequent conditional GET.
func (r *Response) Checkpoint() Checkpoint {
	if r == nil {
		return Checkpoint{}
	}
	return Checkpoint{ETag: r.ETag, LastModified: r.LastModified}
}

// HTTPStatusError reports a bounded non-success HTTP response.
type HTTPStatusError struct {
	StatusCode int
	URL        string
	RetryAfter string
}

func (e *HTTPStatusError) Error() string {
	if e.RetryAfter != "" {
		return fmt.Sprintf("public HTTP GET %s returned %d (Retry-After: %s)", e.URL, e.StatusCode, e.RetryAfter)
	}
	return fmt.Sprintf("public HTTP GET %s returned %d", e.URL, e.StatusCode)
}

// ClientOptions configures bounded retrieval. Zero values receive production
// defaults. Policy and Resolver should refer to the same resolver; when Policy
// is nil, the client constructs one from Resolver.
type ClientOptions struct {
	Policy       *PublicHTTPPolicy
	Resolver     Resolver
	Dialer       Dialer
	MaxBodyBytes int64
	Timeout      time.Duration
	MaxRedirects int
	UserAgent    string
}

// Client safely retrieves untrusted public HTTP resources.
type Client struct {
	policy       *PublicHTTPPolicy
	httpClient   *http.Client
	maxBodyBytes int64
	userAgent    string
}

// NewClient constructs a client with proxying and automatic decompression
// disabled, bounded connection pools, redirect checks, and DNS-pinned dialing.
func NewClient(options ClientOptions) *Client {
	resolver := options.Resolver
	if resolver == nil {
		resolver = net.DefaultResolver
	}
	policy := options.Policy
	if policy == nil {
		policy = NewPublicHTTPPolicy(resolver)
	}
	dialer := options.Dialer
	if dialer == nil {
		dialer = &net.Dialer{Timeout: 10 * time.Second, KeepAlive: 30 * time.Second}
	}
	maxBodyBytes := options.MaxBodyBytes
	if maxBodyBytes <= 0 {
		maxBodyBytes = DefaultMaxBodyBytes
	}
	timeout := options.Timeout
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	maxRedirects := options.MaxRedirects
	if maxRedirects <= 0 {
		maxRedirects = DefaultMaxRedirects
	}
	userAgent := strings.TrimSpace(options.UserAgent)
	if userAgent == "" {
		userAgent = "CanadaOpportunityGraph/1.0 (public-data discovery)"
	}

	client := &Client{policy: policy, maxBodyBytes: maxBodyBytes, userAgent: userAgent}
	transport := &http.Transport{
		Proxy:                 nil,
		DialContext:           client.safeDialContext(dialer),
		ForceAttemptHTTP2:     true,
		DisableCompression:    true,
		MaxIdleConns:          32,
		MaxIdleConnsPerHost:   4,
		MaxConnsPerHost:       4,
		IdleConnTimeout:       60 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 15 * time.Second,
		ExpectContinueTimeout: time.Second,
	}
	client.httpClient = &http.Client{
		Transport: transport,
		Timeout:   timeout,
		CheckRedirect: func(request *http.Request, via []*http.Request) error {
			if len(via) >= maxRedirects {
				return fmt.Errorf("stopped after %d redirects", maxRedirects)
			}
			_, _, err := policy.Validate(request.Context(), request.URL.String())
			if err != nil {
				return fmt.Errorf("redirect rejected: %w", err)
			}
			return nil
		},
	}
	return client
}

// Validate performs the same destination validation used before requests and
// redirects without sending traffic.
func (c *Client) Validate(ctx context.Context, rawURL string) (*url.URL, error) {
	parsed, _, err := c.policy.Validate(ctx, rawURL)
	return parsed, err
}

// Get performs a conditional GET using the supplied checkpoint.
func (c *Client) Get(ctx context.Context, rawURL string, checkpoint Checkpoint) (*Response, error) {
	return c.Do(ctx, Request{URL: rawURL, Checkpoint: checkpoint})
}

// Do performs a bounded GET. Custom headers are copied; hop-by-hop and host
// override headers are rejected.
func (c *Client) Do(ctx context.Context, input Request) (*Response, error) {
	parsed, _, err := c.policy.Validate(ctx, input.URL)
	if err != nil {
		return nil, err
	}
	if err := validateCheckpoint(input.Checkpoint); err != nil {
		return nil, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create public HTTP request: %w", err)
	}
	for key, values := range input.Header {
		if forbiddenRequestHeader(key) {
			return nil, fmt.Errorf("forbidden public HTTP request header %q", key)
		}
		for _, value := range values {
			if strings.ContainsAny(value, "\r\n") {
				return nil, fmt.Errorf("invalid public HTTP request header %q", key)
			}
			request.Header.Add(key, value)
		}
	}
	request.Header.Set("User-Agent", c.userAgent)
	if input.Checkpoint.ETag != "" {
		request.Header.Set("If-None-Match", input.Checkpoint.ETag)
	}
	if input.Checkpoint.LastModified != "" {
		request.Header.Set("If-Modified-Since", input.Checkpoint.LastModified)
	}

	httpResponse, err := c.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("public HTTP GET %s: %w", parsed.Redacted(), err)
	}
	defer httpResponse.Body.Close()

	response := &Response{
		FinalURL:     httpResponse.Request.URL.String(),
		StatusCode:   httpResponse.StatusCode,
		Header:       httpResponse.Header.Clone(),
		NotModified:  httpResponse.StatusCode == http.StatusNotModified,
		ETag:         strings.TrimSpace(httpResponse.Header.Get("ETag")),
		LastModified: strings.TrimSpace(httpResponse.Header.Get("Last-Modified")),
		RetrievedAt:  time.Now().UTC(),
	}
	if response.ETag == "" {
		response.ETag = input.Checkpoint.ETag
	}
	if response.LastModified == "" {
		response.LastModified = input.Checkpoint.LastModified
	}
	if response.NotModified {
		return response, nil
	}

	if httpResponse.ContentLength > c.maxBodyBytes {
		return response, fmt.Errorf("%w: Content-Length %d exceeds %d", ErrBodyTooLarge, httpResponse.ContentLength, c.maxBodyBytes)
	}
	body, readErr := io.ReadAll(io.LimitReader(httpResponse.Body, c.maxBodyBytes+1))
	if readErr != nil {
		return response, fmt.Errorf("read public HTTP response: %w", readErr)
	}
	if int64(len(body)) > c.maxBodyBytes {
		return response, fmt.Errorf("%w: read more than %d bytes", ErrBodyTooLarge, c.maxBodyBytes)
	}
	response.Body = body
	if httpResponse.StatusCode < http.StatusOK || httpResponse.StatusCode >= http.StatusMultipleChoices {
		return response, &HTTPStatusError{
			StatusCode: httpResponse.StatusCode,
			URL:        response.FinalURL,
			RetryAfter: strings.TrimSpace(httpResponse.Header.Get("Retry-After")),
		}
	}
	return response, nil
}

func (c *Client) safeDialContext(dialer Dialer) func(context.Context, string, string) (net.Conn, error) {
	return func(ctx context.Context, network, address string) (net.Conn, error) {
		host, portText, err := net.SplitHostPort(address)
		if err != nil {
			return nil, fmt.Errorf("validate dial address: %w", err)
		}
		portValue, err := strconv.ParseUint(portText, 10, 16)
		if err != nil || portValue == 0 {
			return nil, fmt.Errorf("%w: invalid dial port %q", ErrUnsafeDestination, portText)
		}
		if _, allowed := c.policy.allowedPorts[uint16(portValue)]; !allowed {
			return nil, fmt.Errorf("%w: dial port %d is not allowed", ErrUnsafeDestination, portValue)
		}
		host = normalizeHost(strings.Trim(host, "[]"))
		if err := c.policy.validateHostname(host); err != nil {
			return nil, err
		}
		addresses, err := c.policy.resolvePublic(ctx, host)
		if err != nil {
			return nil, err
		}
		var failures []error
		for _, resolved := range addresses {
			connection, dialErr := dialer.DialContext(ctx, network, net.JoinHostPort(resolved.String(), portText))
			if dialErr == nil {
				return connection, nil
			}
			failures = append(failures, dialErr)
		}
		return nil, fmt.Errorf("dial public host %q: %w", host, errors.Join(failures...))
	}
}

func validateCheckpoint(checkpoint Checkpoint) error {
	if strings.ContainsAny(checkpoint.ETag, "\r\n") || strings.ContainsAny(checkpoint.LastModified, "\r\n") {
		return fmt.Errorf("invalid conditional HTTP checkpoint")
	}
	return nil
}

func forbiddenRequestHeader(key string) bool {
	switch strings.ToLower(strings.TrimSpace(key)) {
	case "host", "connection", "proxy-authorization", "proxy-connection", "transfer-encoding", "upgrade":
		return true
	default:
		return false
	}
}
