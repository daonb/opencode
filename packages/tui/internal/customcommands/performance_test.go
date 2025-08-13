package customcommands

import (
	"os"
	"path/filepath"
	"testing"
)

func BenchmarkRegistryWithoutCache(b *testing.B) {
	// Create a temporary directory with multiple command files
	tempDir, err := os.MkdirTemp("", "customcommands_bench")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create multiple test command files
	for i := 0; i < 10; i++ {
		commandContent := `---
name: test-command-%d
description: A test command %d
trigger: ["test%d"]
---

This is test command %d.
`
		commandFile := filepath.Join(tempDir, "test%d.md")
		if err := os.WriteFile(commandFile, []byte(commandContent), 0644); err != nil {
			b.Fatalf("Failed to write test command file: %v", err)
		}
	}

	// Create scanner
	scanner := NewScanner(tempDir)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate old behavior - always scan from disk
		_, _ = scanner.ScanCommands()
	}
}

func BenchmarkRegistryWithCache(b *testing.B) {
	// Create a temporary directory with multiple command files
	tempDir, err := os.MkdirTemp("", "customcommands_bench")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create multiple test command files
	for i := 0; i < 10; i++ {
		commandContent := `---
name: test-command-%d
description: A test command %d
trigger: ["test%d"]
---

This is test command %d.
`
		commandFile := filepath.Join(tempDir, "test%d.md")
		if err := os.WriteFile(commandFile, []byte(commandContent), 0644); err != nil {
			b.Fatalf("Failed to write test command file: %v", err)
		}
	}

	// Create registry with caching
	scanner := NewScanner(tempDir)
	registry := NewRegistryWithScanner(nil, scanner)

	// Prime the cache
	_ = registry.GetCustomCommands()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// This should use the cache after the first call
		_ = registry.GetCustomCommands()
	}
}

func BenchmarkLazyLoadingFirstAccess(b *testing.B) {
	// Create a temporary directory with multiple command files
	tempDir, err := os.MkdirTemp("", "customcommands_bench")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create multiple test command files
	for i := 0; i < 10; i++ {
		commandContent := `---
name: test-command-%d
description: A test command %d
trigger: ["test%d"]
---

This is test command %d.
`
		commandFile := filepath.Join(tempDir, "test%d.md")
		if err := os.WriteFile(commandFile, []byte(commandContent), 0644); err != nil {
			b.Fatalf("Failed to write test command file: %v", err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create new registry each time to test first access
		scanner := NewScanner(tempDir)
		registry := NewRegistryWithScanner(nil, scanner)
		_ = registry.GetCustomCommands()
	}
}
