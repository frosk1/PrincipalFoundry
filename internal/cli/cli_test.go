package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/frosk1/PrincipalFoundry/docs/prompts"
)

func TestPromptArchitectLoopRendersRepositoryURL(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := Run([]string{
		"prompt",
		"architect-loop",
		"--repo",
		"https://github.com/frosk1/Principalfoundry",
	}, &stdout, &stderr)
	if err != nil {
		t.Fatal(err)
	}

	got := stdout.String()
	if strings.Contains(got, prompts.RepositoryURLPlaceholder) {
		t.Fatal("rendered prompt still contains repository placeholder")
	}
	if !strings.Contains(got, "GitHub repository: https://github.com/frosk1/Principalfoundry") {
		t.Fatal("rendered prompt missing repository URL")
	}
	if !strings.HasPrefix(got, "# Architect Development Loop Prompt") {
		t.Fatalf("stdout should contain only the prompt, got %q", got[:min(len(got), 80)])
	}
}

func TestPromptArchitectLoopRequiresRepositoryURL(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := Run([]string{"prompt", "architect-loop"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected missing --repo to fail")
	}
	if !strings.Contains(err.Error(), "--repo") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPromptArchitectLoopRejectsNonGitHubRepository(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := Run([]string{
		"prompt",
		"architect-loop",
		"--repo",
		"https://gitlab.com/example/project",
	}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected non-GitHub repository to fail")
	}
}
