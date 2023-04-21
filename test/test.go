package test

import (
	_ "dalang/setup"

	"os"

	"testing"
)

func setup() {
}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	os.Exit(code)
}
