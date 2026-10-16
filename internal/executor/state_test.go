package executor

import "testing"

func readyFacts() Facts {
	return Facts{
		Phase:       "AUTONOMOUS",
		TaskPresent: true,
		TaskID:      "TASK-0012",
		TaskStatus:  taskStatusReady,
		Revision:    1,
	}
}

func TestDecide(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Facts)
		want   State
	}{
		{"ready task without evidence", func(f *Facts) {}, StateReadyToExecute},
		{"ready task with completed evidence for same revision", func(f *Facts) {
			f.EvidenceMatches = true
			f.EvidenceStatus = evidenceCompleted
		}, StateWaitingArchitect},
		{"higher revision is new work", func(f *Facts) {
			f.Revision = 2
			f.EvidenceMatches = false
		}, StateReadyToExecute},
		{"next task id is new work", func(f *Facts) {
			f.TaskID = "TASK-0013"
			f.EvidenceMatches = false
		}, StateReadyToExecute},
		{"paused factory", func(f *Facts) { f.Phase = "PAUSED" }, StatePaused},
		{"forbidden discovery phase", func(f *Facts) { f.Phase = "DISCOVERY" }, StatePhaseBlocked},
		{"forbidden scaffolding phase", func(f *Facts) { f.Phase = "SCAFFOLDING" }, StatePhaseBlocked},
		{"blocked human task", func(f *Facts) { f.TaskStatus = taskStatusBlockedHuman }, StateBlockedHuman},
		{"blocked task", func(f *Facts) { f.TaskStatus = taskStatusBlocked }, StateBlockedHuman},
		{"idle task", func(f *Facts) { f.TaskStatus = taskStatusIdle }, StateNoActiveTask},
		{"no task present", func(f *Facts) { f.TaskPresent = false }, StateNoActiveTask},
		{"executor lock held", func(f *Facts) { f.LockHeld = true }, StateAlreadyRunning},
		{"unsafe git state", func(f *Facts) { f.GitError = "worktree has uncommitted changes" }, StateError},
		{"failed evidence", func(f *Facts) {
			f.EvidenceMatches = true
			f.EvidenceStatus = evidenceFailed
		}, StateError},
		{"architect decision evidence", func(f *Facts) {
			f.EvidenceMatches = true
			f.EvidenceStatus = evidenceArchitectDecision
		}, StateBlockedHuman},
		{"ready without task id", func(f *Facts) { f.TaskID = "" }, StateError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			facts := readyFacts()
			tt.mutate(&facts)
			got, reason := Decide(facts)
			if got != tt.want {
				t.Fatalf("Decide() = %s (%s), want %s", got, reason, tt.want)
			}
			if reason == "" {
				t.Fatal("Decide() returned an empty reason")
			}
		})
	}
}

func TestPhaseAllowsAutonomous(t *testing.T) {
	if !PhaseAllowsAutonomous("AUTONOMOUS") || !PhaseAllowsAutonomous("autonomous") {
		t.Fatal("AUTONOMOUS should permit execution")
	}
	for _, phase := range []string{"", "DISCOVERY", "DESIGN", "REPOSITORY_BOOTSTRAP", "SCAFFOLDING", "MANUAL_VALIDATION", "PAUSED"} {
		if PhaseAllowsAutonomous(phase) {
			t.Errorf("phase %q should not permit autonomous execution", phase)
		}
	}
}

func TestEvidenceMatches(t *testing.T) {
	evidence := Evidence{Present: true, TaskID: "TASK-0001", Revision: 1, Status: evidenceCompleted}

	if !evidence.Matches(Task{Present: true, ID: "TASK-0001", Revision: 1}) {
		t.Fatal("same task and revision should match")
	}
	if evidence.Matches(Task{Present: true, ID: "TASK-0001", Revision: 2}) {
		t.Fatal("higher revision should not match")
	}
	if evidence.Matches(Task{Present: true, ID: "TASK-0002", Revision: 1}) {
		t.Fatal("different task id should not match")
	}
	if evidence.Matches(Task{Present: false, ID: "TASK-0001", Revision: 1}) {
		t.Fatal("absent task should not match")
	}
	if (Evidence{Present: false}).Matches(Task{Present: true, ID: "TASK-0001", Revision: 1}) {
		t.Fatal("absent evidence should not match")
	}
}
