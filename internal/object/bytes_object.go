// pylearn/internal/object/bytes_object.go
package object

import (
	"bytes"
	"fmt"
	"hash/fnv"

	"github.com/deniskipeles/pylearn/internal/constants"
)

const (
	BYTEARRAY_OBJ ObjectType = "BYTEARRAY"
)

// --- Bytes (Immutable) ---
type Bytes struct{ Value []byte }

func (b *Bytes) Type() ObjectType { return BYTES_OBJ }
func (b *Bytes) Inspect() string {
	var out bytes.Buffer
	out.WriteString(constants.OBJECT_BYTES_INSPECT_PREFIX)
	for _, byteVal := range b.Value {
		switch {
		case byteVal == constants.SingleQuoteByte:
			out.WriteString(constants.OBJECT_BYTES_INSPECT_SINGLE_QUOTE)
		case byteVal == constants.BackslashByte:
			out.WriteString(constants.OBJECT_BYTES_INSPECT_BACKSLASH)
		case byteVal == constants.TabByte:
			out.WriteString(constants.OBJECT_BYTES_INSPECT_TAB)
		case byteVal == constants.NewlineByte:
			out.WriteString(constants.OBJECT_BYTES_INSPECT_NEWLINE)
		case byteVal == constants.CarriageReturnByte:
			out.WriteString(constants.OBJECT_BYTES_INSPECT_CARRIAGE_RETURN)
		case byteVal >= 32 && byteVal < 127: // Printable ASCII
			out.WriteByte(byteVal)
		default:
			fmt.Fprintf(&out, constants.OBJECT_BYTES_INSPECT_HEX_FORMAT, byteVal)
		}
	}
	out.WriteString(constants.OBJECT_BYTES_INSPECT_SUFFIX)
	return out.String()
}
func (b *Bytes) HashKey() (HashKey, error) {
	h := fnv.New64a()
	_, _ = h.Write(b.Value)
	return HashKey{Type: b.Type(), Value: h.Sum64()}, nil
}

// Bytes Item Access (Read-only, returns Integer)
func (b *Bytes) GetObjectItem(key Object) Object {
	idxObj, ok := key.(*Integer)
	if !ok {
		return NewError(constants.TypeError, constants.OBJECT_BYTES_INDEX_TYPE_ERROR, key.Type())
	}
	idx := idxObj.Value
	bytesLen := int64(len(b.Value))
	if idx < 0 {
		idx += bytesLen
	} // Handle negative index
	if idx < 0 || idx >= bytesLen {
		return NewError(constants.IndexError, constants.OBJECT_BYTES_INDEX_OUT_OF_RANGE)
	}
	return &Integer{Value: int64(b.Value[idx])} // Python bytes indexing returns int
}

// pyBytesContainsFn implements bytes.__contains__(sub)
func pyBytesContainsFn(ctx ExecutionContext, args ...Object) Object {
	if len(args) != 2 {
		return NewError(constants.TypeError, "__contains__() takes exactly one argument (%d given)", len(args)-1)
	}

	selfBytes, ok := args[0].(*Bytes)
	if !ok {
		return NewError(constants.InternalError, "__contains__ called on non-Bytes object")
	}

	// Python's `in` for bytes accepts an integer or a bytes-like object
	switch sub := args[1].(type) {
	case *Integer:
		if sub.Value < 0 || sub.Value > 255 {
			return NewError(constants.ValueError, "byte must be in range(0, 256)")
		}
		return NativeBoolToBooleanObject(bytes.Contains(selfBytes.Value, []byte{byte(sub.Value)}))
	case *Bytes:
		return NativeBoolToBooleanObject(bytes.Contains(selfBytes.Value, sub.Value))
	case *ByteArray:
		return NativeBoolToBooleanObject(bytes.Contains(selfBytes.Value, sub.Value))
	default:
		return NewError(constants.TypeError, "a bytes-like object is required for 'in' operator, not '%s'", args[1].Type())
	}
}

// pyBytesDecodeFn implements bytes.decode(encoding='utf-8', errors='strict')
func pyBytesDecodeFn(ctx ExecutionContext, args ...Object) Object {
	if len(args) < 1 || len(args) > 2 { // self, [encoding]
		return NewError(constants.TypeError, "decode() takes at most 1 argument (%d given)", len(args)-1)
	}
	selfBytes, ok := args[0].(*Bytes)
	if !ok {
		return NewError(constants.TypeError, "decode() must be called on a bytes object")
	}
	// For simplicity, we ignore the encoding argument and assume utf-8
	return &String{Value: string(selfBytes.Value)}
}

// pyBytesJoinFn implements bytes.join(iterable)
func pyBytesJoinFn(ctx ExecutionContext, args ...Object) Object {
	// args[0] is self (the separator Bytes object)
	// args[1] is the iterable
	if len(args) != 2 {
		return NewError(constants.TypeError, "join() takes exactly one argument (%d given)", len(args)-1)
	}
	separator, ok := args[0].(*Bytes)
	if !ok {
		return NewError(constants.TypeError, "join() must be called on a bytes object")
	}
	iterable, ok := args[1].(*List) // For now, assume the iterable is a List
	if !ok {
		return NewError(constants.TypeError, "join() argument must be an iterable of bytes, not %s", args[1].Type())
	}

	var buffer bytes.Buffer
	for i, item := range iterable.Elements {
		itemBytes, ok := item.(*Bytes)
		if !ok {
			return NewError(constants.TypeError, "sequence item %d: expected bytes instance, %s found", i, item.Type())
		}
		if i > 0 {
			buffer.Write(separator.Value)
		}
		buffer.Write(itemBytes.Value)
	}
	return &Bytes{Value: buffer.Bytes()}
}

// pyBytesMulFn implements bytes * int and int * bytes
func pyBytesMulFn(ctx ExecutionContext, args ...Object) Object {
	// For both bytes * int and int * bytes, the interpreter's dunder logic
	// will ensure args are ordered as (bytes, int).
	selfBytes, okBytes := args[0].(*Bytes)
	countObj, okInt := args[1].(*Integer)

	if !okBytes || !okInt {
		// This tells the interpreter that this operation is not implemented for these types,
		// allowing it to raise a standard TypeError.
		return NOT_IMPLEMENTED
	}

	count := int(countObj.Value)
	if count < 0 {
		count = 0 // Multiplying by a negative number results in empty bytes
	}
    
    // Use the efficient bytes.Repeat function
	repeated := bytes.Repeat(selfBytes.Value, count)
	return &Bytes{Value: repeated}
}

// GetObjectAttribute for Bytes to expose methods
func (b *Bytes) GetObjectAttribute(ctx ExecutionContext, name string) (Object, bool) {
	makeBytesMethod := func(methodName string, goFn BuiltinFunction) *Builtin {
		return &Builtin{
			Name: "bytes." + methodName,
			Fn: func(callCtx ExecutionContext, scriptProvidedArgs ...Object) Object {
				methodArgs := make([]Object, 0, 1+len(scriptProvidedArgs))
				methodArgs = append(methodArgs, b)
				methodArgs = append(methodArgs, scriptProvidedArgs...)
				return goFn(callCtx, methodArgs...)
			},
		}
	}
	switch name {
	case constants.DunderContains:
		return makeBytesMethod(constants.DunderContains, pyBytesContainsFn), true
	case "decode": // Use literal string name for the method
		return makeBytesMethod("decode", pyBytesDecodeFn), true
	case "join":
		return makeBytesMethod("join", pyBytesJoinFn), true
	case constants.DunderMul, constants.DunderRMul:
		// Both regular and reflected multiplication map to the same Go function.
		return makeBytesMethod(name, pyBytesMulFn), true
	}
	return nil, false
}

var _ Object = (*Bytes)(nil)
var _ Hashable = (*Bytes)(nil)
var _ ItemGetter = (*Bytes)(nil)
var _ AttributeGetter = (*Bytes)(nil)


// --- ByteArray (Mutable) ---

type ByteArray struct{ Value []byte }

func (ba *ByteArray) Type() ObjectType { return BYTEARRAY_OBJ }
func (ba *ByteArray) Inspect() string {
	// Re-use the Bytes inspect logic by creating a temporary Bytes object, then wrap in bytearray()
	return "bytearray(" + (&Bytes{Value: ba.Value}).Inspect() + ")"
}

// NOTE: Item setting/deleting and other mutable methods (append, extend, pop, etc.)
// would be implemented here, similar to the List object. For now, we are only
// implementing the constructor via the built-in bytearray().

var _ Object = (*ByteArray)(nil)
// ByteArray is mutable, so it is NOT Hashable.