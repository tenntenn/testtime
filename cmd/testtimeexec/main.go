package main

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
)

func main() {
	out, err := exec.Command("testtime").Output()
	if err != nil {
		log.Fatal(err)
	}

	data, err := os.ReadFile(strings.TrimSpace(string(out)))
	if err != nil {
		log.Fatal(err)
	}

	var overlay struct {
		Replace map[string]string `json:"Replace"`
	}
	if err := json.Unmarshal(data, &overlay); err != nil {
		log.Fatal(err)
	}

	tool, args := os.Args[1], os.Args[2:]

	// Modify version output to isolate build cache
	if slices.Contains(args, "-V=full") {
		out, _ := exec.Command(tool, args...).Output()
		h := fnv.New64()
		json.NewEncoder(h).Encode(overlay.Replace)
		fmt.Printf("%s testtime:%x\n", strings.TrimSpace(string(out)), h.Sum64())
		return
	}

	// Rewrite source file paths for compile command
	if filepath.Base(tool) == "compile" {
		for i, arg := range args {
			if abs, err := filepath.Abs(arg); err == nil {
				if to, ok := overlay.Replace[abs]; ok {
					args[i] = to
				}
			}
		}
	}

	cmd := exec.Command(tool, args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		if e, ok := err.(*exec.ExitError); ok {
			os.Exit(e.ExitCode())
		}
		os.Exit(1)
	}
}
