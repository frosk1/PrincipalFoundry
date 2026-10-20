package prompts

import (
	"strings"
	"testing"
)

func TestArchitectDevelopmentLoopTemplateContainsRepositoryPlaceholder(t *testing.T) {
	if !strings.Contains(ArchitectDevelopmentLoopTemplate(), RepositoryURLPlaceholder) {
		t.Fatalf("template missing %s", RepositoryURLPlaceholder)
	}
}

func TestArchitectDevelopmentLoopTemplateTracksMilestones(t *testing.T) {
	template := ArchitectDevelopmentLoopTemplate()
	for _, marker := range []string{".factory/MILESTONE.md", "# Milestone-driven product delivery", "does not override the READY/no-new-evidence NO_ACTION rule"} {
		if !strings.Contains(template, marker) {
			t.Fatalf("architect prompt missing %q", marker)
		}
	}
}

func TestRenderArchitectDevelopmentLoopInjectsNormalizedRepositoryURL(t *testing.T) {
	got, err := RenderArchitectDevelopmentLoop(" https://github.com/frosk1/Principalfoundry.git/ ")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(got, RepositoryURLPlaceholder) {
		t.Fatalf("rendered prompt still contains placeholder")
	}
	if !strings.Contains(got, "GitHub repository: https://github.com/frosk1/Principalfoundry") {
		t.Fatalf("rendered prompt missing repository URL")
	}
}

func TestRenderArchitectDevelopmentLoopRejectsNonGitHubURL(t *testing.T) {
	_, err := RenderArchitectDevelopmentLoop("https://gitlab.com/example/project")
	if err == nil {
		t.Fatal("expected non-GitHub URL to fail")
	}
}
