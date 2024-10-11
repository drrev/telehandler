package utils

import (
	"os"
	"strings"
	"testing"
)

func NoError(t *testing.T) func(e error) bool {
	t.Helper()
	return func(e error) bool {
		return e == nil
	}
}

func ErrorTextContains(t *testing.T, str string) func(e error) bool {
	t.Helper()
	return func(e error) bool {
		if e == nil {
			return len(str) == 0
		}
		return strings.Contains(e.Error(), str)
	}
}

func TempDir(t *testing.T) string {
	tmp, err := os.MkdirTemp(os.TempDir(), "test")
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { os.RemoveAll(tmp) })
	return tmp
}
