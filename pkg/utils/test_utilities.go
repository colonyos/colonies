package utils

import (
	"testing"
)

func CheckError(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

func Fatal(t *testing.T, message string) {
	t.Fatal(message)
}
