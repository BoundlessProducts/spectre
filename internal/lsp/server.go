package lsp

// server.go — LSP server loop: reads JSON-RPC messages from r, dispatches,
// writes responses to w.  Designed to be run on stdin/stdout.

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Server holds the LSP server state.
type Server struct {
	r          io.Reader
	w          io.Writer
	docs       map[string]string // uri → content
	shutdown   bool
}

// NewServer creates a server reading from r and writing to w.
func NewServer(r io.Reader, w io.Writer) *Server {
	return &Server{
		r:    r,
		w:    w,
		docs: map[string]string{},
	}
}

// Run starts the server loop and blocks until exit or an I/O error.
func (s *Server) Run() error {
	br := bufio.NewReader(s.r)
	for {
		msg, err := readMessage(br)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if stop := s.dispatch(msg); stop {
			return nil
		}
	}
}

// dispatch handles one message and returns true if the server should exit.
func (s *Server) dispatch(msg *Message) bool {
	// Notifications (no ID) and requests (have ID).
	switch msg.Method {
	case "initialize":
		s.handleInitialize(msg)
	case "initialized":
		// no-op
	case "shutdown":
		s.shutdown = true
		s.sendResult(msg.ID, nil)
	case "exit":
		return true
	case "textDocument/didOpen":
		var p DidOpenTextDocumentParams
		if json.Unmarshal(msg.Params, &p) == nil {
			s.docs[p.TextDocument.URI] = p.TextDocument.Text
			s.publishDiagnostics(p.TextDocument.URI, p.TextDocument.Text)
		}
	case "textDocument/didChange":
		var p DidChangeTextDocumentParams
		if json.Unmarshal(msg.Params, &p) == nil && len(p.ContentChanges) > 0 {
			text := p.ContentChanges[len(p.ContentChanges)-1].Text
			s.docs[p.TextDocument.URI] = text
			s.publishDiagnostics(p.TextDocument.URI, text)
		}
	case "textDocument/didClose":
		var p DidCloseTextDocumentParams
		if json.Unmarshal(msg.Params, &p) == nil {
			delete(s.docs, p.TextDocument.URI)
			// Clear diagnostics.
			s.sendNotification("textDocument/publishDiagnostics", PublishDiagnosticsParams{
				URI:         p.TextDocument.URI,
				Diagnostics: []Diagnostic{},
			})
		}
	case "textDocument/hover":
		var p TextDocumentPositionParams
		if json.Unmarshal(msg.Params, &p) == nil {
			s.handleHover(msg.ID, p)
		} else {
			s.sendResult(msg.ID, nil)
		}
	case "textDocument/completion":
		var p CompletionParams
		if json.Unmarshal(msg.Params, &p) == nil {
			s.handleCompletion(msg.ID, p)
		} else {
			s.sendResult(msg.ID, []CompletionItem{})
		}
	case "textDocument/definition":
		var p TextDocumentPositionParams
		if json.Unmarshal(msg.Params, &p) == nil {
			s.handleDefinition(msg.ID, p)
		} else {
			s.sendResult(msg.ID, nil)
		}
	default:
		if msg.ID != nil {
			s.sendError(msg.ID, ErrMethodNotFound, fmt.Sprintf("method not found: %s", msg.Method))
		}
	}
	return false
}

// ── handlers ──────────────────────────────────────────────────────────────────

func (s *Server) handleInitialize(msg *Message) {
	s.sendResult(msg.ID, InitializeResult{
		Capabilities: ServerCapabilities{
			TextDocumentSync:   1, // Full sync
			HoverProvider:      true,
			DefinitionProvider: true,
			CompletionProvider: &CompletionOptions{
				TriggerCharacters: []string{".", " "},
			},
		},
		ServerInfo: ServerInfo{Name: "spectre-lsp", Version: "0.1.0"},
	})
}

func (s *Server) publishDiagnostics(uri, content string) {
	result := ParseDocument(uri, content)
	s.sendNotification("textDocument/publishDiagnostics", PublishDiagnosticsParams{
		URI:         uri,
		Diagnostics: result.Errors,
	})
}

func (s *Server) handleHover(id json.RawMessage, p TextDocumentPositionParams) {
	content, ok := s.docs[p.TextDocument.URI]
	if !ok {
		s.sendResult(id, nil)
		return
	}
	result := ParseDocument(p.TextDocument.URI, content)
	hover := HoverInfo(result.File, content, p.Position)
	s.sendResult(id, hover)
}

func (s *Server) handleCompletion(id json.RawMessage, p CompletionParams) {
	content, ok := s.docs[p.TextDocument.URI]
	if !ok {
		s.sendResult(id, []CompletionItem{})
		return
	}
	result := ParseDocument(p.TextDocument.URI, content)
	items := CompletionItems(result.File)
	s.sendResult(id, items)
}

func (s *Server) handleDefinition(id json.RawMessage, p TextDocumentPositionParams) {
	content, ok := s.docs[p.TextDocument.URI]
	if !ok {
		s.sendResult(id, nil)
		return
	}
	result := ParseDocument(p.TextDocument.URI, content)
	loc := DefinitionLocation(result.File, p.TextDocument.URI, content, p.Position)
	s.sendResult(id, loc)
}

// ── transport ─────────────────────────────────────────────────────────────────

func readMessage(r *bufio.Reader) (*Message, error) {
	contentLen := -1
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break
		}
		if strings.HasPrefix(line, "Content-Length: ") {
			n, parseErr := strconv.Atoi(strings.TrimPrefix(line, "Content-Length: "))
			if parseErr == nil {
				contentLen = n
			}
		}
	}
	if contentLen <= 0 {
		return nil, fmt.Errorf("missing or invalid Content-Length")
	}
	body := make([]byte, contentLen)
	if _, err := io.ReadFull(r, body); err != nil {
		return nil, err
	}
	var msg Message
	if err := json.Unmarshal(body, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

func (s *Server) writeMessage(v interface{}) {
	body, _ := json.Marshal(v)
	fmt.Fprintf(s.w, "Content-Length: %d\r\n\r\n", len(body))
	s.w.Write(body) //nolint
}

func (s *Server) sendResult(id json.RawMessage, result interface{}) {
	s.writeMessage(Message{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	})
}

func (s *Server) sendError(id json.RawMessage, code int, message string) {
	s.writeMessage(Message{
		JSONRPC: "2.0",
		ID:      id,
		Error:   &RPCError{Code: code, Message: message},
	})
}

func (s *Server) sendNotification(method string, params interface{}) {
	body, _ := json.Marshal(params)
	s.writeMessage(Message{
		JSONRPC: "2.0",
		Method:  method,
		Params:  body,
	})
}
