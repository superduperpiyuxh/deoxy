package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// runRoot is a simple helper that executes rootCmd with the given args.
func runRoot(args []string) (string, error) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()
	rootCmd.SetOut(nil)
	rootCmd.SetErr(nil)
	return buf.String(), err
}

// TestRootHelp checks the root command's help output.
func TestRootHelp(t *testing.T) {
	output, err := runRoot([]string{"--help"})
	if err != nil {
		t.Fatalf("root --help failed: %v", err)
	}
	if !strings.Contains(output, "deoxy") {
		t.Errorf("expected help to contain 'deoxy'")
	}
	if !strings.Contains(output, "generate") {
		t.Errorf("expected help to mention generate command")
	}
	if !strings.Contains(output, "init") {
		t.Errorf("expected help to mention init command")
	}
	if !strings.Contains(output, "watch") {
		t.Errorf("expected help to mention watch command")
	}
}

// TestGenerateHelp tests the generate command's help output.
func TestGenerateHelp(t *testing.T) {
	output, err := runRoot([]string{"generate", "--help"})
	if err != nil {
		t.Fatalf("generate --help failed: %v", err)
	}
	if !strings.Contains(output, "--diff") {
		t.Errorf("expected --diff flag in help")
	}
	if !strings.Contains(output, "--dry-run") {
		t.Errorf("expected --dry-run flag in help")
	}
	if !strings.Contains(output, "--force") {
		t.Errorf("expected --force flag in help")
	}
	if !strings.Contains(output, "--config") {
		t.Errorf("expected --config flag in help")
	}
}

// TestInitHelp tests the init command's help output.
func TestInitHelp(t *testing.T) {
	output, err := runRoot([]string{"init", "--help"})
	if err != nil {
		t.Fatalf("init --help failed: %v", err)
	}
	if !strings.Contains(output, ".deoxy.yaml") {
		t.Errorf("expected init help to mention .deoxy.yaml")
	}
}

// TestWatchHelp tests the watch command's help output.
func TestWatchHelp(t *testing.T) {
	output, err := runRoot([]string{"watch", "--help"})
	if err != nil {
		t.Fatalf("watch --help failed: %v", err)
	}
	if !strings.Contains(output, "watch") {
		t.Errorf("expected watch help to contain 'watch'")
	}
}

// TestInitCreatesConfig tests that 'deoxy init' creates a .deoxy.yaml file.
func TestInitCreatesConfig(t *testing.T) {
	dir := t.TempDir()

	// Directly invoke initCmd RunE instead of routing through rootCmd
	// to avoid cross-test state issues with cobra's global root command.
	buf := new(bytes.Buffer)
	initCmd.SetOut(buf)
	initCmd.SetErr(buf)
	err := initCmd.RunE(initCmd, []string{dir})
	if err != nil {
		t.Fatalf("init failed: %v (output: %s)", err, buf.String())
	}

	configPath := filepath.Join(dir, ".deoxy.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatalf("expected .deoxy.yaml to be created at %s", configPath)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read created config: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "deoxy") {
		t.Errorf("expected config to mention deoxy")
	}
	if !strings.Contains(content, "default_style") {
		t.Errorf("expected config to contain default_style")
	}
}

// TestInitFailsIfExists tests that 'deoxy init' fails when .deoxy.yaml exists.
func TestInitFailsIfExists(t *testing.T) {
	dir := t.TempDir()
	existingPath := filepath.Join(dir, ".deoxy.yaml")
	if err := os.WriteFile(existingPath, []byte("existing: true\n"), 0644); err != nil {
		t.Fatalf("failed to write existing config: %v", err)
	}

	buf := new(bytes.Buffer)
	initCmd.SetOut(buf)
	initCmd.SetErr(buf)
	err := initCmd.RunE(initCmd, []string{dir})
	if err == nil {
		t.Fatal("expected init to fail when .deoxy.yaml already exists")
	}
}

// TestInitFailsWithNonExistentDir tests error handling for bad paths.
func TestInitFailsWithNonExistentDir(t *testing.T) {
	buf := new(bytes.Buffer)
	initCmd.SetOut(buf)
	initCmd.SetErr(buf)
	err := initCmd.RunE(initCmd, []string{"/nonexistent/path/that/does/not/exist"})
	if err == nil {
		t.Fatal("expected init to fail with non-existent directory")
	}
}

// runGenerate executes the generateCmd's RunE directly with the given flag values.
// This avoids cross-test state pollution from cobra's global root command.
func runGenerate(t *testing.T, args []string, force, dryRun, diff bool, config string) (string, error) {
	t.Helper()

	// Set global flags
	forceFlag = force
	dryRunFlag = dryRun
	diffFlag = diff
	configFlag = config

	// Capture output
	buf := new(bytes.Buffer)
	generateCmd.SetOut(buf)
	generateCmd.SetErr(buf)

	err := generateCmd.RunE(generateCmd, args)
	return buf.String(), err
}

// TestGenerateWithNonexistentPath tests error handling for bad generate paths.
func TestGenerateWithNonexistentPath(t *testing.T) {
	output, err := runGenerate(t, []string{"/nonexistent/path"}, false, false, false, "")
	if err != nil {
		t.Fatalf("generate should handle errors gracefully, got: %v output: %s", err, output)
	}
}

// TestGenerateDryRun tests that --dry-run doesn't modify files.
func TestGenerateDryRun(t *testing.T) {
	if os.Getenv("CGO_ENABLED") == "0" {
		t.Skip("skipping test: CGO_ENABLED=0")
	}
	dir := t.TempDir()
	src := "package test\n\nfunc Add(a int, b int) int {\n\treturn a + b\n}\n"
	goFile := filepath.Join(dir, "main.go")
	if err := os.WriteFile(goFile, []byte(src), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	output, err := runGenerate(t, []string{dir}, false, true, false, "")
	if err != nil {
		t.Fatalf("generate --dry-run failed: %v", err)
	}

	content, err := os.ReadFile(goFile)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if string(content) != src {
		t.Errorf("file was modified despite --dry-run")
	}
	if !strings.Contains(output, "dry-run") {
		t.Errorf("expected output to mention dry-run, got: %s", output)
	}
}

// TestGenerateDiff tests that --diff shows changes without modifying files.
func TestGenerateDiff(t *testing.T) {
	if os.Getenv("CGO_ENABLED") == "0" {
		t.Skip("skipping test: CGO_ENABLED=0")
	}
	dir := t.TempDir()
	src := "package test\n\nfunc Add(a int, b int) int {\n\treturn a + b\n}\n"
	goFile := filepath.Join(dir, "main.go")
	if err := os.WriteFile(goFile, []byte(src), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	output, err := runGenerate(t, []string{dir}, false, false, true, "")
	if err != nil {
		t.Fatalf("generate --diff failed: %v", err)
	}

	content, err := os.ReadFile(goFile)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if string(content) != src {
		t.Errorf("file was modified despite --diff")
	}
	if !strings.Contains(output, "+ Insert comment") && !strings.Contains(output, "+++") {
		t.Errorf("expected diff output, got: %s", output)
	}
}

// TestGenerateForce tests that --force overwrites existing comments.
func TestGenerateForce(t *testing.T) {
	if os.Getenv("CGO_ENABLED") == "0" {
		t.Skip("skipping test: CGO_ENABLED=0")
	}
	dir := t.TempDir()
	src := "package test\n\n// Old comment\nfunc Add(a int, b int) int {\n\treturn a + b\n}\n"
	goFile := filepath.Join(dir, "main.go")
	if err := os.WriteFile(goFile, []byte(src), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	output, err := runGenerate(t, []string{dir}, true, false, false, "")
	if err != nil {
		t.Fatalf("generate --force failed: %v", err)
	}
	_ = output

	content, err := os.ReadFile(goFile)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if !strings.Contains(string(content), "Add adds") {
		t.Errorf("expected generated comment containing 'Add adds', got:\n%s", string(content))
	}
	if strings.Contains(string(content), "Old comment") {
		t.Errorf("expected old comment to be replaced, but still found it:\n%s", string(content))
	}
}

// TestGenerateSkipsExistingByDefault tests that existing comments are preserved.
func TestGenerateSkipsExistingByDefault(t *testing.T) {
	if os.Getenv("CGO_ENABLED") == "0" {
		t.Skip("skipping test: CGO_ENABLED=0")
	}
	dir := t.TempDir()
	src := "package test\n\n// Old manual comment\nfunc Add(a int, b int) int {\n\treturn a + b\n}\n"
	goFile := filepath.Join(dir, "main.go")
	if err := os.WriteFile(goFile, []byte(src), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	output, err := runGenerate(t, []string{dir}, false, false, false, "")
	if err != nil {
		t.Fatalf("generate failed: %v", err)
	}
	_ = output

	content, err := os.ReadFile(goFile)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if !strings.Contains(string(content), "Old manual comment") {
		t.Errorf("expected existing comment to be preserved, got:\n%s", string(content))
	}
}
