// Package scanner discovers source files by walking directories or matching
// glob patterns, then groups them by detected language. It uses the parser
// registry's extension-to-language mapping — no tree-sitter dependency.
package scanner

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/superduperpiyuxh/deoxy/internal/parser"
)

// FileEntry represents a single discovered source file with its language identity.
type FileEntry struct {
	// Path is the absolute file path.
	Path string
	// Language is the canonical language name ("go", "python", etc.).
	Language string
	// RelPath is the path relative to the scan root.
	RelPath string
}

// Result holds scan results grouped by language.
type Result struct {
	// Files is a flat list of all discovered files.
	Files []FileEntry
	// ByLanguage maps language names to their file entries.
	ByLanguage map[string][]FileEntry
}

// Scan discovers source files at the given file/directory paths.
// It walks directories recursively, matching file extensions against the
// parser registry. Hidden files/directories (starting with ".") and
// vendor/ and node_modules/ directories are skipped.
// Duplicate files are included only once.
func Scan(paths ...string) (*Result, error) {
	exts := parser.SupportedExtensions()
	return ScanWithExtensions(paths, exts)
}

// ScanWithExtensions scans the given paths but only returns files matching
// the provided extensions list.
func ScanWithExtensions(paths []string, extensions []string) (*Result, error) {
	extSet := make(map[string]struct{}, len(extensions))
	for _, ext := range extensions {
		extSet[ext] = struct{}{}
	}

	result := &Result{
		ByLanguage: make(map[string][]FileEntry),
	}
	seen := make(map[string]bool)

	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			return nil, err
		}

		if info.IsDir() {
			err = filepath.Walk(path, func(fp string, fi os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				// Skip hidden files/directories
				if strings.HasPrefix(fi.Name(), ".") && fp != path {
					if fi.IsDir() {
						return filepath.SkipDir
					}
					return nil
				}

				// Skip vendor and node_modules directories
				if fi.IsDir() && (fi.Name() == "vendor" || fi.Name() == "node_modules") {
					return filepath.SkipDir
				}

				if fi.IsDir() {
					return nil
				}

				ext := filepath.Ext(fi.Name())
				if _, ok := extSet[ext]; !ok {
					return nil
				}

				absPath, err := filepath.Abs(fp)
				if err != nil {
					return err
				}

				if seen[absPath] {
					return nil
				}
				seen[absPath] = true

				lang, ok := parser.GetConfigForExtension(ext)
				if !ok {
					return nil
				}

				relPath, _ := filepath.Rel(path, fp)
				entry := FileEntry{
					Path:     absPath,
					Language: lang.Name,
					RelPath:  relPath,
				}
				result.Files = append(result.Files, entry)
				result.ByLanguage[lang.Name] = append(result.ByLanguage[lang.Name], entry)
				return nil
			})
			if err != nil {
				return nil, err
			}
		} else {
			ext := filepath.Ext(info.Name())
			if _, ok := extSet[ext]; !ok {
				continue
			}

			absPath, err := filepath.Abs(path)
			if err != nil {
				return nil, err
			}

			if seen[absPath] {
				continue
			}
			seen[absPath] = true

			lang, ok := parser.GetConfigForExtension(ext)
			if !ok {
				continue
			}

			entry := FileEntry{
				Path:     absPath,
				Language: lang.Name,
				RelPath:  filepath.Base(path),
			}
			result.Files = append(result.Files, entry)
			result.ByLanguage[lang.Name] = append(result.ByLanguage[lang.Name], entry)
		}
	}

	return result, nil
}

// DetectLanguage returns the language name for a file path based on its extension,
// or ("", false) if the extension is unknown.
func DetectLanguage(path string) (string, bool) {
	ext := filepath.Ext(path)
	cfg, ok := parser.GetConfigForExtension(ext)
	if !ok {
		return "", false
	}
	return cfg.Name, true
}
