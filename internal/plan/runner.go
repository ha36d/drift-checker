package plan

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type RunnerKind string

const (
	RunnerTofu      RunnerKind = "tofu"
	RunnerTerraform RunnerKind = "terraform"
)

func SelectRunner() (RunnerKind, error) {
	if _, err := exec.LookPath("tofu"); err == nil {
		return RunnerTofu, nil
	}
	if _, err := exec.LookPath("terraform"); err == nil {
		return RunnerTerraform, nil
	}
	return "", errors.New("neither 'tofu' nor 'terraform' found in PATH")
}

// MakeRefreshOnlyPlanJSON always uses two-step method for reliability:
//   <runner> -chdir=<path> plan -refresh-only -out=drift-checker.plan
//   <runner> -chdir=<path> show -json drift-checker.plan
//
// This works across Terraform and OpenTofu versions without depending on streaming -json.
func MakeRefreshOnlyPlanJSON(ctx context.Context, runner RunnerKind, path string) ([]byte, error) {
	chdir := []string{"-chdir=" + path}

	planFile := "drift-checker.plan"
	planPath := filepath.Join(path, planFile)

	// 1) plan -refresh-only -out=...
	planCmd := exec.CommandContext(
		ctx,
		string(runner),
		append(chdir, "plan", "-refresh-only", "-out="+planFile)...,
	)
	planCmd.Env = os.Environ()

	var stderrPlan bytes.Buffer
	planCmd.Stderr = &stderrPlan
	if err := planCmd.Run(); err != nil {
		return nil, fmt.Errorf("%s plan failed: %v\n%s", runner, err, stderrPlan.String())
	}
	defer func() { _ = os.Remove(planPath) }()

	// 2) show -json <planFile>
	showCmd := exec.CommandContext(
		ctx,
		string(runner),
		append(chdir, "show", "-json", planFile)...,
	)
	showCmd.Env = os.Environ()

	var stdoutShow, stderrShow bytes.Buffer
	showCmd.Stdout = &stdoutShow
	showCmd.Stderr = &stderrShow
	if err := showCmd.Run(); err != nil {
		return nil, fmt.Errorf("%s show -json failed: %v\n%s", runner, err, stderrShow.String())
	}

	return stdoutShow.Bytes(), nil
}