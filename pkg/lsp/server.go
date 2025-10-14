package lsp

import (
	"context"
	"fmt"
	"sync"

	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"
)

type Server struct {
	conn    jsonrpc2.Conn
	cache   map[protocol.DocumentURI]*DocumentState
	cacheMu sync.RWMutex
}

type DocumentState struct {
	URI     protocol.DocumentURI
	Content string
	Version int32
}

func NewServer(conn jsonrpc2.Conn) *Server {
	return &Server{
		conn:  conn,
		cache: make(map[protocol.DocumentURI]*DocumentState),
	}
}

func (s *Server) Initialize(ctx context.Context, params *protocol.InitializeParams) (*protocol.InitializeResult, error) {
	return &protocol.InitializeResult{
		Capabilities: protocol.ServerCapabilities{
			TextDocumentSync: protocol.TextDocumentSyncOptions{
				OpenClose: true,
				Change:    protocol.TextDocumentSyncKindFull,
			},
			CompletionProvider: &protocol.CompletionOptions{
				TriggerCharacters: []string{"{", ":", " "},
			},
			HoverProvider: true,
		},
		ServerInfo: &protocol.ServerInfo{
			Name:    "gxc-language-server",
			Version: "0.1.0",
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
	return nil
}

func (s *Server) DidOpen(ctx context.Context, params *protocol.DidOpenTextDocumentParams) error {
	s.cacheMu.Lock()
	s.cache[params.TextDocument.URI] = &DocumentState{
		URI:     params.TextDocument.URI,
		Content: params.TextDocument.Text,
		Version: params.TextDocument.Version,
	}
	s.cacheMu.Unlock()

	go s.publishDiagnostics(ctx, params.TextDocument.URI, params.TextDocument.Text)

	return nil
}

func (s *Server) DidChange(ctx context.Context, params *protocol.DidChangeTextDocumentParams) error {
	if len(params.ContentChanges) == 0 {
		return nil
	}

	change := params.ContentChanges[0]
	newContent := change.Text

	s.cacheMu.Lock()
	if state, ok := s.cache[params.TextDocument.URI]; ok {
		state.Content = newContent
		state.Version = params.TextDocument.Version
	}
	s.cacheMu.Unlock()

	go s.publishDiagnostics(ctx, params.TextDocument.URI, newContent)

	return nil
}

func (s *Server) DidClose(ctx context.Context, params *protocol.DidCloseTextDocumentParams) error {
	s.cacheMu.Lock()
	delete(s.cache, params.TextDocument.URI)
	s.cacheMu.Unlock()

	return nil
}

func (s *Server) DidSave(ctx context.Context, params *protocol.DidSaveTextDocumentParams) error {
	return nil
}

func (s *Server) publishDiagnostics(ctx context.Context, uri protocol.DocumentURI, content string) {
	diagnostics := s.analyze(content)

	err := s.conn.Notify(ctx, "textDocument/publishDiagnostics", &protocol.PublishDiagnosticsParams{
		URI:         uri,
		Diagnostics: diagnostics,
	})

	if err != nil {
		fmt.Printf("Error publishing diagnostics: %v\n", err)
	}
}

func (s *Server) Completion(ctx context.Context, params *protocol.CompletionParams) (*protocol.CompletionList, error) {
	s.cacheMu.RLock()
	state, ok := s.cache[params.TextDocument.URI]
	s.cacheMu.RUnlock()

	if !ok {
		return &protocol.CompletionList{Items: []protocol.CompletionItem{}}, nil
	}

	items := s.getCompletions(state.Content, params.Position)

	return &protocol.CompletionList{
		IsIncomplete: false,
		Items:        items,
	}, nil
}

func (s *Server) Hover(ctx context.Context, params *protocol.HoverParams) (*protocol.Hover, error) {
	s.cacheMu.RLock()
	state, ok := s.cache[params.TextDocument.URI]
	s.cacheMu.RUnlock()

	if !ok {
		return nil, nil
	}

	hover := s.getHover(state.Content, params.Position)
	return hover, nil
}
