package main

import (
	"bytes"
	"flag"
	"fmt"
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

func WithWorkingDirectory(path string, mkdir bool, f func()) {
	// deleteDirectoryAfterwards := false
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if mkdir == true {
			os.Mkdir(path, 0755)
		} else {
			panic(err)
		}
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

func TestCommand_noCommandArg(t *testing.T) {
	StubFlagArgs(nil)
	err := command()
	assert.Equal(t, "Unknown command: \n", err.Error())
}

func TestCommand_badCommandArg(t *testing.T) {
	StubFlagArgs([]string{"doesnotexist"})
	err := command()
	assert.Equal(t, "Unknown command: doesnotexist\n", err.Error())
}

func TestCommand_link_noArgs(t *testing.T) {
	StubFlagArgs([]string{"link"})

	WithWorkingDirectory("/tmp/puma-dev-example-command-link-noargs", true, func() {
		actual := WithStdoutCaptured(func() {
			command()
		})

		assert.Equal(t, "+ App 'puma-dev-example-command-link-noargs' created, linked to '/private/tmp/puma-dev-example-command-link-noargs'\n", actual)
	})

	RemovePumaDevSymlink("puma-dev-example-command-link-noargs")
}

func TestCommand_link_withNameOverride(t *testing.T) {
	tmpCwd := "/tmp/puma-dev-example-command-link-noargs"

	StubFlagArgs([]string{"link", "-n", "anothername", tmpCwd})

	WithWorkingDirectory(tmpCwd, true, func() {
		actual := WithStdoutCaptured(func() {
			command()
		})

		assert.Equal(t, "+ App 'anothername' created, linked to '/tmp/puma-dev-example-command-link-noargs'\n", actual)
	})

	RemovePumaDevSymlink("anothername")
}

func TestCommand_link_invalidDirectory(t *testing.T) {
	StubFlagArgs([]string{"link", "/this/path/does/not/exist"})

	err := command()

	assert.Equal(t, "Invalid directory: /this/path/does/not/exist", err.Error())
}

func TestCommand_link_reassignExistingApp(t *testing.T) {
	appDir1 := "/tmp/puma-dev-test-command-link-reassign-existing-app-one"
	appDir2 := "/tmp/puma-dev-test-command-link-reassign-existing-app-two"

	StubFlagArgs([]string{"link", "-n", "existing-app", appDir1})
	os.Mkdir(appDir1, 0755)
	actual1 := WithStdoutCaptured(func() {
		if err := command(); err != nil {
			assert.Fail(t, err.Error())
		}
	})
	expected1 := fmt.Sprintf("+ App 'existing-app' created, linked to '%s'\n", appDir1)
	assert.Equal(t, expected1, actual1)

	StubFlagArgs([]string{"link", "-n", "existing-app", appDir2})
	os.Mkdir(appDir2, 0755)
	actual2 := WithStdoutCaptured(func() {
		if err := command(); err != nil {
			assert.Fail(t, err.Error())
		}
	})
	expected2 := fmt.Sprintf("! App 'existing-app' already exists, pointed at '%s'\n", appDir1)
	assert.Equal(t, expected2, actual2)

	RemovePumaDevSymlink("existing-app")
}
