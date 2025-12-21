package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunCmd(t *testing.T) {
	testDataDir := filepath.Join("testdata", "env")
	echoScript := filepath.Join("testdata", "echo.sh")

	if _, err := os.Stat(testDataDir); os.IsNotExist(err) {
		t.Skipf("Test data directory not found: %s", testDataDir)
	}
	if _, err := os.Stat(echoScript); os.IsNotExist(err) {
		t.Skipf("Echo script not found: %s", echoScript)
	}

	env, err := ReadDir(testDataDir)
	if err != nil {
		t.Fatalf("ReadDir failed: %v", err)
	}

	env["ADDED"] = EnvValue{
		Value:      "from added",
		NeedRemove: false,
	}

	cmd := []string{echoScript, "first", "second", "third"}

	oldStdout := os.Stdout
	oldStderr := os.Stderr

	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w

	returnCode := RunCmd(cmd, env)

	w.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if returnCode != 0 {
		t.Errorf("Expected exit code 0, got %d", returnCode)
	}

	expectedLines := []string{
		"HELLO is (\"hello\")",
		"BAR is (bar)",
		"FOO is (   foo",
		"with new line)",
		"UNSET is ()",
		"ADDED is (from added)",
		"EMPTY is ()",
		"arguments are first second third",
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != len(expectedLines) {
		t.Errorf("Expected %d lines of output, got %d", len(expectedLines), len(lines))
		t.Logf("Output:\n%s", output)
	}

	for _, expectedLine := range expectedLines {
		found := false
		for _, line := range lines {
			if strings.Contains(line, expectedLine) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected line not found: %s", expectedLine)
			t.Logf("Full output:\n%s", output)
		}
	}
}
