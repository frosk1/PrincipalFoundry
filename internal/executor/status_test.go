package executor

import (
	"encoding/json"
	"testing"
)

func TestStatusReadyToExecute(t *testing.T) {
	f := newFixture(t)
	f.setupAutonomous("TASK-0001", 1)

	report, err := Status(f.dir, "origin")
	if err != nil {
		t.Fatal(err)
	}
	if report.State != StateReadyToExecute {
		t.Fatalf("state = %s (%s)", report.State, report.Reason)
	}
	if report.Phase != "AUTONOMOUS" {
		t.Fatalf("phase = %s", report.Phase)
	}
	if report.TaskID != "TASK-0001" || report.Revision != 1 {
		t.Fatalf("task = %s r%d", report.TaskID, report.Revision)
	}
	if report.LastResult != "NONE" {
		t.Fatalf("last result = %s", report.LastResult)
	}
}

func TestStatusPhaseBlocked(t *testing.T) {
	f := newFixture(t)
	f.write(".factory/FACTORY_STATE.md", "---\nphase: SCAFFOLDING\narchitect: chatgpt\nexecutor: opencode\n---\n")
	f.write(".factory/CURRENT_TASK.md", "---\ntask_id: TASK-0001\nrevision: 1\nstatus: READY\n---\n\n# TASK-0001\n")
	f.write(".factory/IMPLEMENTATION_REPORT.md", "# Implementation Report\n")
	f.commit("scaffolding")

	report, err := Status(f.dir, "origin")
	if err != nil {
		t.Fatal(err)
	}
	if report.State != StatePhaseBlocked {
		t.Fatalf("state = %s (%s)", report.State, report.Reason)
	}
}

func TestStatusWaitingArchitect(t *testing.T) {
	f := newFixture(t)
	f.setupAutonomous("TASK-0001", 1)
	f.setEvidence("TASK-0001", 1, evidenceCompleted, "abc1234")
	f.commit("executor: complete")

	report, err := Status(f.dir, "origin")
	if err != nil {
		t.Fatal(err)
	}
	if report.State != StateWaitingArchitect {
		t.Fatalf("state = %s (%s)", report.State, report.Reason)
	}
	if report.LastResult != evidenceCompleted {
		t.Fatalf("last result = %s", report.LastResult)
	}
	if report.ResultCommit != "abc1234" {
		t.Fatalf("result commit = %s", report.ResultCommit)
	}
}

func TestStatusJSONIsStable(t *testing.T) {
	f := newFixture(t)
	f.setupAutonomous("TASK-0001", 1)

	report, err := Status(f.dir, "origin")
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("status JSON is not parseable: %v", err)
	}
	for _, key := range []string{"repository", "branch", "phase", "task_id", "revision", "state", "reason", "last_result", "result_commit"} {
		if _, ok := parsed[key]; !ok {
			t.Errorf("status JSON missing stable key %q", key)
		}
	}
	if parsed["state"] != string(StateReadyToExecute) {
		t.Errorf("state = %v", parsed["state"])
	}
	if parsed["phase"] != "AUTONOMOUS" {
		t.Errorf("phase = %v", parsed["phase"])
	}
	if parsed["task_id"] != "TASK-0001" {
		t.Errorf("task_id = %v", parsed["task_id"])
	}
}
