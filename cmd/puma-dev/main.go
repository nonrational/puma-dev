package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
)

var (
	Continue       = CommandResult{-1, false}
	EarlyExitClean = CommandResult{0, true}
	EarlyExitError = CommandResult{1, true}
	Version        = "devel"

	fVersion = flag.Bool("V", false, "display version info")
)

// func (msg string, vars ...interface{}) {

// }

// func Cerrf(msg string, vars ...interface{}) {

// }

type CommandResult struct {
	exitStatusCode int
	shouldExit     bool
}

func allCheck() {
	if result := execWithExitStatus(); result.shouldExit {
		os.Exit(result.exitStatusCode)
	}
}

func execWithExitStatus() CommandResult {
	if *fVersion {
		fmt.Printf("Version: %s (%s)\n", Version, runtime.Version())
		return EarlyExitClean
	}

	if flag.NArg() > 0 {
		err := command()

		if err != nil {
			fmt.Printf("Error: %s\n", err)
			return EarlyExitError
		}

		return EarlyExitClean
	}

	return Continue
}

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()

		fmt.Fprintf(os.Stderr, "\nAvailable subcommands: link\n")
	}
}
