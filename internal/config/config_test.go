package config

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestLoadValidatedProductionSecuritySettings(t *testing.T) {
	clearConfigEnvironment(t)
	t.Setenv("ENV", "production")
	t.Setenv("BIND_ADDRESS", "127.0.0.1")
	t.Setenv("PORT", "9090")
	t.Setenv("DATABASE_URL", "postgres://admin:super-secret@db.internal/cog?sslmode=require&token=also-secret")
	t.Setenv("CORS_ORIGINS", "https://planner.gc.ca/, https://investor.example")
	t.Setenv("TRUSTED_PROXY_CIDRS", "10.0.0.0/8, 2001:db8::/32")
	t.Setenv("REQUEST_TIMEOUT", "12s")
	t.Setenv("RATE_LIMIT_PER_MINUTE", "300")
	t.Setenv("RATE_LIMIT_BURST", "40")

	cfg, err := LoadValidated()
	if err != nil {
		t.Fatalf("LoadValidated: %v", err)
	}
	if !cfg.EnableHSTS {
		t.Fatal("production configuration did not enable HSTS by default")
	}
	if actual := cfg.ListenAddress(); actual != "127.0.0.1:9090" {
		t.Fatalf("ListenAddress = %q", actual)
	}
	if cfg.RequestTimeout != 12*time.Second {
		t.Fatalf("RequestTimeout = %s", cfg.RequestTimeout)
	}
	wantOrigins := []string{"https://planner.gc.ca", "https://investor.example"}
	if !reflect.DeepEqual(cfg.CORSOrigins, wantOrigins) {
		t.Fatalf("CORSOrigins = %#v, want %#v", cfg.CORSOrigins, wantOrigins)
	}

	summaryJSON, err := json.Marshal(cfg.SafeSummary())
	if err != nil {
		t.Fatalf("marshal SafeSummary: %v", err)
	}
	formatted := strings.Join([]string{
		string(summaryJSON),
		fmt.Sprintf("%v", *cfg),
		fmt.Sprintf("%+v", *cfg),
		fmt.Sprintf("%#v", *cfg),
	}, "\n")
	for _, secret := range []string{"super-secret", "also-secret", "postgres://"} {
		if strings.Contains(formatted, secret) {
			t.Fatalf("safe configuration output leaked %q: %s", secret, formatted)
		}
	}
	if !cfg.SafeSummary().DatabaseConfigured {
		t.Fatal("safe summary should indicate that a database is configured")
	}
}

func TestLoadValidatedSupportsLegacyCORSOrigin(t *testing.T) {
	clearConfigEnvironment(t)
	t.Setenv("CORS_ORIGIN", "https://legacy.example")

	cfg, err := LoadValidated()
	if err != nil {
		t.Fatalf("LoadValidated: %v", err)
	}
	if !reflect.DeepEqual(cfg.CORSOrigins, []string{"https://legacy.example"}) {
		t.Fatalf("CORSOrigins = %#v", cfg.CORSOrigins)
	}
}

func TestLoadValidatedRejectsUnsafeOrMalformedSettings(t *testing.T) {
	tests := map[string]struct {
		key   string
		value string
	}{
		"invalid port":             {"PORT", "70000"},
		"invalid duration":         {"REQUEST_TIMEOUT", "forever"},
		"weak header ceiling":      {"MAX_HEADER_BYTES", "1024"},
		"invalid log level":        {"LOG_LEVEL", "verbose"},
		"invalid public URL":       {"PUBLIC_URL", "https://example.ca/private/path"},
		"invalid trusted proxy":    {"TRUSTED_PROXY_CIDRS", "10.0.0.1"},
		"mixed wildcard origins":   {"CORS_ORIGINS", "*,https://planner.gc.ca"},
		"credentialed CORS origin": {"CORS_ORIGINS", "https://user:password@planner.gc.ca"},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			clearConfigEnvironment(t)
			t.Setenv(test.key, test.value)
			if _, err := LoadValidated(); err == nil {
				t.Fatalf("LoadValidated accepted %s=%q", test.key, test.value)
			}
		})
	}
}

func TestLoadValidatedAllowsExplicitRateLimitDisableAndHSTSOverride(t *testing.T) {
	clearConfigEnvironment(t)
	t.Setenv("ENV", "production")
	t.Setenv("ENABLE_HSTS", "false")
	t.Setenv("RATE_LIMIT_PER_MINUTE", "0")

	cfg, err := LoadValidated()
	if err != nil {
		t.Fatalf("LoadValidated: %v", err)
	}
	if cfg.EnableHSTS {
		t.Fatal("explicit HSTS override was ignored")
	}
	if cfg.RateLimitPerMinute != 0 {
		t.Fatalf("RateLimitPerMinute = %d", cfg.RateLimitPerMinute)
	}
}

func TestLoadFailsClosedOnInvalidConfiguration(t *testing.T) {
	clearConfigEnvironment(t)
	t.Setenv("PORT", "not-a-port")
	defer func() {
		if recover() == nil {
			t.Fatal("Load did not fail closed")
		}
	}()
	_ = Load()
}

func clearConfigEnvironment(t *testing.T) {
	t.Helper()
	for _, key := range []string{
		"ENV",
		"BIND_ADDRESS",
		"PORT",
		"DATABASE_URL",
		"LOG_LEVEL",
		"CORS_ORIGIN",
		"CORS_ORIGINS",
		"PUBLIC_URL",
		"ENABLE_HSTS",
		"TRUSTED_PROXY_CIDRS",
		"MAX_HEADER_BYTES",
		"MAX_REQUEST_BODY_BYTES",
		"RATE_LIMIT_PER_MINUTE",
		"RATE_LIMIT_BURST",
		"REQUEST_TIMEOUT",
		"READINESS_TIMEOUT",
		"READ_HEADER_TIMEOUT",
		"READ_TIMEOUT",
		"WRITE_TIMEOUT",
		"IDLE_TIMEOUT",
		"SHUTDOWN_TIMEOUT",
		"INITIAL_INGEST_TIMEOUT",
	} {
		t.Setenv(key, "")
	}
}
