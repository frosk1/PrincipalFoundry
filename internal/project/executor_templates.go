package project

import (
	"fmt"
	"strings"
)

const (
	openCodeTesterPath   = ".opencode/agents/tester.md"
	openCodeReviewerPath = ".opencode/agents/reviewer.md"
)

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

func applyExecutorTemplates(files map[string]string, executor string) map[string]string {
	tester := files[openCodeTesterPath]
	reviewer := files[openCodeReviewerPath]
	delete(files, openCodeTesterPath)
	delete(files, openCodeReviewerPath)

	switch executor {
	case "opencode":
		files[openCodeTesterPath] = tester
		files[openCodeReviewerPath] = reviewer
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
