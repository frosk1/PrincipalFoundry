package prompts

import (
	_ "embed"
	"fmt"
	"net/url"
	"strings"
)

const RepositoryURLPlaceholder = "{{GITHUB_REPOSITORY_URL}}"

//go:embed architect-development-loop.md
var architectDevelopmentLoop string

func ArchitectDevelopmentLoopTemplate() string {
	return architectDevelopmentLoop
}

func RenderArchitectDevelopmentLoop(repositoryURL string) (string, error) {
	normalized, err := normalizeGitHubRepositoryURL(repositoryURL)
	if err != nil {
		return "", err
	}
	if !strings.Contains(architectDevelopmentLoop, RepositoryURLPlaceholder) {
		return "", fmt.Errorf("architect development loop prompt is missing repository placeholder %s", RepositoryURLPlaceholder)
	}
	return strings.ReplaceAll(architectDevelopmentLoop, RepositoryURLPlaceholder, normalized), nil
}

func normalizeGitHubRepositoryURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("GitHub repository URL is required")
	}

	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("invalid GitHub repository URL %q: %w", raw, err)
	}
	if u.Scheme != "https" || !strings.EqualFold(u.Host, "github.com") {
		return "", fmt.Errorf("repository must be an HTTPS GitHub URL like https://github.com/owner/repo")
	}

	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", fmt.Errorf("repository must point to a GitHub repository root like https://github.com/owner/repo")
	}

	repo := strings.TrimSuffix(parts[1], ".git")
	if repo == "" {
		return "", fmt.Errorf("repository name is required")
	}

	return "https://github.com/" + parts[0] + "/" + repo, nil
}
