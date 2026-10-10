package gitrepo

import (
	"fmt"
	"os/exec"
	"strings"
)

const initialCommitMessage = "chore: initialize PrincipalFoundry project"

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
