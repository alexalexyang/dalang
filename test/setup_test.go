package test

// This just runs the init function in setup/setup.go

import (
	_ "dalang/setup"

	"os"

	"testing"
)

func TestMain(m *testing.M) {
	code := m.Run()
	os.Exit(code)
}
