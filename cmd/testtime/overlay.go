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
	"path/filepath"
	"runtime"

	"golang.org/x/tools/go/ast/astutil"
)

//go:embed _partials/testtime.go
var testtime string

func createOverlay(update bool, dir string) (string, error) {

	ver := build.Default.ReleaseTags[len(build.Default.ReleaseTags)-1]

	overlay := filepath.Join(dir, fmt.Sprintf("overlay_%s.json", ver))
	_, err := os.Stat(overlay)
	switch {
	case err == nil:
		if !update {
			return overlay, nil
		}
	case !errors.Is(err, os.ErrNotExist):
		return "", err
	}

	if err := os.MkdirAll(dir, 0o700); err != nil {
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

	new := filepath.Join(dir, fmt.Sprintf("time_%s.go", ver))
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
