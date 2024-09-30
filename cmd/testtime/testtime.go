package main

import (
	"flag"
	"fmt"
	"go/build"
	"os"
	"path/filepath"
	"strings"
)

var (
	flagCacheDir string
	flagUpdate   bool
)

func init() {
	flag.StringVar(&flagCacheDir, "dir", defaultCacheDir(), "cache directory for testtime")
	flag.BoolVar(&flagUpdate, "u", false, "update exsiting files")
}

func main() {
	flag.Parse()
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func run() error {

	modroot := flag.Arg(0)
	if modroot == "" {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		modroot = wd
	}

	overlay, err := createOverlay(flagUpdate, modroot, flagCacheDir)
	if err != nil {
		return err
	}
	fmt.Println(overlay)

	return nil
}

func defaultCacheDir() string {
	if envGOPATH := os.Getenv("GOPATH"); envGOPATH != "" {
		gopath := strings.Split(envGOPATH, string(os.PathListSeparator))
		return filepath.Join(gopath[0], "pkg", "testtime")
	}
	return filepath.Join(build.Default.GOPATH, "pkg", "testtime")
}
