// path: cmd/gate_strict_test.go
package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// Helper-process pattern to observe os.Exit(2) without killing 'go test'.
func TestGate_Strict_Destructive_ExitCode2(t *testing.T) {
	if os.Getenv("GATE_HELPER") == "1" {
		// child process
		input := filepath.Join("..", "internal", "plan", "testdata", "plan_drift.json")
		// Silence stdout/stderr for the helper run
		devnull, _ := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
		defer devnull.Close()
		stdout := os.Stdout
		stderr := os.Stderr
		os.Stdout = devnull
		os.Stderr = devnull
		defer func() {
			os.Stdout = stdout
			os.Stderr = stderr
		}()

		gateInputPath = input
		gateFormat = "json"
		gateStrict = true
		gateMaxDeletes = -1
		gateMaxReplaces = -1
		gateList = true

		_ = runGate(nil, nil)
		os.Exit(1)
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestGate_Strict_Destructive_ExitCode2")
	cmd.Env = append(os.Environ(), "GATE_HELPER=1")
	err := cmd.Run()
	if err == nil {
		t.Fatalf("expected non-zero error from child process")
	}
	if ee, ok := err.(*exec.ExitError); ok {
		if code := ee.ExitCode(); code != 2 {
			t.Fatalf("expected exit code 2, got %d", code)
		}
	} else {
		t.Fatalf("unexpected error type: %T: %v", err, err)
	}
}

func TestGate_Safe_UpdatesOnly_ExitCode0(t *testing.T) {
	input := filepath.Join("..", "internal", "plan", "testdata", "plan_updates_only.json")

	gateInputPath = input
	gateFormat = "json"
	gateStrict = true
	gateMaxDeletes = -1
	gateMaxReplaces = -1
	gateList = true

	// Capture stdout via a pipe
	r, w, _ := os.Pipe()
	stdout := os.Stdout
	os.Stdout = w

	err := runGate(nil, nil)

	w.Close()
	os.Stdout = stdout
	if err != nil {
		t.Fatalf("runGate error: %v", err)
	}

	var buf bytes.Buffer
	if _, copyErr := io.Copy(&buf, r); copyErr != nil {
		t.Fatalf("io.Copy: %v", copyErr)
	}
	r.Close()

	out := buf.Bytes()
	var payload struct {
		DestructiveTotal int      `json:"destructive_total"`
		Destructive      []string `json:"destructive"`
	}
	if uerr := json.Unmarshal(out, &payload); uerr != nil {
		t.Fatalf("invalid JSON output: %v\n%s", uerr, string(out))
	}
	if payload.DestructiveTotal != 0 {
		t.Fatalf("expected destructive_total 0, got %d", payload.DestructiveTotal)
	}
	if len(payload.Destructive) != 0 {
		t.Fatalf("expected no destructive addresses, got %v", payload.Destructive)
	}
}

func TestGate_Thresholds_TripExit2(t *testing.T) {
	if os.Getenv("GATE_HELPER_THRESH") == "1" {
		// child process
		input := filepath.Join("..", "internal", "plan", "testdata", "plan_drift.json")
		devnull, _ := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
		defer devnull.Close()
		stdout := os.Stdout
		stderr := os.Stderr
		os.Stdout = devnull
		os.Stderr = devnull
		defer func() {
			os.Stdout = stdout
			os.Stderr = stderr
		}()

		gateInputPath = input
		gateFormat = "json"
		gateStrict = true
		// plan_drift.json has 1 delete and 1 replace -> both thresholds 0 will trip
		gateMaxDeletes = 0
		gateMaxReplaces = 0
		gateList = false

		_ = runGate(nil, nil)
		os.Exit(1)
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestGate_Thresholds_TripExit2")
	cmd.Env = append(os.Environ(), "GATE_HELPER_THRESH=1")
	if err := cmd.Run(); err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			if code := ee.ExitCode(); code != 2 {
				t.Fatalf("expected exit code 2, got %d", code)
			}
			return
		}
		t.Fatalf("unexpected error: %v", err)
	}
	t.Fatalf("expected non-zero exit due to os.Exit(2)")
}
