package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	. "github.com/puma/puma-dev/dev/devtest"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	EnsurePumaDevDirectory()
	os.Exit(m.Run())
}

func backgroundPumaDev(t *testing.T) func() {
	StubCommandLineArgs()
	testAppLinkDirPath := "~/.gotest-puma-dev"
	SetFlagOrFail(t, "dir", testAppLinkDirPath)
	SetFlagOrFail(t, "http-port", "32100")
	SetFlagOrFail(t, "d", "pumadevtld")

	info, err := os.Stat("/etc/resolver")

	if err != nil {
		assert.FailNow(t, err.Error())
	}

	assert.True(t, info.Mode().IsDir())
	fmt.Printf("/etc/resolver permissions: %#o\n", info.Mode().Perm())

	go func() {
		main()
	}()

	// REPLACE WITH SOCKET WAIT
	time.Sleep(1 * time.Second)

	return func() {
		RemoveDirectoryOrFail(t, testAppLinkDirPath)
	}
}

func TestMainPumaDev(t *testing.T) {
	defer backgroundPumaDev(t)()

	curlStatus := func() string {
		url := fmt.Sprintf("http://127.0.0.1:%d/status", *fHTTPPort)

		req, _ := http.NewRequest("GET", url, nil)
		req.Host = "puma-dev"

		resp, err := http.DefaultClient.Do(req)

		if err != nil {
			assert.FailNow(t, err.Error())
		}

		defer resp.Body.Close()

		bodyBytes, _ := ioutil.ReadAll(resp.Body)

		return strings.TrimSpace(string(bodyBytes))
	}

	assert.Equal(t, "{}", curlStatus())
}

func TestMain_execWithExitStatus_versionFlag(t *testing.T) {
	StubCommandLineArgs("-V")
	assert.True(t, *fVersion)

	execStdOut := WithStdoutCaptured(func() {
		result := execWithExitStatus()
		assert.Equal(t, 0, result.exitStatusCode)
		assert.Equal(t, true, result.shouldExit)
	})

	assert.Regexp(t, "^Version: devel \\(.+\\)\\n$", execStdOut)
}

func TestMain_execWithExitStatus_noFlag(t *testing.T) {
	StubCommandLineArgs()
	assert.False(t, *fVersion)

	execStdOut := WithStdoutCaptured(func() {
		result := execWithExitStatus()
		assert.Equal(t, false, result.shouldExit)
	})

	assert.Equal(t, "", execStdOut)
}

func TestMain_execWithExitStatus_commandArgs(t *testing.T) {
	StubCommandLineArgs("nosoupforyou")

	execStdOut := WithStdoutCaptured(func() {
		result := execWithExitStatus()
		assert.Equal(t, 1, result.exitStatusCode)
		assert.Equal(t, true, result.shouldExit)
	})

	assert.Equal(t, "Error: Unknown command: nosoupforyou\n\n", execStdOut)
}

func TestMain_allCheck_versionFlag(t *testing.T) {
	if os.Getenv("GO_TEST_SUBPROCESS") == "1" {
		StubCommandLineArgs("-V")
		allCheck()

		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestMain_allCheck_versionFlag")
	cmd.Env = append(os.Environ(), "GO_TEST_SUBPROCESS=1")
	err := cmd.Run()

	assert.Nil(t, err)
}

func TestMain_allCheck_badArg(t *testing.T) {
	if os.Getenv("GO_TEST_SUBPROCESS") == "1" {
		StubCommandLineArgs("-badarg")
		allCheck()

		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestMain_allCheck_badArg")
	cmd.Env = append(os.Environ(), "GO_TEST_SUBPROCESS=1")
	err := cmd.Run()

	exit, ok := err.(*exec.ExitError)

	assert.True(t, ok)
	assert.False(t, exit.Success())
}
