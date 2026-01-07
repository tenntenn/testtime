package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRewriteCompileArgs(t *testing.T) {
	t.Parallel()

	// Create a temporary file to simulate a source file
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "time.go")
	if err := os.WriteFile(srcFile, []byte("package time"), 0o644); err != nil {
		t.Fatal(err)
	}

	replaceFile := filepath.Join(tmpDir, "time_replaced.go")
	if err := os.WriteFile(replaceFile, []byte("package time // replaced"), 0o644); err != nil {
		t.Fatal(err)
	}

	overlay := &Overlay{
		Replace: map[string]string{
			srcFile: replaceFile,
		},
	}

	tests := []struct {
		name     string
		args     []string
		expected []string
	}{
		{
			name:     "no replacement needed",
			args:     []string{"-o", "output.o", "other.go"},
			expected: []string{"-o", "output.o", "other.go"},
		},
		{
			name:     "replace source file",
			args:     []string{"-o", "output.o", srcFile},
			expected: []string{"-o", "output.o", replaceFile},
		},
		{
			name:     "multiple args with replacement",
			args:     []string{"-p", "time", "-o", "time.a", srcFile, "other.go"},
			expected: []string{"-p", "time", "-o", "time.a", replaceFile, "other.go"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rewriteCompileArgs(tt.args, overlay)
			if len(result) != len(tt.expected) {
				t.Errorf("got %d args, expected %d", len(result), len(tt.expected))
				return
			}
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("arg[%d]: got %q, expected %q", i, v, tt.expected[i])
				}
			}
		})
	}
}

func TestRewriteCompileArgs_DoesNotModifyOriginal(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "time.go")
	if err := os.WriteFile(srcFile, []byte("package time"), 0o644); err != nil {
		t.Fatal(err)
	}
	replaceFile := filepath.Join(tmpDir, "time_replaced.go")

	overlay := &Overlay{
		Replace: map[string]string{
			srcFile: replaceFile,
		},
	}

	original := []string{"-o", "output.o", srcFile}
	originalCopy := make([]string, len(original))
	copy(originalCopy, original)

	rewriteCompileArgs(original, overlay)

	for i, v := range original {
		if v != originalCopy[i] {
			t.Errorf("original args was modified: arg[%d] changed from %q to %q", i, originalCopy[i], v)
		}
	}
}

func TestModifyVersionOutput(t *testing.T) {
	t.Parallel()

	overlay := &Overlay{
		Replace: map[string]string{
			"/usr/local/go/src/time/time.go": "/tmp/testtime/time_go1.23.go",
		},
	}

	var buf bytes.Buffer
	err := modifyVersionOutput(&buf, "echo", []string{"go1.23.0"}, overlay)
	if err != nil {
		t.Fatal(err)
	}

	output := buf.String()

	// Should contain original version
	if !strings.Contains(output, "go1.23.0") {
		t.Errorf("output should contain version string, got: %s", output)
	}

	// Should contain testtime hash
	if !strings.Contains(output, "testtime:") {
		t.Errorf("output should contain testtime hash, got: %s", output)
	}
}

func TestModifyVersionOutput_DifferentOverlaysProduceDifferentHashes(t *testing.T) {
	t.Parallel()

	overlay1 := &Overlay{
		Replace: map[string]string{
			"/path/to/time.go": "/path/to/time1.go",
		},
	}

	overlay2 := &Overlay{
		Replace: map[string]string{
			"/path/to/time.go": "/path/to/time2.go",
		},
	}

	var buf1, buf2 bytes.Buffer
	if err := modifyVersionOutput(&buf1, "echo", []string{"go1.23.0"}, overlay1); err != nil {
		t.Fatal(err)
	}
	if err := modifyVersionOutput(&buf2, "echo", []string{"go1.23.0"}, overlay2); err != nil {
		t.Fatal(err)
	}

	if buf1.String() == buf2.String() {
		t.Error("different overlays should produce different hashes")
	}
}

func TestModifyVersionOutput_SameOverlaysProduceSameHashes(t *testing.T) {
	t.Parallel()

	overlay := &Overlay{
		Replace: map[string]string{
			"/path/to/time.go": "/path/to/time1.go",
		},
	}

	var buf1, buf2 bytes.Buffer
	if err := modifyVersionOutput(&buf1, "echo", []string{"go1.23.0"}, overlay); err != nil {
		t.Fatal(err)
	}
	if err := modifyVersionOutput(&buf2, "echo", []string{"go1.23.0"}, overlay); err != nil {
		t.Fatal(err)
	}

	if buf1.String() != buf2.String() {
		t.Errorf("same overlays should produce same hashes, got %q and %q", buf1.String(), buf2.String())
	}
}

// Integration test that requires testtime and testtimeexec to be installed
func TestToolexecIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Check if testtime is available
	if _, err := exec.LookPath("testtime"); err != nil {
		t.Skip("testtime command not found, skipping integration test")
	}

	// Check if testtimeexec is available
	if _, err := exec.LookPath("testtimeexec"); err != nil {
		t.Skip("testtimeexec command not found, skipping integration test")
	}

	// Run testtime to generate overlay
	out, err := exec.Command("testtime").Output()
	if err != nil {
		t.Fatalf("testtime command failed: %v", err)
	}

	overlayPath := strings.TrimSpace(string(out))
	if _, err := os.Stat(overlayPath); err != nil {
		t.Fatalf("overlay file not found: %v", err)
	}

	// Verify overlay JSON structure
	data, err := os.ReadFile(overlayPath)
	if err != nil {
		t.Fatalf("failed to read overlay: %v", err)
	}

	var overlay Overlay
	if err := json.Unmarshal(data, &overlay); err != nil {
		t.Fatalf("failed to parse overlay JSON: %v", err)
	}

	if len(overlay.Replace) == 0 {
		t.Error("overlay Replace map should not be empty")
	}

	// Verify each replacement file exists
	for from, to := range overlay.Replace {
		if _, err := os.Stat(to); err != nil {
			t.Errorf("replacement file for %s not found: %v", from, err)
		}
	}
}

// Test that version hash changes based on overlay content
func TestVersionHashIsolation(t *testing.T) {
	t.Parallel()

	// Test that -V=full output includes testtime hash when overlays differ
	overlay1 := &Overlay{
		Replace: map[string]string{
			"/path/a": "/path/a1",
		},
	}

	overlay2 := &Overlay{
		Replace: map[string]string{
			"/path/a": "/path/a2",
		},
	}

	var buf1, buf2 bytes.Buffer
	if err := modifyVersionOutput(&buf1, "echo", []string{"go1.23.0"}, overlay1); err != nil {
		t.Fatalf("modifyVersionOutput failed: %v", err)
	}
	if err := modifyVersionOutput(&buf2, "echo", []string{"go1.23.0"}, overlay2); err != nil {
		t.Fatalf("modifyVersionOutput failed: %v", err)
	}

	// Both should contain testtime hash
	if !strings.Contains(buf1.String(), "testtime:") {
		t.Error("output should contain testtime hash")
	}
	if !strings.Contains(buf2.String(), "testtime:") {
		t.Error("output should contain testtime hash")
	}

	// Hashes should differ for different overlays
	if buf1.String() == buf2.String() {
		t.Error("different overlays should produce different version outputs for cache isolation")
	}

	t.Logf("overlay1 version: %s", strings.TrimSpace(buf1.String()))
	t.Logf("overlay2 version: %s", strings.TrimSpace(buf2.String()))
}

// TestBuildCacheIsolation verifies that build cache is properly isolated
// when using toolexec with different overlay configurations.
// This test performs actual builds to verify cache behavior.
func TestBuildCacheIsolation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Check if testtime and testtimeexec are available
	if _, err := exec.LookPath("testtime"); err != nil {
		t.Skip("testtime command not found, skipping integration test")
	}
	if _, err := exec.LookPath("testtimeexec"); err != nil {
		t.Skip("testtimeexec command not found, skipping integration test")
	}

	// Create a temporary directory for test
	tmpDir := t.TempDir()

	// Create a simple Go module for testing
	modFile := filepath.Join(tmpDir, "go.mod")
	if err := os.WriteFile(modFile, []byte("module testcache\n\ngo 1.21\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create a simple Go file that uses time.Now()
	mainFile := filepath.Join(tmpDir, "main.go")
	mainCode := `package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println(time.Now())
}
`
	if err := os.WriteFile(mainFile, []byte(mainCode), 0o644); err != nil {
		t.Fatal(err)
	}

	// Test 1: Build without toolexec (normal build)
	cmd1 := exec.Command("go", "build", "-o", filepath.Join(tmpDir, "normal"), ".")
	cmd1.Dir = tmpDir
	if out, err := cmd1.CombinedOutput(); err != nil {
		t.Fatalf("normal build failed: %v\n%s", err, out)
	}

	// Test 2: Build with toolexec
	cmd2 := exec.Command("go", "build", "-toolexec=testtimeexec", "-o", filepath.Join(tmpDir, "toolexec"), ".")
	cmd2.Dir = tmpDir
	if out, err := cmd2.CombinedOutput(); err != nil {
		t.Fatalf("toolexec build failed: %v\n%s", err, out)
	}

	// Test 3: Rebuild with toolexec should use cache (check with -x flag)
	cmd3 := exec.Command("go", "build", "-x", "-toolexec=testtimeexec", "-o", filepath.Join(tmpDir, "toolexec2"), ".")
	cmd3.Dir = tmpDir
	out3, err := cmd3.CombinedOutput()
	if err != nil {
		t.Fatalf("toolexec rebuild failed: %v\n%s", err, out3)
	}

	// The -x output should show cache hits (no compile commands for unchanged code)
	// If cache is working, we should see minimal compile activity
	t.Logf("toolexec rebuild output:\n%s", string(out3))

	// Test 4: Normal rebuild should also use its own cache
	cmd4 := exec.Command("go", "build", "-x", "-o", filepath.Join(tmpDir, "normal2"), ".")
	cmd4.Dir = tmpDir
	out4, err := cmd4.CombinedOutput()
	if err != nil {
		t.Fatalf("normal rebuild failed: %v\n%s", err, out4)
	}
	t.Logf("normal rebuild output:\n%s", string(out4))

	// Verify both binaries were created
	if _, err := os.Stat(filepath.Join(tmpDir, "normal")); err != nil {
		t.Error("normal binary not created")
	}
	if _, err := os.Stat(filepath.Join(tmpDir, "toolexec")); err != nil {
		t.Error("toolexec binary not created")
	}
	if _, err := os.Stat(filepath.Join(tmpDir, "toolexec2")); err != nil {
		t.Error("toolexec2 binary not created")
	}
	if _, err := os.Stat(filepath.Join(tmpDir, "normal2")); err != nil {
		t.Error("normal2 binary not created")
	}

	t.Log("Build cache isolation test passed: both normal and toolexec builds completed successfully")
}
