package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"

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
		return errors.New("usage: principalfoundry new <project-name> [--architect chatgpt] [--executor opencode] [--dir .]")
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

	fmt.Fprintf(stdout, "Created PrincipalFoundry project %q at %s\n", cfg.Name, path)
	fmt.Fprintln(stdout, "Phase: DISCOVERY")
	fmt.Fprintln(stdout, "Next: open the repository with your architect and run the discovery process before issuing implementation tasks.")
	return nil
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, `PrincipalFoundry - bootstrap senior-led AI software teams

Usage:
  principalfoundry new <project-name> [options]
  principalfoundry version

Commands:
  new       Create a PrincipalFoundry project in DISCOVERY phase
  version   Print the CLI version`)
}
