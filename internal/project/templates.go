package project

import "fmt"

func templateFiles(cfg Config) map[string]string {
	return map[string]string{
		"AGENTS.md": `# Agent Instructions

This repository uses PrincipalFoundry.

Implementation agents must follow .factory/CURRENT_TASK.md as the single active implementation contract. Do not invent major architecture or product decisions. If a task requires changing an approved architectural decision, stop and report that an architect decision is required.

Do not begin implementation while .factory/FACTORY_STATE.md is in DISCOVERY or DESIGN unless explicitly instructed by a human.
`,
		"docs/PRODUCT_SPEC.md": `# Product Specification

Status: Not established.

This document becomes the durable source of truth for what the product must do after discovery is approved.
`,
		"docs/ARCHITECTURE.md": `# Architecture

Status: Not established.

The architect records the current approved technical architecture here after discovery.
`,
		"docs/decisions/INDEX.md": `# Architecture Decision Records

No architecture decisions have been recorded yet.
`,
		".factory/FACTORY_STATE.md": fmt.Sprintf(`---
phase: DISCOVERY
architect: %s
executor: %s
---

# Factory State

The project is in DISCOVERY. Human and architect should establish product intent, requirements, constraints, architecture, and initial decisions before implementation begins.
`, cfg.Architect, cfg.Executor),
		".factory/PROJECT_BRIEF.md": fmt.Sprintf(`# Project Brief

Project: %s

Status: Discovery in progress.

Use this document to synthesize the initial human/architect discovery conversation. It is bootstrap context; once approved, durable product requirements belong in docs/PRODUCT_SPEC.md and durable technical decisions belong in docs/ARCHITECTURE.md and docs/decisions/.
`, cfg.Name),
		".factory/CURRENT_TASK.md": `---
status: IDLE
---

# Current Task

No implementation task has been issued. Implementation must not begin while the project is in DISCOVERY.
`,
		".factory/REVIEW.md":                "# Architect Review\n\nNo implementation has been reviewed yet.\n",
		".factory/IMPLEMENTATION_REPORT.md": "# Implementation Report\n\nNo implementation cycle has run yet.\n",
		".factory/TESTER_REPORT.md":         "# Tester Report\n\nNo test cycle has run yet.\n",
		".factory/INTERNAL_REVIEW.md":       "# Internal Review\n\nNo internal review has run yet.\n",
		".factory/VERIFICATION_REPORT.md":   "# Verification Report\n\nNo verification cycle has run yet.\n",
		".opencode/agents/tester.md": `---
description: Independent tester for PrincipalFoundry implementation cycles
mode: subagent
---

Act as an adversarial tester. Read the current task and the actual implementation. Identify impacted behavior, run relevant tests, and report reproducible failures. Do not weaken tests to make the implementation pass.
`,
		".opencode/agents/reviewer.md": `---
description: Independent reviewer for PrincipalFoundry implementation cycles
mode: subagent
---

Review the implementation against the current task, architecture, and evidence. Prefer read-only review. Return PASS, FINDINGS, or ARCHITECT_DECISION_REQUIRED. Do not silently change architecture or acceptance criteria.
`,
	}
}
