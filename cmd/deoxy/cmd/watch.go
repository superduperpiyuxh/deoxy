package cmd

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
)

var watchCmd = &cobra.Command{
	Use:   "watch [path]",
	Short: "Watch files and auto-generate doc comments on change",
	Long: `Watch files in the given directory and automatically regenerate doc
comments when source files are modified.

Uses fsnotify to detect file system events. When a supported source file
is written, deoxy re-processes it and injects any missing doc comments.

Note: This is an experimental feature. Recursive directory watching and
full regeneration pipeline are not yet implemented.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}

		absPath, err := filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("resolving path: %w", err)
		}

		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			return fmt.Errorf("creating watcher: %w", err)
		}
		defer watcher.Close()

		// Add path to watcher
		if err := watcher.Add(absPath); err != nil {
			return fmt.Errorf("watching path %q: %w", absPath, err)
		}

		fmt.Printf("Watching %s for changes... (Ctrl+C to stop)\n", absPath)

		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Printf("Modified: %s (regeneration not yet implemented)\n", event.Name)
					// TODO: regenerate doc comments for this file
					// The full implementation will:
					// 1. Check if the file extension is supported
					// 2. Run the generator pipeline on the changed file
					// 3. Insert any missing doc comments
					// 4. Report what was added
				}
			case err := <-watcher.Errors:
				log.Printf("Watch error: %v\n", err)
			}
		}
	},
}
