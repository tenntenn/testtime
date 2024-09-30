package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/build"
	"go/format"
	"go/parser"
	"go/token"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"golang.org/x/tools/go/ast/astutil"
)

//go:embed _partials/testtime.go
var testtime string

func goVersion(modroot string) (string, error) {
	var stdout bytes.Buffer
	cmd := exec.Command("go", "env", "GOVERSION")
	cmd.Dir = modroot
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		return "", err
	}

	return strings.TrimSpace(stdout.String()), nil
}

func createOverlay(update bool, modroot, output string) (string, error) {

	ver, err := goVersion(modroot)
	if err != nil {
		return "", err
	}

	overlay := filepath.Join(output, fmt.Sprintf("overlay_%s.json", ver))
	_, err = os.Stat(overlay)
	switch {
	case err == nil:
		if !update {
			return overlay, nil
		}
	case !errors.Is(err, os.ErrNotExist):
		return "", err
	}

	if err := os.MkdirAll(output, 0o700); err != nil {
		return "", err
	}

	var buf bytes.Buffer
	old, err := replaceTimeNow(&buf)
	if err != nil {
		return "", err
	}

	fmt.Fprint(&buf, testtime)

	src, err := format.Source(buf.Bytes())
	if err != nil {
		return "", err
	}

	new := filepath.Join(output, fmt.Sprintf("time_%s.go", ver))
	if err := os.WriteFile(new, src, 0o600); err != nil {
		return "", err
	}

	v := struct {
		Replace map[string]string
	}{map[string]string{old: new}}
	jsonBytes, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(overlay, jsonBytes, 0o600); err != nil {
		return "", err
	}

	return overlay, nil
}

func replaceTimeNow(w io.Writer) (string, error) {
	srcDir := filepath.Join(runtime.GOROOT(), "src")
	pkg, err := build.Default.Import("time", srcDir, 0)
	if err != nil {
		return "", err
	}

	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, pkg.Dir, nil, parser.ParseComments)
	if err != nil {
		return "", err
	}

	if pkgs["time"] == nil {
		return "", errors.New("cannot find time package")
	}

	var (
		path   string
		syntax *ast.File
	)
LOOP:
	for name, file := range pkgs["time"].Files {
		for _, decl := range file.Decls {
			decl, _ := decl.(*ast.FuncDecl)
			if decl == nil {
				continue
			}

			if decl.Name.Name == "Now" {
				decl.Name.Name = "_Now"
				path = name
				syntax = file
				break LOOP
			}
		}
	}

	if path == "" || syntax == nil {
		return "", errors.New("cannot find time.Now")
	}

	astutil.AddImport(fset, syntax, "sync")
	astutil.AddImport(fset, syntax, "runtime")

	if err := format.Node(w, fset, syntax); err != nil {
		return "", err
	}

	return path, nil
}
