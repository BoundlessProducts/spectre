package lsp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

// encode builds a Content-Length framed JSON-RPC message.
func encode(t *testing.T, msg interface{}) string {
	t.Helper()
	body, _ := json.Marshal(msg)
	return fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(body), body)
}

// decodeAll reads all complete JSON-RPC responses from output.
func decodeAll(t *testing.T, output string) []Message {
	t.Helper()
	var msgs []Message
	for len(output) > 0 {
		const prefix = "Content-Length: "
		idx := strings.Index(output, prefix)
		if idx < 0 {
			break
		}
		rest := output[idx+len(prefix):]
		nl := strings.Index(rest, "\r\n")
		if nl < 0 {
			break
		}
		var n int
		fmt.Sscanf(rest[:nl], "%d", &n)
		sep := strings.Index(rest, "\r\n\r\n")
		if sep < 0 {
			break
		}
		body := rest[sep+4 : sep+4+n]
		var m Message
		if err := json.Unmarshal([]byte(body), &m); err == nil {
			msgs = append(msgs, m)
		}
		output = rest[sep+4+n:]
	}
	return msgs
}

// runServer feeds input messages to a Server and returns all output messages.
func runServer(t *testing.T, inputs ...interface{}) []Message {
	t.Helper()
	var sb strings.Builder
	for _, in := range inputs {
		sb.WriteString(encode(t, in))
	}
	var out bytes.Buffer
	srv := NewServer(strings.NewReader(sb.String()), &out)
	srv.Run() //nolint
	return decodeAll(t, out.String())
}

func rpcRequest(id int, method string, params interface{}) map[string]interface{} {
	return map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"method":  method,
		"params":  params,
	}
}

func rpcNotification(method string, params interface{}) map[string]interface{} {
	return map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  method,
		"params":  params,
	}
}

// ── initialize ────────────────────────────────────────────────────────────────

func TestServer_Initialize(t *testing.T) {
	msgs := runServer(t,
		rpcRequest(1, "initialize", map[string]interface{}{}),
		rpcNotification("exit", nil),
	)
	if len(msgs) == 0 {
		t.Fatal("expected response to initialize")
	}
	if msgs[0].Error != nil {
		t.Fatalf("initialize error: %v", msgs[0].Error)
	}
	// Result should have capabilities.
	var result InitializeResult
	raw, _ := json.Marshal(msgs[0].Result)
	json.Unmarshal(raw, &result)
	if !result.Capabilities.HoverProvider {
		t.Error("expected HoverProvider = true")
	}
	if result.ServerInfo.Name != "spectre-lsp" {
		t.Errorf("unexpected server name: %q", result.ServerInfo.Name)
	}
}

// ── textDocument/didOpen → publishDiagnostics ─────────────────────────────────

func TestServer_DidOpen_ValidSpec_NoDiagnostics(t *testing.T) {
	msgs := runServer(t,
		rpcRequest(1, "initialize", map[string]interface{}{}),
		rpcNotification("initialized", nil),
		rpcNotification("textDocument/didOpen", DidOpenTextDocumentParams{
			TextDocument: TextDocumentItem{
				URI:  "file:///a.spec",
				Text: "var x: int\ninit { x = 0 }\n",
			},
		}),
		rpcNotification("exit", nil),
	)
	for _, m := range msgs {
		if m.Method == "textDocument/publishDiagnostics" {
			var p PublishDiagnosticsParams
			raw, _ := json.Marshal(m.Params)
			json.Unmarshal(raw, &p)
			if len(p.Diagnostics) != 0 {
				t.Errorf("expected no diagnostics for valid spec, got %v", p.Diagnostics)
			}
			return
		}
	}
	t.Error("expected publishDiagnostics notification")
}

func TestServer_DidOpen_InvalidSpec_HasDiagnostics(t *testing.T) {
	msgs := runServer(t,
		rpcRequest(1, "initialize", map[string]interface{}{}),
		// Unclosed block produces a real parse error.
		rpcNotification("textDocument/didOpen", DidOpenTextDocumentParams{
			TextDocument: TextDocumentItem{
				URI:  "file:///bad.spec",
				Text: "action incr {\n",
			},
		}),
		rpcNotification("exit", nil),
	)
	for _, m := range msgs {
		if m.Method == "textDocument/publishDiagnostics" {
			var p PublishDiagnosticsParams
			raw, _ := json.Marshal(m.Params)
			json.Unmarshal(raw, &p)
			if len(p.Diagnostics) == 0 {
				t.Error("expected at least one diagnostic for invalid spec")
			}
			return
		}
	}
	t.Error("expected publishDiagnostics notification")
}

// ── textDocument/didChange ────────────────────────────────────────────────────

func TestServer_DidChange_UpdatesDiagnostics(t *testing.T) {
	msgs := runServer(t,
		rpcRequest(1, "initialize", map[string]interface{}{}),
		rpcNotification("textDocument/didOpen", DidOpenTextDocumentParams{
			TextDocument: TextDocumentItem{URI: "file:///c.spec", Text: "var x: int\n"},
		}),
		rpcNotification("textDocument/didChange", DidChangeTextDocumentParams{
			TextDocument: VersionedTextDocumentIdentifier{URI: "file:///c.spec", Version: 2},
			ContentChanges: []TextDocumentContentChangeEvent{
				{Text: "var x: int\ninit { x = 0 }\n"},
			},
		}),
		rpcNotification("exit", nil),
	)
	publishCount := 0
	for _, m := range msgs {
		if m.Method == "textDocument/publishDiagnostics" {
			publishCount++
		}
	}
	if publishCount < 2 {
		t.Errorf("expected ≥2 publishDiagnostics (open + change), got %d", publishCount)
	}
}

// ── textDocument/hover ────────────────────────────────────────────────────────

func TestServer_Hover_KnownSymbol(t *testing.T) {
	content := "var counter: int\ninit { counter = 0 }\n"
	msgs := runServer(t,
		rpcRequest(1, "initialize", map[string]interface{}{}),
		rpcNotification("textDocument/didOpen", DidOpenTextDocumentParams{
			TextDocument: TextDocumentItem{URI: "file:///h.spec", Text: content},
		}),
		rpcRequest(2, "textDocument/hover", TextDocumentPositionParams{
			TextDocument: TextDocumentIdentifier{URI: "file:///h.spec"},
			Position:     Position{Line: 0, Character: 4}, // "counter"
		}),
		rpcNotification("exit", nil),
	)
	for _, m := range msgs {
		if m.Method == "" && m.ID != nil {
			var id int
			json.Unmarshal(m.ID, &id)
			if id == 2 {
				if m.Error != nil {
					t.Fatalf("hover error: %v", m.Error)
				}
				// Result may be nil (if hover finds nothing) or a Hover object.
				return
			}
		}
	}
	t.Error("expected response to hover request")
}

// ── textDocument/completion ───────────────────────────────────────────────────

func TestServer_Completion(t *testing.T) {
	content := "var x: int\naction incr { x' = x + 1 }\n"
	msgs := runServer(t,
		rpcRequest(1, "initialize", map[string]interface{}{}),
		rpcNotification("textDocument/didOpen", DidOpenTextDocumentParams{
			TextDocument: TextDocumentItem{URI: "file:///cp.spec", Text: content},
		}),
		rpcRequest(2, "textDocument/completion", CompletionParams{
			TextDocument: TextDocumentIdentifier{URI: "file:///cp.spec"},
			Position:     Position{Line: 1, Character: 0},
		}),
		rpcNotification("exit", nil),
	)
	for _, m := range msgs {
		if m.Method == "" && m.ID != nil {
			var id int
			json.Unmarshal(m.ID, &id)
			if id == 2 {
				if m.Error != nil {
					t.Fatalf("completion error: %v", m.Error)
				}
				var items []CompletionItem
				raw, _ := json.Marshal(m.Result)
				json.Unmarshal(raw, &items)
				labels := map[string]bool{}
				for _, it := range items {
					labels[it.Label] = true
				}
				if !labels["x"] || !labels["incr"] {
					t.Errorf("expected x and incr in completions, got %v", labels)
				}
				return
			}
		}
	}
	t.Error("expected completion response")
}

// ── textDocument/definition ───────────────────────────────────────────────────

func TestServer_Definition(t *testing.T) {
	content := "var x: int\ninit { x = 0 }\n"
	msgs := runServer(t,
		rpcRequest(1, "initialize", map[string]interface{}{}),
		rpcNotification("textDocument/didOpen", DidOpenTextDocumentParams{
			TextDocument: TextDocumentItem{URI: "file:///d.spec", Text: content},
		}),
		rpcRequest(2, "textDocument/definition", TextDocumentPositionParams{
			TextDocument: TextDocumentIdentifier{URI: "file:///d.spec"},
			Position:     Position{Line: 0, Character: 4}, // "x"
		}),
		rpcNotification("exit", nil),
	)
	for _, m := range msgs {
		if m.Method == "" && m.ID != nil {
			var id int
			json.Unmarshal(m.ID, &id)
			if id == 2 {
				if m.Error != nil {
					t.Fatalf("definition error: %v", m.Error)
				}
				return
			}
		}
	}
	t.Error("expected definition response")
}

// ── shutdown / exit ───────────────────────────────────────────────────────────

func TestServer_ShutdownExit(t *testing.T) {
	msgs := runServer(t,
		rpcRequest(1, "initialize", map[string]interface{}{}),
		rpcRequest(2, "shutdown", nil),
		rpcNotification("exit", nil),
	)
	// Should have a response for shutdown (id=2).
	for _, m := range msgs {
		if m.ID != nil {
			var id int
			json.Unmarshal(m.ID, &id)
			if id == 2 && m.Error == nil {
				return
			}
		}
	}
	t.Error("expected successful shutdown response")
}

// ── unknown method ────────────────────────────────────────────────────────────

func TestServer_UnknownMethod(t *testing.T) {
	msgs := runServer(t,
		rpcRequest(1, "initialize", map[string]interface{}{}),
		rpcRequest(2, "unknownMethod/test", nil),
		rpcNotification("exit", nil),
	)
	for _, m := range msgs {
		if m.ID != nil {
			var id int
			json.Unmarshal(m.ID, &id)
			if id == 2 {
				if m.Error == nil {
					t.Error("expected error for unknown method")
				} else if m.Error.Code != ErrMethodNotFound {
					t.Errorf("expected MethodNotFound error, got %d", m.Error.Code)
				}
				return
			}
		}
	}
	t.Error("expected error response for unknown method")
}

// ── textDocument/didClose ─────────────────────────────────────────────────────

func TestServer_DidClose_ClearsDiagnostics(t *testing.T) {
	msgs := runServer(t,
		rpcRequest(1, "initialize", map[string]interface{}{}),
		rpcNotification("textDocument/didOpen", DidOpenTextDocumentParams{
			TextDocument: TextDocumentItem{URI: "file:///close.spec", Text: "var x: int\n"},
		}),
		rpcNotification("textDocument/didClose", DidCloseTextDocumentParams{
			TextDocument: TextDocumentIdentifier{URI: "file:///close.spec"},
		}),
		rpcNotification("exit", nil),
	)
	for _, m := range msgs {
		if m.Method == "textDocument/publishDiagnostics" {
			var p PublishDiagnosticsParams
			raw, _ := json.Marshal(m.Params)
			json.Unmarshal(raw, &p)
			if p.URI == "file:///close.spec" && len(p.Diagnostics) == 0 {
				return
			}
		}
	}
	t.Error("expected empty diagnostics after didClose")
}
