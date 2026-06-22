package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/superduperpiyuxh/deoxy/internal/config"
	"github.com/superduperpiyuxh/deoxy/internal/generator"
	"github.com/superduperpiyuxh/deoxy/internal/writer"
)

var (
	diffFlag   bool
	dryRunFlag bool
	forceFlag  bool
	configFlag string
)

var generateCmd = &cobra.Command{
	Use:   "generate [path...]",
	Short: "Generate doc comments for source files",
	Long: `Generate doc comments for source files in the given paths.

Scans directories recursively for supported source files (Go, Python, C,
C++, Rust), generates idiomatic doc comments using tree-sitter AST parsing,
and injects them into the source files. Existing comments are preserved
unless --force is used.

Supports GoDoc (Go), Google-style docstrings (Python), Doxygen (C/C++),
and Rustdoc (Rust) comment styles.`,
	Args: cobra.MinimumNArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		// 1. Load config
		cfg := config.LoadDefaultConfig()
		if configFlag != "" {
			var err error
			cfg, err = config.LoadConfig(configFlag)
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}
		}

		// 2. Create generator (handles parser + template engine internally)
		gen, err := generator.New(cfg)
		if err != nil {
			return fmt.Errorf("creating generator: %w", err)
		}
		defer gen.Close()

		// 3. Create writer
		w := writer.New(gen, forceFlag, diffFlag, dryRunFlag)

		// 4. Process paths
		paths := args
		if len(paths) == 0 {
			paths = []string{"."}
		}

		totalFiles := 0
		totalInsertions := 0

		for _, p := range paths {
			// Resolve the path
			absPath, err := filepath.Abs(p)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Warning: cannot resolve path %q: %v\n", p, err)
				continue
			}

			edits, err := w.Generate(absPath)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Warning: processing %q: %v\n", p, err)
				continue
			}

			for _, edit := range edits {
				totalFiles++
				totalInsertions += len(edit.Insertions)

				if dryRunFlag {
					// Show what would be done
					relPath := relPathOrName(edit.FilePath, p)
					cmd.Printf("[dry-run] %s: %d comment(s) to insert\n", relPath, len(edit.Insertions))
					for _, ins := range edit.Insertions {
						cmd.Printf("  - %s (line %d)\n", ins.SymbolName, ins.Line)
					}
				} else if diffFlag {
					// Show the diff
					relPath := relPathOrName(edit.FilePath, p)
					cmd.Printf("=== %s ===\n", relPath)
					cmd.Printf("--- a/%s\n", relPath)
					cmd.Printf("+++ b/%s\n", relPath)
					for _, ins := range edit.Insertions {
						cmd.Printf("+ Insert comment for %s at line %d\n", ins.SymbolName, ins.Line)
						for _, line := range strings.Split(ins.Content, "\n") {
							cmd.Printf("+%s\n", line)
						}
					}
				} else {
					// Normal mode - already written by writer
					relPath := relPathOrName(edit.FilePath, p)
					cmd.Printf("\u2713 %s: %d comment(s) inserted\n", relPath, len(edit.Insertions))
				}
			}
		}

		if totalFiles == 0 {
			cmd.Println("No files processed.")
			return nil
		}

		if !diffFlag {
			cmd.Printf("\nProcessed %d file(s), %d comment(s) generated.\n", totalFiles, totalInsertions)
		}

		return nil
	},
}

// relPathOrName returns a relative path if possible, otherwise the base name.
func relPathOrName(absPath, root string) string {
	rel, err := filepath.Rel(root, absPath)
	if err != nil {
		return filepath.Base(absPath)
	}
	return rel
}

func init() {
	generateCmd.Flags().BoolVarP(&diffFlag, "diff", "d", false, "Show proposed changes without writing")
	generateCmd.Flags().BoolVar(&dryRunFlag, "dry-run", false, "Process files without writing")
	generateCmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "Overwrite existing comments")
	generateCmd.Flags().StringVarP(&configFlag, "config", "c", "", "Path to .deoxy.yaml config file")
}
