package devtest

import (
	"bytes"
	"flag"
	"io"
	"os"
	"path/filepath"

	"github.com/puma/puma-dev/homedir"
)

// StubFlagArgs overrides command arguments to pretend as if puma-dev was executed at the commandline.
// ex: StubArgFlags([]string{"-n", "myapp", "path/to/app"}) ->
//   $ puma-dev -n myapp path/to/app
func StubFlagArgs(args []string) {
	os.Args = append([]string{"puma-dev"}, args...)
	flag.Parse()
}

// WithStdoutCaptured executes the passed function and returns a string containing the stdout of the executed function.
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

// WithWorkingDirectory executes the passed function within the context of
// the passed working directory path.
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

// RemoveApp deletes a symlink at ~/.puma-dev/{name} or panics.
func RemoveApp(name string) {
	fDir := "~/.puma-dev"
	path, err := homedir.Expand(filepath.Join(fDir, name))
	if err != nil {
		panic(err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return
	}

	if err := os.Remove(path); err != nil {
		panic(err)
	}
}
