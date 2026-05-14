package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/deniskipeles/pylearn/internal/builtins"
	"github.com/deniskipeles/pylearn/internal/constants"
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

type SemanticTokensParams struct {
	TextDocument struct {
		URI string `json:"uri"`
	} `json:"textDocument"`
}

type SemanticTokensResponse struct {
	Data []int `json:"data"`
}

type CompletionParams struct {
	TextDocument struct {
		URI string `json:"uri"`
	} `json:"textDocument"`
	Position Position `json:"position"`
}

type CompletionItem struct {
	Label  string `json:"label"`
	Kind   int    `json:"kind"` // 3=Function, 6=Variable, 14=Keyword
	Detail string `json:"detail,omitempty"`
}

type CompletionList struct {
	IsIncomplete bool             `json:"isIncomplete"`
	Items        []CompletionItem `json:"items"`
}

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

var documentCache = make(map[string]string)

func main() {
	reader := bufio.NewReader(os.Stdin)

	for {
		var length int
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					return
				}
				panic(err)
			}
			line = strings.TrimSpace(line)
			if line == "" {
				break
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

		buf := make([]byte, length)
		io.ReadFull(reader, buf)

		var msg Message
		json.Unmarshal(buf, &msg)

		switch msg.Method {
		case "initialize":
			sendResponse(msg.ID, map[string]interface{}{
				"capabilities": map[string]interface{}{
					"textDocumentSync": 1,
					"semanticTokensProvider": map[string]interface{}{
						"legend": map[string]interface{}{
							"tokenTypes": []string{
								"keyword", "function", "variable", "string", "number", "operator", "comment",
							},
							"tokenModifiers": []string{},
						},
						"full": true,
					},
					"completionProvider": map[string]interface{}{
						"resolveProvider":   false,
						"triggerCharacters": []string{"."},
					},
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
		case "textDocument/semanticTokens/full":
			var params SemanticTokensParams
			json.Unmarshal(msg.Params, &params)
			sendResponse(msg.ID, SemanticTokensResponse{
				Data: provideSemanticTokens(params.TextDocument.URI),
			})
		case "textDocument/completion":
			var params CompletionParams
			json.Unmarshal(msg.Params, &params)
			sendResponse(msg.ID, provideCompletions(params.TextDocument.URI, params.Position))
		}
	}
}

func lintDocument(uri string, text string) {
	documentCache[uri] = text
	l := lexer.New(text)
	p := parser.New(l)
	p.ParseProgram()

	re := regexp.MustCompile(`line (\d+):(\d+):? (.*)`)
	var diagnostics []Diagnostic
	for _, errStr := range p.Errors() {
		line, col := 0, 0
		msg := errStr

		if matches := re.FindStringSubmatch(errStr); len(matches) >= 4 {
			l, _ := strconv.Atoi(matches[1])
			c, _ := strconv.Atoi(matches[2])
			if l > 0 { line = l - 1 }
			if c > 0 { col = c - 1 }
			msg = strings.TrimSpace(matches[3])
		}

		diagnostics = append(diagnostics, Diagnostic{
			Range: Range{
				Start: Position{Line: line, Character: col},
				End:   Position{Line: line, Character: col + 5},
			},
			Severity: 1,
			Source:   "swalang",
			Message:  msg,
		})
	}
	sendNotification("textDocument/publishDiagnostics", PublishDiagnosticsParams{URI: uri, Diagnostics: diagnostics})
}

func provideCompletions(uri string, pos Position) CompletionList {
	text := documentCache[uri]
	lines := strings.Split(text, "\n")
	
	isDotCompletion := false

	// Check if we are doing a dot completion
	if pos.Line < len(lines) {
		lineText := lines[pos.Line]
		if pos.Character > 0 && pos.Character <= len(lineText) {
			cursorStr := lineText[:pos.Character]
			if strings.HasSuffix(cursorStr, ".") {
				isDotCompletion = true
			}
		}
	}

	var items []CompletionItem

	if isDotCompletion {
		// DOT COMPLETION (Attributes & Methods)
		// We provide standard types methods and magic methods.
		methods := []string{
			"append", "pop", "remove", "insert", "index", "count", "extend", // Lists
			"upper", "lower", "split", "join", "strip", "replace", "format", // Strings
			"keys", "values", "items", "get", "update",                      // Dicts
			"add", "discard", "union", "difference",                         // Sets
			"close", "read", "write", "readlines",                           // Files
		}
		for _, m := range methods {
			items = append(items, CompletionItem{Label: m, Kind: 3, Detail: "Method"})
		}
		
		// Add dunder methods dynamically
		dunders := []string{constants.DunderInit, constants.DunderStr, constants.DunderLen, constants.DunderCall}
		for _, d := range dunders {
			items = append(items, CompletionItem{Label: d, Kind: 3, Detail: "Magic Method"})
		}
	} else {
		// STANDARD COMPLETION (Keywords, Builtins, Locals)

		// 1. Dynamic Keywords (From currently compiled language tags)
		for kw := range lexer.GetKeywords() {
			items = append(items, CompletionItem{Label: kw, Kind: 14, Detail: "Keyword"})
		}

		// 2. Dynamic Built-in Functions
		for name := range builtins.Builtins {
			items = append(items, CompletionItem{Label: name, Kind: 3, Detail: "Built-in Function"})
		}

		// 3. Document Local Variables & Functions
		l := lexer.New(text)
		seen := make(map[string]bool)
		for {
			tok := l.NextToken()
			if tok.Type == lexer.EOF {
				break
			}
			if tok.Type == lexer.IDENT {
				if !seen[tok.Literal] {
					seen[tok.Literal] = true
					// Don't duplicate builtins or keywords
					if _, isBuiltin := builtins.Builtins[tok.Literal]; !isBuiltin {
						if _, isKeyword := lexer.GetKeywords()[tok.Literal]; !isKeyword {
							items = append(items, CompletionItem{
								Label:  tok.Literal,
								Kind:   6, // Variable
								Detail: "Local symbol",
							})
						}
					}
				}
			}
		}
	}

	return CompletionList{IsIncomplete: false, Items: items}
}

func provideSemanticTokens(uri string) []int {
	text := documentCache[uri]
	l := lexer.New(text)
	var data []int
	prevLine, prevChar := 0, 0

	for {
		tok := l.NextToken()
		if tok.Type == lexer.EOF { break }

		tokenType := -1
		if _, ok := lexer.GetKeywords()[tok.Literal]; ok {
			tokenType = 0 // Keyword
		} else {
			switch tok.Type {
			case lexer.IDENT: tokenType = 2
			case lexer.STRING, lexer.FSTRING, lexer.BYTES: tokenType = 3
			case lexer.INT, lexer.FLOAT: tokenType = 4
			case lexer.ASSIGN, lexer.PLUS, lexer.MINUS, lexer.ASTERISK, lexer.SLASH, lexer.EQ, lexer.DOT: tokenType = 5
			case lexer.COMMENT: tokenType = 6
			}
		}

		if tokenType == -1 { continue }

		line := tok.Line - 1
		char := tok.Column - 1
		if line < 0 { line = 0 }
		if char < 0 { char = 0 }

		deltaLine := line - prevLine
		deltaChar := char
		if deltaLine == 0 { deltaChar = char - prevChar }

		if deltaLine < 0 || deltaChar < 0 { continue }
		data = append(data, deltaLine, deltaChar, len(tok.Literal), tokenType, 0)
		prevLine, prevChar = line, char
	}
	return data
}

func sendResponse(id *int, result interface{}) {
	payload := map[string]interface{}{"jsonrpc": "2.0", "id": id, "result": result}
	send(payload)
}

func sendNotification(method string, params interface{}) {
	payload := map[string]interface{}{"jsonrpc": "2.0", "method": method, "params": params}
	send(payload)
}

func send(payload interface{}) {
	data, _ := json.Marshal(payload)
	fmt.Printf("Content-Length: %d\r\n\r\n%s", len(data), data)
}