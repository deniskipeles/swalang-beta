package object

import (
	"fmt"
	"hash/fnv"
	"strconv"
	"strings"
	"unicode/utf8" // For String.__len__ if implemented explicitly, and already used in builtins.pyLenFn

	"github.com/deniskipeles/pylearn/internal/constants"
)

// String
type String struct{ Value string }

func (s *String) Type() ObjectType { return STRING_OBJ }
func (s *String) Inspect() string  { return fmt.Sprintf(constants.STRING_INSPECT_FORMAT, s.Value) }
func (s *String) HashKey() (HashKey, error) {
	h := fnv.New64a()
	_, _ = h.Write([]byte(s.Value))
	return HashKey{Type: s.Type(), Value: h.Sum64()}, nil
}

// --- Implement Attribute Access using Closures within Builtins ---
// ... (within object.go, after String struct definition and its other methods) ...

// GetObjectAttribute for String to expose methods
func (s *String) GetObjectAttribute(ctx ExecutionContext, name string) (Object, bool) {
	// Helper to create the method Builtin objects for String
	makeStringMethod := func(methodName string, goFn BuiltinFunction) *Builtin {
		return &Builtin{
			Name: constants.STRING_METHOD_PREFIX + methodName,
			Fn: func(callCtx ExecutionContext, scriptProvidedArgs ...Object) Object {
				methodArgs := make([]Object, 0, 1+len(scriptProvidedArgs))
				methodArgs = append(methodArgs, s) // Prepend self (the String 's')
				methodArgs = append(methodArgs, scriptProvidedArgs...)
				return goFn(callCtx, methodArgs...)
			},
		}
	}

	switch name {
	case constants.STRING_UPPER_METHOD_NAME:
		return makeStringMethod(constants.STRING_UPPER_METHOD_NAME, func(c ExecutionContext, a ...Object) Object { /* From previous implementation */
			self := a[0].(*String)
			if len(a) != 1 {
				return NewError(constants.TypeError, constants.STRING_UPPER_ARG_COUNT_ERROR)
			}
			return &String{Value: strings.ToUpper(self.Value)}
		}), true
	case constants.STRING_LOWER_METHOD_NAME:
		return makeStringMethod(constants.STRING_LOWER_METHOD_NAME, func(c ExecutionContext, a ...Object) Object { /* From previous implementation */
			self := a[0].(*String)
			if len(a) != 1 {
				return NewError(constants.TypeError, constants.STRING_LOWER_ARG_COUNT_ERROR)
			}
			return &String{Value: strings.ToLower(self.Value)}
		}), true
	case constants.STRING_STARTSWITH_METHOD_NAME: // Already existed, ensure it's using the new makeStringMethod helper or its logic
		return makeStringMethod(constants.STRING_STARTSWITH_METHOD_NAME, pyStringStartswithFn), true
	// New methods:
	case constants.STRING_SPLIT_METHOD_NAME:
		return makeStringMethod(constants.STRING_SPLIT_METHOD_NAME, pyStringSplitFn), true
	case constants.STRING_JOIN_METHOD_NAME:
		return makeStringMethod(constants.STRING_JOIN_METHOD_NAME, pyStringJoinFn), true
	case constants.STRING_STRIP_METHOD_NAME:
		return makeStringMethod(constants.STRING_STRIP_METHOD_NAME, func(c ExecutionContext, a ...Object) Object {
			return pyStringGenericStripFn(constants.STRING_STRIP_METHOD_NAME, c, a...)
		}), true
	case constants.STRING_LSTRIP_METHOD_NAME:
		return makeStringMethod(constants.STRING_LSTRIP_METHOD_NAME, func(c ExecutionContext, a ...Object) Object {
			return pyStringGenericStripFn(constants.STRING_LSTRIP_METHOD_NAME, c, a...)
		}), true
	case constants.STRING_RSTRIP_METHOD_NAME:
		return makeStringMethod(constants.STRING_RSTRIP_METHOD_NAME, func(c ExecutionContext, a ...Object) Object {
			return pyStringGenericStripFn(constants.STRING_RSTRIP_METHOD_NAME, c, a...)
		}), true
	case constants.STRING_ENDSWITH_METHOD_NAME:
		return makeStringMethod(constants.STRING_ENDSWITH_METHOD_NAME, pyStringEndswithFn), true
	case constants.STRING_REPLACE_METHOD_NAME:
		return makeStringMethod(constants.STRING_REPLACE_METHOD_NAME, pyStringReplaceFn), true
	case constants.STRING_FIND_METHOD_NAME:
		return makeStringMethod(constants.STRING_FIND_METHOD_NAME, pyStringFindFn), true
	case constants.STRING_ENCODE_METHOD_NAME:
		return makeStringMethod(constants.STRING_ENCODE_METHOD_NAME, pyStringEncodeFn), true
	case constants.STRING_FORMAT_METHOD_NAME: // New method
		return makeStringMethod(constants.STRING_FORMAT_METHOD_NAME, pyStringFormatFn), true
	case constants.DunderLen:
		return makeStringMethod(constants.DunderLen, pyStringLenFn), true
	case constants.DunderContains:
		return makeStringMethod(constants.DunderContains, pyStringContainsFn), true
	case constants.DunderMul:
		return makeStringMethod(constants.DunderMul, pyStringMulFn), true
	case constants.DunderRMul: // For `int * str`
		return makeStringMethod(constants.DunderRMul, pyStringMulFn), true
		// __add__ is handled by the infix operator logic usually, but can be added here if needed for explicit calls.
	}
	return nil, false
}

// Ensure implementation check is present within the object package
var _ AttributeGetter = (*String)(nil)

// String Item Access (Read-only)
func (s *String) GetObjectItem(key Object) Object {
	idxObj, ok := key.(*Integer)
	if !ok {
		return NewError(constants.TypeError, constants.STRING_ITEM_INDEX_TYPE_ERROR, key.Type())
	}
	idx := idxObj.Value
	// Use runes for proper Unicode indexing
	runes := []rune(s.Value)
	strLen := int64(len(runes))
	if idx < 0 {
		idx += strLen
	} // Handle negative index
	if idx < 0 || idx >= strLen {
		return NewError(constants.IndexError, constants.STRING_ITEM_INDEX_OUT_OF_RANGE)
	}
	return &String{Value: string(runes[idx])}
}

// TODO: String Attribute Access (e.g., .upper(), .lower(), .split())
// func (s *String) GetObjectAttribute(name string) (Object, bool) { ... }
var _ Object = (*String)(nil)
var _ Hashable = (*String)(nil)
var _ ItemGetter = (*String)(nil) // String implements item getting

// --- Go functions for String methods ---

// pyStringMulFn implements string * int
func pyStringMulFn(ctx ExecutionContext, args ...Object) Object {
	// For both str * int and int * str, the arguments will be (string, int)
	// due to the dispatch logic prioritizing the string's dunder method.
	selfStr, okStr := args[0].(*String)
	countObj, okInt := args[1].(*Integer)

	if !okStr || !okInt {
		// This path should ideally not be taken if the infix evaluation logic is correct,
		// but serves as a safeguard.
		return NewError(constants.TypeError, "unsupported operand type(s) for *: '%s' and '%s'", args[0].Type(), args[1].Type())
	}

	count := int(countObj.Value)
	if count < 0 {
		count = 0 // Multiplying by a negative number results in an empty string
	}

	// Use the efficient strings.Repeat function from Go's standard library.
	return &String{Value: strings.Repeat(selfStr.Value, count)}
}

// pyStringEncodeFn implements string.encode(encoding='utf-8', errors='strict')
func pyStringEncodeFn(ctx ExecutionContext, args ...Object) Object {
	// args[0] is self (String object)
	// args[1] is encoding (optional)
	// args[2] is errors (optional, ignored for now)
	if len(args) < 1 {
		return NewError(constants.InternalError, constants.STRING_ENCODE_ARG_COUNT_ERROR, len(args)-1)
	}
	selfStr, ok := args[0].(*String)
	if !ok {
		return NewError(constants.TypeError, constants.STRING_ENCODE_ON_STRING_ERROR)
	}

	// For now, we only support utf-8 and will ignore the encoding argument.
	// A full implementation would check args[1] if it exists.
	// if len(args) > 1 {
	//   // check encoding type and value
	// }

	return &Bytes{Value: []byte(selfStr.Value)}
}

// pyStringSplitFn implements string.split(sep=None, maxsplit=-1)
func pyStringSplitFn(ctx ExecutionContext, args ...Object) Object {
	// args[0] is self (the String object)
	selfStr, ok := args[0].(*String)
	if !ok {
		return NewError(constants.TypeError, constants.STRING_SPLIT_ON_STRING_ERROR)
	}

	var sepOpt Object = nil // Sentinel for None separator
	var maxsplitOpt Object = nil
	numScriptArgs := len(args) - 1 // Number of arguments provided by the script

	if numScriptArgs > 2 {
		return NewError(constants.TypeError, constants.STRING_SPLIT_ARG_COUNT_ERROR, numScriptArgs)
	}
	if numScriptArgs >= 1 {
		sepOpt = args[1]
	}
	if numScriptArgs == 2 {
		maxsplitOpt = args[2]
	}

	// Process separator
	var sepArg string = constants.EmptyString // Empty string means split by whitespace (special case for strings.Fields)
	sepIsNone := true                         // Python's default split by whitespace
	if sepOpt != nil && sepOpt != NULL {
		sepStr, okSep := sepOpt.(*String)
		if !okSep {
			return NewError(constants.TypeError, constants.STRING_SPLIT_SEP_TYPE_ERROR, sepOpt.Type())
		}
		sepArg = sepStr.Value
		sepIsNone = false
	}

	// Process maxsplit
	maxsplitArg := -1 // Default: split all occurrences
	if maxsplitOpt != nil && maxsplitOpt != NULL {
		maxsplitInt, okMax := maxsplitOpt.(*Integer)
		if !okMax {
			return NewError(constants.TypeError, constants.STRING_SPLIT_MAXSPLIT_TYPE_ERROR, maxsplitOpt.Type())
		}
		maxsplitArg = int(maxsplitInt.Value)
	}

	var result []string
	source := selfStr.Value

	if sepIsNone { // Default behavior: split by whitespace, discard empty strings
		if maxsplitArg == -1 {
			result = strings.Fields(source)
		} else {
			// strings.Fields doesn't support maxsplit. We need to emulate.
			// This is a simplified emulation. A more robust one handles consecutive whitespace correctly.
			tempResult := strings.Fields(source)
			if maxsplitArg >= 0 && len(tempResult) > maxsplitArg {
				// Join the remaining parts
				lastPart := strings.Join(tempResult[maxsplitArg:], constants.Space) // This assumes single space original separator
				result = append(tempResult[:maxsplitArg], lastPart)
			} else {
				result = tempResult
			}
		}
	} else { // Specific separator
		if maxsplitArg == -1 {
			result = strings.Split(source, sepArg)
		} else {
			if maxsplitArg < 0 { // Python treats negative maxsplit like -1 (split all)
				result = strings.Split(source, sepArg)
			} else {
				result = strings.SplitN(source, sepArg, maxsplitArg+1)
			}
		}
	}

	elements := make([]Object, len(result))
	for i, s := range result {
		elements[i] = &String{Value: s}
	}
	return &List{Elements: elements}
}

// pyStringJoinFn implements separator_string.join(iterable_of_strings)
func pyStringJoinFn(ctx ExecutionContext, args ...Object) Object {
	// args[0] is self (the separator String object)
	// args[1] is the iterable
	if len(args) != 2 {
		return NewError(constants.TypeError, constants.STRING_JOIN_ARG_COUNT_ERROR, len(args)-1)
	}
	separatorStr, ok := args[0].(*String)
	if !ok {
		return NewError(constants.TypeError, constants.STRING_JOIN_ON_STRING_ERROR)
	}
	iterable := args[1]

	var elementsToJoin []string

	// Handle List and Tuple directly
	switch iter := iterable.(type) {
	case *List:
		elementsToJoin = make([]string, len(iter.Elements))
		for i, item := range iter.Elements {
			itemStr, okItem := item.(*String)
			if !okItem {
				return NewError(constants.TypeError, constants.STRING_SEQUENCE_ITEM_STR_ERROR, i, item.Type())
			}
			elementsToJoin[i] = itemStr.Value
		}
	case *Tuple: // Assuming Tuple exists and is similar to List
		elementsToJoin = make([]string, len(iter.Elements))
		for i, item := range iter.Elements {
			itemStr, okItem := item.(*String)
			if !okItem {
				return NewError(constants.TypeError, constants.STRING_SEQUENCE_ITEM_STR_ERROR, i, item.Type())
			}
			elementsToJoin[i] = itemStr.Value
		}
	default:
		// TODO: Support generic iterables by using the iterator protocol
		return NewError(constants.TypeError, constants.STRING_JOIN_ITERABLE_TYPE_ERROR, iterable.Type())
	}

	return &String{Value: strings.Join(elementsToJoin, separatorStr.Value)}
}

// pyStringStripFn, pyStringLStripFn, pyStringRStripFn
func pyStringGenericStripFn(stripType string, _ ExecutionContext, args ...Object) Object {
	// args[0] is self (String)
	// args[1] (optional) is chars (String)
	selfStr, ok := args[0].(*String)
	if !ok {
		return NewError(constants.TypeError, constants.STRING_STRIP_ON_STRING_ERROR, stripType)
	}
	numScriptArgs := len(args) - 1
	if numScriptArgs > 1 {
		return NewError(constants.TypeError, constants.STRING_STRIP_ARG_COUNT_ERROR, stripType, numScriptArgs)
	}

	charsToStrip := constants.EmptyString // Default is whitespace
	if numScriptArgs == 1 {
		if args[1] != NULL { // None means default whitespace
			charsStr, okChars := args[1].(*String)
			if !okChars {
				return NewError(constants.TypeError, constants.STRING_STRIP_CHARS_TYPE_ERROR, stripType, args[1].Type())
			}
			charsToStrip = charsStr.Value
		}
	}

	var result string
	switch stripType {
	case constants.STRING_STRIP_METHOD_NAME:
		if charsToStrip == constants.EmptyString {
			result = strings.TrimSpace(selfStr.Value)
		} else {
			result = strings.Trim(selfStr.Value, charsToStrip)
		}
	case constants.STRING_LSTRIP_METHOD_NAME:
		if charsToStrip == constants.EmptyString {
			result = strings.TrimLeftFunc(selfStr.Value, func(r rune) bool {
				return r == constants.SpaceRune || r == constants.TabRune || r == constants.NewlineRune || r == constants.CarriageReturnRune || r == constants.VerticalTabRune || r == constants.FormFeedRune
			}) // More precise whitespace
		} else {
			result = strings.TrimLeft(selfStr.Value, charsToStrip)
		}
	case constants.STRING_RSTRIP_METHOD_NAME:
		if charsToStrip == constants.EmptyString {
			result = strings.TrimRightFunc(selfStr.Value, func(r rune) bool {
				return r == constants.SpaceRune || r == constants.TabRune || r == constants.NewlineRune || r == constants.CarriageReturnRune || r == constants.VerticalTabRune || r == constants.FormFeedRune
			})
		} else {
			result = strings.TrimRight(selfStr.Value, charsToStrip)
		}
	default:
		return NewError(constants.InternalError, constants.STRING_STRIP_UNKNOWN_TYPE_ERROR, stripType)
	}
	return &String{Value: result}
}

// pyStringStartswithFn implements string.startswith(prefix, start=0, end=len(string))
func pyStringStartswithFn(ctx ExecutionContext, args ...Object) Object {
	// args[0] self, args[1] prefix, args[2] start (opt), args[3] end (opt)
	selfStr, ok := args[0].(*String)
	if !ok {
		return NewError(constants.TypeError, constants.STRING_STARTSWITH_ON_STRING_ERROR)
	}

	numScriptArgs := len(args) - 1
	if numScriptArgs < 1 || numScriptArgs > 3 {
		return NewError(constants.TypeError, constants.STRING_STARTSWITH_ARG_COUNT_ERROR, numScriptArgs)
	}

	prefixStr, okPrefix := args[1].(*String)
	if !okPrefix {
		// Python allows tuple of strings for prefix, not implemented here yet
		return NewError(constants.TypeError, constants.STRING_STARTSWITH_PREFIX_TYPE_ERROR, args[1].Type())
	}

	sourceRunes := []rune(selfStr.Value)
	startIdx, endIdx := 0, len(sourceRunes)

	if numScriptArgs >= 2 && args[2] != NULL {
		startInt, okStart := args[2].(*Integer)
		if !okStart {
			return NewError(constants.TypeError, constants.STRING_SLICE_INDICES_TYPE_ERROR, args[2].Type())
		}
		startIdx = int(startInt.Value)
	}
	if numScriptArgs == 3 && args[3] != NULL {
		endInt, okEnd := args[3].(*Integer)
		if !okEnd {
			return NewError(constants.TypeError, constants.STRING_SLICE_INDICES_TYPE_ERROR, args[3].Type())
		}
		endIdx = int(endInt.Value)
	}

	// Python slice semantics for start/end
	if startIdx < 0 {
		startIdx = len(sourceRunes) + startIdx
	}
	if startIdx < 0 {
		startIdx = 0
	}
	if startIdx > len(sourceRunes) {
		startIdx = len(sourceRunes)
	}

	if endIdx < 0 {
		endIdx = len(sourceRunes) + endIdx
	}
	if endIdx < 0 {
		endIdx = 0
	}
	if endIdx > len(sourceRunes) {
		endIdx = len(sourceRunes)
	}

	if startIdx > endIdx { // Empty slice if start > end
		return FALSE
	}

	subString := string(sourceRunes[startIdx:endIdx])
	return NativeBoolToBooleanObject(strings.HasPrefix(subString, prefixStr.Value))
}

// pyStringEndswithFn is analogous to pyStringStartswithFn
func pyStringEndswithFn(ctx ExecutionContext, args ...Object) Object {
	selfStr, ok := args[0].(*String)
	if !ok {
		return NewError(constants.TypeError, constants.STRING_ENDSWITH_ON_STRING_ERROR)
	}

	numScriptArgs := len(args) - 1
	if numScriptArgs < 1 || numScriptArgs > 3 {
		return NewError(constants.TypeError, constants.STRING_ENDSWITH_ARG_COUNT_ERROR, numScriptArgs)
	}

	suffixStr, okSuffix := args[1].(*String)
	if !okSuffix {
		return NewError(constants.TypeError, constants.STRING_ENDSWITH_SUFFIX_TYPE_ERROR, args[1].Type())
	}

	sourceRunes := []rune(selfStr.Value)
	startIdx, endIdx := 0, len(sourceRunes)

	if numScriptArgs >= 2 && args[2] != NULL {
		startInt, okStart := args[2].(*Integer)
		if !okStart {
			return NewError(constants.TypeError, constants.STRING_SLICE_INDICES_TYPE_ERROR, args[2].Type())
		}
		startIdx = int(startInt.Value)
	}
	if numScriptArgs == 3 && args[3] != NULL {
		endInt, okEnd := args[3].(*Integer)
		if !okEnd {
			return NewError(constants.TypeError, constants.STRING_SLICE_INDICES_TYPE_ERROR, args[3].Type())
		}
		endIdx = int(endInt.Value)
	}
	// Adjust slice indices (Python semantics)
	if startIdx < 0 {
		startIdx = len(sourceRunes) + startIdx
	}
	if startIdx < 0 {
		startIdx = 0
	}
	if startIdx > len(sourceRunes) {
		startIdx = len(sourceRunes)
	}

	if endIdx < 0 {
		endIdx = len(sourceRunes) + endIdx
	}
	if endIdx < 0 {
		endIdx = 0
	}
	if endIdx > len(sourceRunes) {
		endIdx = len(sourceRunes)
	}

	if startIdx > endIdx {
		return FALSE
	}

	subString := string(sourceRunes[startIdx:endIdx])
	return NativeBoolToBooleanObject(strings.HasSuffix(subString, suffixStr.Value))
}

// pyStringReplaceFn implements string.replace(old, new, count=-1)
func pyStringReplaceFn(ctx ExecutionContext, args ...Object) Object {
	// args[0] self, args[1] old, args[2] new, args[3] count (opt)
	selfStr, ok := args[0].(*String)
	if !ok {
		return NewError(constants.TypeError, constants.STRING_REPLACE_ON_STRING_ERROR)
	}

	numScriptArgs := len(args) - 1
	if numScriptArgs < 2 || numScriptArgs > 3 {
		return NewError(constants.TypeError, constants.STRING_REPLACE_ARG_COUNT_ERROR, numScriptArgs)
	}

	oldSub, okOld := args[1].(*String)
	if !okOld {
		return NewError(constants.TypeError, constants.STRING_REPLACE_OLD_TYPE_ERROR, args[1].Type())
	}
	newSub, okNew := args[2].(*String)
	if !okNew {
		return NewError(constants.TypeError, constants.STRING_REPLACE_NEW_TYPE_ERROR, args[2].Type())
	}

	countArg := -1 // Default: replace all
	if numScriptArgs == 3 {
		if args[3] != NULL {
			countInt, okCount := args[3].(*Integer)
			if !okCount {
				return NewError(constants.TypeError, constants.STRING_REPLACE_COUNT_TYPE_ERROR, args[3].Type())
			}
			countArg = int(countInt.Value)
		}
	}
	return &String{Value: strings.Replace(selfStr.Value, oldSub.Value, newSub.Value, countArg)}
}

// pyStringFindFn implements string.find(sub, start=0, end=len(string))
func pyStringFindFn(ctx ExecutionContext, args ...Object) Object {
	selfStr, ok := args[0].(*String)
	if !ok {
		return NewError(constants.TypeError, constants.STRING_FIND_ON_STRING_ERROR)
	}

	numScriptArgs := len(args) - 1
	if numScriptArgs < 1 || numScriptArgs > 3 {
		return NewError(constants.TypeError, constants.STRING_FIND_ARG_COUNT_ERROR, numScriptArgs)
	}

	subStr, okSub := args[1].(*String)
	if !okSub {
		return NewError(constants.TypeError, constants.STRING_FIND_SUB_TYPE_ERROR, args[1].Type())
	}

	sourceVal := selfStr.Value // Use the full string for Index
	targetSub := subStr.Value

	// Python's find operates on the substring defined by start/end
	// strings.Index operates on the full string, so we slice first

	runes := []rune(sourceVal)
	startIdx, endIdx := 0, len(runes)

	if numScriptArgs >= 2 && args[2] != NULL {
		startInt, okStart := args[2].(*Integer)
		if !okStart {
			return NewError(constants.TypeError, constants.STRING_SLICE_INDICES_TYPE_ERROR, args[2].Type())
		}
		startIdx = int(startInt.Value)
	}
	if numScriptArgs == 3 && args[3] != NULL {
		endInt, okEnd := args[3].(*Integer)
		if !okEnd {
			return NewError(constants.TypeError, constants.STRING_SLICE_INDICES_TYPE_ERROR, args[3].Type())
		}
		endIdx = int(endInt.Value)
	}

	// Python slice semantics for start/end on find
	if startIdx < 0 {
		startIdx = len(runes) + startIdx
	}
	if startIdx < 0 {
		startIdx = 0
	}
	if startIdx > len(runes) {
		startIdx = len(runes)
	} // Can be len, results in empty slice

	if endIdx < 0 {
		endIdx = len(runes) + endIdx
	}
	if endIdx < 0 {
		endIdx = 0
	}
	if endIdx > len(runes) {
		endIdx = len(runes)
	}

	if startIdx > endIdx || startIdx == len(runes) { // If slice is empty or start is beyond end
		if targetSub == constants.EmptyString {
			return &Integer{Value: int64(startIdx)}
		} // "" is found at startIdx if slice is valid but empty
		return &Integer{Value: -1}
	}

	searchSlice := string(runes[startIdx:endIdx])

	idxInSlice := strings.Index(searchSlice, targetSub)
	if idxInSlice == -1 {
		return &Integer{Value: -1}
	}
	return &Integer{Value: int64(startIdx + idxInSlice)} // Adjust index relative to original string
}

// pyStringLenFn implements string.__len__()
func pyStringLenFn(ctx ExecutionContext, args ...Object) Object {
	selfStr, ok := args[0].(*String)
	if !ok {
		return NewError(constants.TypeError, constants.STRING_LEN_ON_STRING_ERROR)
	}
	if len(args) != 1 {
		return NewError(constants.TypeError, constants.LIST_LEN_ON_LIST_ERROR)
	} // Reusing for now
	return &Integer{Value: int64(utf8.RuneCountInString(selfStr.Value))}
}

// pyStringContainsFn implements string.__contains__(substring)
func pyStringContainsFn(ctx ExecutionContext, args ...Object) Object {
	selfStr, ok := args[0].(*String)
	if !ok {
		return NewError(constants.TypeError, constants.STRING_CONTAINS_ON_STRING_ERROR)
	}
	if len(args) != 2 {
		return NewError(constants.TypeError, constants.LIST_CONTAINS_ARG_COUNT_ERROR)
	} // Reusing for now

	subObj, okSub := args[1].(*String)
	if !okSub {
		return NewError(constants.TypeError, constants.STRING_CONTAINS_REQUIRES_STRING, args[1].Type())
	}

	return NativeBoolToBooleanObject(strings.Contains(selfStr.Value, subObj.Value))
}

// pyStringFormatFn implements string.format(*args, **kwargs)
// Simplified version: only positional {} and {index} placeholders. No kwargs, no complex format specifiers.
func pyStringFormatFn(ctx ExecutionContext, args ...Object) Object {
	// args[0] is self (the String object to format)
	// args[1:] are the values to format into the string
	if len(args) < 1 {
		// This should ideally not happen if called as a method, 'self' is always prepended.
		return NewError(constants.InternalError, constants.STRING_FORMAT_ARG_COUNT_ERROR)
	}
	selfStr, ok := args[0].(*String)
	if !ok {
		return NewError(constants.TypeError, constants.STRING_FORMAT_ON_STRING_ERROR)
	}

	formatArgs := args[1:] // Arguments provided to format() by the Pylearn script
	source := selfStr.Value
	var result strings.Builder
	autoIdx := 0 // For automatic indexing with {}

	i := 0
	for i < len(source) {
		char := source[i]
		if char == constants.OpenBraceRune {
			if i+1 < len(source) && source[i+1] == constants.OpenBraceRune { // Escaped {{
				result.WriteByte(byte(constants.OpenBraceRune))
				i += 2
				continue
			}
			// Potential placeholder
			i++ // Move past '{'
			placeholderEnd := strings.IndexByte(source[i:], byte(constants.CloseBraceRune))
			if placeholderEnd == -1 {
				return NewError(constants.ValueError, constants.STRING_FORMAT_SINGLE_BRACE_ERROR)
			}

			placeholderContent := source[i : i+placeholderEnd]
			i += placeholderEnd + 1 // Move past '}'

			var argToFormat Object

			if placeholderContent == constants.EmptyString { // Automatic indexing: {}
				if autoIdx >= len(formatArgs) {
					return NewError(constants.IndexError, constants.STRING_FORMAT_INDEX_OUT_OF_RANGE, autoIdx)
				}
				argToFormat = formatArgs[autoIdx]
				autoIdx++
			} else {
				// Try to parse as explicit index: {0}, {1}, etc.
				idx, err := strconv.Atoi(placeholderContent)
				if err == nil { // Successfully parsed as integer index
					if idx < 0 || idx >= len(formatArgs) {
						return NewError(constants.IndexError, constants.STRING_FORMAT_INDEX_OUT_OF_RANGE, idx)
					}
					argToFormat = formatArgs[idx]
					// If explicit indexing is used, automatic indexing should ideally not continue or reset.
					// For simplicity, this version might allow mixing if not careful, Python's is stricter.
					// Python: "cannot switch from manual field specification to automatic field numbering"
					if autoIdx > 0 && placeholderContent != constants.EmptyString { // if autoIdx was used and now we see explicit
						// This check is basic. Python's rule is more about the *first* one setting the mode.
						// return NewError(constants.ValueError, "cannot switch from automatic field numbering to manual field specification")
					}
				} else {
					// Placeholder content is not empty and not a simple integer index.
					// This could be a keyword, attribute access, etc. Not supported in this simplified version.
					return NewError(constants.ValueError, constants.STRING_FORMAT_UNSUPPORTED_PLACEHOLDER_ERROR, placeholderContent)
				}
			}

			// Convert the Pylearn argument to its string representation
			// We need to call the Pylearn `str()` built-in on argToFormat
			// This requires the ExecutionContext (ctx) and access to the str built-in.

			// Simplification: Use Inspect() for now.
			// A full solution would use ctx.Execute(strBuiltin, argToFormat)
			// (See builtins.pyPrintFn for an example of calling str() via context)
			if argToFormat == nil { // Should not happen if formatArgs are valid Pylearn objects
				result.WriteString(constants.STRING_FORMAT_NIL_ARG_ERROR)
			} else if strVal, isStr := argToFormat.(*String); isStr {
				result.WriteString(strVal.Value)
			} else if argToFormat == NULL {
				result.WriteString(constants.STRING_FORMAT_NONE_KEYWORD) // Python's str(None) is "None"
			} else {
				// For other types, attempt to get their string representation via str() or Inspect()
				// Let's use Inspect() as a fallback.
				// To correctly use str(), the ExecutionContext (ctx) would need to be leveraged.
				// For now, directly using Inspect might differ from Python's str() for some types.
				// For example, str(123) is "123", my_list.Inspect() might be "[1, 2]".

				// Ideal way:
				// strBuiltin, strBuiltinFound := Builtins["str"] // Assuming Builtins is accessible
				// if strBuiltinFound {
				//   strResultObj := ctx.Execute(strBuiltin, argToFormat)
				//   if IsError(strResultObj) { /* handle error from str() */ return strResultObj }
				//   result.WriteString(strResultObj.(*String).Value)
				// } else {
				//   result.WriteString(argToFormat.Inspect()) // Fallback
				// }
				result.WriteString(argToFormat.Inspect()) // Simplified: Using Inspect directly
			}

		} else if char == constants.CloseBraceRune {
			if i+1 < len(source) && source[i+1] == constants.CloseBraceRune { // Escaped }}
				result.WriteByte(byte(constants.CloseBraceRune))
				i += 2
				continue
			}
			return NewError(constants.ValueError, constants.STRING_FORMAT_SINGLE_CLOSING_BRACE_ERROR)
		} else {
			result.WriteByte(byte(char))
			i++
		}
	}

	return &String{Value: result.String()}
}
