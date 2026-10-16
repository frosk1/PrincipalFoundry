package executor

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// Control-file paths relative to the repository root.
const (
	factoryStatePath         = ".factory/FACTORY_STATE.md"
	currentTaskPath          = ".factory/CURRENT_TASK.md"
	implementationReportPath = ".factory/IMPLEMENTATION_REPORT.md"
)

// taskIDPattern extracts a task identifier such as TASK-0012 from a document.
var taskIDPattern = regexp.MustCompile(`TASK-[0-9]+`)

// FactoryState is the parsed .factory/FACTORY_STATE.md control file.
type FactoryState struct {
	Phase     string
	Architect string
	Executor  string
}

// Task is the parsed .factory/CURRENT_TASK.md control file.
type Task struct {
	Present  bool
	ID       string
	Revision int
	Status   string
}

// Evidence is the parsed .factory/IMPLEMENTATION_REPORT.md executor evidence.
type Evidence struct {
	Present      bool
	TaskID       string
	Revision     int
	Status       string
	ResultCommit string
}

func readControlFile(root, rel string) (string, error) {
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ParseFactoryState reads and parses .factory/FACTORY_STATE.md.
func ParseFactoryState(root string) (FactoryState, error) {
	content, err := readControlFile(root, factoryStatePath)
	if err != nil {
		return FactoryState{}, err
	}
	fields := parseFrontmatter(content)
	return FactoryState{
		Phase:     strings.ToUpper(strings.TrimSpace(fields["phase"])),
		Architect: strings.TrimSpace(fields["architect"]),
		Executor:  strings.TrimSpace(fields["executor"]),
	}, nil
}

// ParseTask reads and parses .factory/CURRENT_TASK.md. When the frontmatter does
// not carry a task_id, it is derived from the task title so that existing task
// documents remain usable.
func ParseTask(root string) (Task, error) {
	content, err := readControlFile(root, currentTaskPath)
	if err != nil {
		return Task{}, err
	}
	fields := parseFrontmatter(content)

	task := Task{
		ID:       strings.ToUpper(strings.TrimSpace(fields["task_id"])),
		Revision: parseRevision(fields["revision"]),
		Status:   strings.ToUpper(strings.TrimSpace(fields["status"])),
	}
	if task.ID == "" {
		if match := taskIDPattern.FindString(content); match != "" {
			task.ID = strings.ToUpper(match)
		}
	}
	if task.ID != "" {
		task.Present = true
	}
	return task, nil
}

// ParseEvidence reads and parses .factory/IMPLEMENTATION_REPORT.md.
func ParseEvidence(root string) (Evidence, error) {
	content, err := readControlFile(root, implementationReportPath)
	if err != nil {
		return Evidence{}, err
	}
	fields := parseFrontmatter(content)

	evidence := Evidence{
		TaskID:       strings.ToUpper(strings.TrimSpace(fields["task_id"])),
		Revision:     parseRevision(fields["revision"]),
		Status:       strings.ToUpper(strings.TrimSpace(fields["status"])),
		ResultCommit: strings.TrimSpace(fields["result_commit"]),
	}
	if evidence.TaskID != "" && evidence.Status != "" {
		evidence.Present = true
	}
	return evidence, nil
}

// Matches reports whether the evidence identifies the given task and revision.
func (e Evidence) Matches(task Task) bool {
	return e.Present && task.Present && e.TaskID == task.ID && e.Revision == task.Revision
}

func parseRevision(value string) int {
	value = strings.TrimSpace(value)
	if value == "" {
		return 1
	}
	n, err := strconv.Atoi(value)
	if err != nil || n < 1 {
		return 1
	}
	return n
}
