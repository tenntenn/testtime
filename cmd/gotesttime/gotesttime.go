package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

func main() {

	overlay, cleanup, err := replace()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}

	tags := strings.Join(append(parseTags(), "overlaytesttime"), ",")
	args := append([]string{"test", "-tags", tags, "-overlay", overlay}, os.Args[1:]...)
	cmd := exec.Command("go", args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	cmd.Run()

	if err := cleanup(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}

	os.Exit(cmd.ProcessState.ExitCode())
}

func parseTags() []string {
	flags := flag.NewFlagSet("", flag.ContinueOnError)
	flags.Usage = func() {}
	flags.SetOutput(io.Discard)
	tags := flags.String("tags", "", "tags")
	err := flags.Parse(os.Args[1:])
	if err != nil {
		return nil
	}
	return strings.Split(*tags, ",")
}
