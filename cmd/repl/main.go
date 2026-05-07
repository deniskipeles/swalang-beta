package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/chzyer/readline"
	"github.com/deniskipeles/pylearn/internal/ast"
	"github.com/deniskipeles/pylearn/internal/builtins"
	"github.com/deniskipeles/pylearn/internal/constants"
	"github.com/deniskipeles/pylearn/internal/interpreter"
	"github.com/deniskipeles/pylearn/internal/lexer"
	"github.com/deniskipeles/pylearn/internal/object"
	"github.com/deniskipeles/pylearn/internal/parser"
)

func main() {
	historyFile := filepath.Join(os.TempDir(), ".swalang_history")
	if home := os.Getenv("HOME"); home != "" {
		historyFile = filepath.Join(home, ".swalang_history")
	}

	l, err := readline.NewEx(&readline.Config{
		Prompt:          constants.CmdReplMainPrompt,
		HistoryFile:     historyFile,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing readline: %v\n", err)
		os.Exit(1)
	}
	defer l.Close()

	env := object.NewEnvironment()
	for name, builtin := range builtins.Builtins {
		env.Set(name, builtin)
	}

	fmt.Println(constants.CmdReplMainWelcomeMessage)
	fmt.Println(constants.CmdReplMainExitMessage)

	var lines []string

	for {
		line, err := l.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				fmt.Println("KeyboardInterrupt")
				lines = nil
				l.SetPrompt(constants.CmdReplMainPrompt)
				continue
			} else {
				continue
			}
		} else if err == io.EOF {
			break
		}

		if line == "" && len(lines) == 0 {
			continue
		}

		if line == "" {
			// Empty line finishes multi-line input
			if len(lines) > 0 {
				evaluate(strings.Join(lines, "\n"), env)
				lines = nil
				l.SetPrompt(constants.CmdReplMainPrompt)
			}
			continue
		}

		lines = append(lines, line)
		fullInput := strings.Join(lines, "\n")

		if isIncomplete(fullInput) || isBlock(fullInput) {
			l.SetPrompt("... ")
			continue
		}

		// It might be a complete statement on a single line, or a finished multi-line expression (not block)
		evaluate(fullInput, env)
		lines = nil
		l.SetPrompt(constants.CmdReplMainPrompt)
	}

	fmt.Println(constants.CmdReplMainGoodbyeMessage)
}

func evaluate(input string, env *object.Environment) {
	lx := lexer.New(input)
	p := parser.New(lx)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		printParserErrors(os.Stderr, p.Errors())
		return
	}

	mainCtx := &interpreter.InterpreterContext{Env: env}
	evaluated := interpreter.Eval(program, mainCtx)

	if evaluated != nil && evaluated.Type() == object.ERROR_OBJ {
		errObj := evaluated.(*object.Error)
		fmt.Fprintf(os.Stderr, constants.CmdReplMainRuntimeErrorPrefix, errObj.Message)
		return
	}

	if evaluated != nil && evaluated.Type() != object.NULL_OBJ {
		fmt.Println(evaluated.Inspect())
	}
}

func isIncomplete(input string) bool {
	l := lexer.New(input)
	paren := 0
	bracket := 0
	brace := 0
	for {
		tok := l.NextToken()
		if tok.Type == lexer.EOF {
			break
		}
		switch tok.Type {
		case lexer.LPAREN:
			paren++
		case lexer.RPAREN:
			paren--
		case lexer.LBRACKET:
			bracket++
		case lexer.RBRACKET:
			bracket--
		case lexer.LBRACE:
			brace++
		case lexer.RBRACE:
			brace--
		}
	}
	if paren > 0 || bracket > 0 || brace > 0 {
		return true
	}
	return false
}

func isBlock(input string) bool {
	lx := lexer.New(input)
	p := parser.New(lx)
	program := p.ParseProgram()
	if len(program.Statements) == 0 {
		// Might still end with a colon which starts a block
		trimmed := strings.TrimSpace(input)
		return strings.HasSuffix(trimmed, ":")
	}
	lastStmt := program.Statements[len(program.Statements)-1]
	switch lastStmt.(type) {
	case *ast.IfStatement, *ast.WhileStatement, *ast.ForStatement, *ast.ClassStatement, *ast.TryStatement, *ast.WithStatement:
		return true
	case *ast.LetStatement:
		ls := lastStmt.(*ast.LetStatement)
		if _, ok := ls.Value.(*ast.FunctionLiteral); ok {
			return true
		}
	}
	return false
}

func printParserErrors(out io.Writer, errors []string) {
	fmt.Fprintln(out, constants.CmdReplMainParserErrorsEncountered)
	for _, msg := range errors {
		fmt.Fprintf(out, constants.CmdReplMainParserErrorFormat, msg)
	}
}
