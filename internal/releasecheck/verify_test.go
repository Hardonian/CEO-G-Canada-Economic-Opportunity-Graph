package releasecheck

import (
	"path/filepath"
	"testing"
)

func TestCheckedInRelease(t *testing.T) {
	repositoryRoot := filepath.Clean(filepath.Join("..", ".."))
	if err := Verify(repositoryRoot); err != nil {
		t.Fatal(err)
	}
}
