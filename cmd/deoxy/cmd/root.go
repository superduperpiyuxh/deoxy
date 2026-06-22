package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "deoxy",
	Short: "Multi-language documentation comment generator",
	Long: `deoxy generates documentation comments for your source code.
It uses tree-sitter to parse Go, Python, C, C++, and Rust files
and inserts idiomatic doc comments (GoDoc, Doxygen, JSDoc, etc.).`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(watchCmd)
}
