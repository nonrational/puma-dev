package main

import (
	"bytes"
	"flag"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/puma/puma-dev/homedir"
	"github.com/stretchr/testify/assert"
)

func StubFlagArgs(args []string) {
	os.Args = append([]string{"puma-dev"}, args...)
	flag.Parse()
}

func WithStdoutCaptured(f func()) string {
	osStdout := os.Stdout
	r, w, err := os.Pipe()

	if err != nil {
		panic(err)
	}

	os.Stdout = w

	outC := make(chan string)

	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	f()

	w.Close()
	os.Stdout = osStdout
	out := <-outC

	return out
}

func WithWorkingDirectory(path string, f func()) {
	// deleteDirectoryAfterwards := false
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, 0755)
	}

	originalPath, _ := os.Getwd()

	os.Chdir(path)
	f()
	os.Chdir(originalPath)
}

func RemovePumaDevSymlink(name string) {
	path, err := homedir.Expand(filepath.Join(*fDir, name))

	if err != nil {
		panic(err)
	}

	if err := os.Remove(path); err != nil {
		panic(err)
	}
}

/**
func RemovePumaDevCommandTestLinks() error {
	splatSymlinks, err := homedir.Expand(filepath.Join(*fDir, "puma-dev-command-test*"))

	if err != nil {
		return err
	}

	files, err := filepath.Glob(splatSymlinks)

	if err != nil {
		return err
	}

	for _, f := range files {
		if err := os.Remove(f); err != nil {
			panic(err)
		}
	}

	return nil
}
*/

func TestCommand_noargs(t *testing.T) {
	StubFlagArgs(nil)

	err := command()

	if err == nil || err.Error() != "Unknown command: \n" {
		t.Error("unexpected error message", err.Error())
	}
}

func TestCommand_badargs(t *testing.T) {
	StubFlagArgs([]string{"doesnotexist"})

	err := command()

	if err == nil || err.Error() != "Unknown command: doesnotexist\n" {
		t.Error("unexpected error message -> ", err.Error())
	}
}

func TestCommand_link_noargs(t *testing.T) {
	StubFlagArgs([]string{"link"})

	WithWorkingDirectory("/tmp/puma-dev-example-command-link-noargs", func() {
		actual := WithStdoutCaptured(func() {
			command()
		})

		assert.Equal(t, "+ App 'puma-dev-example-command-link-noargs' created, linked to '/private/tmp/puma-dev-example-command-link-noargs'\n", actual)
	})

	RemovePumaDevSymlink("puma-dev-example-command-link-noargs")
}

func TestCommand_link_namedargs(t *testing.T) {
	tmpCwd := "/tmp/puma-dev-example-command-link-noargs"

	StubFlagArgs([]string{"link", "-n", "anothername", tmpCwd})

	WithWorkingDirectory(tmpCwd, func() {
		actual := WithStdoutCaptured(func() {
			command()
		})

		assert.Equal(t, "+ App 'anothername' created, linked to '/tmp/puma-dev-example-command-link-noargs'\n", actual)
	})

	RemovePumaDevSymlink("anothername")
}
