package user

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	exitCode := func() int {
		return m.Run()
	}()
	os.Exit(exitCode)
}
