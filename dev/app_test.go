package dev_test

import (
	"testing"
	"time"
	"github.com/stretchr/testify/mock"
	. "github.com/puma/puma-dev/dev"
)

type MockAppPool struct{
  mock.Mock
}

type MockCommand struct{
	mock.Mock
}

type MockEvents struct {
	mock.Mock
}

func TestAppKill(t *testing.T) {
	stdout, err := cmd.StdoutPipe()

	if err != nil {
		t.Fatal(err)
	}

	mockPool := new(MockAppPool)
	mockCmd := new(MockCommand)
	mockEvents := new(MockEvents)

	mockPool.On("Events").Return(mockEvents, nil)

	app := &App{
		Name:      "app",
		Command:   mockCmd,
		Events:    mockPool.Events,
		stdout:    stdout,
		dir:       "/",
		pool:      mockPool,
		readyChan: make(chan struct{}),
		lastUse:   time.Now().Add(time.Duration(-60) * time.Minute),
	}

	t.Logf(app.Name)

	app.Kill("just because")
}
