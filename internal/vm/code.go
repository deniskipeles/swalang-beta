package vm

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// Instructions is the compiled bytecode sequence
type Instructions []byte

// String makes Instructions readable (for debugging)
func (ins Instructions) String() string {
	var out bytes.Buffer
	i := 0
	for i < len(ins) {
		def, err := Lookup(ins[i]) // ins[i] is the opcode byte
		if err != nil {
			fmt.Fprintf(&out, "ERROR: %s\n", err)
			// Attempt to recover/continue printing? Maybe just print the byte and move on.
			fmt.Fprintf(&out, "%04d Error reading opcode %d\n", i, ins[i])
			i++ // Move past the problematic byte
			continue
		}
		operands, read := ReadOperands(def, ins[i+1:])
		fmt.Fprintf(&out, "%04d %s\n", i, ins.fmtInstruction(def, operands))
		i += 1 + read // Move past opcode and operands
	}
	return out.String()
}

// fmtInstruction formats a single instruction with its operands.
func (ins Instructions) fmtInstruction(def *Definition, operands []int) string {
	operandCount := len(def.OperandWidths)
	if len(operands) != operandCount {
		return fmt.Sprintf("ERROR: operand len %d does not match defined %d for %s\n",
			len(operands), operandCount, def.Name)
	}
	switch operandCount {
	case 0:
		return def.Name
	case 1:
		return fmt.Sprintf("%s %d", def.Name, operands[0])
	case 2:
		return fmt.Sprintf("%s %d %d", def.Name, operands[0], operands[1])
		// Add case 3 etc. if needed for opcodes with more operands
	}
	return fmt.Sprintf("ERROR: unhandled operandCount %d for %s\n", operandCount, def.Name)
}

// Opcode is a single byte representing an operation
type Opcode byte

// Define Opcodes based on interpreter capabilities
const (
	// --- Constants and Literals ---
	OpConstant Opcode = iota // Push constant from pool (index uint16)
	OpTrue                   // Push True object
	OpFalse                  // Push False object
	OpNull                   // Push Null object (None)
	OpDup

	// --- Stack Manipulation ---
	OpPop // Pop the top element

	// --- Arithmetic Operations ---
	OpAdd      // TOS1 + TOS -> Push result
	OpSubtract // TOS1 - TOS -> Push result
	OpMultiply // TOS1 * TOS -> Push result
	OpDivide   // TOS1 / TOS -> Push result (potentially float)
	OpModulo   // TOS1 % TOS -> Push result

	// --- Comparison Operations ---
	OpEqual            // TOS1 == TOS -> Push Boolean
	OpNotEqual         // TOS1 != TOS -> Push Boolean
	OpGreaterThan      // TOS1 > TOS -> Push Boolean
	OpGreaterThanEqual // TOS1 >= TOS -> Push Boolean
	OpLesserThan       // TOS1 < TOS -> Push Boolean (Added for completeness)
	OpLesserThanEqual  // TOS1 <= TOS -> Push Boolean (Added for completeness)
	OpIn               // TOS is in TOS1 -> Push Boolean

	// --- Logical Operations ---
	// 'and'/'or' are handled via jumps, 'not' uses OpBang
	OpBang // !TOS -> Push Boolean (Logical NOT)

	// --- Prefix Operations ---
	OpMinus // -TOS -> Push result (Unary minus)
	OpPos   // +TOS -> Push result (Unary plus)

	// --- Jumps ---
	OpJump          // Unconditional jump (offset uint16)
	OpJumpNotTruthy // Pop TOS; Jump if falsy (offset uint16)
	OpJumpIfTruthy  // <<< NEW: Pop TOS; Jump if truthy (offset uint16)
	// OpJumpIfTrue  // Pop TOS; Jump if truthy (offset uint16) - Alternative/optional

	// --- Variable Access ---
	OpSetGlobal  // Pop TOS; globals[uint16] = TOS;
	OpGetGlobal  // Push globals[uint16]
	OpSetLocal   // Pop TOS; frame.locals[uint8] = TOS;
	OpGetLocal   // Push frame.locals[uint8]
	OpSetFree    // Pop TOS; frame.closure.freeVars[uint8] = TOS;
	OpGetFree    // Push frame.closure.freeVars[uint8]
	OpGetBuiltin // Push builtins[uint8]
	OpStoreName  // Like SetGlobal/Local but uses name index from consts? Needed? Let's use Set for now.
	OpImportName // Import module by name const[uint16], push module obj
	OpLoadName   // Generic load? Maybe replace GetGlobal/Local? Stick with specific for now.

	// --- Collection/Data Structure Operations ---
	OpArray          // Build list from stack (count uint16)
	OpTuple          // Build tuple from stack (count uint16)
	OpSet            // Build set from stack (count uint16)
	OpHash           // Build dict from stack (pair count uint16 -> 2*count items)
	OpIndex          // Pop index, pop collection; Push collection[index]
	OpSetIndex       // Pop value, pop index, pop collection; collection[index] = value
	OpGetIter        // Pop iterable; Push iterator
	OpForIter        // Pop iterator; Call next(); If StopIteration jump(uint16); else push value; push iterator back
	OpUnpackSequence // Pop sequence; Push elements onto stack (count uint8) - For assignment like a,b=iterable

	// --- Function/Method/Class Operations ---
	OpCall         // Pop N args (uint8), pop callable; Call callable; Push result
	OpReturnValue  // Pop TOS; Return from current frame
	OpReturn       // Return Null from current frame
	OpClosure      // Push closure object (const func index uint16, free var count uint8)
	OpClass        // Define class (name index uint16, superclass count uint8?); Pop methods/vars; Push class obj
	OpGetAttribute // Pop object; Push object.name (name index uint16)
	OpSetAttribute // Pop value, pop object; Set object.name = value (name index uint16)

	OpBuildClass Opcode = iota // Pop namespace dict, pop name obj, build class, push class obj.

	// --- NEW: Exception Handling Opcodes ---
	OpPushExceptionHandler // Push handler info (ExceptTarget uint16, FinallyTarget uint16)
	OpPopExceptionHandler  // Pop the current exception handler
	OpRaise                // Re-raise the exception currently on stack (or create new one?) - Placeholder

	// --- Helpers used by Exception Handling ---
	OpSwap       // Swaps top two stack items
	OpIsInstance // Pops type, pops object, pushes isinstance(obj, type)

	// --- Helper for Except Blocks (Optional but helpful) ---
	// OpMatchException // Peeks exception, pops type, pushes bool? (Or combine logic in VM)
	// Let's skip OpMatchException for now and do checks in VM when needed.

	// --- Exception Handling (Basic) ---
	OpSetupExcept // Push exception handler block (jump offset uint16)
	OpPopBlock    // Pop exception handler block from stack
	// OpRaise       // Raise exception? (Not implemented in interpreter yet)
)

// Definition describes an opcode (name and operand widths in bytes)
type Definition struct {
	Name          string
	OperandWidths []int // Number of bytes each operand takes
}

// definitions maps Opcodes to their Definitions.
// Operand widths: 1 = uint8, 2 = uint16
var definitions = map[Opcode]*Definition{
	// --- Constants and Literals ---
	OpConstant: {"OpConstant", []int{2}}, // uint16 index
	OpTrue:     {"OpTrue", []int{}},
	OpFalse:    {"OpFalse", []int{}},
	OpNull:     {"OpNull", []int{}},
	OpDup:      {"OpDup", []int{}},

	// --- Stack Manipulation ---
	OpPop: {"OpPop", []int{}},

	// --- Arithmetic Operations ---
	OpAdd:      {"OpAdd", []int{}},
	OpSubtract: {"OpSubtract", []int{}},
	OpMultiply: {"OpMultiply", []int{}},
	OpDivide:   {"OpDivide", []int{}},
	OpModulo:   {"OpModulo", []int{}},

	// --- Comparison Operations ---
	OpEqual:            {"OpEqual", []int{}},
	OpNotEqual:         {"OpNotEqual", []int{}},
	OpGreaterThan:      {"OpGreaterThan", []int{}},
	OpGreaterThanEqual: {"OpGreaterThanEqual", []int{}},
	OpLesserThan:       {"OpLesserThan", []int{}},
	OpLesserThanEqual:  {"OpLesserThanEqual", []int{}},
	OpIn:               {"OpIn", []int{}},

	// --- Logical Operations ---
	OpBang: {"OpBang", []int{}},

	// --- Prefix Operations ---
	OpMinus: {"OpMinus", []int{}},
	OpPos:   {"OpPos", []int{}},

	// --- Jumps ---
	OpJump:          {"OpJump", []int{2}},          // uint16 offset
	OpJumpNotTruthy: {"OpJumpNotTruthy", []int{2}}, // uint16 offset
	OpJumpIfTruthy:  {"OpJumpIfTruthy", []int{2}},  // <<< NEW DEFINITION (uint16 offset)

	// --- Variable Access ---
	OpSetGlobal:  {"OpSetGlobal", []int{2}},  // uint16 index
	OpGetGlobal:  {"OpGetGlobal", []int{2}},  // uint16 index
	OpSetLocal:   {"OpSetLocal", []int{1}},   // uint8 index
	OpGetLocal:   {"OpGetLocal", []int{1}},   // uint8 index
	OpSetFree:    {"OpSetFree", []int{1}},    // uint8 index
	OpGetFree:    {"OpGetFree", []int{1}},    // uint8 index
	OpGetBuiltin: {"OpGetBuiltin", []int{1}}, // uint8 index
	OpImportName: {"OpImportName", []int{2}}, // uint16 name const index

	// --- Collection/Data Structure Operations ---
	OpArray:          {"OpArray", []int{2}}, // uint16 count
	OpTuple:          {"OpTuple", []int{2}}, // uint16 count
	OpSet:            {"OpSet", []int{2}},   // uint16 count
	OpHash:           {"OpHash", []int{2}},  // uint16 pair count
	OpIndex:          {"OpIndex", []int{}},
	OpSetIndex:       {"OpSetIndex", []int{}},
	OpGetIter:        {"OpGetIter", []int{}},
	OpForIter:        {"OpForIter", []int{2}},        // uint16 jump offset if iteration stops
	OpUnpackSequence: {"OpUnpackSequence", []int{1}}, // uint8 count

	// --- Function/Method/Class Operations ---
	OpCall:         {"OpCall", []int{1}}, // uint8 argc
	OpReturnValue:  {"OpReturnValue", []int{}},
	OpReturn:       {"OpReturn", []int{}},
	OpClosure:      {"OpClosure", []int{2, 1}},   // uint16 func const index, uint8 free var count
	OpClass:        {"OpClass", []int{2}},        // uint16 name const index
	OpGetAttribute: {"OpGetAttribute", []int{2}}, // uint16 name const index
	OpSetAttribute: {"OpSetAttribute", []int{2}}, // uint16 name const index

	// --- BEGIN FIX ---
	// OpBuildClass takes one 2-byte operand: the number of methods pushed before it
	OpBuildClass: {"OpBuildClass", []int{2}},
	// --- END FIX ---

	// --- NEW: Exception Handling Definitions ---
	// Target IP if exception occurs in try, Target IP for finally (0 if none)
	OpPushExceptionHandler: {"OpPushExceptionHandler", []int{2, 2}},
	OpPopExceptionHandler:  {"OpPopExceptionHandler", []int{}},
	OpRaise:                {"OpRaise", []int{}}, // May take operand later for `raise ErrorType`

	OpSwap:       {"OpSwap", []int{}},
	OpIsInstance: {"OpIsInstance", []int{}},

	// --- Exception Handling (Basic) ---
	OpSetupExcept: {"OpSetupExcept", []int{2}}, // uint16 jump offset
	OpPopBlock:    {"OpPopBlock", []int{}},
}

// Lookup finds the Definition for a given opcode byte
func Lookup(op byte) (*Definition, error) {
	def, ok := definitions[Opcode(op)]
	if !ok {
		return nil, fmt.Errorf("opcode %d undefined", op)
	}
	return def, nil
}

// Make creates a bytecode instruction slice from an opcode and operands.
// It packs operands according to the widths defined for the opcode.
func Make(op Opcode, operands ...int) []byte {
	def, ok := definitions[op]
	if !ok {
		// Return empty slice or potentially panic if definition is missing
		// for a known Opcode constant. Returning empty might hide errors.
		// Panic might be better during development.
		panic(fmt.Sprintf("opcode %d definition not found", op))
		// return []byte{}
	}

	// Calculate total instruction length
	instructionLen := 1 // Start with 1 byte for the opcode itself
	if len(operands) != len(def.OperandWidths) {
		// Panic if the number of provided operands doesn't match definition
		panic(fmt.Sprintf("operand count mismatch for %s: expected %d, got %d",
			def.Name, len(def.OperandWidths), len(operands)))
	}
	for _, w := range def.OperandWidths {
		instructionLen += w
	}

	instruction := make([]byte, instructionLen)
	instruction[0] = byte(op) // Set the opcode byte

	offset := 1
	for i, o := range operands {
		width := def.OperandWidths[i]
		switch width {
		case 1: // uint8
			if o < 0 || o > 255 {
				panic(fmt.Sprintf("operand %d value %d out of range for uint8 width in %s", i, o, def.Name))
			}
			instruction[offset] = byte(o)
		case 2: // uint16
			if o < 0 || o > 65535 {
				panic(fmt.Sprintf("operand %d value %d out of range for uint16 width in %s", i, o, def.Name))
			}
			binary.BigEndian.PutUint16(instruction[offset:], uint16(o))
		// Add case 4 for uint32 etc. if needed later
		default:
			// Panic if an unsupported operand width is defined
			panic(fmt.Sprintf("unsupported operand width %d defined for %s", width, def.Name))
		}
		offset += width
	}
	return instruction
}

// ReadOperands decodes operands from a bytecode instruction slice based on Definition.
// Returns a slice of operands (as int) and the number of bytes read.
func ReadOperands(def *Definition, ins Instructions) ([]int, int) {
	operands := make([]int, len(def.OperandWidths))
	offset := 0
	for i, width := range def.OperandWidths {
		// Check if there are enough bytes left in `ins`
		if offset+width > len(ins) {
			// This indicates truncated bytecode, return error or handle?
			// For now, maybe return partial results and signal error via read count?
			// Let's return what we could read and let the caller decide.
			// Or just return an empty slice and 0?
			// Returning partials might be confusing. Let's return nil and 0.
			// fmt.Printf("WARN: ReadOperands insufficient bytes for width %d at offset %d (len=%d)\n", width, offset, len(ins))
			return nil, 0 // Or maybe return operands[:i], offset?
		}

		switch width {
		case 1:
			operands[i] = int(ReadUint8(ins[offset:]))
		case 2:
			operands[i] = int(ReadUint16(ins[offset:]))
		// Add case 4 etc. if needed later
		default:
			// Should not happen if definitions are correct
			// fmt.Printf("WARN: ReadOperands unsupported width %d\n", width)
			return nil, 0 // Indicate error
		}
		offset += width
	}
	return operands, offset
}

// --- Read Helpers ---

// ReadUint16 reads a uint16 from a byte slice using BigEndian order.
// Assumes the slice has at least 2 bytes.
func ReadUint16(ins Instructions) uint16 {
	return binary.BigEndian.Uint16(ins)
}

// ReadUint8 reads a uint8 (byte) from a byte slice.
// Assumes the slice has at least 1 byte.
func ReadUint8(ins Instructions) uint8 {
	return ins[0]
}
