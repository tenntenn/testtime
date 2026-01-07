package main

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
)

// Overlay represents the overlay JSON structure
type Overlay struct {
	Replace map[string]string `json:"Replace"`
}

func main() {
	if err := run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		if e, ok := err.(*exec.ExitError); ok {
			os.Exit(e.ExitCode())
		}
		log.Fatal(err)
	}
}

// loadOverlay executes testtime command and loads the overlay JSON
func loadOverlay() (*Overlay, error) {
	out, err := exec.Command("testtime").Output()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(strings.TrimSpace(string(out)))
	if err != nil {
		return nil, err
	}

	var overlay Overlay
	if err := json.Unmarshal(data, &overlay); err != nil {
		return nil, err
	}

	return &overlay, nil
}

// modifyVersionOutput appends a hash of overlay to the tool's version output
// This ensures build cache isolation for different overlay configurations
func modifyVersionOutput(w io.Writer, tool string, args []string, overlay *Overlay) error {
	out, err := exec.Command(tool, args...).Output()
	if err != nil {
		return err
	}

	h := fnv.New64()
	json.NewEncoder(h).Encode(overlay.Replace)
	fmt.Fprintf(w, "%s testtime:%x\n", strings.TrimSpace(string(out)), h.Sum64())
	return nil
}

// rewriteCompileArgs rewrites source file paths in compile command arguments
// according to the overlay Replace map
func rewriteCompileArgs(args []string, overlay *Overlay) []string {
	newArgs := make([]string, len(args))
	copy(newArgs, args)

	for i, arg := range newArgs {
		if abs, err := filepath.Abs(arg); err == nil {
			if to, ok := overlay.Replace[abs]; ok {
				newArgs[i] = to
			}
		}
	}
	return newArgs
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: testtimeexec <tool> [args...]")
	}

	overlay, err := loadOverlay()
	if err != nil {
		return err
	}

	tool := args[0]
	toolArgs := args[1:]

	// Modify version output to isolate build cache
	if slices.Contains(toolArgs, "-V=full") {
		return modifyVersionOutput(stdout, tool, toolArgs, overlay)
	}

	// Rewrite source file paths for compile command
	if filepath.Base(tool) == "compile" {
		toolArgs = rewriteCompileArgs(toolArgs, overlay)
	}

	cmd := exec.Command(tool, toolArgs...)
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return cmd.Run()
}
