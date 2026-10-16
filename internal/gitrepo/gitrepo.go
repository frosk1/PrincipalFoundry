package gitrepo

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

const initialCommitMessage = "chore: initialize PrincipalFoundry project"

// InitialCommitMessage is the message used for the repository bootstrap commit.
func InitialCommitMessage() string {
	return initialCommitMessage
}

func Init(path string) error {
	if _, err := exec.LookPath("git"); err != nil {
		return fmt.Errorf("git is required but was not found in PATH")
	}

	if err := run(path, "init"); err != nil {
		return err
	}
	if err := run(path, "branch", "-M", "main"); err != nil {
		return err
	}
	if err := run(path, "add", "."); err != nil {
		return err
	}
	if err := run(path,
		"-c", "user.name=PrincipalFoundry",
		"-c", "user.email=pfoundry@local",
		"commit", "-m", initialCommitMessage,
	); err != nil {
		return err
	}
	return nil
}

func RemoteInstructions(path string) string {
	return fmt.Sprintf("Next: create an empty remote repository, then connect and push it:\n\n  cd %q\n  git remote add origin REMOTE_URL\n  git push -u origin main\n\nDo not initialize the remote with a README, .gitignore, or license; the local repository already has its bootstrap commit.", path)
}

func run(path string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = path
	output, err := cmd.CombinedOutput()
	if err != nil {
		detail := strings.TrimSpace(string(output))
		if detail == "" {
			return fmt.Errorf("git %s failed: %w", strings.Join(args, " "), err)
		}
		return fmt.Errorf("git %s failed: %s", strings.Join(args, " "), detail)
	}
	return nil
}

// Run runs git in dir and returns an error when the command fails.
func Run(dir string, args ...string) error {
	return run(dir, args...)
}

// Output runs git in dir and returns the trimmed standard output of the command.
func Output(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		detail := strings.TrimSpace(stderr.String())
		if detail == "" {
			detail = strings.TrimSpace(stdout.String())
		}
		if detail == "" {
			return "", fmt.Errorf("git %s failed: %w", strings.Join(args, " "), err)
		}
		return "", fmt.Errorf("git %s failed: %s", strings.Join(args, " "), detail)
	}
	return strings.TrimSpace(stdout.String()), nil
}
