package runner

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func RunSteps(label, scriptDir string, steps []string) error {
	script := "set -e\n" + strings.Join(steps, "\n")
	cmd := exec.Command("bash", "-c", script)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Env = append(os.Environ(), "SCRIPT_DIR="+scriptDir)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("[%s] failed: %w", label, err)
	}
	return nil
}
