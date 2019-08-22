package dev

import (
	"os/exec"
	"testing"
	"time"
	// "github.com/stretchr/testify/mock"
	// . "github.com/puma/puma-dev/dev"
)

func TestAppKillAppPoolRemoval(t *testing.T) {
	cmd := exec.Command("bash", "-c", "sleep 600")
	err := cmd.Start()

	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%+v\n", cmd.Process)
	t.Log(cmd.Process.Pid)

	var events Events
	var pool AppPool

	pool.Dir = "/tmp/puma-dev"
	pool.IdleTime = 15 * 60 * time.Second
	pool.Events = &events

	app := &App{
		Name:      "app",
		Command:   cmd,
		Events:    pool.Events,
		readyChan: make(chan struct{}),
	}

	t.Logf("%+v\n", app)
	t.Logf("%+v\n", app.Events)

	app.Kill("just because")
}
