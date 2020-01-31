package main

import (
	"fmt"
	"os"
	"testing"

	homedir "github.com/mitchellh/go-homedir"
	. "github.com/puma/puma-dev/dev/devtest"

	"github.com/stretchr/testify/assert"
)

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

	appDir, _ := homedir.Expand("~/my-test-puma-dev-application")

	WithWorkingDirectory(appDir, true, func() {
		actual := WithStdoutCaptured(func() {
			command()
		})

		expected := fmt.Sprintf("+ App 'my-test-puma-dev-application' created, linked to '%s'\n", appDir)
		assert.Equal(t, expected, actual)
	})

	RemoveApp("my-test-puma-dev-application")
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

	RemoveApp("anothername")
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

	RemoveApp("existing-app")

	if err1 := os.RemoveAll(appDir1); err1 != nil {
		assert.Fail(t, err1.Error())
	}

	if err2 := os.RemoveAll(appDir2); err2 != nil {
		assert.Fail(t, err2.Error())
	}
}
