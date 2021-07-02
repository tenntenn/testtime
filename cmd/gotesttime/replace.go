package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"os"
	"path/filepath"

	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/ast/inspector"
	"golang.org/x/tools/go/packages"
)

//go:embed _partials/testtime.go
var testtime string

func replace() (string, func() error, error) {

	cfg := &packages.Config{Mode: packages.NeedFiles | packages.NeedSyntax | packages.NeedTypes}
	pkgs, err := packages.Load(cfg, "time")
	if err != nil {
		return "", nil, err
	}

	var timepkg *packages.Package
	for _, p := range pkgs {
		if p.ID == "time" {
			timepkg = p
			break
		}
	}
	if timepkg == nil {
		return "", nil, errors.New("cannot find time pacakge")
	}

	var timego struct {
		Path   string
		Syntax *ast.File
	}
	inspect := inspector.New(timepkg.Syntax)
	inspect.Preorder([]ast.Node{(*ast.FuncDecl)(nil)}, func(n ast.Node) {
		decl, _ := n.(*ast.FuncDecl)
		if decl == nil || decl.Name.Name != "Now" {
			return
		}
		f := timepkg.Fset.File(decl.Pos())
		if f == nil {
			return
		}
		// replacce Now => _Now
		decl.Name.Name = "_Now"
		timego.Path = f.Name()
		timego.Syntax = file(timepkg.Syntax, decl.Pos())
	})

	if timego.Path == "" || timego.Syntax == nil {
		return "", nil, errors.New("cannot find time.go")
	}

	astutil.AddImport(timepkg.Fset, timego.Syntax, "sync")
	astutil.AddImport(timepkg.Fset, timego.Syntax, "runtime")

	var buf bytes.Buffer
	if err := format.Node(&buf, timepkg.Fset, timego.Syntax); err != nil {
		return "", nil, err
	}

	fmt.Fprintln(&buf, testtime)

	src, err := format.Source(buf.Bytes())
	if err != nil {
		return "", nil, err
	}

	dir, err := os.MkdirTemp("", "gotesttime*")
	if err != nil {
		return "", nil, err
	}

	new := filepath.Join(dir, "time.go")
	if err := os.WriteFile(new, src, 0o600); err != nil {
		return "", nil, err
	}

	overlay := filepath.Join(dir, "overlay.json")
	v := struct {
		Replace map[string]string
	}{map[string]string{timego.Path: new}}
	jsonBytes, err := json.Marshal(v)
	if err != nil {
		return "", nil, err
	}
	if err := os.WriteFile(overlay, jsonBytes, 0o600); err != nil {
		return "", nil, err
	}

	cleanup := func() error {
		return os.RemoveAll(dir)
	}

	return overlay, cleanup, nil
}

func file(files []*ast.File, pos token.Pos) *ast.File {
	for _, f := range files {
		if f.Pos() <= pos && pos <= f.End() {
			return f
		}
	}
	return nil
}
