package adaptersandbox

import (
	"fmt"
	"testing"
	"time"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/domain"
)

func TestNewRegistry_Empty(t *testing.T) {
	r := NewRegistry()
	if r == nil {
		t.Fatal("expected non-nil registry")
	}
	if r.Count() != 0 {
		t.Fatalf("expected 0 entries, got %d", r.Count())
	}
	if len(r.All()) != 0 {
		t.Fatal("expected empty All()")
	}
	if len(r.Names()) != 0 {
		t.Fatal("expected empty Names()")
	}
	if len(r.Approved()) != 0 {
		t.Fatal("expected empty Approved()")
	}
}

func TestRegister_ValidEntry(t *testing.T) {
	r := NewRegistry()
	entry := &AdapterEntry{
		Name:         "municipal_adapter",
		Version:      "1.0",
		SourceURL:    "https://example.com/adapter",
		Tier:         domain.SourceTier3,
		Contact:      "dev@example.com",
		Description:  "Test adapter",
		Capabilities: []string{"scraping"},
		RegisteredAt: time.Now().UTC(),
		Approved:     true,
	}
	if err := r.Register(entry); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if r.Count() != 1 {
		t.Fatalf("expected 1 entry, got %d", r.Count())
	}
	got := r.Get("municipal_adapter")
	if got == nil {
		t.Fatal("expected non-nil entry")
	}
	if got.Name != "municipal_adapter" {
		t.Errorf("name = %q, want municipal_adapter", got.Name)
	}
}

func TestRegister_DuplicateName(t *testing.T) {
	r := NewRegistry()
	entry := &AdapterEntry{
		Name:        "dup",
		Version:     "1.0",
		SourceURL:   "https://example.com",
		Tier:        domain.SourceTier1,
		Contact:     "dev@example.com",
		Description: "Test",
		RegisteredAt: time.Now().UTC(),
	}
	if err := r.Register(entry); err != nil {
		t.Fatalf("first Register() error = %v", err)
	}
	// Replace with a new entry.
	entry2 := &AdapterEntry{
		Name:        "dup",
		Version:     "2.0",
		SourceURL:   "https://example.com/v2",
		Tier:        domain.SourceTier2,
		Contact:     "dev@example.com",
		Description: "Updated",
		RegisteredAt: time.Now().UTC(),
	}
	if err := r.Register(entry2); err != nil {
		t.Fatalf("second Register() error = %v", err)
	}
	got := r.Get("dup")
	if got == nil {
		t.Fatal("expected non-nil entry")
	}
	if got.Version != "2.0" {
		t.Errorf("version = %q, want 2.0 (replacement should win)", got.Version)
	}
	// Count should still be 1.
	if r.Count() != 1 {
		t.Fatalf("expected 1 entry after replacement, got %d", r.Count())
	}
}

func TestRegister_InvalidName(t *testing.T) {
	r := NewRegistry()
	entry := &AdapterEntry{
		Name:        "",
		Version:     "1.0",
		SourceURL:   "https://example.com",
		Tier:        domain.SourceTier1,
		Contact:     "dev@example.com",
		Description: "Test",
		RegisteredAt: time.Now().UTC(),
	}
	if err := r.Register(entry); err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestRegister_InvalidURL(t *testing.T) {
	r := NewRegistry()
	entry := &AdapterEntry{
		Name:        "bad",
		Version:     "1.0",
		SourceURL:   "not-a-url",
		Tier:        domain.SourceTier1,
		Contact:     "dev@example.com",
		Description: "Test",
		RegisteredAt: time.Now().UTC(),
	}
	if err := r.Register(entry); err == nil {
		t.Fatalf("expected error for invalid URL, got nil")
	}
	// Also test a URL with no scheme.
	entry2 := &AdapterEntry{
		Name:        "bad2",
		Version:     "1.0",
		SourceURL:   "example.com/path",
		Tier:        domain.SourceTier1,
		Contact:     "dev@example.com",
		Description: "Test",
		RegisteredAt: time.Now().UTC(),
	}
	if err := r.Register(entry2); err == nil {
		t.Fatalf("expected error for URL without scheme, got nil")
	}
}

func TestRegister_NilEntry(t *testing.T) {
	r := NewRegistry()
	if err := r.Register(nil); err == nil {
		t.Fatal("expected error for nil entry")
	}
}

func TestApprovedFiltering(t *testing.T) {
	r := NewRegistry()
	now := time.Now().UTC()
	r.Register(&AdapterEntry{
		Name:        "approved_adapter",
		Version:     "1.0",
		SourceURL:   "https://example.com/approved",
		Tier:        domain.SourceTier1,
		Contact:     "dev@example.com",
		Description: "Approved",
		RegisteredAt: now,
		Approved:    true,
	})
	r.Register(&AdapterEntry{
		Name:        "pending_adapter",
		Version:     "1.0",
		SourceURL:   "https://example.com/pending",
		Tier:        domain.SourceTier2,
		Contact:     "dev@example.com",
		Description: "Pending",
		RegisteredAt: now,
		Approved:    false,
	})

	approved := r.Approved()
	if len(approved) != 1 {
		t.Fatalf("expected 1 approved entry, got %d", len(approved))
	}
	if approved[0].Name != "approved_adapter" {
		t.Errorf("approved[0].name = %q, want approved_adapter", approved[0].Name)
	}
}

func TestAllAndNamesOrder(t *testing.T) {
	r := NewRegistry()
	now := time.Now().UTC()
	for _, name := range []string{"z_adapter", "a_adapter", "m_adapter"} {
		r.Register(&AdapterEntry{
			Name:        name,
			Version:     "1.0",
			SourceURL:   "https://example.com/" + name,
			Tier:        domain.SourceTier1,
			Contact:     "dev@example.com",
			Description: "Test",
			RegisteredAt: now,
		})
	}
	names := r.Names()
	if len(names) != 3 {
		t.Fatalf("expected 3 names, got %d", len(names))
	}
	// Insertion order should be preserved.
	if names[0] != "z_adapter" || names[1] != "a_adapter" || names[2] != "m_adapter" {
		t.Errorf("names = %v, want [z_adapter a_adapter m_adapter]", names)
	}
	all := r.All()
	if len(all) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(all))
	}
}

func TestGetAbsent(t *testing.T) {
	r := NewRegistry()
	if got := r.Get("nonexistent"); got != nil {
		t.Fatalf("expected nil for absent entry, got %v", got)
	}
}

func TestRegistryVersion(t *testing.T) {
	if RegistryVersion != "adapter-sandbox-v1.0" {
		t.Errorf("RegistryVersion = %q, want adapter-sandbox-v1.0", RegistryVersion)
	}
}

func TestApprove_Success(t *testing.T) {
	r := NewRegistry()
	now := time.Now().UTC()
	r.Register(&AdapterEntry{
		Name:         "pending_adapter",
		Version:      "1.0",
		SourceURL:    "https://example.com/pending",
		Tier:         domain.SourceTier2,
		Contact:      "dev@example.com",
		Description:  "Pending review",
		RegisteredAt: now,
		Approved:     false,
	})
	// Adapter should not be in Approved() initially.
	if len(r.Approved()) != 0 {
		t.Fatal("expected 0 approved entries before approval")
	}
	if err := r.Approve("pending_adapter"); err != nil {
		t.Fatalf("Approve() error = %v", err)
	}
	// Now it should be approved.
	approved := r.Approved()
	if len(approved) != 1 {
		t.Fatalf("expected 1 approved entry, got %d", len(approved))
	}
	if approved[0].Name != "pending_adapter" {
		t.Errorf("approved[0].Name = %q, want pending_adapter", approved[0].Name)
	}
}

func TestApprove_NotFound(t *testing.T) {
	r := NewRegistry()
	if err := r.Approve("nonexistent"); err == nil {
		t.Fatal("expected error approving nonexistent adapter")
	}
}

func TestApprove_Idempotent(t *testing.T) {
	r := NewRegistry()
	now := time.Now().UTC()
	r.Register(&AdapterEntry{
		Name:         "idem_adapter",
		Version:      "1.0",
		SourceURL:    "https://example.com/idem",
		Tier:         domain.SourceTier1,
		Contact:      "dev@example.com",
		Description:  "Idempotency test",
		RegisteredAt: now,
		Approved:     false,
	})
	r.Approve("idem_adapter")
	// Approve again — should not error.
	if err := r.Approve("idem_adapter"); err != nil {
		t.Fatalf("second Approve() error = %v", err)
	}
	if len(r.Approved()) != 1 {
		t.Error("expected exactly 1 approved entry")
	}
}

func TestReject_Success(t *testing.T) {
	r := NewRegistry()
	now := time.Now().UTC()
	r.Register(&AdapterEntry{
		Name:         "reject_me",
		Version:      "1.0",
		SourceURL:    "https://example.com/reject",
		Tier:         domain.SourceTier3,
		Contact:      "dev@example.com",
		Description:  "To be rejected",
		RegisteredAt: now,
	})
	if r.Count() != 1 {
		t.Fatal("precondition: expected 1 entry")
	}
	if !r.Reject("reject_me") {
		t.Fatal("Reject() returned false, expected true")
	}
	if r.Count() != 0 {
		t.Errorf("expected 0 entries after reject, got %d", r.Count())
	}
	if got := r.Get("reject_me"); got != nil {
		t.Error("rejected entry should not be retrievable")
	}
}

func TestReject_NotFound(t *testing.T) {
	r := NewRegistry()
	if r.Reject("ghost") {
		t.Fatal("Reject() returned true for nonexistent entry")
	}
}

func TestDelete_AliasForReject(t *testing.T) {
	r := NewRegistry()
	now := time.Now().UTC()
	r.Register(&AdapterEntry{
		Name:         "delete_me",
		Version:      "1.0",
		SourceURL:    "https://example.com/delete",
		Tier:         domain.SourceTier1,
		Contact:      "dev@example.com",
		Description:  "To be deleted",
		RegisteredAt: now,
	})
	if !r.Delete("delete_me") {
		t.Fatal("Delete() returned false, expected true")
	}
	if r.Count() != 0 {
		t.Errorf("expected 0 entries after delete, got %d", r.Count())
	}
}

func TestReject_PreservesOrder(t *testing.T) {
	r := NewRegistry()
	now := time.Now().UTC()
	for _, name := range []string{"first", "middle", "last"} {
		r.Register(&AdapterEntry{
			Name:         name,
			Version:      "1.0",
			SourceURL:    "https://example.com/" + name,
			Tier:         domain.SourceTier1,
			Contact:      "dev@example.com",
			Description:  name,
			RegisteredAt: now,
		})
	}
	r.Reject("middle")
	names := r.Names()
	if len(names) != 2 {
		t.Fatalf("expected 2 remaining, got %d", len(names))
	}
	if names[0] != "first" || names[1] != "last" {
		t.Errorf("names = %v, want [first last]", names)
	}
}

func TestConcurrentAccess(t *testing.T) {
	r := NewRegistry()
	done := make(chan bool)
	// Spawn writers.
	for i := 0; i < 10; i++ {
		go func(idx int) {
			r.Register(&AdapterEntry{
				Name:         fmt.Sprintf("adapter-%d", idx),
				Version:      "1.0",
				SourceURL:    fmt.Sprintf("https://example.com/%d", idx),
				Tier:         domain.SourceTier1,
				Contact:      "dev@example.com",
				Description:  "concurrent test",
				RegisteredAt: time.Now().UTC(),
			})
			done <- true
		}(i)
	}
	// Spawn readers.
	for i := 0; i < 10; i++ {
		go func() {
			_ = r.All()
			_ = r.Names()
			_ = r.Approved()
			_ = r.Count()
			done <- true
		}()
	}
	for i := 0; i < 20; i++ {
		<-done
	}
	if r.Count() != 10 {
		t.Errorf("expected 10 entries after concurrent writes, got %d", r.Count())
	}
}