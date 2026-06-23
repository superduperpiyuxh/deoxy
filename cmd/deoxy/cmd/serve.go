package cmd

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/superduperpiyuxh/deoxy/internal/config"
	"github.com/superduperpiyuxh/deoxy/internal/generator"
	"github.com/superduperpiyuxh/deoxy/internal/lsp"
	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"
)

type stdioReadWriter struct {
	io.Reader
	io.Writer
}

func (rw *stdioReadWriter) Close() error { return nil }

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start LSP language server (stdin/stdout for IDE integration)",
	Long: `Start the deoxy LSP language server.

Implements a minimal LSP over stdin/stdout supporting textDocument/codeAction
for doc comment generation.

Supported LSP methods:
  initialize
  textDocument/didOpen, textDocument/didChange, textDocument/didClose
  textDocument/codeAction
  shutdown

Designed for VS Code extension via vscode-languageclient.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Fprintf(os.Stderr, "deoxy serve starting\n")

		cfg := config.LoadDefaultConfig()
		gen, err := generator.New(cfg)
		if err != nil {
			return fmt.Errorf("creating generator: %w", err)
		}
		defer gen.Close()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		server := lsp.NewServer(gen, cancel)
		stream := jsonrpc2.NewStream(&stdioReadWriter{Reader: os.Stdin, Writer: os.Stdout})
		_, conn, _ := protocol.NewServer(ctx, server, stream)

		fmt.Fprintf(os.Stderr, "deoxy serve exited: %v\n", <-conn.Done())
		return nil
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
