package lsp_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/superduperpiyuxh/deoxy/internal/config"
	"github.com/superduperpiyuxh/deoxy/internal/generator"
	"github.com/superduperpiyuxh/deoxy/internal/lsp"
	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

func TestServer_Initialize(t *testing.T) {
	gen := newGenerator(t)
	defer gen.Close()
	_, cancel := context.WithCancel(context.Background())
	defer cancel()
	srv := lsp.NewServer(gen, cancel)

	result, err := srv.Initialize(context.Background(), &protocol.InitializeParams{})
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	if result.Capabilities.TextDocumentSync == nil {
		t.Fatal("TextDocumentSync capability is nil")
	}
	if result.Capabilities.CodeActionProvider == nil {
		t.Fatal("CodeActionProvider capability is nil")
	}
}

func TestServer_DidOpenDidChangeDidClose(t *testing.T) {
	gen := newGenerator(t)
	defer gen.Close()
	_, cancel := context.WithCancel(context.Background())
	defer cancel()
	srv := lsp.NewServer(gen, cancel)

	docURI := uri.URI("file:///test.go")
	content := `package test

func hello() string {
	return "hello"
}
`

	err := srv.DidOpen(context.Background(), &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:        docURI,
			LanguageID: "go",
			Version:    1,
			Text:       content,
		},
	})
	if err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	t.Run("code_action_on_function", func(t *testing.T) {
		actions, err := srv.CodeAction(context.Background(), &protocol.CodeActionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: docURI},
			Range: protocol.Range{
				Start: protocol.Position{Line: 2, Character: 0},
				End:   protocol.Position{Line: 2, Character: 10},
			},
			Context: protocol.CodeActionContext{},
		})
		if err != nil {
			t.Fatalf("CodeAction failed: %v", err)
		}
		if len(actions) == 0 {
			t.Fatal("expected at least one code action for function")
		}
	})

	t.Run("code_action_outside_symbol", func(t *testing.T) {
		actions, err := srv.CodeAction(context.Background(), &protocol.CodeActionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: docURI},
			Range: protocol.Range{
				Start: protocol.Position{Line: 0, Character: 0},
				End:   protocol.Position{Line: 0, Character: 1},
			},
			Context: protocol.CodeActionContext{},
		})
		if err != nil {
			t.Fatalf("CodeAction failed: %v", err)
		}
		if len(actions) != 0 {
			t.Errorf("expected no code action for package line, got %d", len(actions))
		}
	})

	t.Run("code_action_no_document", func(t *testing.T) {
		unknownURI := uri.URI("file:///unknown.go")
		actions, err := srv.CodeAction(context.Background(), &protocol.CodeActionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: unknownURI},
			Range: protocol.Range{
				Start: protocol.Position{Line: 0, Character: 0},
				End:   protocol.Position{Line: 0, Character: 1},
			},
			Context: protocol.CodeActionContext{},
		})
		if err != nil {
			t.Fatalf("CodeAction failed: %v", err)
		}
		if len(actions) != 0 {
			t.Errorf("expected no code action for unknown document, got %d", len(actions))
		}
	})

	t.Run("did_change_updates_content", func(t *testing.T) {
		newContent := `package test

// hello prints hello
func hello() string {
	return "hello"
}
`
		err := srv.DidChange(context.Background(), &protocol.DidChangeTextDocumentParams{
			TextDocument: protocol.VersionedTextDocumentIdentifier{
				TextDocumentIdentifier: protocol.TextDocumentIdentifier{URI: docURI},
				Version:               2,
			},
			ContentChanges: []protocol.TextDocumentContentChangeEvent{
				&protocol.TextDocumentContentChangeWholeDocument{
					Text: newContent,
				},
			},
		})
		if err != nil {
			t.Fatalf("DidChange failed: %v", err)
		}

		actions, err := srv.CodeAction(context.Background(), &protocol.CodeActionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: docURI},
			Range: protocol.Range{
				Start: protocol.Position{Line: 3, Character: 0},
				End:   protocol.Position{Line: 3, Character: 10},
			},
			Context: protocol.CodeActionContext{},
		})
		if err != nil {
			t.Fatalf("CodeAction after DidChange failed: %v", err)
		}
		if len(actions) != 0 {
			t.Errorf("expected no code action for symbol with existing comment, got %d", len(actions))
		}
	})

	t.Run("did_close_removes_content", func(t *testing.T) {
		err := srv.DidClose(context.Background(), &protocol.DidCloseTextDocumentParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: docURI},
		})
		if err != nil {
			t.Fatalf("DidClose failed: %v", err)
		}

		actions, err := srv.CodeAction(context.Background(), &protocol.CodeActionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: docURI},
			Range: protocol.Range{
				Start: protocol.Position{Line: 2, Character: 0},
				End:   protocol.Position{Line: 2, Character: 10},
			},
			Context: protocol.CodeActionContext{},
		})
		if err != nil {
			t.Fatalf("CodeAction after DidClose failed: %v", err)
		}
		if len(actions) != 0 {
			t.Errorf("expected no code action after close, got %d", len(actions))
		}
	})
}

func TestServer_CodeAction_MultipleLanguages(t *testing.T) {
	gen := newGenerator(t)
	defer gen.Close()
	_, cancel := context.WithCancel(context.Background())
	defer cancel()
	srv := lsp.NewServer(gen, cancel)

	tests := []struct {
		name    string
		uri     uri.URI
		content string
		cursor  protocol.Position
		want    bool
	}{
		{
			name: "go function",
			uri:  "file:///test.go",
			content: `package test

func add(a int, b int) int {
	return a + b
}
`,
			cursor: protocol.Position{Line: 2, Character: 0},
			want:   true,
		},
		{
			name: "python function",
			uri:  "file:///test.py",
			content: `def greet(name):
    return f"Hello, {name}"
`,
			cursor: protocol.Position{Line: 0, Character: 0},
			want:   true,
		},
		{
			name: "c function",
			uri:  "file:///test.c",
			content: `int multiply(int x, int y) {
    return x * y;
}
`,
			cursor: protocol.Position{Line: 0, Character: 0},
			want:   true,
		},
		{
			name: "cpp function",
			uri:  "file:///test.cpp",
			content: `int divide(int a, int b) {
    return a / b;
}
`,
			cursor: protocol.Position{Line: 0, Character: 0},
			want:   true,
		},
		{
			name: "rust function",
			uri:  "file:///test.rs",
			content: `fn compute(x: i32) -> i32 {
    x * 2
}
`,
			cursor: protocol.Position{Line: 0, Character: 0},
			want:   true,
		},
		{
			name: "unsupported language (javascript)",
			uri:  "file:///test.js",
			content: `function hello() {
    return "hello";
}
`,
			cursor: protocol.Position{Line: 0, Character: 0},
			want:   false,
		},
		{
			name:    "empty document",
			uri:     "file:///test.go",
			content: ``,
			cursor:  protocol.Position{Line: 0, Character: 0},
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv.DidOpen(context.Background(), &protocol.DidOpenTextDocumentParams{
				TextDocument: protocol.TextDocumentItem{
					URI:        tt.uri,
					LanguageID: "unknown",
					Version:    1,
					Text:       tt.content,
				},
			})
			defer srv.DidClose(context.Background(), &protocol.DidCloseTextDocumentParams{
				TextDocument: protocol.TextDocumentIdentifier{URI: tt.uri},
			})

			actions, err := srv.CodeAction(context.Background(), &protocol.CodeActionParams{
				TextDocument: protocol.TextDocumentIdentifier{URI: tt.uri},
				Range: protocol.Range{
					Start: tt.cursor,
					End:   tt.cursor,
				},
				Context: protocol.CodeActionContext{},
			})
			if err != nil {
				t.Fatalf("CodeAction failed: %v", err)
			}

			got := len(actions) > 0
			if got != tt.want {
				t.Errorf("CodeAction presence = %v, want %v (got %d actions)", got, tt.want, len(actions))
			}
		})
	}
}

func TestServer_Shutdown(t *testing.T) {
	gen := newGenerator(t)
	defer gen.Close()
	_, cancel := context.WithCancel(context.Background())
	defer cancel()
	srv := lsp.NewServer(gen, cancel)

	if err := srv.Shutdown(context.Background()); err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}
	if err := srv.Exit(context.Background()); err != nil {
		t.Errorf("Exit failed: %v", err)
	}
}

func newGenerator(t *testing.T) *generator.Generator {
	t.Helper()

	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, ".deoxy.yaml")
	cfgContent := `version: "1"
default_style: godoc
languages:
  go:
    docstyle: godoc
  python:
    docstyle: pydoc
  c:
    docstyle: doxygen
  cpp:
    docstyle: doxygen
  rust:
    docstyle: rustdoc
`
	if err := os.WriteFile(cfgPath, []byte(cfgContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	gen, err := generator.New(cfg)
	if err != nil {
		t.Fatalf("generator.New failed: %v", err)
	}
	return gen
}
