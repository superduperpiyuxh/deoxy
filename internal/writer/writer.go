// Package writer handles injecting generated doc comments into source files.
// It detects existing comments, manages indentation, and supports dry-run
// and diff-only modes for safe preview before writing.
package writer

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/superduperpiyuxh/deoxy/internal/generator"
	"github.com/superduperpiyuxh/deoxy/internal/scanner"
)

// FileEdit represents a set of insertions to make in a single file.
type FileEdit struct {
	FilePath   string
	Insertions []Insertion
	// Original is the original file content for diff calculation.
	Original []byte
	// Modified is the content after all insertions (nil if no changes).
	Modified []byte
}

// Insertion represents a single doc comment to insert before a symbol.
type Insertion struct {
	Line       int    // 0-indexed line number to insert BEFORE
	Content    string // the doc comment to insert
	SymbolName string // for logging
}

// Writer manages the process of injecting doc comments into source files.
type Writer struct {
	gen      *generator.Generator
	force    bool
	diffOnly bool
	dryRun   bool
}

// New creates a Writer that uses the given Generator to produce doc comments.
//
//   - force: overwrite existing comments (default: skip)
//   - diffOnly: show proposed changes without writing
//   - dryRun: process files but do not write
func New(gen *generator.Generator, force, diffOnly, dryRun bool) *Writer {
	return &Writer{
		gen:      gen,
		force:    force,
		diffOnly: diffOnly,
		dryRun:   dryRun,
	}
}

// Generate processes a single file by running it through the generator pipeline,
// then determining which comments need insertion. Returns the computed edits
// without writing unless !dryRun && !diffOnly.
//
// When diffOnly is true, modified content is still computed (for display) but
// not written. When dryRun is true, no modification occurs at all.
func (w *Writer) Generate(filePath string) ([]FileEdit, error) {
	// Ensure path exists
	info, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("writer: cannot access %q: %w", filePath, err)
	}

	if info.IsDir() {
		return w.GenerateDir(filePath)
	}

	// Run the generator pipeline on this single file
	result, err := w.gen.Run([]string{filePath})
	if err != nil {
		return nil, fmt.Errorf("writer: generator failed for %q: %w", filePath, err)
	}

	if len(result.Files) == 0 {
		return nil, nil
	}

	// Read the original source
	src, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("writer: failed to read %q: %w", filePath, err)
	}

	fileResult := result.Files[0]
	if fileResult.ParseError != nil {
		return nil, fmt.Errorf("writer: parse error for %q: %w", filePath, fileResult.ParseError)
	}

	if len(fileResult.Comments) == 0 {
		// No comments generated — nothing to do
		return nil, nil
	}

	return w.applyComments(filePath, src, fileResult.Comments)
}

// GenerateDir processes all supported files in a directory.
func (w *Writer) GenerateDir(dirPath string) ([]FileEdit, error) {
	// Scan the directory for source files
	scanResult, err := scanner.Scan(dirPath)
	if err != nil {
		return nil, fmt.Errorf("writer: scan failed for %q: %w", dirPath, err)
	}

	if len(scanResult.Files) == 0 {
		return nil, nil
	}

	return w.GenerateAll(scanResult)
}

// GenerateAll processes all files from a scanner result, running each through
// the generator and computing edits.
func (w *Writer) GenerateAll(scanResult *scanner.Result) ([]FileEdit, error) {
	if scanResult == nil || len(scanResult.Files) == 0 {
		return nil, nil
	}

	// Collect all file paths
	paths := make([]string, len(scanResult.Files))
	for i, f := range scanResult.Files {
		paths[i] = f.Path
	}

	// Run the generator on all discovered files
	result, err := w.gen.Run(paths)
	if err != nil {
		return nil, fmt.Errorf("writer: generator failed: %w", err)
	}

	if len(result.Files) == 0 {
		return nil, nil
	}

	var allEdits []FileEdit

	for _, fileResult := range result.Files {
		if fileResult.ParseError != nil {
			// Skip files with parse errors but don't abort
			continue
		}

		if len(fileResult.Comments) == 0 {
			continue
		}

		edits, err := w.applyComments(fileResult.Path, fileResult.Source, fileResult.Comments)
		if err != nil {
			return nil, err
		}
		if edits != nil {
			allEdits = append(allEdits, edits[0])
		}
	}

	return allEdits, nil
}

// applyComments does the core work: for a single file's source and its
// generated comments, determine which comments need insertion, compute
// the resulting modified source, and optionally write it to disk.
func (w *Writer) applyComments(filePath string, src []byte, comments []generator.SymbolComment) ([]FileEdit, error) {
	// Convert source to lines, preserving original endings and trailing newline
	lines, sep, trailingNewline := toLines(src)

	// Sort comments by StartLine descending so we insert bottom-up,
	// preserving line number validity as we go.
	sorted := make([]generator.SymbolComment, len(comments))
	copy(sorted, comments)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].StartLine > sorted[j].StartLine
	})

	var insertions []Insertion
	modifiedLines := make([]string, len(lines))
	copy(modifiedLines, lines)

	for _, sc := range sorted {
		if sc.DocComment == "" {
			continue
		}

		// Check if there's already an existing comment above this symbol
		hasComment := hasExistingComment(modifiedLines, sc.StartLine)

		if hasComment {
			if !w.force {
				// Skip — preserve existing comment
				continue
			}
			// Force mode: remove the existing comment block first.
			// The insertion point becomes where the old comment started.
			commentStart, commentEnd := findCommentBlock(modifiedLines, sc.StartLine)
			if commentStart >= 0 {
				nRemoved := commentEnd - commentStart
				modifiedLines = append(modifiedLines[:commentStart], modifiedLines[commentEnd:]...)
				sc.StartLine -= nRemoved
			}
		}

		// Determine indentation from the target line
		indent := extractIndent(modifiedLines[sc.StartLine])

		// Insert the comment
		modifiedLines = insertComment(modifiedLines, sc.StartLine, sc.DocComment, indent)

		insertions = append(insertions, Insertion{
			Line:       sc.StartLine,
			Content:    sc.DocComment,
			SymbolName: sc.Symbol.Name,
		})
	}

	if len(insertions) == 0 {
		return nil, nil
	}

	// Reconstruct modified content
	modified := fromLines(modifiedLines, sep, trailingNewline)

	edit := FileEdit{
		FilePath:   filePath,
		Insertions: insertions,
		Original:   src,
		Modified:   modified,
	}

	// Write to disk if not dry-run and not diff-only
	if !w.dryRun && !w.diffOnly {
		if err := os.WriteFile(filePath, modified, 0644); err != nil {
			return nil, fmt.Errorf("writer: failed to write %q: %w", filePath, err)
		}
	}

	return []FileEdit{edit}, nil
}

// isBuildDirective returns true for Go compiler directives (//go:build, //go:generate,
// //go:embed, // +build) that start with // but are not doc comments.
func isBuildDirective(line string) bool {
	trimmed := strings.TrimSpace(line)
	return strings.HasPrefix(trimmed, "//go:") ||
		strings.HasPrefix(trimmed, "// +") ||
		strings.HasPrefix(trimmed, "//export")
}

// isCommentLine returns true if the line is a doc comment line.
// Supports Go (//), Rust (///), Python (#, """, '''), and C/C++ (/*, *, */) styles.
// Excludes compiler directives like //go:build.
func isCommentLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	if isBuildDirective(line) {
		return false
	}
	return strings.HasPrefix(trimmed, "//") ||
		strings.HasPrefix(trimmed, "#") ||
		strings.HasPrefix(trimmed, "\"\"\"") ||
		strings.HasPrefix(trimmed, "'''") ||
		strings.HasPrefix(trimmed, "/*") ||
		strings.HasPrefix(trimmed, "*") ||
		strings.HasPrefix(trimmed, "*/")
}

// hasExistingComment checks whether there is already a doc comment above the
// given line index. It looks backwards from lineIdx, skipping blank lines,
// and returns true if a comment line is found.
func hasExistingComment(lines []string, lineIdx int) bool {
	if lineIdx <= 0 {
		return false
	}

	i := lineIdx - 1
	for i >= 0 && strings.TrimSpace(lines[i]) == "" {
		i--
	}

	return i >= 0 && isCommentLine(lines[i])
}

// findCommentBlock finds the range of an existing comment block above the given
// line index. Returns (startLine, endLine) where startLine is the first line of
// the comment block (inclusive) and endLine is the first line after the block
// (exclusive). Returns (-1, -1) if no comment block is found.
func findCommentBlock(lines []string, lineIdx int) (int, int) {
	if lineIdx <= 0 {
		return -1, -1
	}

	i := lineIdx - 1
	for i >= 0 && strings.TrimSpace(lines[i]) == "" {
		i--
	}

	if i < 0 || !isCommentLine(lines[i]) {
		return -1, -1
	}

	// endLine is exclusive — it's right past the last comment line (before any blank lines)
	endLine := i + 1

	startLine := i
	for startLine > 0 {
		prev := strings.TrimSpace(lines[startLine-1])
		if prev == "" || isCommentLine(lines[startLine-1]) {
			startLine--
		} else {
			break
		}
	}

	return startLine, endLine
}

// insertComment inserts a doc comment before the given line index, preserving
// indentation. It adds a blank line after the comment if the target line is
// not already preceded by a blank line.
func insertComment(lines []string, lineIdx int, comment string, indent string) []string {
	// Split the comment into lines (already has comment prefix like //, ///, etc.)
	commentLines := strings.Split(comment, "\n")

	// Calculate result size: original lines + comment lines + optional blank line
	extraLines := len(commentLines)

	// Check if we need an extra blank line before the comment
	needsLeadingBlank := lineIdx > 0 && strings.TrimSpace(lines[lineIdx-1]) != ""

	if needsLeadingBlank {
		extraLines++
	}

	result := make([]string, 0, len(lines)+extraLines)

	// Copy lines before the insertion point
	result = append(result, lines[:lineIdx]...)

	// Add leading blank line if needed
	if needsLeadingBlank {
		result = append(result, "")
	}

	// Add indented comment lines
	for _, cl := range commentLines {
		if cl == "" {
			result = append(result, indent)
		} else {
			result = append(result, indent+cl)
		}
	}

	// Copy remaining lines (from lineIdx onward)
	result = append(result, lines[lineIdx:]...)

	return result
}

// fromLines joins lines back into source bytes using the detected line separator
// and restoring the trailing newline if the original had one.
func fromLines(lines []string, sep string, trailingNewline bool) []byte {
	result := strings.Join(lines, sep)
	if trailingNewline {
		result += sep
	}
	return []byte(result)
}

// extractIndent returns the leading whitespace of the given line.
func extractIndent(line string) string {
	for i, c := range line {
		if c != ' ' && c != '\t' {
			return line[:i]
		}
	}
	return line
}

// toLines splits source bytes into lines, preserving information about
// the original line ending convention and trailing newline.
// Returns the lines, the line separator, and whether the source ended with
// a trailing newline.
func toLines(src []byte) (lines []string, sep string, trailingNewline bool) {
	if len(src) == 0 {
		return nil, "\n", false
	}

	// Detect line ending
	sep = "\n"
	if bytes.Contains(src, []byte("\r\n")) {
		sep = "\r\n"
	}

	s := string(src)
	trailingNewline = strings.HasSuffix(s, sep)

	// Trim trailing separator for consistent split
	if trailingNewline && len(s) >= len(sep) {
		s = s[:len(s)-len(sep)]
	}

	lines = strings.Split(s, sep)
	return lines, sep, trailingNewline
}
