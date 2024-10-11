package utils

import (
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
