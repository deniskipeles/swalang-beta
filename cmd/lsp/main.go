package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/deniskipeles/pylearn/internal/lexer"
	"github.com/deniskipeles/pylearn/internal/parser"
)

// LSP JSON-RPC Base Message
type Message struct {
	RPC    string          `json:"jsonrpc"`
	ID     *int            `json:"id,omitempty"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params,omitempty"`
}

// Params for textDocument/didOpen and textDocument/didChange
type DidOpenTextDocumentParams struct {
	TextDocument struct {
		URI  string `json:"uri"`
		Text string `json:"text"`
	} `json:"textDocument"`
}

type DidChangeTextDocumentParams struct {
	TextDocument struct {
		URI string `json:"uri"`
	} `json:"textDocument"`
	ContentChanges []struct {
		Text string `json:"text"`
	} `json:"contentChanges"`
}

// Notification payload to send Diagnostics (Parser Errors)
type PublishDiagnosticsParams struct {
	URI         string       `json:"uri"`
	Diagnostics []Diagnostic `json:"diagnostics"`
}

type Diagnostic struct {
	Range    Range  `json:"range"`
	Severity int    `json:"severity"`
	Source   string `json:"source"`
	Message  string `json:"message"`
}

type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

func main() {
	reader := bufio.NewReader(os.Stdin)

	for {
		// 1. Read LSP Header (Content-Length: X)
		var length int
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					return // IDE closed the connection
				}
				panic(err)
			}
			line = strings.TrimSpace(line)
			if line == "" {
				break // End of headers
			}
			if strings.HasPrefix(line, "Content-Length:") {
				parts := strings.Split(line, ":")
				if len(parts) == 2 {
					length, _ = strconv.Atoi(strings.TrimSpace(parts[1]))
				}
			}
		}

		if length == 0 {
			continue
		}

		// 2. Read the JSON Payload
		buf := make([]byte, length)
		_, err := io.ReadFull(reader, buf)
		if err != nil {
			panic(err)
		}

		var msg Message
		if err := json.Unmarshal(buf, &msg); err != nil {
			continue
		}

		// 3. Route the message
		switch msg.Method {
		case "initialize":
			// Acknowledge initialization
			sendResponse(msg.ID, map[string]interface{}{
				"capabilities": map[string]interface{}{
					"textDocumentSync": 1, // Full document sync for now
				},
			})
		case "textDocument/didOpen":
			var params DidOpenTextDocumentParams
			json.Unmarshal(msg.Params, &params)
			lintDocument(params.TextDocument.URI, params.TextDocument.Text)
		case "textDocument/didChange":
			var params DidChangeTextDocumentParams
			json.Unmarshal(msg.Params, &params)
			if len(params.ContentChanges) > 0 {
				lintDocument(params.TextDocument.URI, params.ContentChanges[0].Text)
			}
		}
	}
}

func lintDocument(uri string, text string) {
	// Re-use our robust Lexer and Parser!
	l := lexer.New(text)
	p := parser.New(l)
	p.ParseProgram() // We don't need the AST yet, just the errors

	var diagnostics []Diagnostic
	for _, errStr := range p.Errors() {
		// Create a basic diagnostic pointing to line 1 for now,
		// in the future you can parse the line/col from the error string!
		diag := Diagnostic{
			Range: Range{
				Start: Position{Line: 0, Character: 0},
				End:   Position{Line: 0, Character: 100},
			},
			Severity: 1, // 1 = Error
			Source:   "swalang",
			Message:  errStr,
		}
		diagnostics = append(diagnostics, diag)
	}

	sendNotification("textDocument/publishDiagnostics", PublishDiagnosticsParams{
		URI:         uri,
		Diagnostics: diagnostics,
	})
}

// Helper to write RPC responses back to stdout
func sendResponse(id *int, result interface{}) {
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"result":  result,
	}
	send(payload)
}

func sendNotification(method string, params interface{}) {
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  method,
		"params":  params,
	}
	send(payload)
}

func send(payload interface{}) {
	data, _ := json.Marshal(payload)
	fmt.Printf("Content-Length: %d\r\n\r\n%s", len(data), data)
}
