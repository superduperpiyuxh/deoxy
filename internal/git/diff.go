package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

type ChangedFileSet map[string]bool

func GetRepoRoot(path string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = path
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git: not a git repository or git unavailable: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func ChangedFiles(path string) (ChangedFileSet, error) {
	root, err := GetRepoRoot(path)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command("git", "diff", "--name-only", "HEAD")
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git: diff failed: %w", err)
	}

	files := make(ChangedFileSet)
	for _, line := range bytes.Split(bytes.TrimSpace(out), []byte("\n")) {
		rel := strings.TrimSpace(string(line))
		if rel == "" {
			continue
		}
		abs := filepath.Join(root, rel)
		abs, err := filepath.Abs(abs)
		if err != nil {
			continue
		}
		files[abs] = true
	}
	return files, nil
}

func (c ChangedFileSet) Matches(absPath string) bool {
	if c == nil {
		return true
	}
	return c[absPath]
}
