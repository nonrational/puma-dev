package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/puma/puma-dev/homedir"
)

func StubFlagArgs(args []string) {
	os.Args = append([]string{"puma-dev"}, args...)
	flag.Parse()
}

func StubCwd() string {
	var path = "/tmp/puma-dev-command-test-cwd"

	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, 0777)
	}

	os.Chdir(path)

	return path
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

// func RemovePumaDevCommandTestLinks() error {
// 	splatSymlinks, err := homedir.Expand(filepath.Join(*fDir, "puma-dev-command-test*"))
//
// 	if err != nil {
// 		return err
// 	}
//
// 	files, err := filepath.Glob(splatSymlinks)
//
// 	if err != nil {
// 		return err
// 	}
//
// 	for _, f := range files {
// 		if err := os.Remove(f); err != nil {
// 			panic(err)
// 		}
// 	}
//
// 	return nil
// }

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

func ExampleCommand_link_noargs() {
	StubFlagArgs([]string{"link"})
	cwd := StubCwd()

	command()

	RemovePumaDevSymlink(filepath.Base(cwd))

	// Output:
	// + App 'puma-dev-command-test-cwd' created, linked to '/private/tmp/puma-dev-command-test-cwd'
}

func ExampleCommand_link_namedargs() {
	cwd := StubCwd()
	StubFlagArgs([]string{"link", cwd, "-n", "anothername"})

	err := command()

	fmt.Println(err)

	// RemovePumaDevSymlink("anothername")

	// Output:
	// + App 'anothername' created, linked to '/private/tmp/puma-dev-command-test-cwd'
}
