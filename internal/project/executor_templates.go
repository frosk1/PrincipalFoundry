package project

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	openCodeTesterPath   = ".opencode/agents/tester.md"
	openCodeReviewerPath = ".opencode/agents/reviewer.md"
)

// OpenCode model syntax supports nested provider paths and an optional variant.
// Restricting frontmatter values prevents YAML injection and accidental settings.
var openCodeModelPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*(/[A-Za-z0-9._:-]+)+(#[A-Za-z0-9._-]+)?$`)

func validateSubagentModels(cfg Config) error {
	for _, role := range []struct{ name, model string }{
		{"tester", cfg.TesterModel}, {"reviewer", cfg.ReviewerModel},
	} {
		model := strings.TrimSpace(role.model)
		if model == "" {
			continue
		}
		if cfg.Executor != "opencode" {
			return fmt.Errorf("%s model override requires --executor opencode", role.name)
		}
		if !openCodeModelPattern.MatchString(model) {
			return fmt.Errorf("invalid %s model %q: use provider/model[#variant]", role.name, model)
		}
	}
	return nil
}

func withOpenCodeModel(content, model string) string {
	model = strings.TrimSpace(model)
	if model == "" {
		return content // inherit executor model unchanged
	}
	return strings.Replace(content, "mode: subagent\n", "mode: subagent\nmodel: "+model+"\n", 1)
}

func normalizeExecutor(executor string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(executor)) {
	case "", "opencode":
		return "opencode", nil
	case "claude-code":
		return "claude-code", nil
	case "codex":
		return "codex", nil
	case "none":
		return "none", nil
	default:
		return "", fmt.Errorf("unsupported executor %q: supported executors are opencode, claude-code, codex, none", executor)
	}
}

func applyExecutorTemplates(files map[string]string, cfg Config) map[string]string {
	tester := files[openCodeTesterPath]
	reviewer := files[openCodeReviewerPath]
	delete(files, openCodeTesterPath)
	delete(files, openCodeReviewerPath)

	switch cfg.Executor {
	case "opencode":
		files[openCodeTesterPath] = withOpenCodeModel(tester, cfg.TesterModel)
		files[openCodeReviewerPath] = withOpenCodeModel(reviewer, cfg.ReviewerModel)
	case "claude-code":
		files[".claude/agents/tester.md"] = claudeAgentMarkdown(
			"tester",
			"Adversarial tester for PrincipalFoundry implementation cycles",
			stripMarkdownFrontmatter(tester),
			"Read, Grep, Glob, Bash, Write, Edit",
		)
		files[".claude/agents/reviewer.md"] = claudeAgentMarkdown(
			"reviewer",
			"Independent reviewer for PrincipalFoundry implementation cycles",
			stripMarkdownFrontmatter(reviewer),
			"Read, Grep, Glob, Bash",
		)
	case "codex":
		files[".codex/agents/tester.toml"] = codexAgentTOML(
			"tester",
			"Adversarial tester for PrincipalFoundry implementation cycles",
			stripMarkdownFrontmatter(tester),
			false,
		)
		files[".codex/agents/reviewer.toml"] = codexAgentTOML(
			"reviewer",
			"Independent reviewer for PrincipalFoundry implementation cycles",
			stripMarkdownFrontmatter(reviewer),
			true,
		)
	case "none":
		// The repository-level AGENTS.md remains the executor contract,
		// but no executor-specific subagent definitions are generated.
	}

	return files
}

func stripMarkdownFrontmatter(content string) string {
	parts := strings.SplitN(content, "---\n", 3)
	if len(parts) != 3 {
		return strings.TrimSpace(content) + "\n"
	}
	return strings.TrimSpace(parts[2]) + "\n"
}

func claudeAgentMarkdown(name, description, body, tools string) string {
	return fmt.Sprintf(`---
name: %s
description: %s
tools: %s
---

%s`, name, description, tools, body)
}

func codexAgentTOML(name, description, body string, readOnly bool) string {
	sandbox := ""
	if readOnly {
		sandbox = "sandbox_mode = \"read-only\"\n"
	}

	return fmt.Sprintf(`name = "%s"
description = "%s"
%sdeveloper_instructions = """
%s"""
`, name, description, sandbox, body)
}
