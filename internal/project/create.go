package project

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
)

type Config struct {
	Name      string
	Architect string
	Executor  string
	ParentDir string
}

var validName = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)

func Create(cfg Config) (string, error) {
	root, files, err := prepare(cfg)
	if err != nil {
		return "", err
	}
	if err := writeFiles(root, files); err != nil {
		return "", err
	}
	return absolutePath(root), nil
}

// prepare validates the configuration and returns the destination root and the
// set of repository-relative files to write. It validates the name, normalizes
// the executor, and rejects an existing destination, but does not write files.
func prepare(cfg Config) (string, map[string]string, error) {
	if !validName.MatchString(cfg.Name) {
		return "", nil, fmt.Errorf("invalid project name %q: use letters, numbers, '.', '_' or '-'", cfg.Name)
	}
	normalizedExecutor, err := normalizeExecutor(cfg.Executor)
	if err != nil {
		return "", nil, err
	}
	cfg.Executor = normalizedExecutor

	if cfg.ParentDir == "" {
		cfg.ParentDir = "."
	}

	root := filepath.Join(cfg.ParentDir, cfg.Name)
	if _, err := os.Stat(root); err == nil {
		return "", nil, fmt.Errorf("destination already exists: %s", root)
	} else if !os.IsNotExist(err) {
		return "", nil, err
	}

	return root, templateFiles(cfg), nil
}

// writeFiles writes repository-relative files under root in a deterministic
// order, creating parent directories as needed.
func writeFiles(root string, files map[string]string) error {
	paths := make([]string, 0, len(files))
	for path := range files {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	for _, path := range paths {
		fullPath := filepath.Join(root, filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(fullPath, []byte(files[path]), 0o644); err != nil {
			return err
		}
	}
	return nil
}

func absolutePath(root string) string {
	abs, err := filepath.Abs(root)
	if err != nil {
		return root
	}
	return abs
}
