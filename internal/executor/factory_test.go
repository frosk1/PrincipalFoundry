package executor

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFactoryFile(t *testing.T, root, rel, content string) {
	t.Helper()
	full := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestParseFrontmatter(t *testing.T) {
	content := "---\ntask_id: TASK-0012\nrevision: 2\nstatus: READY\nrequired_gates:\n  - BUILD\n  - UNIT\n---\n\n# TASK-0012\n"
	fields := parseFrontmatter(content)

	if fields["task_id"] != "TASK-0012" {
		t.Errorf("task_id = %q", fields["task_id"])
	}
	if fields["revision"] != "2" {
		t.Errorf("revision = %q", fields["revision"])
	}
	if fields["status"] != "READY" {
		t.Errorf("status = %q", fields["status"])
	}
}

func TestParseFrontmatterIgnoresContentWithoutBlock(t *testing.T) {
	if fields := parseFrontmatter("# Heading\n\nNo frontmatter here.\n"); len(fields) != 0 {
		t.Fatalf("expected no fields, got %#v", fields)
	}
}

func TestParseFactoryState(t *testing.T) {
	root := t.TempDir()
	writeFactoryFile(t, root, factoryStatePath, "---\nphase: autonomous\narchitect: chatgpt\nexecutor: opencode\n---\n\n# Factory State\n")

	state, err := ParseFactoryState(root)
	if err != nil {
		t.Fatal(err)
	}
	if state.Phase != "AUTONOMOUS" || state.Architect != "chatgpt" || state.Executor != "opencode" {
		t.Fatalf("unexpected factory state %#v", state)
	}
}

func TestParseTaskFromFrontmatter(t *testing.T) {
	root := t.TempDir()
	writeFactoryFile(t, root, currentTaskPath, "---\ntask_id: TASK-0007\nrevision: 3\nstatus: READY\n---\n\n# TASK-0007\n")

	task, err := ParseTask(root)
	if err != nil {
		t.Fatal(err)
	}
	if !task.Present || task.ID != "TASK-0007" || task.Revision != 3 || task.Status != taskStatusReady {
		t.Fatalf("unexpected task %#v", task)
	}
}

func TestParseTaskDerivesIDFromTitle(t *testing.T) {
	root := t.TempDir()
	writeFactoryFile(t, root, currentTaskPath, "---\nstatus: READY\n---\n\n# TASK-0042 — Some bounded outcome\n\nBody.\n")

	task, err := ParseTask(root)
	if err != nil {
		t.Fatal(err)
	}
	if task.ID != "TASK-0042" || task.Revision != 1 || task.Status != taskStatusReady {
		t.Fatalf("unexpected task %#v", task)
	}
}

func TestParseTaskIdlePlaceholder(t *testing.T) {
	root := t.TempDir()
	writeFactoryFile(t, root, currentTaskPath, "---\nstatus: IDLE\n---\n\n# Current Task\n\nNo implementation task has been issued.\n")

	task, err := ParseTask(root)
	if err != nil {
		t.Fatal(err)
	}
	if task.Present {
		t.Fatalf("idle placeholder should not be present: %#v", task)
	}
}

func TestParseEvidence(t *testing.T) {
	root := t.TempDir()
	writeFactoryFile(t, root, implementationReportPath, "---\ntask_id: TASK-0007\nrevision: 2\nstatus: COMPLETED\nresult_commit: deadbee\n---\n\n# Implementation Report\n")

	evidence, err := ParseEvidence(root)
	if err != nil {
		t.Fatal(err)
	}
	if !evidence.Present || evidence.TaskID != "TASK-0007" || evidence.Revision != 2 || evidence.Status != evidenceCompleted || evidence.ResultCommit != "deadbee" {
		t.Fatalf("unexpected evidence %#v", evidence)
	}
}

func TestParseEvidencePlaceholderIsNotPresent(t *testing.T) {
	root := t.TempDir()
	writeFactoryFile(t, root, implementationReportPath, "# Implementation Report\n\nNo implementation cycle has run yet.\n")

	evidence, err := ParseEvidence(root)
	if err != nil {
		t.Fatal(err)
	}
	if evidence.Present {
		t.Fatalf("placeholder report should not be present: %#v", evidence)
	}
}
