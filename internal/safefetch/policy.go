// Package safefetch provides a bounded HTTP client for retrieving untrusted
// public data. It deliberately rejects destinations that could reach local,
// private, link-local, metadata, or otherwise non-public network space.
package safefetch

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"net/url"
	"strconv"
	"strings"
)

// ErrUnsafeDestination identifies URLs or resolved addresses rejected by the
// public-network egress policy.
var ErrUnsafeDestination = errors.New("unsafe public HTTP destination")

// Resolver is the DNS surface used by the egress policy. Keeping it small
// makes DNS and rebinding behaviour deterministic in tests.
type Resolver interface {
	LookupIP(ctx context.Context, network, host string) ([]net.IP, error)
}

// PublicHTTPPolicy permits only HTTP(S) requests to DNS names and addresses
// that resolve exclusively to public unicast space. Plain HTTP is enabled
// because some public-data publishers still expose it; callers may select the
// HTTPS-only option for stricter deployments.
type PublicHTTPPolicy struct {
	resolver        Resolver
	allowHTTP       bool
	allowedPorts    map[uint16]struct{}
	blockedSuffixes []string
}

// PolicyOption customizes a PublicHTTPPolicy.
type PolicyOption func(*PublicHTTPPolicy)

// HTTPSOnly disables plain HTTP destinations.
func HTTPSOnly() PolicyOption {
	return func(policy *PublicHTTPPolicy) { policy.allowHTTP = false }
}

// AllowPorts replaces the default port allowlist (80 and 443). An empty list
// retains the secure defaults.
func AllowPorts(ports ...uint16) PolicyOption {
	return func(policy *PublicHTTPPolicy) {
		if len(ports) == 0 {
			return
		}
		policy.allowedPorts = make(map[uint16]struct{}, len(ports))
		for _, port := range ports {
			if port != 0 {
				policy.allowedPorts[port] = struct{}{}
			}
		}
	}
}

// NewPublicHTTPPolicy constructs a deny-by-default destination validator.
func NewPublicHTTPPolicy(resolver Resolver, options ...PolicyOption) *PublicHTTPPolicy {
	if resolver == nil {
		resolver = net.DefaultResolver
	}
	policy := &PublicHTTPPolicy{
		resolver:     resolver,
		allowHTTP:    true,
		allowedPorts: map[uint16]struct{}{80: {}, 443: {}},
		blockedSuffixes: []string{
			".localhost", ".local", ".internal", ".home.arpa",
			".test", ".example", ".invalid",
		},
	}
	for _, option := range options {
		if option != nil {
			option(policy)
		}
	}
	return policy
}

// Validate resolves and validates a URL before a request or redirect is sent.
// It returns a normalized copy of the URL and the public addresses observed.
func (p *PublicHTTPPolicy) Validate(ctx context.Context, rawURL string) (*url.URL, []netip.Addr, error) {
	if p == nil {
		return nil, nil, fmt.Errorf("%w: missing policy", ErrUnsafeDestination)
	}
	if len(rawURL) == 0 || len(rawURL) > 8192 {
		return nil, nil, fmt.Errorf("%w: URL length is invalid", ErrUnsafeDestination)
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: parse URL: %v", ErrUnsafeDestination, err)
	}
	parsed.Scheme = strings.ToLower(parsed.Scheme)
	switch parsed.Scheme {
	case "https":
	case "http":
		if !p.allowHTTP {
			return nil, nil, fmt.Errorf("%w: plain HTTP is disabled", ErrUnsafeDestination)
		}
	default:
		return nil, nil, fmt.Errorf("%w: scheme %q is not HTTP(S)", ErrUnsafeDestination, parsed.Scheme)
	}
	if parsed.User != nil {
		return nil, nil, fmt.Errorf("%w: URL user information is forbidden", ErrUnsafeDestination)
	}
	if parsed.Hostname() == "" {
		return nil, nil, fmt.Errorf("%w: URL has no host", ErrUnsafeDestination)
	}
	if parsed.Opaque != "" {
		return nil, nil, fmt.Errorf("%w: opaque URLs are forbidden", ErrUnsafeDestination)
	}
	port, err := effectivePort(parsed)
	if err != nil {
		return nil, nil, err
	}
	if _, allowed := p.allowedPorts[port]; !allowed {
		return nil, nil, fmt.Errorf("%w: port %d is not allowed", ErrUnsafeDestination, port)
	}

	host := normalizeHost(parsed.Hostname())
	if err := p.validateHostname(host); err != nil {
		return nil, nil, err
	}
	addresses, err := p.resolvePublic(ctx, host)
	if err != nil {
		return nil, nil, err
	}
	parsed.Host = normalizedAuthority(host, parsed.Port())
	return parsed, addresses, nil
}

func (p *PublicHTTPPolicy) validateHostname(host string) error {
	if host == "" {
		return fmt.Errorf("%w: empty host", ErrUnsafeDestination)
	}
	lower := strings.ToLower(strings.TrimSuffix(host, "."))
	if lower == "localhost" || lower == "metadata" || lower == "metadata.google.internal" {
		return fmt.Errorf("%w: blocked hostname %q", ErrUnsafeDestination, host)
	}
	if net.ParseIP(lower) == nil && !strings.Contains(lower, ".") {
		return fmt.Errorf("%w: single-label hostname %q is not public", ErrUnsafeDestination, host)
	}
	for _, suffix := range p.blockedSuffixes {
		if strings.HasSuffix(lower, suffix) {
			return fmt.Errorf("%w: blocked hostname suffix in %q", ErrUnsafeDestination, host)
		}
	}
	return nil
}

func (p *PublicHTTPPolicy) resolvePublic(ctx context.Context, host string) ([]netip.Addr, error) {
	if literal, err := netip.ParseAddr(strings.Trim(host, "[]")); err == nil {
		literal = literal.Unmap()
		if !isPublicAddress(literal) {
			return nil, fmt.Errorf("%w: address %s is not public", ErrUnsafeDestination, literal)
		}
		return []netip.Addr{literal}, nil
	}

	resolved, err := p.resolver.LookupIP(ctx, "ip", host)
	if err != nil {
		return nil, fmt.Errorf("resolve public host %q: %w", host, err)
	}
	if len(resolved) == 0 {
		return nil, fmt.Errorf("resolve public host %q: no addresses", host)
	}
	addresses := make([]netip.Addr, 0, len(resolved))
	seen := make(map[netip.Addr]struct{}, len(resolved))
	for _, value := range resolved {
		address, ok := netip.AddrFromSlice(value)
		if !ok {
			return nil, fmt.Errorf("%w: invalid resolved address for %q", ErrUnsafeDestination, host)
		}
		address = address.Unmap()
		if !isPublicAddress(address) {
			// Reject the whole destination when DNS mixes public and private
			// answers; selecting only the public answer is unsafe under rebinding.
			return nil, fmt.Errorf("%w: host %q resolved to non-public address %s", ErrUnsafeDestination, host, address)
		}
		if _, exists := seen[address]; !exists {
			seen[address] = struct{}{}
			addresses = append(addresses, address)
		}
	}
	return addresses, nil
}

func effectivePort(parsed *url.URL) (uint16, error) {
	if value := parsed.Port(); value != "" {
		port, err := strconv.ParseUint(value, 10, 16)
		if err != nil || port == 0 {
			return 0, fmt.Errorf("%w: invalid port %q", ErrUnsafeDestination, value)
		}
		return uint16(port), nil
	}
	if parsed.Scheme == "https" {
		return 443, nil
	}
	return 80, nil
}

func normalizedAuthority(host, explicitPort string) string {
	authority := host
	if strings.Contains(host, ":") {
		authority = "[" + strings.Trim(host, "[]") + "]"
	}
	if explicitPort != "" {
		authority += ":" + explicitPort
	}
	return authority
}

func normalizeHost(host string) string {
	return strings.ToLower(strings.TrimSuffix(strings.TrimSpace(host), "."))
}

var nonPublicPrefixes = []netip.Prefix{
	// IPv4 special-purpose and non-forwardable ranges not completely covered
	// by netip's IsPrivate/IsLoopback/IsLinkLocal helpers.
	netip.MustParsePrefix("0.0.0.0/8"),
	netip.MustParsePrefix("100.64.0.0/10"),
	netip.MustParsePrefix("192.0.0.0/24"),
	netip.MustParsePrefix("192.0.2.0/24"),
	netip.MustParsePrefix("198.18.0.0/15"),
	netip.MustParsePrefix("198.51.100.0/24"),
	netip.MustParsePrefix("203.0.113.0/24"),
	netip.MustParsePrefix("240.0.0.0/4"),
	// IPv6 documentation and discard-only prefixes.
	netip.MustParsePrefix("100::/64"),
	netip.MustParsePrefix("2001:db8::/32"),
}

func isPublicAddress(address netip.Addr) bool {
	if !address.IsValid() {
		return false
	}
	address = address.Unmap()
	if !address.IsGlobalUnicast() || address.IsPrivate() || address.IsLoopback() ||
		address.IsLinkLocalUnicast() || address.IsLinkLocalMulticast() ||
		address.IsMulticast() || address.IsUnspecified() {
		return false
	}
	for _, prefix := range nonPublicPrefixes {
		if prefix.Contains(address) {
			return false
		}
	}
	return true
}
