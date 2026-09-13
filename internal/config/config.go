package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultPort               = 8080
	defaultMaxHeaderBytes     = 16 << 10
	defaultMaxRequestBody     = 4 << 10
	defaultRateLimitPerMinute = 600
	defaultRateLimitBurst     = 100
)

// Config holds validated runtime configuration values. Secrets such as
// DatabaseURL must never be included in logs; use SafeSummary for diagnostics.
type Config struct {
	Env                  string
	BindAddress          string
	Port                 int
	DatabaseURL          string `json:"-"`
	LogLevel             string
	CORSOrigin           string // Raw, backwards-compatible CORS setting.
	CORSOrigins          []string
	PublicURL            string
	EnableHSTS           bool
	TrustedProxyCIDRs    []string
	MaxHeaderBytes       int
	MaxRequestBodyBytes  int64
	RateLimitPerMinute   int
	RateLimitBurst       int
	RequestTimeout       time.Duration
	ReadinessTimeout     time.Duration
	ReadHeaderTimeout    time.Duration
	ReadTimeout          time.Duration
	WriteTimeout         time.Duration
	IdleTimeout          time.Duration
	ShutdownTimeout      time.Duration
	InitialIngestTimeout time.Duration
}

// RuntimeSummary is safe to emit to application logs. It deliberately reports
// only whether a database is configured, never its URL, credentials, or query.
type RuntimeSummary struct {
	Environment        string   `json:"environment"`
	ListenAddress      string   `json:"listen_address"`
	PublicURL          string   `json:"public_url"`
	CORSOrigins        []string `json:"cors_origins"`
	DatabaseConfigured bool     `json:"database_configured"`
	HSTS               bool     `json:"hsts"`
	RateLimitPerMinute int      `json:"rate_limit_per_minute"`
	RateLimitBurst     int      `json:"rate_limit_burst"`
}

// Load retains the original convenience API and fails closed on malformed
// security settings. New executables should prefer LoadValidated so they can
// report the error through their normal startup path.
func Load() *Config {
	cfg, err := LoadValidated()
	if err != nil {
		panic(fmt.Errorf("invalid runtime configuration: %w", err))
	}
	return cfg
}

// LoadValidated reads and validates runtime configuration. Invalid values are
// returned as startup errors instead of silently weakening production limits.
func LoadValidated() (*Config, error) {
	cfg := defaults()
	cfg.Env = strings.ToLower(strings.TrimSpace(getEnv("ENV", cfg.Env)))
	cfg.BindAddress = strings.TrimSpace(getEnv("BIND_ADDRESS", cfg.BindAddress))
	cfg.DatabaseURL = getEnv("DATABASE_URL", "")
	cfg.LogLevel = strings.ToLower(strings.TrimSpace(getEnv("LOG_LEVEL", cfg.LogLevel)))
	cfg.PublicURL = strings.TrimSpace(getEnv("PUBLIC_URL", cfg.PublicURL))

	corsRaw := getEnv("CORS_ORIGINS", getEnv("CORS_ORIGIN", cfg.CORSOrigin))
	corsOrigins, err := parseOrigins(corsRaw)
	if err != nil {
		return nil, fmt.Errorf("CORS_ORIGINS: %w", err)
	}
	cfg.CORSOrigin = corsRaw
	cfg.CORSOrigins = corsOrigins

	cfg.Port, err = intEnv("PORT", cfg.Port, 1, 65535)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(os.Getenv("PUBLIC_URL")) == "" {
		cfg.PublicURL = fmt.Sprintf("http://localhost:%d", cfg.Port)
	}
	cfg.MaxHeaderBytes, err = intEnv("MAX_HEADER_BYTES", cfg.MaxHeaderBytes, 8<<10, 1<<20)
	if err != nil {
		return nil, err
	}
	maxBody, err := intEnv("MAX_REQUEST_BODY_BYTES", int(cfg.MaxRequestBodyBytes), 1, 1<<20)
	if err != nil {
		return nil, err
	}
	cfg.MaxRequestBodyBytes = int64(maxBody)
	cfg.RateLimitPerMinute, err = intEnv("RATE_LIMIT_PER_MINUTE", cfg.RateLimitPerMinute, 0, 1_000_000)
	if err != nil {
		return nil, err
	}
	cfg.RateLimitBurst, err = intEnv("RATE_LIMIT_BURST", cfg.RateLimitBurst, 1, 100_000)
	if err != nil {
		return nil, err
	}

	cfg.RequestTimeout, err = durationEnv("REQUEST_TIMEOUT", cfg.RequestTimeout, 100*time.Millisecond, 2*time.Minute)
	if err != nil {
		return nil, err
	}
	cfg.ReadinessTimeout, err = durationEnv("READINESS_TIMEOUT", cfg.ReadinessTimeout, 100*time.Millisecond, 30*time.Second)
	if err != nil {
		return nil, err
	}
	cfg.ReadHeaderTimeout, err = durationEnv("READ_HEADER_TIMEOUT", cfg.ReadHeaderTimeout, time.Second, time.Minute)
	if err != nil {
		return nil, err
	}
	cfg.ReadTimeout, err = durationEnv("READ_TIMEOUT", cfg.ReadTimeout, time.Second, 5*time.Minute)
	if err != nil {
		return nil, err
	}
	cfg.WriteTimeout, err = durationEnv("WRITE_TIMEOUT", cfg.WriteTimeout, time.Second, 5*time.Minute)
	if err != nil {
		return nil, err
	}
	cfg.IdleTimeout, err = durationEnv("IDLE_TIMEOUT", cfg.IdleTimeout, time.Second, 10*time.Minute)
	if err != nil {
		return nil, err
	}
	cfg.ShutdownTimeout, err = durationEnv("SHUTDOWN_TIMEOUT", cfg.ShutdownTimeout, time.Second, 2*time.Minute)
	if err != nil {
		return nil, err
	}
	cfg.InitialIngestTimeout, err = durationEnv("INITIAL_INGEST_TIMEOUT", cfg.InitialIngestTimeout, time.Second, 30*time.Minute)
	if err != nil {
		return nil, err
	}

	hstsDefault := strings.EqualFold(cfg.Env, "production")
	cfg.EnableHSTS, err = boolEnv("ENABLE_HSTS", hstsDefault)
	if err != nil {
		return nil, err
	}

	cfg.TrustedProxyCIDRs = splitCSV(getEnv("TRUSTED_PROXY_CIDRS", ""))
	for _, cidr := range cfg.TrustedProxyCIDRs {
		if _, _, parseErr := net.ParseCIDR(cidr); parseErr != nil {
			return nil, fmt.Errorf("TRUSTED_PROXY_CIDRS: invalid CIDR %q", cidr)
		}
	}

	if !validName(cfg.Env) {
		return nil, fmt.Errorf("ENV must contain only letters, digits, hyphens, or underscores")
	}
	if !validBindAddress(cfg.BindAddress) {
		return nil, fmt.Errorf("BIND_ADDRESS must be a host or IP address")
	}
	if !oneOf(cfg.LogLevel, "debug", "info", "warn", "error") {
		return nil, fmt.Errorf("LOG_LEVEL must be one of debug, info, warn, or error")
	}
	if err := validateHTTPOrigin(cfg.PublicURL); err != nil {
		return nil, fmt.Errorf("PUBLIC_URL: %w", err)
	}

	return cfg, nil
}

func defaults() *Config {
	return &Config{
		Env:                  "development",
		BindAddress:          "0.0.0.0",
		Port:                 defaultPort,
		LogLevel:             "info",
		CORSOrigin:           "*",
		CORSOrigins:          []string{"*"},
		PublicURL:            "http://localhost:8080",
		MaxHeaderBytes:       defaultMaxHeaderBytes,
		MaxRequestBodyBytes:  defaultMaxRequestBody,
		RateLimitPerMinute:   defaultRateLimitPerMinute,
		RateLimitBurst:       defaultRateLimitBurst,
		RequestTimeout:       15 * time.Second,
		ReadinessTimeout:     2 * time.Second,
		ReadHeaderTimeout:    5 * time.Second,
		ReadTimeout:          20 * time.Second,
		WriteTimeout:         30 * time.Second,
		IdleTimeout:          60 * time.Second,
		ShutdownTimeout:      15 * time.Second,
		InitialIngestTimeout: 2 * time.Minute,
	}
}

func (c *Config) ListenAddress() string {
	return net.JoinHostPort(c.BindAddress, strconv.Itoa(c.Port))
}

func (c *Config) SafeSummary() RuntimeSummary {
	return RuntimeSummary{
		Environment:        c.Env,
		ListenAddress:      c.ListenAddress(),
		PublicURL:          c.PublicURL,
		CORSOrigins:        append([]string(nil), c.CORSOrigins...),
		DatabaseConfigured: strings.TrimSpace(c.DatabaseURL) != "",
		HSTS:               c.EnableHSTS,
		RateLimitPerMinute: c.RateLimitPerMinute,
		RateLimitBurst:     c.RateLimitBurst,
	}
}

// String and GoString prevent common formatted logging from exposing
// DatabaseURL. Callers needing the full configuration should access fields
// explicitly and keep secret-bearing values out of logs.
func (c Config) String() string {
	summary := c.SafeSummary()
	return fmt.Sprintf("{environment:%q listen_address:%q public_url:%q cors_origins:%q database_configured:%t hsts:%t rate_limit_per_minute:%d rate_limit_burst:%d}",
		summary.Environment,
		summary.ListenAddress,
		summary.PublicURL,
		summary.CORSOrigins,
		summary.DatabaseConfigured,
		summary.HSTS,
		summary.RateLimitPerMinute,
		summary.RateLimitBurst,
	)
}

func (c Config) GoString() string {
	return c.String()
}

func parseOrigins(raw string) ([]string, error) {
	origins := splitCSV(raw)
	if len(origins) == 0 {
		return nil, nil
	}
	if len(origins) > 1 {
		for _, origin := range origins {
			if origin == "*" {
				return nil, fmt.Errorf("wildcard origin cannot be combined with explicit origins")
			}
		}
	}

	seen := make(map[string]struct{}, len(origins))
	result := make([]string, 0, len(origins))
	for _, origin := range origins {
		if origin != "*" {
			if err := validateHTTPOrigin(origin); err != nil {
				return nil, err
			}
			origin = strings.TrimSuffix(origin, "/")
		}
		if _, ok := seen[origin]; ok {
			continue
		}
		seen[origin] = struct{}{}
		result = append(result, origin)
	}
	return result, nil
}

func validateHTTPOrigin(raw string) error {
	if len(raw) > 2048 || strings.ContainsAny(raw, "\r\n\t") {
		return fmt.Errorf("origin is malformed")
	}
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return fmt.Errorf("origin %q must be an absolute http(s) URL", raw)
	}
	if u.User != nil || u.RawQuery != "" || u.Fragment != "" || (u.Path != "" && u.Path != "/") {
		return fmt.Errorf("origin %q must not contain credentials, a path, query, or fragment", raw)
	}
	return nil
}

func validBindAddress(host string) bool {
	if net.ParseIP(host) != nil {
		return true
	}
	if host == "" || len(host) > 253 || strings.HasPrefix(host, ".") || strings.HasSuffix(host, ".") {
		return false
	}
	for _, label := range strings.Split(host, ".") {
		if label == "" || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for _, char := range label {
			if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') ||
				(char >= '0' && char <= '9') || char == '-' {
				continue
			}
			return false
		}
	}
	return true
}

func validName(value string) bool {
	if value == "" || len(value) > 32 {
		return false
	}
	for _, char := range value {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '-' || char == '_' {
			continue
		}
		return false
	}
	return true
}

func intEnv(key string, defaultValue, min, max int) (int, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return defaultValue, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < min || value > max {
		return 0, fmt.Errorf("%s must be an integer between %d and %d", key, min, max)
	}
	return value, nil
}

func durationEnv(key string, defaultValue, min, max time.Duration) (time.Duration, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return defaultValue, nil
	}
	value, err := time.ParseDuration(raw)
	if err != nil || value < min || value > max {
		return 0, fmt.Errorf("%s must be a duration between %s and %s", key, min, max)
	}
	return value, nil
}

func boolEnv(key string, defaultValue bool) (bool, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return defaultValue, nil
	}
	value, err := strconv.ParseBool(raw)
	if err != nil {
		return false, fmt.Errorf("%s must be a boolean", key)
	}
	return value, nil
}

func splitCSV(raw string) []string {
	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if value := strings.TrimSpace(part); value != "" {
			result = append(result, value)
		}
	}
	return result
}

func oneOf(value string, allowed ...string) bool {
	for _, candidate := range allowed {
		if value == candidate {
			return true
		}
	}
	return false
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
