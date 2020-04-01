package main

import (
	"flag"
	"log"
	"os"
	"runtime"
)

var (
	EarlyExitClean = CommandResult{0, true}
	EarlyExitError = CommandResult{1, true}
	Continue       = CommandResult{-1, false}

	fVersion = flag.Bool("V", false, "display version info")
	Version  = "devel"
	StdLog   = log.New(os.Stdout, "", 1)
	ErrLog   = log.New(os.Stderr, "", 1)
)

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
		StdLog.Printf("Version: %s (%s)\n", Version, runtime.Version())
		return EarlyExitClean
	}

	if flag.NArg() > 0 {
		err := command()

		if err != nil {
			StdLog.Printf("Error: %s\n", err)
			return EarlyExitError
		}

		return EarlyExitClean
	}

	return Continue
}

func init() {
	StdLog.SetFlags(0)
	ErrLog.SetFlags(0)

	flag.Usage = func() {
		ErrLog.Printf("Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()

		ErrLog.Printf("\nAvailable subcommands: link\n")
	}
}
