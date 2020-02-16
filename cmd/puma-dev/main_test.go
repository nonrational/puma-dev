package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	dev "github.com/puma/puma-dev/dev"
	. "github.com/puma/puma-dev/dev/devtest"
	"github.com/puma/puma-dev/homedir"
	"github.com/stretchr/testify/assert"
)

var testAppLinkDirPath = homedir.MustExpand("~/.gotest-puma-dev")

func TestMain(m *testing.M) {
	EnsurePumaDevDirectory()
	os.Exit(m.Run())
}

func generateLivePumaDevCertIfNotExist(t *testing.T) {
	liveSupportPath := homedir.MustExpand(dev.SupportDir)
	liveCertPath := filepath.Join(liveSupportPath, "cert.pem")
	liveKeyPath := filepath.Join(liveSupportPath, "key.pem")

	if !FileExists(liveCertPath) || !FileExists(liveKeyPath) {
		MakeDirectoryOrFail(t, liveSupportPath)

		if err := dev.GeneratePumaDevCertificateAuthority(liveCertPath, liveKeyPath); err != nil {
			assert.FailNow(t, err.Error())
		}
	}
}

func backgroundPumaDev(t *testing.T) func() {
	StubCommandLineArgs()

	SetFlagOrFail(t, "dir", testAppLinkDirPath)

	generateLivePumaDevCertIfNotExist(t)

	go func() {
		main()
	}()

	// REPLACE WITH SOCKET WAIT
	time.Sleep(1 * time.Second)

	return func() {
		RemoveDirectoryOrFail(t, testAppLinkDirPath)
	}
}

func getUrlWithHost(t *testing.T, url string, host string) string {
	req, _ := http.NewRequest("GET", url, nil)
	req.Host = host

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		assert.FailNow(t, err.Error())
	}

	defer resp.Body.Close()

	bodyBytes, _ := ioutil.ReadAll(resp.Body)

	return strings.TrimSpace(string(bodyBytes))
}

func TestMainPumaDev(t *testing.T) {
	defer backgroundPumaDev(t)()

	projectRoot := Basepath
	fmt.Println(projectRoot)
	rackAppPath := filepath.Join(projectRoot, "etc", "rack-hi-puma")
	hipumaLinkPath := filepath.Join(homedir.MustExpand(testAppLinkDirPath), "hipuma")
	os.Symlink(rackAppPath, hipumaLinkPath)

	t.Run("status", func(t *testing.T) {
		statusUrl := fmt.Sprintf("http://localhost:%d/status", *fHTTPPort)
		statusHost := "puma-dev"

		assert.Equal(t, "{}", getUrlWithHost(t, statusUrl, statusHost))
	})

	t.Run("hipuma", func(t *testing.T) {
		statusUrl := fmt.Sprintf("http://localhost:%d/", *fHTTPPort)
		statusHost := "hipuma"

		assert.Equal(t, "Hi Puma!", getUrlWithHost(t, statusUrl, statusHost))
	})

	t.Run("unknown app", func(t *testing.T) {
		statusUrl := fmt.Sprintf("http://localhost:%d/", *fHTTPPort)
		statusHost := "doesnotexist"

		assert.Equal(t, "unknown app", getUrlWithHost(t, statusUrl, statusHost))
	})
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
