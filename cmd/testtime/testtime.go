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
	flagDir    string
	flagUpdate bool
)

func init() {
	flag.StringVar(&flagDir, "dir", defaultDir(), "working directory for testtime")
	flag.BoolVar(&flagUpdate, "u", false, "update exsiting files")
}

func main() {
	flag.Parse()
	overlay, err := createOverlay(flagUpdate, flagDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
	fmt.Println(overlay)
}

func defaultDir() string {
	if envGOPATH := os.Getenv("GOPATH"); envGOPATH != "" {
		gopath := strings.Split(envGOPATH, string(os.PathListSeparator))
		return filepath.Join(gopath[0], "pkg", "testtime")
	}
	return filepath.Join(build.Default.GOPATH, "pkg", "testtime")
}
