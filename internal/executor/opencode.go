package executor

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// Runner runs an external executor command. It is abstracted so deterministic
// tests can substitute a fake without a real LLM or OpenCode server.
type Runner interface {
	Run(ctx context.Context, dir string, command []string, stdout, stderr io.Writer) error
}

// RunnerFunc adapts a function to the Runner interface.
type RunnerFunc func(ctx context.Context, dir string, command []string, stdout, stderr io.Writer) error

// Run implements Runner.
func (f RunnerFunc) Run(ctx context.Context, dir string, command []string, stdout, stderr io.Writer) error {
	return f(ctx, dir, command, stdout, stderr)
}

// ExecRunner runs the executor command as a real process.
type ExecRunner struct{}

// Run implements Runner.
func (ExecRunner) Run(ctx context.Context, dir string, command []string, stdout, stderr io.Writer) error {
	if len(command) == 0 {
		return fmt.Errorf("empty executor command")
	}
	cmd := exec.CommandContext(ctx, command[0], command[1:]...)
	cmd.Dir = dir
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("executor command %q failed: %w", command[0], err)
	}
	return nil
}

// executorCommand builds the process command line for the supported executors.
// The attach URL, model, and auto-approval flag are supplied explicitly rather
// than being globally hard-coded. The model uses the OpenCode form
// provider/model#variant.
func executorCommand(executor, attach, model string, auto bool, prompt string) ([]string, error) {
	switch strings.ToLower(strings.TrimSpace(executor)) {
	case "", "opencode":
		command := []string{"opencode", "run"}
		if strings.TrimSpace(attach) != "" {
			command = append(command, "--server", strings.TrimSpace(attach))
		}
		if strings.TrimSpace(model) != "" {
			command = append(command, "--model", strings.TrimSpace(model))
		}
		if auto {
			command = append(command, "--auto")
		}
		return append(command, prompt), nil
	default:
		return nil, fmt.Errorf("executor %q does not support autonomous execution", executor)
	}
}

// executionPrompt points OpenCode at the durable repository contract instead of
// re-encoding the whole implementation contract here.
func executionPrompt(repoURL string, task Task) string {
	subject := "the active task"
	if task.ID != "" {
		subject = fmt.Sprintf("%s revision %d", task.ID, task.Revision)
	}

	var b strings.Builder
	b.WriteString("You are the PrincipalFoundry implementation executor for this repository.\n\n")
	if repoURL != "" {
		fmt.Fprintf(&b, "Repository: %s\n\n", repoURL)
	}
	b.WriteString("Work in the current project repository.\n\n")
	fmt.Fprintf(&b, "Read AGENTS.md first, then .factory/CURRENT_TASK.md, and execute exactly %s.\n\n", subject)
	b.WriteString("Obey docs/PRODUCT_SPEC.md and docs/ARCHITECTURE.md (including the ADRs in docs/decisions/).\n")
	b.WriteString("Perform the required tester, reviewer, and deterministic verification loop defined by AGENTS.md.\n")
	b.WriteString("Update the required .factory evidence files: IMPLEMENTATION_REPORT.md, TESTER_REPORT.md, INTERNAL_REVIEW.md, and VERIFICATION_REPORT.md.\n\n")
	b.WriteString("Do not change product requirements. Do not make unauthorized architecture decisions; return ARCHITECT_DECISION_REQUIRED when one is genuinely required. Do not start another task.\n\n")
	b.WriteString("Commit the completed executor handoff, and push it to the shared branch when the repository's Git policy permits.\n")
	return b.String()
}
