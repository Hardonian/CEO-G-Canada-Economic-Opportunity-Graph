package main

import (
	"fmt"
	"os"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/releasecheck"
)

func main() {
	repositoryRoot := "."
	if len(os.Args) > 2 {
		fmt.Fprintln(os.Stderr, "usage: releasecheck [repository-root]")
		os.Exit(2)
	}
	if len(os.Args) == 2 {
		repositoryRoot = os.Args[1]
	}
	if err := releasecheck.Verify(repositoryRoot); err != nil {
		fmt.Fprintln(os.Stderr, "[ERROR] release integrity check failed:", err)
		os.Exit(1)
	}
	fmt.Println("[OK] release hashes, counts, ordering, CEGS conformance, and references verified")
}
