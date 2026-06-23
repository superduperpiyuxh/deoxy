package lsp

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"sync"

	"github.com/superduperpiyuxh/deoxy/internal/generator"
	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

const maxDocumentSize = 100 * 1024 * 1024

type documentState struct {
	content []byte
	version int32
}

type Server struct {
	protocol.UnimplementedServer

	gen       *generator.Generator
	documents sync.Map
	cancel    context.CancelFunc
}

func NewServer(gen *generator.Generator, cancel context.CancelFunc) *Server {
	return &Server{gen: gen, cancel: cancel}
}

func (s *Server) Initialize(ctx context.Context, params *protocol.InitializeParams) (*protocol.InitializeResult, error) {
	return &protocol.InitializeResult{
		Capabilities: protocol.ServerCapabilities{
			TextDocumentSync: protocol.TextDocumentSyncKindFull,
			CodeActionProvider: &protocol.CodeActionOptions{
				CodeActionKinds: []protocol.CodeActionKind{protocol.CodeActionKindQuickFix},
			},
		},
	}, nil
}

func (s *Server) Initialized(ctx context.Context, params *protocol.InitializedParams) error {
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	return nil
}

func (s *Server) Exit(ctx context.Context) error {
	if s.cancel != nil {
		s.cancel()
	}
	return nil
}

func (s *Server) SetTrace(ctx context.Context, params *protocol.SetTraceParams) error {
	return nil
}

func (s *Server) Progress(ctx context.Context, params *protocol.ProgressParams) error {
	return nil
}

func (s *Server) WorkDoneProgressCancel(ctx context.Context, params *protocol.WorkDoneProgressCancelParams) error {
	return nil
}

func (s *Server) DidOpen(ctx context.Context, params *protocol.DidOpenTextDocumentParams) error {
	if len(params.TextDocument.Text) > maxDocumentSize {
		log.Printf("lsp: document too large (%d bytes), rejecting", len(params.TextDocument.Text))
		return nil
	}
	if !isFileURI(params.TextDocument.URI) {
		log.Printf("lsp: ignoring non-file URI %q", params.TextDocument.URI)
		return nil
	}
	s.documents.Store(params.TextDocument.URI, &documentState{
		content: []byte(params.TextDocument.Text),
		version: params.TextDocument.Version,
	})
	return nil
}

func (s *Server) DidChange(ctx context.Context, params *protocol.DidChangeTextDocumentParams) error {
	if len(params.ContentChanges) == 0 {
		return nil
	}
	if !isFileURI(params.TextDocument.URI) {
		return nil
	}

	raw, ok := s.documents.Load(params.TextDocument.URI)
	if !ok {
		return nil
	}
	state, ok := raw.(*documentState)
	if !ok {
		return nil
	}

	if state.version > params.TextDocument.Version {
		return nil
	}

	text := textFromChangeEvent(params.ContentChanges[len(params.ContentChanges)-1])
	if len(text) > maxDocumentSize {
		log.Printf("lsp: updated document too large (%d bytes), rejecting", len(text))
		return nil
	}

	state.content = []byte(text)
	state.version = params.TextDocument.Version
	return nil
}

func (s *Server) WillSave(ctx context.Context, params *protocol.WillSaveTextDocumentParams) error {
	return nil
}

func (s *Server) WillSaveWaitUntil(ctx context.Context, params *protocol.WillSaveTextDocumentParams) ([]protocol.TextEdit, error) {
	return nil, nil
}

func (s *Server) DidSave(ctx context.Context, params *protocol.DidSaveTextDocumentParams) error {
	return nil
}

func (s *Server) DidClose(ctx context.Context, params *protocol.DidCloseTextDocumentParams) error {
	s.documents.Delete(params.TextDocument.URI)
	return nil
}

func (s *Server) CodeAction(ctx context.Context, params *protocol.CodeActionParams) ([]protocol.CommandOrCodeAction, error) {
	docURI := params.TextDocument.URI
	if !isFileURI(docURI) {
		return nil, nil
	}

	raw, ok := s.documents.Load(docURI)
	if !ok {
		return nil, nil
	}

	state, ok := raw.(*documentState)
	if !ok {
		log.Println("lsp: stored document state has unexpected type")
		return nil, nil
	}

	path, err := filePathFromURI(docURI)
	if err != nil {
		log.Printf("lsp: invalid URI %q: %v", docURI, err)
		return nil, nil
	}

	lang := detectLanguage(path)
	if lang == "" {
		return nil, nil
	}

	cursorLine := int(params.Range.Start.Line)

	comments, err := s.gen.ProcessContent(path, state.content, lang)
	if err != nil {
		log.Printf("lsp: ProcessContent(%q): %v", path, err)
		return nil, nil
	}
	if len(comments) == 0 {
		return nil, nil
	}

	var target *generator.SymbolComment
	for i, c := range comments {
		if cursorLine >= c.StartLine && cursorLine <= c.EndLine {
			target = &comments[i]
			break
		}
	}
	if target == nil {
		return nil, nil
	}

	if hasExistingComment(state.content, target.StartLine) {
		return nil, nil
	}

	kind := protocol.CodeActionKindQuickFix
	preferred := true

	return []protocol.CommandOrCodeAction{
		&protocol.CodeAction{
			Title: fmt.Sprintf("Generate doc comment for %s", target.Symbol.Name),
			Kind:  &kind,
			Edit: &protocol.WorkspaceEdit{
				Changes: map[uri.URI][]protocol.TextEdit{
					docURI: {
						{
							Range: protocol.Range{
								Start: protocol.Position{Line: uint32(target.StartLine), Character: 0},
								End:   protocol.Position{Line: uint32(target.StartLine), Character: 0},
							},
							NewText: target.DocComment + "\n",
						},
					},
				},
			},
			IsPreferred: &preferred,
		},
	}, nil
}

func isFileURI(docURI uri.URI) bool {
	s := string(docURI)
	return strings.HasPrefix(s, "file://") || strings.HasPrefix(s, "file:")
}

func filePathFromURI(docURI uri.URI) (string, error) {
	path := docURI.FsPath()
	if path != "" {
		return path, nil
	}
	path = docURI.Path()
	if path != "" {
		return path, nil
	}
	return "", fmt.Errorf("cannot extract file path from %q", docURI)
}

func textFromChangeEvent(event protocol.TextDocumentContentChangeEvent) string {
	switch e := event.(type) {
	case *protocol.TextDocumentContentChangeWholeDocument:
		return e.Text
	case *protocol.TextDocumentContentChangePartial:
		return e.Text
	default:
		log.Printf("lsp: unknown change event type %T", event)
		return ""
	}
}

func detectLanguage(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".go":
		return "go"
	case ".py":
		return "python"
	case ".c", ".h":
		return "c"
	case ".cpp", ".cc", ".cxx", ".hpp":
		return "cpp"
	case ".rs":
		return "rust"
	default:
		return ""
	}
}

func hasExistingComment(content []byte, startLine int) bool {
	if len(content) == 0 {
		return false
	}

	lines := strings.Split(string(content), "\n")
	if startLine <= 0 || startLine >= len(lines) {
		return false
	}

	for line := startLine - 1; line >= 0; line-- {
		trimmed := strings.TrimSpace(lines[line])
		if trimmed == "" {
			continue
		}
		if isCommentLine(trimmed) {
			return true
		}
		return false
	}
	return false
}

func isCommentLine(line string) bool {
	if len(line) == 0 {
		return false
	}
	if isBuildDirective(line) {
		return false
	}
	return strings.HasPrefix(line, "//") ||
		strings.HasPrefix(line, "#") ||
		strings.HasPrefix(line, "\"\"\"") ||
		strings.HasPrefix(line, "'''") ||
		strings.HasPrefix(line, "/*") ||
		strings.HasPrefix(line, "*") ||
		strings.HasPrefix(line, "*/")
}

func isBuildDirective(line string) bool {
	trimmed := strings.TrimSpace(line)
	return strings.HasPrefix(trimmed, "//go:") ||
		strings.HasPrefix(trimmed, "// +") ||
		strings.HasPrefix(trimmed, "//export")
}
