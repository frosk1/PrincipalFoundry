package cli

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/frosk1/PrincipalFoundry/docs/prompts"
	"github.com/frosk1/PrincipalFoundry/internal/executor"
	"github.com/frosk1/PrincipalFoundry/internal/gitrepo"
	"github.com/frosk1/PrincipalFoundry/internal/project"
)

const version = "0.1.0-dev"

func Run(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		printUsage(stdout)
		return nil
	}

	switch args[0] {
	case "new":
		return runNew(args[1:], stdout, stderr)
	case "init":
		return runInit(args[1:], stdout, stderr)
	case "prompt":
		return runPrompt(args[1:], stdout, stderr)
	case "executor":
		return runExecutor(args[1:], stdout, stderr)
	case "version", "--version", "-v":
		fmt.Fprintln(stdout, version)
		return nil
	case "help", "--help", "-h":
		printUsage(stdout)
		return nil
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func runNew(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("new", flag.ContinueOnError)
	fs.SetOutput(stderr)

	architect := fs.String("architect", "chatgpt", "architect agent (for example: chatgpt, claude, none)")
	executor := fs.String("executor", "opencode", "implementation agent (for example: opencode, codex, claude-code)")
	outputDir := fs.String("dir", ".", "parent directory in which the project directory is created")

	if len(args) == 0 {
		return errors.New("usage: pfoundry new <project-name> [--architect chatgpt] [--executor opencode] [--dir .]")
	}
	name := args[0]
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return errors.New("unexpected positional arguments")
	}

	cfg := project.Config{
		Name:      name,
		Architect: *architect,
		Executor:  *executor,
		ParentDir: *outputDir,
	}

	path, err := project.Create(cfg)
	if err != nil {
		return err
	}

	if err := gitrepo.Init(path); err != nil {
		return fmt.Errorf("created project at %s but failed to initialize git: %w", path, err)
	}

	fmt.Fprintf(stdout, "Created PrincipalFoundry project %q at %s\n", cfg.Name, path)
	fmt.Fprintln(stdout, "Initialized Git repository on branch main")
	fmt.Fprintln(stdout, "Created bootstrap commit: chore: initialize PrincipalFoundry project")
	fmt.Fprintln(stdout, "Phase: DISCOVERY")
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, gitrepo.RemoteInstructions(path))
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "After pushing, open the repository with your architect and run discovery before issuing implementation tasks.")
	return nil
}

func runInit(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	fs.SetOutput(stderr)

	architect := fs.String("architect", "chatgpt", "architect agent (for example: chatgpt, claude, none)")
	executor := fs.String("executor", "opencode", "implementation agent (for example: opencode, codex, claude-code)")
	outputDir := fs.String("dir", ".", "parent directory in which the project directory is created")
	from := fs.String("from", "", "approved design package directory produced by the discovery phase")
	task := fs.String("task", "", "optional first implementation task to activate as .factory/CURRENT_TASK.md")

	if len(args) == 0 {
		return errors.New("usage: pfoundry init <project-name> --from <design-package> [--architect chatgpt] [--executor opencode] [--task <task-file>] [--dir .]")
	}
	name := args[0]
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return errors.New("unexpected positional arguments")
	}
	if strings.TrimSpace(*from) == "" {
		return errors.New("--from <design-package> is required")
	}

	cfg := project.InitConfig{
		Name:      name,
		Architect: *architect,
		Executor:  *executor,
		ParentDir: *outputDir,
		DesignDir: *from,
		TaskFile:  *task,
	}

	path, err := project.CreateFromDesign(cfg)
	if err != nil {
		return err
	}

	if err := gitrepo.Init(path); err != nil {
		return fmt.Errorf("created project at %s but failed to initialize git: %w", path, err)
	}

	fmt.Fprintf(stdout, "Created PrincipalFoundry project %q at %s\n", name, path)
	fmt.Fprintf(stdout, "Installed approved design package from %s\n", *from)
	if strings.TrimSpace(*task) != "" {
		fmt.Fprintln(stdout, "Activated first implementation task as .factory/CURRENT_TASK.md (status: READY)")
		fmt.Fprintln(stdout, "Phase: SCAFFOLDING")
	} else {
		fmt.Fprintln(stdout, "Phase: REPOSITORY_BOOTSTRAP")
	}
	fmt.Fprintln(stdout, "Initialized Git repository on branch main")
	fmt.Fprintln(stdout, "Created bootstrap commit: "+gitrepo.InitialCommitMessage())
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, gitrepo.RemoteInstructions(path))
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "The architect should review the imported design and issue the first bounded implementation task before implementation begins.")
	return nil
}

func runPrompt(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		return errors.New("usage: pfoundry prompt architect-loop --repo <github-repository-url>")
	}

	switch args[0] {
	case "architect-loop":
		fs := flag.NewFlagSet("prompt architect-loop", flag.ContinueOnError)
		fs.SetOutput(stderr)
		repositoryURL := fs.String("repo", "", "HTTPS GitHub repository URL, for example https://github.com/owner/repo")

		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if fs.NArg() != 0 {
			return errors.New("unexpected positional arguments")
		}
		if strings.TrimSpace(*repositoryURL) == "" {
			return errors.New("--repo <github-repository-url> is required")
		}

		rendered, err := prompts.RenderArchitectDevelopmentLoop(*repositoryURL)
		if err != nil {
			return err
		}
		fmt.Fprint(stdout, rendered)
		if !strings.HasSuffix(rendered, "\n") {
			fmt.Fprintln(stdout)
		}
		return nil
	default:
		return fmt.Errorf("unknown prompt %q", args[0])
	}
}

func runExecutor(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		return errors.New("usage: pfoundry executor <status|tick> [options]")
	}

	switch args[0] {
	case "status":
		return runExecutorStatus(args[1:], stdout, stderr)
	case "tick":
		return runExecutorTick(args[1:], stdout, stderr)
	default:
		return fmt.Errorf("unknown executor subcommand %q", args[0])
	}
}

func runExecutorStatus(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("executor status", flag.ContinueOnError)
	fs.SetOutput(stderr)

	dir := fs.String("dir", ".", "repository directory to inspect")
	remote := fs.String("remote", "origin", "Git remote used to identify the repository")
	asJSON := fs.Bool("json", false, "print machine-readable JSON")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if fs.NArg() != 0 {
		return errors.New("unexpected positional arguments")
	}

	report, err := executor.Status(*dir, *remote)
	if err != nil {
		return err
	}

	if *asJSON {
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(report)
	}

	fmt.Fprintf(stdout, "Repository: %s\n", report.Repository)
	fmt.Fprintf(stdout, "Factory: %s\n", report.Phase)
	if report.TaskID != "" {
		fmt.Fprintf(stdout, "Task: %s r%d\n", report.TaskID, report.Revision)
	} else {
		fmt.Fprintln(stdout, "Task: (none)")
	}
	fmt.Fprintf(stdout, "State: %s\n", report.State)
	fmt.Fprintf(stdout, "Last executor result: %s\n", report.LastResult)
	fmt.Fprintf(stdout, "Result commit: %s\n", report.ResultCommit)
	if report.Reason != "" {
		fmt.Fprintf(stdout, "Reason: %s\n", report.Reason)
	}
	return nil
}

func runExecutorTick(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("executor tick", flag.ContinueOnError)
	fs.SetOutput(stderr)

	dir := fs.String("dir", ".", "repository directory to execute in")
	remote := fs.String("remote", "origin", "Git remote to synchronize with")
	executorName := fs.String("executor", "", "implementation executor (defaults to the factory state, e.g. opencode)")
	attach := fs.String("attach", "", "attach to an existing executor server, e.g. http://127.0.0.1:4096")
	model := fs.String("model", "", "executor model in provider/model#variant form, e.g. openrouter/deepseek/deepseek-v4.1-flash")
	auto := fs.Bool("auto", false, "auto-approve non-denied executor permissions for unattended runs")
	verbose := fs.Bool("verbose", false, "print the executor command and full output")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if fs.NArg() != 0 {
		return errors.New("unexpected positional arguments")
	}

	result, err := executor.Tick(context.Background(), executor.TickConfig{
		Dir:      *dir,
		Remote:   *remote,
		Executor: *executorName,
		Attach:   *attach,
		Model:    *model,
		Auto:     *auto,
		Runner:   executor.ExecRunner{},
		Stdout:   stdout,
		Stderr:   stderr,
		Verbose:  *verbose,
	})

	if result.TaskID != "" {
		fmt.Fprintf(stdout, "%s: %s (%s r%d)\n", result.Outcome, result.Message, result.TaskID, result.Revision)
	} else {
		fmt.Fprintf(stdout, "%s: %s\n", result.Outcome, result.Message)
	}
	return err
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, `PrincipalFoundry - bootstrap senior-led AI software teams

Usage:
  pfoundry new <project-name> [options]
  pfoundry init <project-name> --from <design-package> [options]
  pfoundry prompt architect-loop --repo <github-repository-url>
  pfoundry executor status [--json] [--dir .] [--remote origin]
  pfoundry executor tick [--executor opencode] [--attach URL] [--model provider/model#variant] [--auto] [--dir .] [--remote origin] [--verbose]
  pfoundry version

Commands:
  new       Create a PrincipalFoundry project in DISCOVERY phase
  init      Bootstrap a project repository from an approved design package
  prompt    Render a ready-to-use agent prompt
  executor  Inspect or run the deterministic executor gate
  version   Print the CLI version`)
}
