package vm

import (
	"fmt"
	"sort" // For sorting dict keys before compilation
	"encoding/binary"

	"github.com/deniskipeles/pylearn/internal/ast"
	"github.com/deniskipeles/pylearn/internal/object"
	"github.com/deniskipeles/pylearn/internal/builtins" // Need access to the builtins map/slice
	
)

// CompilationScope holds the instructions and symbol information for a single scope (e.g., function body).
type CompilationScope struct {
	instructions        Instructions
	lastInstruction     EmittedInstruction
	previousInstruction EmittedInstruction
}

// Compiler takes an AST node and compiles it into Bytecode
type Compiler struct {
	constants []object.Object // Global constant pool for the entire compilation

	symbolTable *SymbolTable // Symbol table for tracking variables

	// Stack of scopes for managing nested functions/blocks
	scopes      []CompilationScope
	scopeIndex  int

	// --- NEW: Loop Context ---
	loopCtx *LoopContext // Pointer to allow nil when not in loop

	lastSymbolTable *SymbolTable // Store the table for the last completed scope
}

// --- NEW: Loop Context Struct ---
type LoopContext struct {
	// For 'continue': bytecode address of the loop's start (condition or iteration check)
	StartPos int
	// For 'break': list of bytecode addresses where OpJump instructions
	// were emitted for break statements, needing patching later.
	BreakPatchPositions []int
	// Optional: Track outer loop context for nested loops later
	Outer *LoopContext
}

// Bytecode holds the compiled instructions and constants
type Bytecode struct {
	Instructions Instructions
	Constants    []object.Object
}

// EmittedInstruction helps track positions for optimizations or patching later
type EmittedInstruction struct {
	Opcode   Opcode
	Position int
	// TODO: Add operand info if needed for complex patching/optimizations
}


// New creates a new Compiler with a global scope.
func NewCompiler() *Compiler {
	mainScope := CompilationScope{
		instructions:        Instructions{},
		lastInstruction:     EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
	}

	symbolTable := NewSymbolTable()

	// Register built-ins in the symbol table
	// Assumes builtins package provides a way to get names/indices
	// For now, let's imagine a simple numbered registration
	// In a real scenario, you'd coordinate this with the VM's builtin handling.
	// for i, v := range object.Builtins { // Assuming object.Builtins is populated map or slice
	// 	symbolTable.DefineBuiltin(i, v.Name) // Need DefineBuiltin method
	// }

	// --- BEGIN FIX ---
    // Register built-ins in a deterministic order
    builtinNames := make([]string, 0, len(builtins.Builtins))
    for name := range builtins.Builtins {
        builtinNames = append(builtinNames, name)
    }
    sort.Strings(builtinNames) // Sort names alphabetically

    for i, name := range builtinNames {
        symbolTable.DefineBuiltin(i, name) // Use the index 'i'
    }
    // --- END FIX ---

	return &Compiler{
		constants:   []object.Object{},
		symbolTable: symbolTable,
		scopes:      []CompilationScope{mainScope},
		scopeIndex:  0,
		loopCtx:     nil, // Initially not in a loop
	}
}

// --- NEW: Helper methods for loop context ---
func (c *Compiler) enterLoop(startPos int) {
	newCtx := &LoopContext{
		StartPos:            startPos,
		BreakPatchPositions: []int{},
		Outer:               c.loopCtx, // Link to outer loop if any
	}
	c.loopCtx = newCtx
}

func (c *Compiler) leaveLoop() *LoopContext {
	if c.loopCtx == nil {
		panic("Compiler error: leaveLoop called when not in a loop")
	}
	leavingCtx := c.loopCtx
	c.loopCtx = leavingCtx.Outer // Restore outer loop context
	return leavingCtx
}

// NewWithState allows creating a compiler with existing state (useful for REPL).
// func NewWithState(s *SymbolTable, constants []object.Object) *Compiler { ... } // TODO if needed

// --- compileFunction Helper ---
// Returns core.CompiledFunction and Symbols
func (c *Compiler) compileFunction(node *ast.FunctionLiteral, name string) (*CompiledFunction, []Symbol, error) {
	c.enterScope()
	if name != "" { c.symbolTable.DefineFunctionName(name) }

	numParams := len(node.Parameters)
	// Define parameters in the new scope's symbol table
	for _, p := range node.Parameters {
		c.symbolTable.Define(p.Name.Value)
	}

	// Compile the function body
	err := c.Compile(node.Body)
	if err != nil { c.leaveScope(); return nil, nil, err }

	// Handle implicit return
	if c.lastInstructionIs(OpPop) { c.replaceLastPopWithReturn() }
	if !c.lastInstructionIs(OpReturnValue) && !c.lastInstructionIs(OpReturn) {
		c.emit(OpReturn)
	}

	freeSymbols := c.symbolTable.FreeSymbols
	numLocals := c.symbolTable.numDefinitions
	instructions := c.leaveScope()

	// Create the shared core.CompiledFunction object
	compiledFn := &CompiledFunction{ // Use core.CompiledFunction
		Instructions:  instructions,
		NumLocals:     numLocals,
		NumParameters: numParams, // Number of params defined in AST
		Name:          name,
	}
	return compiledFn, freeSymbols, nil
}


// Compile is the main entry point for compilation
func (c *Compiler) Compile(node ast.Node) error {
	switch node := node.(type) {
	case *ast.Program:
		for _, s := range node.Statements {
			err := c.Compile(s)
			if err != nil {
				return err
			}
		}
		// --- FIX: Add Implicit Return for Main Scope ---
		// After all statements in the main program are compiled,
		// add an OpReturn to ensure the main frame terminates correctly
		// and implicitly returns None.
		c.emit(OpReturn)

	case *ast.BlockStatement: // Process statements within the block
		for _, s := range node.Statements {
			err := c.Compile(s)
			if err != nil {
				return err
			}
		}

	
	case *ast.ExpressionStatement:
		// --- BEGIN FIX ---
		// Check if the expression is an assignment (Infix with '=')
		isAssignment := false
		if infixExpr, ok := node.Expression.(*ast.InfixExpression); ok {
			if infixExpr.Operator == "=" {
				isAssignment = true
			}
		}

		err := c.Compile(node.Expression) // Compile the inner expression
		if err != nil {
			return err
		}

		// Pop the result ONLY if it wasn't an assignment expression
		if !isAssignment {
			c.emit(OpPop)
		}
		// --- END FIX ---

	// --- Literals ---
	case *ast.IntegerLiteral:
		integer := &object.Integer{Value: node.Value}
		c.emit(OpConstant, c.addConstant(integer))

	case *ast.FloatLiteral:
		float := &object.Float{Value: node.Value}
		c.emit(OpConstant, c.addConstant(float))

	case *ast.StringLiteral:
		str := &object.String{Value: node.Value}
		c.emit(OpConstant, c.addConstant(str))

	case *ast.BytesLiteral:
		bytesObj := &object.Bytes{Value: node.Value} // AST already has []byte
		c.emit(OpConstant, c.addConstant(bytesObj))

	case *ast.BooleanLiteral:
		if node.Value {
			c.emit(OpTrue)
		} else {
			c.emit(OpFalse)
		}

	case *ast.NilLiteral:
		c.emit(OpNull)

	// --- Prefix & Infix Expressions ---
	case *ast.PrefixExpression:
		err := c.Compile(node.Right)
		if err != nil { return err }
		switch node.Operator {
		case "-": c.emit(OpMinus)
		case "+": c.emit(OpPos) // Added OpPos
		case "!": c.emit(OpBang) // Logical not
		case "not": c.emit(OpBang) // Treat 'not' keyword same as '!'
		default:
			return fmt.Errorf("unknown prefix operator: %s", node.Operator)
		}

	case *ast.InfixExpression:
		// --- BEGIN FIX ---
		// Handle assignment explicitly if the operator is '='
		if node.Operator == "=" {
			// 1. Check if the left side is a valid target
			fmt.Println(">>> Compiler: Handling Infix Assignment '='")
			var targetSymbol Symbol // Used only for Identifier targets
			isAttributeAssignment := false
			isIndexAssignment := false
            var attributeTargetNode *ast.DotExpression // Store the node for later use

			switch targetNode := node.Left.(type) {
			case *ast.Identifier:
				// Standard variable assignment: `name = value`
				targetSymbol = c.symbolTable.Define(targetNode.Value) // Define or update symbol
			case *ast.DotExpression:
				fmt.Printf("    Target is DotExpression: %s.%s\n", targetNode.Left, targetNode.Identifier)
				isAttributeAssignment = true
                attributeTargetNode = targetNode // <<< STORE the DotExpression node
				// Compile the object part first (e.g., compile 'self')
				err := c.Compile(targetNode.Left)
				if err != nil { return fmt.Errorf("error compiling object for attribute assignment: %w", err) }
				fmt.Printf("    Compiled DotExpression Left part. SP should have [object].\n")
                // Attribute name ('factor') is handled *after* the value
			case *ast.IndexExpression:
				fmt.Printf("    Target is IndexExpression: %s[%s]\n", targetNode.Left, targetNode.Index)
				isIndexAssignment = true
				err := c.Compile(targetNode.Left); if err != nil { return fmt.Errorf("error compiling collection for index assignment: %w", err) }
				fmt.Printf("    Compiled IndexExpression Left part. SP should have [collection].\n")
				err = c.Compile(targetNode.Index); if err != nil { return fmt.Errorf("error compiling index for index assignment: %w", err) }
				fmt.Printf("    Compiled IndexExpression Index part. SP should have [collection, index].\n")
			default:
				// Invalid assignment target
				return fmt.Errorf("invalid assignment target: cannot assign to %T", node.Left)
			}

			// 2. Compile the right-hand side value
			err := c.Compile(node.Right) // Compile the value expression
			if err != nil { return err }

			// 3. Emit Store Instruction
			if isAttributeAssignment {
                // <<< --- THIS IS THE FIX for DotExpression --- >>>
                if attributeTargetNode == nil { // Safety check
                     return fmt.Errorf("internal compiler error: attributeTargetNode is nil for attribute assignment")
                }
                // Get the attribute name (e.g., "factor")
				attrName := attributeTargetNode.Identifier.Value
				fmt.Printf("    Adding constant for attribute name: %s\n", attrName) // LOG
                // Add the name string to the constant pool
				nameIndex := c.addConstant(&object.String{Value: attrName})
				fmt.Printf("    Added constant '%s' at index %d\n", attrName, nameIndex)
				fmt.Printf("    Emitting OpSetAttribute (NameIdx: %d, Name: %s)\n", nameIndex, attrName)
                // Emit the opcode with the constant index of the name
				c.emit(OpSetAttribute, nameIndex)
                // Stack after OpSetAttribute: [] (consumes object and value)
                // <<< --- END FIX --- >>>

			} else if isIndexAssignment {
				fmt.Println("    Emitting OpSetIndex")
				c.emit(OpSetIndex)
                 // Stack after OpSetIndex: [] (consumes collection, index, value)
			} else {
				// Must be identifier assignment
				fmt.Printf("    Storing symbol %s (Scope: %s, Index: %d)\n", targetSymbol.Name, targetSymbol.Scope, targetSymbol.Index)
				c.storeSymbol(targetSymbol) // Emits OpSetGlobal/Local etc.
                // Stack after OpSet*: [] (consumes value)
			}

			// Assignment itself doesn't push a result used by other expressions typically
			// If the assignment was part of an ExpressionStatement, OpPop will be added there.
			// Don't add OpPop here.

			return nil // Handled assignment, exit this case
		}
		// --- END FIX ---

		if node.Operator == "and" {
			// Logic: x and y
			// If x is Falsy, result is x.
			// If x is Truthy, result is y.

			// 1. Compile x
			err := c.Compile(node.Left)
			if err != nil { return err }
			// Stack: [x]

			// 2. Duplicate x
			c.emit(OpDup)
			// Stack: [x, x]

			// 3. Jump if the *top* copy of x is Falsy.
			// OpJumpNotTruthy pops the top element it tests.
			// If falsy, jumps to 'afterYLabel'. Stack at target: [original_falsy_x]
			jumpIfFalsePos_AND := c.emit(OpJumpNotTruthy, 9999)

			// 4. If x was True (didn't jump):
			// Stack: [original_truthy_x, truthy_x]. OpJumpNotTruthy popped the top one.
			// Stack: [original_truthy_x]. We don't want this x, we want y. Pop it.
			c.emit(OpPop)
			// Stack: []. Now compile y.
			err = c.Compile(node.Right) // Compile y
			if err != nil { return err }
			// Stack: [y]. This is the result for the 'true' path.

			// 5. Define the target label for the falsy jump from step 3
			afterYLabel_AND := len(c.currentInstructions())
			c.changeOperand(jumpIfFalsePos_AND,0, afterYLabel_AND)
			// Execution lands here.
			// Stack: [original_falsy_x] (if x was false) OR [y] (if x was true).
			// The correct result is now on top of the stack.

			return nil // Handled 'and'

		} else if node.Operator == "or" {
			// Logic: x or y
			// If x is Truthy, result is x.
			// If x is Falsy, result is y.

			// 1. Compile x
			err := c.Compile(node.Left)
			if err != nil { return err }
			// Stack: [x]

			// 2. Duplicate x
			c.emit(OpDup)
			// Stack: [x, x]

			// 3. Jump if *top* x is Falsy.
			// OpJumpNotTruthy pops the top element it tests.
			// If falsy, jumps to 'afterXLabel'. Stack at target: [original_falsy_x]
			jumpIfFalsePos_OR := c.emit(OpJumpNotTruthy, 9999)

			// 4. If x was True (didn't jump):
			// Stack: [original_truthy_x, truthy_x]. OpJumpNotTruthy popped the top one.
			// Stack: [original_truthy_x]. This *is* the result for 'or'.
			// Jump over the 'y' evaluation part to the end.
			jumpEndPos_OR := c.emit(OpJump, 9999)

			// 5. Define the target label for the falsy jump from step 3
			afterXLabel_OR := len(c.currentInstructions())
			c.changeOperand(jumpIfFalsePos_OR,0, afterXLabel_OR)
			// Execution lands here if x was false.
			// Stack: [original_falsy_x]. We don't want this, we want y. Pop it.
			c.emit(OpPop)
			// Stack: []. Now compile y.
			err = c.Compile(node.Right) // Compile y
			if err != nil { return err }
			// Stack: [y]. This is the result for the 'false' path.

			// 6. Define the end label (target for the jump in step 4)
			endLabel_OR := len(c.currentInstructions())
			c.changeOperand(jumpEndPos_OR,0, endLabel_OR)
			// Execution lands here.
			// Stack: [original_truthy_x] (if x was true) OR [y] (if x was false).
			// The correct result is now on top of the stack.

			return nil // Handled 'or'
		}

		// --- Original Infix Logic (for +, -, *, ==, etc.) ---
		// Re-order for comparison operators < and <= etc. (optional optimization)
		err := c.Compile(node.Left)
		if err != nil { return err }
		err = c.Compile(node.Right)
		if err != nil { return err }

		switch node.Operator {
		case "+": c.emit(OpAdd)
		case "-": c.emit(OpSubtract)
		// ... rest of the non-assignment infix operators ...
		case "*": c.emit(OpMultiply)
		case "/": c.emit(OpDivide)
		case "%": c.emit(OpModulo)
		case "==": c.emit(OpEqual)
		case "!=": c.emit(OpNotEqual)
		case ">": c.emit(OpGreaterThan)
		case ">=": c.emit(OpGreaterThanEqual)
		case "<": c.emit(OpLesserThan)
		case "<=": c.emit(OpLesserThanEqual)
		case "in":
			c.emit(OpIn)
		case "not in":
			c.emit(OpIn)   // First, check if it's 'in'
			c.emit(OpBang) // Then, negate the result
		default:
			// Should not happen if '=' is handled above
			return fmt.Errorf("unknown infix operator (post-assignment check): %s", node.Operator)
		}

	// --- Variables ---
	case *ast.Identifier:
		symbol, ok := c.symbolTable.Resolve(node.Value)
		if !ok {
			return fmt.Errorf("undefined variable: %s", node.Value)
		}
		c.loadSymbol(symbol) // Emits appropriate Get instruction

	case *ast.LetStatement: // Assignment
		symbol := c.symbolTable.Define(node.Name.Value) // Define or update symbol
		err := c.Compile(node.Value)                     // Compile the value first
		if err != nil { return err }
		c.storeSymbol(symbol) // Emits appropriate Set instruction


	case *ast.IfStatement:
		// --- Compile IF part ---
		// 1. Compile the 'if' condition
		err := c.Compile(node.Condition)
		if err != nil { return err }
		// Stack: [..., if_condition_result]

		// 2. Emit OpJumpNotTruthy with placeholder for skipping 'if' body
		// Jumps if the 'if' condition is FALSE
		ifJumpNotTruthyPos := c.emit(OpJumpNotTruthy, 9999)
		// Stack: [] (OpJumpNotTruthy consumes the condition result)

		// 3. Compile the 'if' consequence block
		err = c.Compile(node.Consequence)
		if err != nil { return err }
		// // Ensure stack is clean after consequence (remove potential leftover value)
		// if !c.lastInstructionIs(OpReturnValue) && !c.lastInstructionIs(OpReturn) { // Added check
		// 	c.emit(OpPop)
		// }


		// 4. Emit OpJump to skip over 'elif'/'else' blocks (placeholder)
		// This jump executes ONLY if the 'if' consequence was executed.
		ifEndJumpPos := c.emit(OpJump, 9999)

		// 5. Patch the 'if' jump: OpJumpNotTruthy jumps here if condition was false
		afterIfConsequencePos := len(c.currentInstructions())
		c.changeOperand(ifJumpNotTruthyPos,0, afterIfConsequencePos)

		// --- Compile ELIF parts ---
		elifEndJumpPositions := []int{} // Store end jump positions for each elif block

		for _, elifBlock := range node.ElifBlocks {
			// Position where this elif block starts execution
			// currentElifStartPos := len(c.currentInstructions())

			// 1. Compile the 'elif' condition
			err = c.Compile(elifBlock.Condition)
			if err != nil { return err }
			// Stack: [..., elif_condition_result]

			// 2. Emit OpJumpNotTruthy for this elif (placeholder)
			// Jumps if the 'elif' condition is FALSE
			elifJumpNotTruthyPos := c.emit(OpJumpNotTruthy, 9999)
			// Stack: []

			// 3. Compile the 'elif' consequence block
			err = c.Compile(elifBlock.Consequence)
			if err != nil { return err }
			// if !c.lastInstructionIs(OpReturnValue) && !c.lastInstructionIs(OpReturn) { // Added check
			// 	c.emit(OpPop) // Clean up stack
			// }


			// 4. Emit OpJump to skip remaining elif/else (placeholder)
			// Executes if this elif consequence ran
			elifEndJumpPos := c.emit(OpJump, 9999)
			elifEndJumpPositions = append(elifEndJumpPositions, elifEndJumpPos)

			// 5. Patch the 'elif' jump: OpJumpNotTruthy jumps here if condition was false
			afterElifConsequencePos := len(c.currentInstructions())
			c.changeOperand(elifJumpNotTruthyPos,0, afterElifConsequencePos)
		}

		// --- Compile ELSE part (if exists) ---
		if node.Alternative != nil {
			// Position where else block starts (target for last if/elif jump)
			// This is the current position.

			// Compile the 'else' block
			err = c.Compile(node.Alternative)
			if err != nil { return err }
			// if !c.lastInstructionIs(OpReturnValue) && !c.lastInstructionIs(OpReturn) { // Added check
			// 	c.emit(OpPop) // Clean up stack
			// }

		} else {
			// If there's no 'else' block, and all if/elif conditions were false,
			// execution jumps to the end. We don't need to explicitly push NULL here,
			// as the VM stack should be empty if no block executed and popped its condition.
			// If the entire if/elif/else was an ExpressionStatement, the outer
			// OpPop would handle it. If it's just a statement, nothing needs to be pushed.
		}

		// --- Patch END Jumps ---
		// Position after the entire if/elif/else structure
		afterAllPos := len(c.currentInstructions())

		// Patch the jump from the 'if' consequence
		c.changeOperand(ifEndJumpPos,0, afterAllPos)

		// Patch the jumps from each 'elif' consequence
		for _, pos := range elifEndJumpPositions {
			c.changeOperand(pos,0, afterAllPos)
		}


	case *ast.ReturnStatement:
		if node.ReturnValue == nil { // `return` (implicitly None)
			c.emit(OpReturn) // Special opcode for returning Null
		} else {
			err := c.Compile(node.ReturnValue)
			if err != nil { return err }
			c.emit(OpReturnValue) // Return value on top of stack
		}

	// --- Collections ---
	case *ast.ListLiteral:
		for _, el := range node.Elements {
			err := c.Compile(el)
			if err != nil { return err }
		}
		c.emit(OpArray, len(node.Elements))

	case *ast.TupleLiteral:
		for _, el := range node.Elements {
			err := c.Compile(el)
			if err != nil { return err }
		}
		c.emit(OpTuple, len(node.Elements))

	case *ast.SetLiteral:
		for _, el := range node.Elements {
			err := c.Compile(el)
			if err != nil { return err }
		}
		c.emit(OpSet, len(node.Elements))

	case *ast.DictLiteral:
		// Sort keys for deterministic compilation/hashing order
		keys := make([]ast.Expression, 0, len(node.Pairs))
		for k := range node.Pairs { keys = append(keys, k) }
		sort.Slice(keys, func(i, j int) bool {
			return keys[i].String() < keys[j].String() // Simple string comparison of AST nodes
		})

		for _, k := range keys {
			v := node.Pairs[k]
			err := c.Compile(k) // Push key
			if err != nil { return err }
			err = c.Compile(v) // Push value
			if err != nil { return err }
		}
		c.emit(OpHash, len(node.Pairs)) // Emit with number of key-value PAIRS

	case *ast.IndexExpression: // e.g., my_list[index]
		err := c.Compile(node.Left) // Collection
		if err != nil { return err }
		err = c.Compile(node.Index) // Index
		if err != nil { return err }
		c.emit(OpIndex)

	// --- Attribute Access ---
	case *ast.DotExpression: // e.g., object.attribute
		err := c.Compile(node.Left) // Compile the object
		if err != nil { return err }
		nameIndex := c.addConstant(&object.String{Value: node.Identifier.Value}) // Add name to constants
		c.emit(OpGetAttribute, nameIndex)                                  // Emit GetAttribute

	// --- Function Calls ---
	case *ast.CallExpression:
		err := c.Compile(node.Function) // Compile the callable object
		if err != nil { return err }
		for _, arg := range node.Arguments { // Compile arguments
			err := c.Compile(arg)
			if err != nil { return err }
		}
		c.emit(OpCall, len(node.Arguments)) // Emit call with argument count


	// --- Imports ---
	case *ast.ImportStatement:
		moduleName := node.Name.Value
		nameIndex := c.addConstant(&object.String{Value: moduleName})
		c.emit(OpImportName, nameIndex)
		// After import, we need to store the returned module object.
		// Assume import pushes the module object onto the stack.
		symbol := c.symbolTable.Define(moduleName)
		c.storeSymbol(symbol) // Store the module object
	
	


	// --- Placeholders for unsupported/complex features ---
	case *ast.FunctionLiteral: // Standalone functions or lambdas
		funcName := ""
		if node.Name != nil { funcName = node.Name.Value }
		compiledFn, freeSymbols, err := c.compileFunction(node, funcName) // Use helper
		if err != nil { return err }
		fnConstIdx := c.addConstant(compiledFn) // Add core.CompiledFunction
		for _, s := range freeSymbols { c.loadSymbol(s) } // Load free vars
		c.emit(OpClosure, fnConstIdx, len(freeSymbols)) // Emit OpClosure


	// ... make sure the *ast.LetStatement case still exists ...
	// It will compile the FunctionLiteral (which emits OpClosure)
	// and then storeSymbol will emit OpSetGlobal/Local for the function name.
	case *ast.ClassStatement:
		className := node.Name.Value
		nameConstIdx := c.addConstant(&object.String{Value: className})

		// --- Push Class Name FIRST ---
		fmt.Printf(">>> Compiler: Pushing Class Name Constant: %s (idx %d)\n", className, nameConstIdx)
		c.emit(OpConstant, nameConstIdx)
		// Stack Runtime Goal: [..., ClassNameString]

		numMethods := 0
		// Compile methods defined in the body
		fmt.Printf(">>> Compiler: Starting method compilation for class %s\n", className)
		for _, stmt := range node.Body.Statements {
			// Expect methods to be defined via LetStatement(Name, FunctionLiteral) due to parser
			if letStmt, ok := stmt.(*ast.LetStatement); ok {
				if fnLit, okLit := letStmt.Value.(*ast.FunctionLiteral); okLit {
					methodName := letStmt.Name.Value
					fmt.Printf(">>> Compiler: Compiling method: %s\n", methodName)
					if fnLit.Name == nil { fnLit.Name = letStmt.Name } // Ensure AST node has name
					if fnLit.Name.Value != methodName {
						return fmt.Errorf("internal compiler error: Mismatched AST names for method %s", methodName)
					}

					// Compile the method function (gets core.CompiledFunction)
					compiledFn, freeSymbols, err := c.compileFunction(fnLit, methodName)
					if err != nil { return fmt.Errorf("error compiling method %s: %w", methodName, err) }
					fnConstIdx := c.addConstant(compiledFn) // Add core.CompiledFunction
					fmt.Printf("    Compiled method %s -> core.CompiledFunction constIdx=%d\n", methodName, fnConstIdx)

					// Load free vars needed by this method's closure onto runtime stack
					if len(freeSymbols) > 0 {
						fmt.Printf("    Loading %d free vars for method %s\n", len(freeSymbols), methodName)
						for _, s := range freeSymbols { c.loadSymbol(s) }
					}

					// --- Push Name FIRST, then Closure ---
					// 1. Push the method name constant
					methodNameIdx := c.addConstant(&object.String{Value: methodName})
					fmt.Printf("    Emitting OpConstant for method name '%s' (idx %d)\n", methodName, methodNameIdx)
					c.emit(OpConstant, methodNameIdx)
					// Stack Runtime Goal: [..., ClassNameString, ..., MethodNameString]

					// 2. Emit OpClosure for the method (pushes vm.Closure at runtime)
					fmt.Printf("    Emitting OpClosure for method '%s' (fnIdx %d, free %d)\n", methodName, fnConstIdx, len(freeSymbols))
					c.emit(OpClosure, fnConstIdx, len(freeSymbols))
					// Stack Runtime Goal: [..., ClassNameString, ..., MethodNameString, methodVmClosure]

					numMethods++
				} else {
					// Handle non-function definitions if needed (e.g., class variables)
					fmt.Printf("Warning: Compiler ignoring non-method LetStatement in class body: %s = %T\n", letStmt.Name.Value, letStmt.Value)
				}
			} else if _, isPass := stmt.(*ast.PassStatement); !isPass {
				// Handle other statements (like docstrings) if needed, or ignore/warn
				fmt.Printf("Warning: Compiler ignoring non-assignment/non-def/non-pass statement in class body: %T\n", stmt)
			}
		} // end loop over body statements
		fmt.Printf(">>> Compiler: Finished method compilation for class %s. Found %d methods.\n", className, numMethods)


		// Emit OpBuildClass (operand is number of methods pushed)
		fmt.Printf(">>> Compiler: Emitting OpBuildClass(%d)\n", numMethods)
		c.emit(OpBuildClass, numMethods)
		// Stack Runtime Goal after OpBuildClass: [..., VmClassObject]

		// Store the resulting Class object in the outer scope
		classSymbol := c.symbolTable.Define(className) // Define class name in outer scope
		fmt.Printf(">>> Compiler: Storing class %s using symbol %+v\n", className, classSymbol)
		c.storeSymbol(classSymbol) // Emit OpSetGlobal/Local

	case *ast.ListComprehension:
		// Compile the comprehension as a hidden function that is immediately called.
		// The hidden function takes the iterable as its only argument.
		err := c.compileComprehension(
			"<listcomp>",
			node.Element,
			node.Variable,
			node.Iterable,
			node.Condition,
			OpArray,
			"append",
		)
		if err != nil {
			return err
		}

	case *ast.SetComprehension:
		// Compile the comprehension as a hidden function that is immediately called.
		// The hidden function takes the iterable as its only argument.
		err := c.compileComprehension(
			"<setcomp>",
			node.Element,
			node.Variable,
			node.Iterable,
			node.Condition,
			OpSet,
			"add",
		)
		if err != nil {
			return err
		}

	case *ast.ForStatement:
		// 1. Compile the iterable expression
		err := c.Compile(node.Iterable)
		if err != nil { return err }
		// Stack: [..., iterable_object]

		// 2. Emit OpGetIter to get the iterator
		c.emit(OpGetIter)
		// Stack: [..., iterator_object]

		// --- Loop Setup ---
		// 3. Mark the start of the loop iteration logic (target for continue)
		loopStartPos := len(c.currentInstructions())

		// 4. Emit OpForIter with a placeholder jump offset (for loop exhaustion)
		opForIterPos := c.emit(OpForIter, 9999)
		// If continues: Stack: [..., iterator_object, next_item]
		// If exhausted: Jumps to afterLoopPos (defined below), pops iterator

		// --- Enter Loop Context ---
		c.enterLoop(loopStartPos) // Provide the 'continue' target address
		// --------------------------

		// --- Assignment Logic ---
		if len(node.Variables) > 0 {
			// 5a. Unpack multiple variables
			numVars := len(node.Variables)
			c.emit(OpUnpackSequence, numVars)
			// Stack: [..., iterator_object, var_N, ..., var_1, var_0]

			// 6a. Store unpacked variables in reverse order
			for i := numVars - 1; i >= 0; i-- {
				symbol := c.symbolTable.Define(node.Variables[i].Value)
				c.storeSymbol(symbol)
			}
			// Stack: [..., iterator_object]
		} else if node.Variable != nil {
			// 5b. Define the single loop variable
			symbol := c.symbolTable.Define(node.Variable.Value)

			// 6b. Emit instruction to store the next_item into the loop variable
			c.storeSymbol(symbol) // Pops next_item
			// Stack: [..., iterator_object]
		} else {
			// Should not happen with a valid parser
			return fmt.Errorf("internal compiler error: ForStatement has no loop variables")
		}
		// --- End Assignment ---

		// --- Compile Body (Now inside loop context) ---
		// 7. Compile the loop body. break/continue will use c.loopCtx.
		err = c.Compile(node.Body)
		if err != nil {
			// Important: Leave loop context even if body compilation fails,
			// otherwise subsequent code might think it's still in a loop.
			_ = c.leaveLoop() // Discard context data on error
			return err
		}
		// --------------------------------------------

		// --- Loop Continuation ---
		// 8. Emit jump back to the start of the loop iteration (to OpForIter)
		c.emit(OpJump, loopStartPos)

		// --- Loop Exit & Patching ---
		// 9. Mark the position *after* the loop (target for OpForIter jump on exhaustion AND break jumps)
		afterLoopPos := len(c.currentInstructions())

		// 10. Patch the OpForIter jump offset to point after the loop
		c.changeOperand(opForIterPos,0, afterLoopPos)

		// --- Leave Loop Context and Patch Breaks ---
		// 11. Pop the loop context and patch any break jumps emitted within the body
		leavingCtx := c.leaveLoop()
		for _, breakJumpPos := range leavingCtx.BreakPatchPositions {
			c.changeOperand(breakJumpPos,0, afterLoopPos) // Patch breaks to jump after the loop
		}
		// -------------------------------------------


	case *ast.WhileStatement:
		conditionPos := len(c.currentInstructions()) // Target for continue/loop jump

		// --- Enter Loop Context ---
		c.enterLoop(conditionPos)
		// --------------------------

		err := c.Compile(node.Condition)
		if err != nil { /* c.leaveLoop(); */ return err } // Handle error
		jumpNotTruthyPos := c.emit(OpJumpNotTruthy, 9999)

		// --- Compile Body (Now inside loop context) ---
		err = c.Compile(node.Body)
		if err != nil { /* c.leaveLoop(); */ return err } // Handle error
		// --------------------------------------------

		// --- Jump back ---
		c.emit(OpJump, conditionPos)

		afterLoopPos := len(c.currentInstructions())
		c.changeOperand(jumpNotTruthyPos,0, afterLoopPos)

		// --- Leave Loop Context and Patch Breaks ---
		leavingCtx := c.leaveLoop()
		afterLoopJumpTarget := len(c.currentInstructions()) // Breaks jump here
		for _, breakJumpPos := range leavingCtx.BreakPatchPositions {
			c.changeOperand(breakJumpPos,0, afterLoopJumpTarget)
		}

	case *ast.PassStatement:
		// No operation needed
		break // Do nothing

	// case *ast.TryStatement:
	// 	// Requires OpSetupExcept, OpPopBlock, complex jump patching
	// 	return fmt.Errorf("compilation for 'try...except' not yet implemented")

	case *ast.TryStatement:
		// --- Compile Try Block ---
		// 1. Emit OpPushExceptionHandler with placeholder targets
		// Target 1: Where to jump if an exception occurs in the 'try' body (start of except block logic)
		// Target 2: Where 'finally' block starts (0 for now, as 'finally' not implemented)
		pushHandlerPos := c.emit(OpPushExceptionHandler, 9999, 0) // except_target=9999, finally_target=0

		// 2. Compile the 'try' block body
		err := c.Compile(node.Body)
		if err != nil { return err } // Error during compilation of try body

		// 3. Emit OpPopExceptionHandler (signals end of protected block *if no error*)
		c.emit(OpPopExceptionHandler)

		// 4. Emit OpJump to skip over all 'except' blocks if 'try' was successful
		afterTryJumpPos := c.emit(OpJump, 9999) // Placeholder end target

		// --- Compile Except Blocks ---
		// 5. Mark the start of the 'except' block execution path (target for jump-on-error)
		exceptStartPos := len(c.currentInstructions())
		// Patch the first operand (except_target) of OpPushExceptionHandler
		c.changeOperand(pushHandlerPos,0, exceptStartPos) // Change operand at index 0 (operand 1)

		// This code path is only entered if the VM jumped here due to an exception.
		// The VM is expected to have placed the exception object on the stack.

		handlerEndJumpPositions := []int{} // Store jumps from end of each successful handler

		if len(node.Handlers) == 0 {
			// If there are no handlers, we still need something at the exceptStartPos.
			// We need to re-raise the exception.
			c.emit(OpRaise) // Assume OpRaise pops the exception and propagates it
		} else {
			for _, handler := range node.Handlers {
				// Each handler section starts here.
				// We need to check if the current exception matches this handler.
				var jumpToNextHandlerPos = -1

				if handler.Type != nil {
					// --- Type Matching Logic ---
					// a. Push the expected exception type onto the stack
					//    This usually involves loading a class (global/builtin)
					err := c.Compile(handler.Type)
					if err != nil { return err }

					// b. We need the VM to perform the check. Let's assume the VM implicitly
					//    checks the pushed type against the caught exception (which is below it).
					//    A dedicated opcode would be cleaner, but let's try without first.
					//    We'll emit OpJumpNotTruthy, assuming the VM leaves True/False on stack
					//    after an *implicit* check. This is fragile.
					//    *** REVISION: Let's add an opcode for clarity ***
					//    Okay, let's stick without a dedicated opcode for now and refine VM logic.
					//    Assume VM leaves True/False after implicit check.

					//    *** Let's refine the VM approach instead: ***
					//    The VM will handle matching. We just compile the handler body.
					//    The jump logic needs careful thought.

					//    *** Alternative Compiler Strategy: ***
					//    Emit code to get the type of the exception on stack, compare with handler type.
					//    This requires runtime type() and ==. Let's avoid this for now.

					//    *** Simplest VM-Assisted Approach: ***
					//    The VM jumps to exceptStartPos. The code here needs to figure out which handler runs.
					//    Maybe OpPushExceptionHandler should push *multiple* targets, one per handler? Complex.

					//    *** Let's stick to the linear check: ***
					//    Assume exception is on stack [..., exc_obj]
					err = c.Compile(handler.Type) // Compile type class. Stack: [..., exc_obj, TypeClass]
					if err != nil { return err }
					c.emit(OpSwap) // Stack: [..., TypeClass, exc_obj] - Need OpSwap! Add it.
					c.emit(OpIsInstance) // Needs OpIsInstance! Add it. Pops both, pushes bool.
					// Stack: [..., bool_result]
					jumpToNextHandlerPos = c.emit(OpJumpNotTruthy, 9999) // Jumps if NO match
					// If matched: Stack: [] (OpJumpNotTruthy pops)
				}
				// If handler.Type was nil (bare except), we don't emit type checks; it always matches.

				// --- Handler Body Execution ---
				// If we are here, the exception matched (or it was a bare except).
				// The exception object is *still* on the stack if we didn't emit checks,
				// or consumed if checks were emitted.
				// Let's simplify: Assume VM makes exception available.

				if handler.Var != nil {
					// TODO: Emit code to somehow get the exception object and store it
					// Option 1: VM pushes it before jumping here. We just store it.
					// Option 2: A special opcode OpLoadCaughtException pushes it.
					// Let's assume Option 1 for now.
					symbol := c.symbolTable.Define(handler.Var.Value)
					c.storeSymbol(symbol) // Pops the exception object from stack
				} else {
					// If no 'as e', pop the exception object the VM put on the stack
					c.emit(OpPop)
				}

				// Compile the handler body
				err = c.Compile(handler.Body)
				if err != nil { return err }

				// After handler body, pop its exception handler marker
				// c.emit(OpPopExceptionHandler) // Pop the handler pushed by OpPushExceptionHandler

				// Jump past all other handlers and the potential re-raise
				handlerEndJump := c.emit(OpJump, 9999) // Placeholder end target
				handlerEndJumpPositions = append(handlerEndJumpPositions, handlerEndJump)

				// --- Patching for Type Check ---
				if jumpToNextHandlerPos != -1 {
					// Mark the position for the *next* handler check
					nextHandlerPos := len(c.currentInstructions())
					c.changeOperand(jumpToNextHandlerPos,0, nextHandlerPos) // Patch jump if type didn't match
				}
			} // End loop through handlers

			// If execution reaches here, no handler matched. Re-raise.
			c.emit(OpRaise)
		} // End if len(handlers) > 0

		// --- End Jump Patching ---
		// Mark the position after the entire try...except structure
		endPos := len(c.currentInstructions())
		// Patch the jump from the successful 'try' block
		c.changeOperand(afterTryJumpPos,0, endPos)
		// Patch jumps from the end of each successful 'except' handler
		for _, jumpPos := range handlerEndJumpPositions {
			c.changeOperand(jumpPos,0, endPos)
		}


	case *ast.BreakStatement:
		// Check if inside a loop
		if c.loopCtx == nil {
			// Use the break token for location info
			return fmt.Errorf("line %d:%d: 'break' outside loop", node.Token.Line, node.Token.Column)
		}

		// Emit OpJump with placeholder, store position for later patching
		breakJumpPos := c.emit(OpJump, 9999)
		c.loopCtx.BreakPatchPositions = append(c.loopCtx.BreakPatchPositions, breakJumpPos)

	case *ast.ContinueStatement:
		// Check if inside a loop
		if c.loopCtx == nil {
			// Use the continue token for location info
			return fmt.Errorf("line %d:%d: 'continue' outside loop", node.Token.Line, node.Token.Column)
		}

		// Emit OpJump to the stored start position of the current loop
		// NOTE: We assume the stack is clean *before* the continue statement
		// executes. If 'continue' could follow an expression whose value
		// needs popping, we'd emit OpPop before OpJump here. For simplicity,
		// assume stack is clean (e.g., `if cond: continue`).
		c.emit(OpJump, c.loopCtx.StartPos)


	default:
		return fmt.Errorf("unsupported node type for compilation: %T", node)
	}
	return nil
}

// Bytecode returns the compiled bytecode for the main scope.
func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.scopes[c.scopeIndex].instructions,
		Constants:    c.constants,
	}
}

// --- Helper Methods ---

// addConstant adds an object to the *global* constant pool and returns its index.
func (c *Compiler) addConstant(obj object.Object) int {
	c.constants = append(c.constants, obj)
	return len(c.constants) - 1
}

// emit generates bytecode instructions and adds them to the *current* scope.
func (c *Compiler) emit(op Opcode, operands ...int) int {
	ins := Make(op, operands...)
	pos := c.addInstruction(ins)
	c.setLastInstruction(op, pos)
	return pos
}

// addInstruction appends instructions to the *current* scope's byte
func (c *Compiler) addInstruction(ins []byte) int {
	posNewInstruction := len(c.currentInstructions())
	c.scopes[c.scopeIndex].instructions = append(c.currentInstructions(), ins...)
	return posNewInstruction
}

// setLastInstruction updates tracking info for the *current* scope.
func (c *Compiler) setLastInstruction(op Opcode, pos int) {
	previous := c.scopes[c.scopeIndex].lastInstruction
	last := EmittedInstruction{Opcode: op, Position: pos}
	c.scopes[c.scopeIndex].previousInstruction = previous
	c.scopes[c.scopeIndex].lastInstruction = last
}

// lastInstructionIs checks the last emitted instruction in the *current* scope.
func (c *Compiler) lastInstructionIs(op Opcode) bool {
    if len(c.currentInstructions()) == 0 { return false }
	return c.scopes[c.scopeIndex].lastInstruction.Opcode == op
}

// removeLastPop removes the last instruction if it was OpPop in the *current* scope.
func (c *Compiler) removeLastPop() {
	if c.lastInstructionIs(OpPop) {
		last := c.scopes[c.scopeIndex].lastInstruction
		c.scopes[c.scopeIndex].instructions = c.currentInstructions()[:last.Position]
		c.scopes[c.scopeIndex].lastInstruction = c.scopes[c.scopeIndex].previousInstruction
	}
}

// currentInstructions gets the instructions for the current compilation scope.
func (c *Compiler) currentInstructions() Instructions {
	return c.scopes[c.scopeIndex].instructions
}

// enterScope creates a new compilation scope (e.g., for functions).
func (c *Compiler) enterScope() {
	scope := CompilationScope{
		instructions:        Instructions{},
		lastInstruction:     EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
	}
	c.scopes = append(c.scopes, scope)
	c.scopeIndex++
	// Create and push a new symbol table for the nested scope
	c.symbolTable = NewEnclosedSymbolTable(c.symbolTable)
}

// leaveScope pops the current compilation scope and returns its instructions.
func (c *Compiler) leaveScope() Instructions {
	instructions := c.currentInstructions()

	c.lastSymbolTable = c.symbolTable // Store table *before* popping

	c.scopes = c.scopes[:len(c.scopes)-1]
	c.scopeIndex--
	// Pop the symbol table
	c.symbolTable = c.symbolTable.Outer
	return instructions
}

// GetFinalSymbolTable returns the symbol table of the *last completed scope*.
// This is typically the outermost scope after compilation finishes.
func (c *Compiler) GetFinalSymbolTable() *SymbolTable {
	// After Compile finishes, the main scope's table should be in lastSymbolTable
	// If Compile ran successfully, c.symbolTable might be nil if root scope was popped.
	if c.lastSymbolTable != nil {
		return c.lastSymbolTable
	}
	// Fallback if leaveScope wasn't called correctly (e.g., compilation error)
	return c.symbolTable
}
// GetNumDefinitions returns the number of definitions from the
// symbol table of the *last completed scope* (usually the main program
// or the last function compiled).
// GetNumDefinitions still uses the logic based on the stored table
func (c *Compiler) GetNumDefinitions() int {
    st := c.GetFinalSymbolTable() // Use the new getter
    if st != nil {
        return st.numDefinitions
    }
    return 0
}

// compileComprehension is a generic helper to compile list/set/dict comprehensions.
func (c *Compiler) compileComprehension(
	compName string,
	element ast.Expression,
	variable *ast.Identifier,
	iterable ast.Expression,
	condition ast.Expression,
	collectionOp Opcode, // OpArray, OpSet, OpHash
	addMethodName string, // "append" for lists, "add" for sets
) error {
	// 1. Compile the comprehension's logic into a function that takes the iterable as an argument.
	c.enterScope()
	c.symbolTable.DefineFunctionName(compName)
	c.symbolTable.Define("_iterable_arg") // This becomes local variable 0 in the new scope

	// a. Create the empty result collection and store it in a local.
	c.emit(collectionOp, 0)
	resultSymbol := c.symbolTable.Define("_result") // This becomes local variable 1
	c.storeSymbol(resultSymbol)

	// b. Get the iterator from the iterable argument (which is local 0).
	iterableSymbol, _ := c.symbolTable.Resolve("_iterable_arg")
	c.loadSymbol(iterableSymbol)
	c.emit(OpGetIter)

	// c. The main loop setup.
	loopStartPos := len(c.currentInstructions())
	opForIterPos := c.emit(OpForIter, 9999) // Pops iterator on finish and jumps to afterLoopPos.

	// d. Inside the loop, store the iterated value in the loop variable.
	loopVarSymbol := c.symbolTable.Define(variable.Value) // This becomes local variable 2
	c.storeSymbol(loopVarSymbol)                          // Pops item from `OpForIter`

	// e. Handle the optional `if` condition.
	var jumpPastAddPos = -1
	if condition != nil {
		err := c.Compile(condition)
		if err != nil {
			c.leaveScope()
			return err
		}
		jumpPastAddPos = c.emit(OpJumpNotTruthy, 9999) // If false, jump over the add logic.
	}

	// f. The logic to add the element to the collection.
	addNameIdx := c.addConstant(&object.String{Value: addMethodName})
	c.loadSymbol(resultSymbol)            // Push the result collection.
	c.emit(OpGetAttribute, addNameIdx)    // Get its `append` or `add` method.
	err := c.Compile(element)             // Compile and push the element to add.
	if err != nil {
		c.leaveScope()
		return err
	}
	c.emit(OpCall, 1) // Call the add/append method.
	c.emit(OpPop)     // Pop the `None` result from the method call.

	// g. Patch the `if` jump (if it exists).
	if condition != nil {
		afterAddPos := len(c.currentInstructions())
		c.changeOperand(jumpPastAddPos, 0, afterAddPos)
	}

	// h. Jump back to the start of the loop.
	c.emit(OpJump, loopStartPos)

	// i. Patch the `for` loop's exit jump to land here.
	afterLoopPos := len(c.currentInstructions())
	c.changeOperand(opForIterPos, 0, afterLoopPos)

	// j. Return the final collection from the hidden function.
	c.loadSymbol(resultSymbol)
	c.emit(OpReturnValue)
	// --- End of hidden function body ---

	// 2. Finalize the compiled function object.
	freeSymbols := c.symbolTable.FreeSymbols
	numLocals := c.symbolTable.numDefinitions
	instructions := c.leaveScope()

	compiledFn := &CompiledFunction{
		Instructions:  instructions,
		NumLocals:     numLocals,
		NumParameters: 1, // Takes one argument: the iterable
		Name:          compName,
	}
	fnConstIdx := c.addConstant(compiledFn)

	// 3. In the outer scope, emit code to create the closure for the hidden function.
	for _, s := range freeSymbols {
		c.loadSymbol(s)
	}
	c.emit(OpClosure, fnConstIdx, len(freeSymbols)) // Pushes the closure onto the stack.

	// 4. Now, compile the iterable expression, pushing the iterable object onto the stack.
	err = c.Compile(iterable)
	if err != nil {
		return err
	}
	// Runtime stack is now: [..., closure, iterable]

	// 5. Call the closure with the iterable as its argument.
	c.emit(OpCall, 1)
	// The result (the final collection) is now on top of the stack.

	return nil
}


// changeOperand modifies a specific operand of an instruction at a given position.
// opPos: The starting byte position of the instruction in the byte
// operandIndex: The zero-based index of the operand within the instruction to modify (0 for first, 1 for second, etc.).
// newValue: The new integer value for the operand.
func (c *Compiler) changeOperand(opPos int, operandIndex int, newValue int) {
	instructions := c.currentInstructions()

	// 1. Basic Bounds Check for opPos
	if opPos < 0 || opPos >= len(instructions) {
		panic(fmt.Sprintf("changeOperand: invalid opPos %d (instruction length %d)", opPos, len(instructions)))
	}

	op := Opcode(instructions[opPos])
	def, err := Lookup(byte(op)) // Get Definition and error
	if err != nil {
		panic(fmt.Sprintf("changeOperand: error looking up opcode %d at pos %d: %v", op, opPos, err))
	}

	// 2. Check if the operandIndex is valid for this opcode
	if operandIndex < 0 || operandIndex >= len(def.OperandWidths) {
		panic(fmt.Sprintf("changeOperand: invalid operand index %d for opcode %s (has %d operands)",
			operandIndex, def.Name, len(def.OperandWidths)))
	}

	// 3. Calculate the byte offset within the instruction for the target operand
	operandOffset := 1 // Start after the opcode byte itself
	for i := 0; i < operandIndex; i++ {
		operandOffset += def.OperandWidths[i] // Add widths of preceding operands
	}

	// 4. Get the width (in bytes) of the target operand
	operandWidth := def.OperandWidths[operandIndex]

	// 5. Check bounds for writing the operand
	writeStartPos := opPos + operandOffset
	writeEndPos := writeStartPos + operandWidth
	if writeEndPos > len(instructions) {
		panic(fmt.Sprintf("changeOperand: calculated write position [%d:%d) is out of bounds for instruction %s at pos %d (total instruction length %d)",
			writeStartPos, writeEndPos, def.Name, opPos, len(instructions)))
	}

	// 6. Write the new value based on the operand width
	switch operandWidth {
	case 1: // uint8
		if newValue < 0 || newValue > 255 {
			panic(fmt.Sprintf("changeOperand: new value %d out of range for uint8 operand %d in %s", newValue, operandIndex, def.Name))
		}
		instructions[writeStartPos] = byte(newValue)
	case 2: // uint16
		if newValue < 0 || newValue > 65535 {
			panic(fmt.Sprintf("changeOperand: new value %d out of range for uint16 operand %d in %s", newValue, operandIndex, def.Name))
		}
		binary.BigEndian.PutUint16(instructions[writeStartPos:writeEndPos], uint16(newValue))
	// Add case 4 for uint32 etc. if needed later
	default:
		// Panic if an unsupported operand width is defined
		panic(fmt.Sprintf("changeOperand: unsupported operand width %d defined for %s", operandWidth, def.Name))
	}
}


// replaceLastPopWithReturn replaces the last OpPop with OpReturn.
func (c *Compiler) replaceLastPopWithReturn() {
	lastPos := c.scopes[c.scopeIndex].lastInstruction.Position
	c.replaceInstruction(lastPos, Make(OpReturn)) // Assumes OpPop and OpReturn have same length (0 operands)
	// Update last instruction tracking
	c.scopes[c.scopeIndex].lastInstruction.Opcode = OpReturn
}

// replaceInstruction replaces bytecode at a given position.
// NOTE: This basic version assumes newInstruction has the same length as replaced one.
func (c *Compiler) replaceInstruction(pos int, newInstruction []byte) {
	ins := c.currentInstructions()
	for i := 0; i < len(newInstruction); i++ {
		if pos+i < len(ins) {
			ins[pos+i] = newInstruction[i]
		} else {
			// This case should ideally not happen if replacing same-length opcodes
			panic("replaceInstruction resulted in out-of-bounds write")
		}
	}
}


// loadSymbol emits the correct Get opcode based on the symbol's scope.
func (c *Compiler) loadSymbol(s Symbol) {
	switch s.Scope {
	case GlobalScope:
		c.emit(OpGetGlobal, s.Index)
	case LocalScope:
		c.emit(OpGetLocal, s.Index)
	case BuiltinScope:
		c.emit(OpGetBuiltin, s.Index)
	case FreeScope:
		c.emit(OpGetFree, s.Index)
	// case FunctionScope: - Represents the function name itself? Usually handled differently.
	default:
		// Should not happen if symbol table is correct
		panic(fmt.Sprintf("Unknown symbol scope: %s for %s", s.Scope, s.Name))
	}
}

// storeSymbol emits the correct Set opcode based on the symbol's scope.
func (c *Compiler) storeSymbol(s Symbol) {
	switch s.Scope {
	case GlobalScope:
		c.emit(OpSetGlobal, s.Index)
	case LocalScope:
		c.emit(OpSetLocal, s.Index)
	case FreeScope:
		c.emit(OpSetFree, s.Index)
	default:
		// Cannot store to Builtin or Function scopes
		panic(fmt.Sprintf("Cannot store symbol with scope %s for %s", s.Scope, s.Name))
	}
}


// --- Symbol Table (Enhanced for Scopes) ---

// Symbol represents a variable or function name in a scope.
type Symbol struct {
	Name  string
	Scope SymbolScope
	Index int // Index within its scope (local index, global index, builtin index, free index)
}

// SymbolScope indicates where a symbol is defined.
type SymbolScope string

const (
	GlobalScope   SymbolScope = "GLOBAL"
	LocalScope    SymbolScope = "LOCAL"
	BuiltinScope  SymbolScope = "BUILTIN"
	FreeScope     SymbolScope = "FREE"
	FunctionScope SymbolScope = "FUNCTION" // For the function's own name
)

// SymbolTable manages symbols for a single scope.
type SymbolTable struct {
	Outer *SymbolTable       // Enclosing scope's table
	store map[string]Symbol  // Symbols defined in *this* scope
	numDefinitions int       // Count of symbols defined *in this scope*
	FreeSymbols    []Symbol  // List of free variables used in this scope
}

// Add necessary getters/methods to SymbolTable if fields aren't exported
func (s *SymbolTable) NumDefinitions() int { return s.numDefinitions } // Example Getter

// Add method to get the root/outermost scope's table
func (s *SymbolTable) OuterMost() *SymbolTable {
    current := s
    for current.Outer != nil {
        current = current.Outer
    }
    return current
}

// Add method to expose the store for iteration (read-only copy recommended)
func (s *SymbolTable) Store() map[string]Symbol {
     // Return a copy to prevent external modification? Or trust caller?
     // For importer, direct access might be okay, but copy is safer.
     storeCopy := make(map[string]Symbol, len(s.store))
     for k, v := range s.store {
         storeCopy[k] = v
     }
     return storeCopy
    // return s.store // Direct access (less safe)
}

// NewSymbolTable creates a root symbol table.
func NewSymbolTable() *SymbolTable {
	s := make(map[string]Symbol)
	free := []Symbol{}
	return &SymbolTable{store: s, FreeSymbols: free}
}

// NewEnclosedSymbolTable creates a nested symbol table.
func NewEnclosedSymbolTable(outer *SymbolTable) *SymbolTable {
	s := NewSymbolTable()
	s.Outer = outer
	return s
}

// Define creates a new symbol in the *current* scope.
func (s *SymbolTable) Define(name string) Symbol {
	// --- BEGIN FIX ---
	// Check if symbol already exists in the *current* scope's store
	symbol, ok := s.store[name]
	if ok {
		// If it exists, just return the existing symbol.
		// We don't redefine or change its index/scope here.
		// Reassignment logic happens via OpSetGlobal/OpSetLocal in the VM.
		return symbol
	}
	// --- END FIX ---
	scope := GlobalScope
	if s.Outer != nil {
		scope = LocalScope // Assume non-global scope is local
	}
	symbol = Symbol{Name: name, Scope: scope, Index: s.numDefinitions}
	s.store[name] = symbol
	s.numDefinitions++
	// // --- ADD DEBUG ---
	// fmt.Printf("<<< DEFINE >>> Name=%s, Scope=%s, Index=%d, Table=%p\n", name, symbol.Scope, symbol.Index, s)
	// // fmt.Printf("  Store contents: %v\n", s.store) // Optional: Print whole map
	// // --- END DEBUG ---
	return symbol
}

// DefineBuiltin defines a built-in name (always in the outermost/global scope ideally).
func (s *SymbolTable) DefineBuiltin(index int, name string) Symbol {
    symbol := Symbol{Name: name, Scope: BuiltinScope, Index: index}
    s.store[name] = symbol // Store in the current table (assumed global for builtins)
    return symbol
}


// DefineFunctionName defines the name for the function itself in its own scope.
// func (s *SymbolTable) DefineFunctionName(name string) Symbol { ... } // TODO if needed
// --- Add DefineFunctionName to SymbolTable ---
// (Optional, but good practice for allowing recursive calls)
func (s *SymbolTable) DefineFunctionName(name string) Symbol {
	symbol := Symbol{Name: name, Scope: FunctionScope, Index: 0} // Index 0 for function name scope
	s.store[name] = symbol
	// Note: This does NOT increment numDefinitions for locals
	return symbol
}

// --- REVISED Resolve ---
func (s *SymbolTable) Resolve(name string) (Symbol, bool) {
	// // --- ADD DEBUG ---
	// fmt.Printf("<<< RESOLVE >>> Name=%s, Table=%p\n", name, s)
	// // fmt.Printf("  Store contents: %v\n", s.store) // Optional: Print whole map
	// // --- END DEBUG ---

	// Try finding in the current scope's store first.
	obj, ok := s.store[name]
	if ok {
		// --- ADD CHECK FOR FUNCTION SCOPE ---
        // Don't mark the function's own name as free within its body
        if obj.Scope == FunctionScope {
            return obj, true // Found the function's own name
        }
        // --- END CHECK ---
		return obj, true // Found locally (could be local, global defined here, builtin defined here, free defined here)
	}

	// If not in local store, and we are in the global scope, it's undefined.
	if s.Outer == nil {
		return obj, false // Not defined globally or as a builtin in this table
	}

	// --- Resolve in Outer Scope (for Local/Free Variables) ---
	// Recursively resolve in the outer scope.
	outerObj, outerOk := s.Outer.Resolve(name)
	if !outerOk {
		return outerObj, false // Not found anywhere up the chain
	}

	// If the object found in the outer scope is Global or Builtin, return it directly.
	// It doesn't become "Free" in the current scope. These scopes are directly accessible.
	if outerObj.Scope == GlobalScope || outerObj.Scope == BuiltinScope {
		return outerObj, true
	}

	// --- Define as Free Variable ---
	// If found in an outer scope and it's NOT Global/Builtin,
	// it means it's a local variable of an enclosing function scope.
	// Therefore, it's a "Free" variable in *this* current scope.
	// We need to define it as such in the *current* scope's table
	// so that the compiler knows to load it using OpGetFree.
	freeSymbol := s.defineFree(outerObj) // defineFree adds it to s.FreeSymbols and s.store
	return freeSymbol, true
}

// --- REVISED defineFree ---
// defineFree adds a resolved outer symbol to the current scope's FreeSymbols list
// and adds a corresponding FreeScope entry to the local store for faster lookups.
func (s *SymbolTable) defineFree(original Symbol) Symbol {
    // Check if already defined as free in this scope
    if existingSymbol, ok := s.store[original.Name]; ok && existingSymbol.Scope == FreeScope {
        return existingSymbol // Already marked as free here
    }

    // Add original symbol to FreeSymbols list for the current scope
    s.FreeSymbols = append(s.FreeSymbols, original)

    // Create a new Symbol entry in the *current* scope's store,
    // marking it as FreeScope and giving it an index within FreeSymbols.
    symbol := Symbol{
        Name:  original.Name,
        Scope: FreeScope,               // Mark it as Free for the current scope
        Index: len(s.FreeSymbols) - 1, // Index in the FreeSymbols slice
    }
    s.store[original.Name] = symbol // Add to *this* scope's store for quick future lookups

    return symbol
}
