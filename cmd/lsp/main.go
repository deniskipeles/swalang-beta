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
	Kind   int    `json:"kind"` // 3=Function, 6=Variable, 14=Keyword, etc.
	Detail string `json:"detail,omitempty"`
}

type CompletionList struct {
	IsIncomplete bool             `json:"isIncomplete"`
	Items        []CompletionItem `json:"items"`
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
		case "textDocument/completion":
			var params CompletionParams
			json.Unmarshal(msg.Params, &params)
			sendResponse(msg.ID, provideCompletions(params.TextDocument.URI, params.Position))
		case "textDocument/semanticTokens/full":
			var params SemanticTokensParams
			json.Unmarshal(msg.Params, &params)
			// Send back syntax highlighting data
			sendResponse(msg.ID, SemanticTokensResponse{
				Data: provideSemanticTokens(params.TextDocument.URI),
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
	// Cache the text for the semantic token request
	documentCache[uri] = text

	// Re-use our robust Lexer and Parser!
	l := lexer.New(text)
	p := parser.New(l)
	p.ParseProgram() // We don't need the AST yet, just the errors

	// Regex to extract line and column from Pylearn parser errors (e.g. "line 46:9: ...")
	re := regexp.MustCompile(`line (\d+):(\d+):? (.*)`)

	var diagnostics []Diagnostic
	for _, errStr := range p.Errors() {
		line := 0
		col := 0
		msg := errStr

		if matches := re.FindStringSubmatch(errStr); len(matches) >= 4 {
			parsedLine, _ := strconv.Atoi(matches[1])
			parsedCol, _ := strconv.Atoi(matches[2])
			// LSP is 0-indexed, so we subtract 1
			if parsedLine > 0 { line = parsedLine - 1 }
			if parsedCol > 0 { col = parsedCol - 1 }
			msg = strings.TrimSpace(matches[3])
		}

		diag := Diagnostic{
			Range: Range{
				Start: Position{Line: line, Character: col},
				// Highlight approx 5 characters from the error start
				End:   Position{Line: line, Character: col + 5}, 
			},
			Severity: 1, // 1 = Error
			Source:   "swalang",
			Message:  msg,
		}
		diagnostics = append(diagnostics, diag)
	}

	sendNotification("textDocument/publishDiagnostics", PublishDiagnosticsParams{
		URI:         uri,
		Diagnostics: diagnostics,
	})
}

// provideCompletions supplies autocomplete suggestions
func provideCompletions(uri string, pos Position) CompletionList {
	items := []CompletionItem{
		// Keywords
		{Label: "def", Kind: 14, Detail: "Define function"},
		{Label: "class", Kind: 14, Detail: "Define class"},
		{Label: "return", Kind: 14, Detail: "Return statement"},
		{Label: "import", Kind: 14, Detail: "Import module"},
		{Label: "if", Kind: 14}, {Label: "elif", Kind: 14}, {Label: "else", Kind: 14},
		{Label: "for", Kind: 14}, {Label: "while", Kind: 14}, {Label: "in", Kind: 14},
		{Label: "try", Kind: 14}, {Label: "except", Kind: 14}, {Label: "finally", Kind: 14},
		{Label: "async", Kind: 14}, {Label: "await", Kind: 14},

		// Built-ins
		{Label: "print", Kind: 3, Detail: "print(*args)"},
		{Label: "len", Kind: 3, Detail: "len(object)"},
		{Label: "type", Kind: 3, Detail: "type(object)"},
		{Label: "format_str", Kind: 3, Detail: "Format string like f\"\""},
		{Label: "int", Kind: 3}, {Label: "float", Kind: 3}, {Label: "str", Kind: 3},
		{Label: "list", Kind: 3}, {Label: "dict", Kind: 3}, {Label: "set", Kind: 3},
		{Label: "range", Kind: 3},
	}

	// Basic dynamic extraction of local variables from the document
	text := documentCache[uri]
	l := lexer.New(text)
	seen := make(map[string]bool)

	for {
		tok := l.NextToken()
		if tok.Type == lexer.EOF {
			break
		}
		if tok.Type == lexer.IDENT {
			// Don't add if it's already in our static list
			if !seen[tok.Literal] {
				seen[tok.Literal] = true
				items = append(items, CompletionItem{
					Label:  tok.Literal,
					Kind:   6, // 6 = Variable
					Detail: "Local identifier",
				})
			}
		}
	}

	return CompletionList{
		IsIncomplete: false,
		Items:        items,
	}
}

// Global cache to hold the latest document text for semantic tokens
var documentCache = make(map[string]string)

func provideSemanticTokens(uri string) []int {
	text := documentCache[uri]
	l := lexer.New(text)

	var data []int
	prevLine := 0
	prevChar := 0

	for {
		tok := l.NextToken()
		if tok.Type == lexer.EOF {
			break
		}

		tokenType := getLspTokenType(tok.Type)
		if tokenType == -1 {
			continue // Skip tokens we don't highlight
		}

		line := tok.Line - 1
		char := tok.Column - 1
		if line < 0 { line = 0 }
		if char < 0 { char = 0 }

		deltaLine := line - prevLine
		deltaChar := char
		if deltaLine == 0 {
			deltaChar = char - prevChar
		}

		// Ensure no negative deltas
		if deltaLine < 0 || deltaChar < 0 {
			continue
		}

		length := len(tok.Literal)

		data = append(data, deltaLine, deltaChar, length, tokenType, 0)

		prevLine = line
		prevChar = char
	}

	return data
}

func getLspTokenType(t lexer.TokenType) int {
	switch t {
	case lexer.FUNCTION, lexer.IF, lexer.ELSE, lexer.ELIF, lexer.RETURN, lexer.FOR, lexer.WHILE, lexer.IN, lexer.CLASS, lexer.IMPORT, lexer.FROM, lexer.AS, lexer.TRY, lexer.EXCEPT, lexer.FINALLY, lexer.RAISE, lexer.PASS, lexer.BREAK, lexer.CONTINUE, lexer.YIELD, lexer.ASYNC, lexer.AWAIT, lexer.WITH, lexer.GLOBAL, lexer.DEL, lexer.ASSERT, lexer.IS, lexer.NOT, lexer.AND, lexer.OR, lexer.TRUE, lexer.FALSE, lexer.NIL:
		return 0 // keyword
	case lexer.IDENT:
		return 2 // variable
	case lexer.STRING, lexer.FSTRING, lexer.BYTES:
		return 3 // string
	case lexer.INT, lexer.FLOAT:
		return 4 // number
	case lexer.ASSIGN, lexer.PLUS, lexer.MINUS, lexer.ASTERISK, lexer.SLASH, lexer.EQ, lexer.NOT_EQ, lexer.LT, lexer.GT, lexer.LT_EQ, lexer.GT_EQ, lexer.POW, lexer.FLOOR_DIV, lexer.PERCENT:
		return 5 // operator
	case lexer.COMMENT:
		return 6 // comment
	default:
		return -1 // skip
	}
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
