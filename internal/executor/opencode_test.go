package executor

import (
	"reflect"
	"strings"
	"testing"
)

func TestExecutorCommandOpenCode(t *testing.T) {
	tests := []struct {
		name     string
		executor string
		attach   string
		model    string
		prompt   string
		want     []string
	}{
		{
			name:     "plain",
			executor: "opencode",
			prompt:   "do the task",
			want:     []string{"opencode", "run", "do the task"},
		},
		{
			name:     "attach and model",
			executor: "opencode",
			attach:   "http://127.0.0.1:4096",
			model:    "openrouter/deepseek/deepseek-v4.1-flash",
			prompt:   "do the task",
			want:     []string{"opencode", "run", "--server", "http://127.0.0.1:4096", "--model", "openrouter/deepseek/deepseek-v4.1-flash", "do the task"},
		},
		{
			name:     "default executor is opencode",
			executor: "",
			model:    "openrouter/deepseek/deepseek-v4.1-flash#high",
			prompt:   "p",
			want:     []string{"opencode", "run", "--model", "openrouter/deepseek/deepseek-v4.1-flash#high", "p"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := executorCommand(tt.executor, tt.attach, tt.model, tt.prompt)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("command = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestExecutorCommandRejectsUnsupported(t *testing.T) {
	if _, err := executorCommand("codex", "", "", "p"); err == nil {
		t.Fatal("expected unsupported executor to fail")
	}
}

func TestExecutionPromptReferencesContract(t *testing.T) {
	prompt := executionPrompt("https://github.com/frosk1/dockspur", Task{Present: true, ID: "TASK-0001", Revision: 1})

	for _, marker := range []string{
		"AGENTS.md",
		".factory/CURRENT_TASK.md",
		"TASK-0001 revision 1",
		"docs/PRODUCT_SPEC.md",
		"docs/ARCHITECTURE.md",
		"ARCHITECT_DECISION_REQUIRED",
		"push",
	} {
		if !strings.Contains(prompt, marker) {
			t.Errorf("execution prompt missing %q", marker)
		}
	}
}
