// path: cmd/scan_strict_test.go
package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/ha36d/drift-checker/internal/drift"
	"github.com/ha36d/drift-checker/internal/plan"
	"github.com/spf13/cobra"
)

// This test uses the standard helper-process pattern to catch os.Exit(2) without killing the test runner.
func TestStrictExitCode_OnDrift(t *testing.T) {
	if os.Getenv("HELPER_PROCESS") == "1" {
		// ---- child process path ----
		// Stub runner selection and plan JSON so we avoid calling real binaries.
		drift.SelectRunner = func() (plan.RunnerKind, error) { return plan.RunnerTofu, nil }
		drift.PlanJSON = func(_ context.Context, _ plan.RunnerKind, _ string) ([]byte, error) {
			// IMPORTANT: when running tests in the cmd package, the CWD is ./cmd,
			// so we need to go up one directory to reach internal/plan/testdata.
			fixture := filepath.Join("..", "internal", "plan", "testdata", "plan_drift.json")
			b, err := os.ReadFile(fixture)
			if err != nil {
				fmt.Fprintln(os.Stderr, "fixture read error:", err)
				os.Exit(1)
			}
			return b, nil
		}

		// Force strict mode; format doesn't matter, but use json to keep stdout minimal.
		strictFlag = true
		formatFlag = "json"
		pathFlag = "."

		// Silence output in the helper by redirecting to /dev/null.
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

		_ = runScan(&cobra.Command{}, nil) // runScan will call os.Exit(2) on drift when --strict is set
		// If we ever get here, something went wrong; exit 1 so parent fails.
		os.Exit(1)
		return
	}

	// ---- parent process path ----
	cmd := exec.Command(os.Args[0], "-test.run=TestStrictExitCode_OnDrift")
	cmd.Env = append(os.Environ(), "HELPER_PROCESS=1")
	err := cmd.Run()
	if err == nil {
		t.Fatalf("expected non-zero exit due to os.Exit(2), got nil error")
	}
	if ee, ok := err.(*exec.ExitError); ok {
		if code := ee.ExitCode(); code != 2 {
			t.Fatalf("expected exit code 2 for drift in --strict, got %d", code)
		}
	} else {
		t.Fatalf("unexpected error type: %T: %v", err, err)
	}
}
