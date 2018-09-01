package main

import (
	"strings"
	"testing"
)

func TestReadLines(t *testing.T) {
	testData := []string{
		"/etc",
		"/tmp",
		"/usr/local",
		"/usr/home",
		"1",
		"2",
		"3",
		"/mnt",
		"/usr/local/etc",
	}

	r := strings.NewReader(strings.Join(testData, "\n"))

	idx := 0
	for line := range readLines(r, 1) {
		t.Logf("%s", line)
		if line != testData[idx] {
			t.Fatalf("readLines() returned %s at index %d; expected %s", line, idx, testData[idx])
		}
		idx++
	}
}
