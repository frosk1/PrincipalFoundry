package project

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

type Config struct {
	Name      string
	Architect string
	Executor  string
	ParentDir string
}

var validName = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)

func Create(cfg Config) (string, error) {
	if !validName.MatchString(cfg.Name) {
		return "", fmt.Errorf("invalid project name %q: use letters, numbers, '.', '_' or '-'", cfg.Name)
	}
	if cfg.ParentDir == "" {
		cfg.ParentDir = "."
	}

	root := filepath.Join(cfg.ParentDir, cfg.Name)
	if _, err := os.Stat(root); err == nil {
		return "", fmt.Errorf("destination already exists: %s", root)
	} else if !os.IsNotExist(err) {
		return "", err
	}

	files := templateFiles(cfg)
	for path, content := range files {
		fullPath := filepath.Join(root, filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			return "", err
		}
		if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
			return "", err
		}
	}

	abs, err := filepath.Abs(root)
	if err != nil {
		return root, nil
	}
	return abs, nil
}
