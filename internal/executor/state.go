package executor

import (
	"fmt"
	"strings"
)

// State is the effective executor state for a repository. It answers the single
// question: may the executor run right now, and if not, why not?
type State string

const (
	StateReadyToExecute   State = "READY_TO_EXECUTE"
	StateWaitingArchitect State = "WAITING_ARCHITECT"
	StateNoActiveTask     State = "NO_ACTIVE_TASK"
	StatePaused           State = "PAUSED"
	StateBlockedHuman     State = "BLOCKED_HUMAN"
	StateAlreadyRunning   State = "ALREADY_RUNNING"
	StatePhaseBlocked     State = "PHASE_BLOCKED"
	StateError            State = "ERROR"
)

// autonomousPhase is the lifecycle phase required for scheduled autonomous
// execution. Other phases are recorded and reported but never executed.
const autonomousPhase = "AUTONOMOUS"

// Task statuses understood in .factory/CURRENT_TASK.md frontmatter.
const (
	taskStatusIdle         = "IDLE"
	taskStatusReady        = "READY"
	taskStatusBlocked      = "BLOCKED"
	taskStatusBlockedHuman = "BLOCKED_HUMAN"
)

// Evidence statuses understood in .factory/IMPLEMENTATION_REPORT.md frontmatter.
const (
	evidenceCompleted         = "COMPLETED"
	evidenceFailed            = "FAILED"
	evidenceBlocked           = "BLOCKED"
	evidenceBlockedHuman      = "BLOCKED_HUMAN"
	evidenceArchitectDecision = "ARCHITECT_DECISION_REQUIRED"
)

// Facts are the deterministic inputs to the executor state decision. They are
// gathered from repository state and from the local executor lock, never from
// wall-clock time.
type Facts struct {
	LockHeld        bool
	GitError        string
	Phase           string
	TaskPresent     bool
	TaskID          string
	TaskStatus      string
	Revision        int
	EvidenceMatches bool
	EvidenceStatus  string
}

// PhaseAllowsAutonomous reports whether a factory lifecycle phase permits
// scheduled autonomous execution.
func PhaseAllowsAutonomous(phase string) bool {
	return strings.EqualFold(strings.TrimSpace(phase), autonomousPhase)
}

// Decide returns the effective state and a short human-readable reason.
//
// The order is deliberate and deterministic:
//  1. an already-running cycle is reported before anything else;
//  2. any unsafe Git condition is an error;
//  3. a paused factory never executes;
//  4. missing or idle tasks mean there is no active work;
//  5. human/architect-blocked tasks never execute;
//  6. evidence for the same task revision means the executor handoff already
//     exists, so the loop is waiting for the architect rather than re-running;
//  7. only a READY task in an autonomous phase is executable.
func Decide(f Facts) (State, string) {
	if f.LockHeld {
		return StateAlreadyRunning, "an executor cycle is already running"
	}
	if f.GitError != "" {
		return StateError, f.GitError
	}
	if strings.EqualFold(strings.TrimSpace(f.Phase), "PAUSED") {
		return StatePaused, "factory phase is PAUSED"
	}
	if !f.TaskPresent || f.TaskStatus == "" || f.TaskStatus == taskStatusIdle {
		return StateNoActiveTask, "no active implementation task"
	}
	switch f.TaskStatus {
	case taskStatusBlockedHuman:
		return StateBlockedHuman, "task is blocked pending a human decision"
	case taskStatusBlocked:
		return StateBlockedHuman, "task is blocked"
	}
	if f.EvidenceMatches {
		switch f.EvidenceStatus {
		case evidenceFailed:
			return StateError, "the previous executor attempt failed for " + taskRef(f)
		case evidenceBlocked, evidenceBlockedHuman, evidenceArchitectDecision:
			return StateBlockedHuman, "the executor requires an architect or human decision for " + taskRef(f)
		default:
			return StateWaitingArchitect, "waiting for architect review of " + taskRef(f)
		}
	}
	if f.TaskStatus == taskStatusReady {
		if f.TaskID == "" {
			return StateError, "READY task has no task_id, so execution cannot be made idempotent"
		}
		if !PhaseAllowsAutonomous(f.Phase) {
			return StatePhaseBlocked, fmt.Sprintf("factory phase %s does not permit autonomous execution", phaseLabel(f.Phase))
		}
		return StateReadyToExecute, taskRef(f) + " is ready to execute"
	}
	return StateNoActiveTask, "task status " + f.TaskStatus + " is not executable"
}

func taskRef(f Facts) string {
	id := f.TaskID
	if id == "" {
		id = "task"
	}
	if f.Revision > 0 {
		return fmt.Sprintf("%s r%d", id, f.Revision)
	}
	return id
}

func phaseLabel(phase string) string {
	phase = strings.TrimSpace(phase)
	if phase == "" {
		return "(unset)"
	}
	return phase
}
