package huhx_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// e2e tests build the examples/deploy binary once and exercise it as a
// real subprocess. Stdin is closed (not a TTY) so the runner takes the
// non-interactive path even without explicit --non-interactive or CI=1
// where applicable, but the tests force CI=1 to make intent explicit.
//
// Run with `go test -run TestE2E ./...`. Skipped under `-short`.

var deployBin string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "huhx-e2e-*")
	if err != nil {
		panic(err)
	}
	deployBin = filepath.Join(dir, "deploy")
	build := exec.Command("go", "build", "-o", deployBin, "./examples/deploy")
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		_ = os.RemoveAll(dir)
		panic("failed to build deploy example: " + err.Error())
	}
	code := m.Run()
	_ = os.RemoveAll(dir)
	os.Exit(code)
}

// runDeploy executes the deploy binary with the given args and env
// additions. Stdin is /dev/null so isatty reports false in the
// non-interactive autodetect path.
func runDeploy(t *testing.T, env []string, args ...string) (stdout, stderr string, code int) {
	t.Helper()
	cmd := exec.Command(deployBin, args...)
	cmd.Env = append(os.Environ(), env...)
	cmd.Stdin = nil
	var out, errb bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errb
	err := cmd.Run()
	if exitErr, ok := err.(*exec.ExitError); ok {
		return out.String(), errb.String(), exitErr.ExitCode()
	}
	if err != nil {
		t.Fatalf("unexpected exec failure: %v", err)
	}
	return out.String(), errb.String(), 0
}

func TestE2E_DeployHappyPath_Flags(t *testing.T) {
	if testing.Short() || deployBin == "" {
		t.Skip("e2e requires built binary; skipped under -short")
	}
	stdout, stderr, code := runDeploy(t,
		[]string{"CI=1"},
		"--answer", "name=myapp",
		"--answer", "environment=prod",
		"--answer", "all-regions=true",
	)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d. stdout=%q stderr=%q", code, stdout, stderr)
	}
	want := `Deploying "myapp" to prod (all regions: true)`
	if !strings.Contains(stdout, want) {
		t.Errorf("expected stdout to contain %q, got %q", want, stdout)
	}
}

func TestE2E_DeployHappyPath_Env(t *testing.T) {
	if testing.Short() || deployBin == "" {
		t.Skip("e2e requires built binary; skipped under -short")
	}
	stdout, stderr, code := runDeploy(t,
		[]string{
			"CI=1",
			"DEPLOY_NAME=envapp",
			"DEPLOY_ENVIRONMENT=prod",
			"DEPLOY_ALL_REGIONS=true",
		},
	)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d. stdout=%q stderr=%q", code, stdout, stderr)
	}
	want := `Deploying "envapp" to prod (all regions: true)`
	if !strings.Contains(stdout, want) {
		t.Errorf("expected stdout to contain %q, got %q", want, stdout)
	}
}

func TestE2E_DeployHiddenGroupSkipped(t *testing.T) {
	if testing.Short() || deployBin == "" {
		t.Skip("e2e requires built binary; skipped under -short")
	}
	// environment=staging hides the all-regions group via WithHide,
	// so leaving --answer all-regions unset must NOT error.
	stdout, stderr, code := runDeploy(t,
		[]string{"CI=1"},
		"--answer", "name=stage-app",
		"--answer", "environment=staging",
	)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d. stdout=%q stderr=%q", code, stdout, stderr)
	}
	want := `Deploying "stage-app" to staging (all regions: false)`
	if !strings.Contains(stdout, want) {
		t.Errorf("expected stdout to contain %q, got %q", want, stdout)
	}
}

func TestE2E_DeployMissingRequired(t *testing.T) {
	if testing.Short() || deployBin == "" {
		t.Skip("e2e requires built binary; skipped under -short")
	}
	stdout, stderr, code := runDeploy(t,
		[]string{"CI=1"},
		"--answer", "name=just-the-name",
	)
	if code == 0 {
		t.Fatalf("expected non-zero exit, got 0. stdout=%q stderr=%q", stdout, stderr)
	}
	if !strings.Contains(stderr, "missing required answers for:") {
		t.Errorf("expected stderr to contain missing-answers header, got %q", stderr)
	}
	if !strings.Contains(stderr, "--environment") {
		t.Errorf("expected stderr to list --environment, got %q", stderr)
	}
	if !strings.Contains(stderr, "(env: DEPLOY_ENVIRONMENT)") {
		t.Errorf("expected stderr to surface env hint, got %q", stderr)
	}
}

func TestE2E_DeployInvalidSelectOption(t *testing.T) {
	if testing.Short() || deployBin == "" {
		t.Skip("e2e requires built binary; skipped under -short")
	}
	stdout, stderr, code := runDeploy(t,
		[]string{"CI=1"},
		"--answer", "name=myapp",
		"--answer", "environment=production",
	)
	if code == 0 {
		t.Fatalf("expected non-zero exit, got 0. stdout=%q stderr=%q", stdout, stderr)
	}
	if !strings.Contains(stderr, `field "environment"`) {
		t.Errorf("expected stderr to surface field name, got %q", stderr)
	}
	if !strings.Contains(stderr, `"production" is not a valid option`) {
		t.Errorf("expected stderr to surface invalid-option error, got %q", stderr)
	}
}

func TestE2E_DeployAnswerFile(t *testing.T) {
	if testing.Short() || deployBin == "" {
		t.Skip("e2e requires built binary; skipped under -short")
	}
	tmp := t.TempDir()
	path := filepath.Join(tmp, "answers.yaml")
	body := "name: file-app\nenvironment: prod\nall-regions: true\n"
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	stdout, stderr, code := runDeploy(t,
		[]string{"CI=1"},
		"--answer-file", path,
	)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d. stdout=%q stderr=%q", code, stdout, stderr)
	}
	want := `Deploying "file-app" to prod (all regions: true)`
	if !strings.Contains(stdout, want) {
		t.Errorf("expected stdout to contain %q, got %q", want, stdout)
	}
}
