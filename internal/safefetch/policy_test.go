package safefetch

import (
	"context"
	"errors"
	"net"
	"testing"
)

type resolverFunc func(context.Context, string, string) ([]net.IP, error)

func (fn resolverFunc) LookupIP(ctx context.Context, network, host string) ([]net.IP, error) {
	return fn(ctx, network, host)
}

func staticResolver(records map[string][]string) Resolver {
	return resolverFunc(func(_ context.Context, _ string, host string) ([]net.IP, error) {
		values, ok := records[host]
		if !ok {
			return nil, errors.New("host not in test DNS")
		}
		addresses := make([]net.IP, 0, len(values))
		for _, value := range values {
			addresses = append(addresses, net.ParseIP(value))
		}
		return addresses, nil
	})
}

func TestPublicHTTPPolicyAcceptsPublicHTTPAndHTTPS(t *testing.T) {
	policy := NewPublicHTTPPolicy(staticResolver(map[string][]string{
		"data.canada.ca": {"93.184.216.34", "2606:2800:220:1:248:1893:25c8:1946"},
	}))
	for _, rawURL := range []string{
		"https://DATA.CANADA.CA./catalog?q=energy",
		"http://data.canada.ca/feed.xml",
	} {
		parsed, addresses, err := policy.Validate(context.Background(), rawURL)
		if err != nil {
			t.Fatalf("Validate(%q) error = %v", rawURL, err)
		}
		if parsed.Hostname() != "data.canada.ca" || len(addresses) != 2 {
			t.Fatalf("Validate(%q) = host %q, addresses %v", rawURL, parsed.Hostname(), addresses)
		}
	}
}

func TestPublicHTTPPolicyRejectsUnsafeURLs(t *testing.T) {
	policy := NewPublicHTTPPolicy(staticResolver(map[string][]string{
		"data.canada.ca": {"93.184.216.34"},
	}))
	tests := []string{
		"",
		"ftp://data.canada.ca/file",
		"http://localhost/admin",
		"http://service.internal/data",
		"http://metadata.google.internal/computeMetadata/v1/",
		"http://singlelabel/path",
		"http://127.0.0.1/",
		"http://10.1.2.3/",
		"http://169.254.169.254/latest/meta-data/",
		"http://[::1]/",
		"http://[::ffff:127.0.0.1]/",
		"https://user:password@data.canada.ca/",
		"https://data.canada.ca:22/",
	}
	for _, rawURL := range tests {
		t.Run(rawURL, func(t *testing.T) {
			if _, _, err := policy.Validate(context.Background(), rawURL); !errors.Is(err, ErrUnsafeDestination) {
				t.Fatalf("Validate(%q) error = %v, want ErrUnsafeDestination", rawURL, err)
			}
		})
	}
}

func TestPublicHTTPPolicyRejectsPrivateAndMixedDNSAnswers(t *testing.T) {
	policy := NewPublicHTTPPolicy(staticResolver(map[string][]string{
		"private.data.ca": {"10.0.0.8"},
		"mixed.data.ca":   {"93.184.216.34", "192.168.1.5"},
	}))
	for _, host := range []string{"private.data.ca", "mixed.data.ca"} {
		if _, _, err := policy.Validate(context.Background(), "https://"+host+"/data"); !errors.Is(err, ErrUnsafeDestination) {
			t.Fatalf("Validate(%q) error = %v, want ErrUnsafeDestination", host, err)
		}
	}
}

func TestPublicHTTPPolicyHTTPSOnlyAndExplicitPorts(t *testing.T) {
	resolver := staticResolver(map[string][]string{"data.canada.ca": {"93.184.216.34"}})
	httpsOnly := NewPublicHTTPPolicy(resolver, HTTPSOnly())
	if _, _, err := httpsOnly.Validate(context.Background(), "http://data.canada.ca/"); !errors.Is(err, ErrUnsafeDestination) {
		t.Fatalf("plain HTTP error = %v, want ErrUnsafeDestination", err)
	}

	customPort := NewPublicHTTPPolicy(resolver, AllowPorts(8443))
	if _, _, err := customPort.Validate(context.Background(), "https://data.canada.ca:8443/"); err != nil {
		t.Fatalf("custom port rejected: %v", err)
	}
	if _, _, err := customPort.Validate(context.Background(), "https://data.canada.ca/"); !errors.Is(err, ErrUnsafeDestination) {
		t.Fatalf("default port unexpectedly accepted: %v", err)
	}
}
