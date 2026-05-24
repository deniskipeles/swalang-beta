//go:build linux || darwin || windows

package ffi3

/*
#include "ffi_helpers.h"
// Define the generic function pointer type that CGo needs for casting.
typedef void (*void_fn)();
*/
import "C"
import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"unicode/utf16"
	"unsafe"

	"github.com/deniskipeles/pylearn/internal/constants"
	"github.com/deniskipeles/pylearn/internal/object"
	"github.com/deniskipeles/pylearn/internal/stdlib/platform"
)

// =============================================================================
// ERRORS
// =============================================================================
type FFIErrorCode int

const (
	ErrGeneric FFIErrorCode = iota
	ErrLibNotFound
	ErrFuncNotFound
	ErrBadSignature
	ErrArgCount
	ErrArgMarshal
	ErrRetUnmarshal
	ErrOutOfMemory
)

type FFIError struct {
	Code    FFIErrorCode
	Message string
}

func (e *FFIError) Error() string {
	return fmt.Sprintf(constants.FFI_ERR_PREFIX, e.Message)
}

// =============================================================================
// TYPES
// =============================================================================
type FFIType interface {
	object.Object
	GetFFIType() *C.ffi_type
	Size() uintptr
	Alignment() uintptr
	ToC(pylearnValue object.Object, dest unsafe.Pointer) (func(), error)
	FromC(src unsafe.Pointer) (object.Object, error)
}

type CPrimitiveType struct {
	object.Object
	ffiType *C.ffi_type
	size    uintptr
	name    string
}

func (p *CPrimitiveType) GetObjectAttribute(ctx object.ExecutionContext, name string) (object.Object, bool) {
	if name == constants.FFI_SIZE_METHOD_NAME {
		return &object.Builtin{
			Name: constants.FFI_PRIMITIVE_SIZE_BUILTIN_NAME,
			Fn: func(callCtx object.ExecutionContext, args ...object.Object) object.Object {
				if len(args) != 0 {
					return object.NewError(constants.TypeError, constants.FFI_SIZE_TAKES_NO_ARGS)
				}
				return &object.Integer{Value: int64(p.Size())}
			},
		}, true
	}
	return nil, false
}

func (p *CPrimitiveType) Type() object.ObjectType { return constants.FFI_PRIMITIVE_TYPE_NAME }
func (p *CPrimitiveType) Inspect() string {
	return fmt.Sprintf(constants.FFI_PRIMITIVE_INSPECT, p.name)
}
func (p *CPrimitiveType) GetFFIType() *C.ffi_type { return p.ffiType }
func (p *CPrimitiveType) Size() uintptr           { return p.size }
func (p *CPrimitiveType) Alignment() uintptr      { return p.size }

func (p *CPrimitiveType) ToC(val object.Object, dest unsafe.Pointer) (func(), error) {
	switch p.ffiType {
	case &C.ffi_type_sint8:
		i, ok := val.(*object.Integer)
		if !ok {
			return nil, fmt.Errorf(constants.FFI_EXPECTED_INT, val.Type())
		}
		*(*C.schar)(dest) = C.schar(i.Value)
	case &C.ffi_type_uint8:
		i, ok := val.(*object.Integer)
		if !ok {
			return nil, fmt.Errorf(constants.FFI_EXPECTED_INT, val.Type())
		}
		*(*C.uchar)(dest) = C.uchar(i.Value)
	case &C.ffi_type_sint32:
		i, ok := val.(*object.Integer)
		if !ok {
			return nil, fmt.Errorf(constants.FFI_EXPECTED_INT, val.Type())
		}
		*(*C.int)(dest) = C.int(i.Value)
	case &C.ffi_type_uint32:
		i, ok := val.(*object.Integer)
		if !ok {
			return nil, fmt.Errorf(constants.FFI_EXPECTED_INT, val.Type())
		}
		*(*C.uint)(dest) = C.uint(i.Value)
	case &C.ffi_type_sint64:
		i, ok := val.(*object.Integer)
		if !ok {
			return nil, fmt.Errorf(constants.FFI_EXPECTED_INT, val.Type())
		}
		*(*C.longlong)(dest) = C.longlong(i.Value)
	case &C.ffi_type_uint64:
		i, ok := val.(*object.Integer)
		if !ok {
			return nil, fmt.Errorf(constants.FFI_EXPECTED_INT, val.Type())
		}
		*(*C.ulonglong)(dest) = C.ulonglong(i.Value)
	case &C.ffi_type_float:
		f, ok := val.(*object.Float)
		if !ok {
			return nil, fmt.Errorf(constants.FFI_EXPECTED_FLOAT, val.Type())
		}
		*(*C.float)(dest) = C.float(f.Value)
	case &C.ffi_type_double:
		f, ok := val.(*object.Float)
		if !ok {
			return nil, fmt.Errorf(constants.FFI_EXPECTED_FLOAT, val.Type())
		}
		*(*C.double)(dest) = C.double(f.Value)
	default:
		return nil, fmt.Errorf(constants.FFI_UNSUPPORTED_PRIMITIVE_MARSHAL, p.name)
	}
	return nil, nil
}

func (p *CPrimitiveType) FromC(src unsafe.Pointer) (object.Object, error) {
	switch p.ffiType {
	case &C.ffi_type_sint8:
		return &object.Integer{Value: int64(*(*C.schar)(src))}, nil
	case &C.ffi_type_uint8:
		return &object.Integer{Value: int64(*(*C.uchar)(src))}, nil
	case &C.ffi_type_sint32:
		return &object.Integer{Value: int64(*(*C.int)(src))}, nil
	case &C.ffi_type_uint32:
		return &object.Integer{Value: int64(*(*C.uint)(src))}, nil
	case &C.ffi_type_sint64:
		return &object.Integer{Value: int64(*(*C.longlong)(src))}, nil
	case &C.ffi_type_uint64:
		return &object.Integer{Value: int64(*(*C.ulonglong)(src))}, nil
	case &C.ffi_type_float:
		return &object.Float{Value: float64(*(*C.float)(src))}, nil
	case &C.ffi_type_double:
		return &object.Float{Value: float64(*(*C.double)(src))}, nil
	default:
		return nil, fmt.Errorf(constants.FFI_UNSUPPORTED_PRIMITIVE_UNMARSHAL, p.name)
	}
}

type CPointerType struct {
	object.Object
	Pointee   FFIType
	ArraySize int
}

func (p *CPointerType) GetObjectAttribute(ctx object.ExecutionContext, name string) (object.Object, bool) {
	if name == constants.FFI_SIZE_METHOD_NAME {
		return &object.Builtin{
			Name: constants.FFI_POINTER_SIZE_BUILTIN_NAME,
			Fn: func(callCtx object.ExecutionContext, args ...object.Object) object.Object {
				if len(args) != 0 {
					return object.NewError(constants.TypeError, constants.FFI_SIZE_TAKES_NO_ARGS)
				}
				return &object.Integer{Value: int64(p.Size())}
			},
		}, true
	}
	if name == constants.DunderCall {
		return &object.Builtin{
			Name: constants.FFI_POINTER_CALL_BUILTIN_NAME,
			Fn: func(callCtx object.ExecutionContext, args ...object.Object) object.Object {
				if len(args) == 1 {
					// Allow instantiation with 0 or None to create a NULL pointer
					if i, ok := args[0].(*object.Integer); ok && i.Value == 0 {
						return &Pointer{Address: nil, PtrType: p}
					}
					if args[0] == object.NULL {
						return &Pointer{Address: nil, PtrType: p}
					}
				}
				return object.NewError(constants.TypeError, constants.FFI_POINTER_INSTANTIATION_ERR)
			},
		}, true
	}
	return nil, false
}

func (p *CPointerType) Type() object.ObjectType { return constants.FFI_POINTER_TYPE_NAME }
func (p *CPointerType) Inspect() string {
	if p.Pointee != nil {
		pointeeStr := p.Pointee.Inspect()
		if p.ArraySize > 0 {
			return fmt.Sprintf(constants.FFI_POINTER_ARRAY_INSPECT, pointeeStr, p.ArraySize)
		}
		return fmt.Sprintf(constants.FFI_POINTER_INSPECT, pointeeStr)
	}
	return constants.FFI_VOID_P_INSPECT
}

func (p *CPointerType) GetFFIType() *C.ffi_type { return &C.ffi_type_pointer }
func (p *CPointerType) Size() uintptr           { return unsafe.Sizeof(uintptr(0)) }
func (p *CPointerType) Alignment() uintptr      { return unsafe.Sizeof(uintptr(0)) }

func (p *CPointerType) ToC(val object.Object, dest unsafe.Pointer) (func(), error) {
	if p.ArraySize > 0 {
		switch v := val.(type) {
		case *object.List:
			if len(v.Elements) != p.ArraySize {
				return nil, fmt.Errorf(constants.FFI_ARRAY_LEN_MISMATCH, len(v.Elements), p.ArraySize)
			}
			totalSize := C.size_t(p.ArraySize) * C.size_t(p.Pointee.Size())
			arrayPtr := C.malloc(totalSize)
			if arrayPtr == nil {
				return nil, &FFIError{Code: ErrOutOfMemory, Message: constants.FFI_MALLOC_FIXED_ARRAY_ERR}
			}

			*(*unsafe.Pointer)(dest) = arrayPtr
			cleanupFns := make([]func(), 0)

			elementPtr := arrayPtr
			for _, elem := range v.Elements {
				cleanup, err := p.Pointee.ToC(elem, elementPtr)
				if err != nil {
					for _, fn := range cleanupFns {
						if fn != nil {
							fn()
						}
					}
					C.free(arrayPtr)
					*(*unsafe.Pointer)(dest) = nil
					return nil, fmt.Errorf(constants.FFI_MARSHAL_ARRAY_ELEM_ERR, err)
				}
				if cleanup != nil {
					cleanupFns = append(cleanupFns, cleanup)
				}
				elementPtr = unsafe.Pointer(uintptr(elementPtr) + p.Pointee.Size())
			}
			return func() {
				for _, fn := range cleanupFns {
					if fn != nil {
						fn()
					}
				}
				C.free(arrayPtr)
			}, nil
		default:
			return nil, fmt.Errorf(constants.FFI_CONVERT_ARRAY_ERR, val.Type(), p.ArraySize)
		}
	} else {
		switch v := val.(type) {
		case *Pointer:
			*(*unsafe.Pointer)(dest) = v.Address
			return nil, nil
		case *Callback:
			*(*unsafe.Pointer)(dest) = v.codePtr
			return nil, nil
		case *object.Bytes:
			if p.Pointee == C_CHAR || p.Pointee == nil {
				b := make([]byte, len(v.Value)+1)
				copy(b, v.Value)
				b[len(v.Value)] = 0
				ptr := C.CBytes(b)
				*(*unsafe.Pointer)(dest) = ptr
				return func() { C.free(ptr) }, nil
			} else {
				return nil, fmt.Errorf(constants.FFI_CONVERT_BYTES_PTR_ERR, p.Pointee.Inspect())
			}
		case *object.String:
			if p.Pointee == C_CHAR || p.Pointee == nil {
				ptr := unsafe.Pointer(C.CString(v.Value))
				*(*unsafe.Pointer)(dest) = ptr
				return func() { C.free(ptr) }, nil
			} else if p.Pointee == C_WCHAR_T {
				utf16Codes := utf16.Encode([]rune(v.Value))
				numWChars := len(utf16Codes) + 1
				if C_WCHAR_T.Size() == 2 {
					totalSize := C.size_t(numWChars) * 2
					cWStringPtr := C.malloc(totalSize)
					if cWStringPtr == nil {
						return nil, &FFIError{Code: ErrOutOfMemory, Message: constants.FFI_MALLOC_WCHAR_STR_ERR}
					}
					wcharSlice := (*[1 << 30]C.wchar_t)(cWStringPtr)[:numWChars:numWChars]
					for i, code := range utf16Codes {
						wcharSlice[i] = C.wchar_t(code)
					}
					wcharSlice[len(utf16Codes)] = 0
					*(*unsafe.Pointer)(dest) = cWStringPtr
					return func() { C.free(cWStringPtr) }, nil
				} else {
					return nil, fmt.Errorf(constants.FFI_WCHAR_TOC_NOT_IMPL, C_WCHAR_T.Size())
				}
			} else {
				ptr := unsafe.Pointer(C.CString(v.Value))
				*(*unsafe.Pointer)(dest) = ptr
				return func() { C.free(ptr) }, nil
			}
		case *object.Null:
			*(*unsafe.Pointer)(dest) = nil
			return nil, nil
		default:
			return nil, fmt.Errorf(constants.FFI_CONVERT_PTR_ERR, val.Type())
		}
	}
}

func (p *CPointerType) FromC(src unsafe.Pointer) (object.Object, error) {
	cPtr := *(*unsafe.Pointer)(src)

	if p.ArraySize > 0 {
		if cPtr == nil {
			return object.NewError(constants.ValueError, constants.FFI_NULL_PTR_ARRAY_READ_ERR), nil
		}
		elements := make([]object.Object, p.ArraySize)
		elementPtr := cPtr
		for i := 0; i < p.ArraySize; i++ {
			elem, err := p.Pointee.FromC(elementPtr)
			if err != nil {
				return nil, fmt.Errorf(constants.FFI_UNMARSHAL_ARRAY_ELEM_ERR, i, err)
			}
			elements[i] = elem
			elementPtr = unsafe.Pointer(uintptr(elementPtr) + p.Pointee.Size())
		}
		return &object.List{Elements: elements}, nil
	}

	if cPtr == nil {
		return &Pointer{Address: nil, PtrType: p}, nil
	}

	if p.Pointee == C_CHAR {
		return &object.String{Value: C.GoString((*C.char)(cPtr))}, nil
	} else if p.Pointee == C_WCHAR_T {
		if C_WCHAR_T.Size() == 2 {
			length := 0
			for {
				if *(*C.wchar_t)(unsafe.Pointer(uintptr(cPtr) + uintptr(length*2))) == 0 {
					break
				}
				length++
			}
			uint16Slice := make([]uint16, length)
			wcharSlice := (*[1 << 30]C.wchar_t)(cPtr)[:length:length]
			for i, wc := range wcharSlice {
				uint16Slice[i] = uint16(wc)
			}
			runes := utf16.Decode(uint16Slice)
			return &object.String{Value: string(runes)}, nil
		} else {
			return nil, fmt.Errorf(constants.FFI_WCHAR_FROMC_NOT_IMPL, C_WCHAR_T.Size())
		}
	}

	return &Pointer{Address: cPtr, PtrType: p}, nil
}

var (
	C_INT8    = &CPrimitiveType{name: constants.FFI_TYPE_NAME_C_INT8, ffiType: &C.ffi_type_sint8, size: unsafe.Sizeof(int8(0))}
	C_UINT8   = &CPrimitiveType{name: constants.FFI_TYPE_NAME_C_UINT8, ffiType: &C.ffi_type_uint8, size: unsafe.Sizeof(uint8(0))}
	C_INT32   = &CPrimitiveType{name: constants.FFI_TYPE_NAME_C_INT32, ffiType: &C.ffi_type_sint32, size: unsafe.Sizeof(int32(0))}
	C_UINT32  = &CPrimitiveType{name: constants.FFI_TYPE_NAME_C_UINT32, ffiType: &C.ffi_type_uint32, size: unsafe.Sizeof(uint32(0))}
	C_INT64   = &CPrimitiveType{name: constants.FFI_TYPE_NAME_C_INT64, ffiType: &C.ffi_type_sint64, size: unsafe.Sizeof(int64(0))}
	C_UINT64  = &CPrimitiveType{name: constants.FFI_TYPE_NAME_C_UINT64, ffiType: &C.ffi_type_uint64, size: unsafe.Sizeof(int64(0))}
	C_FLOAT32 = &CPrimitiveType{name: constants.FFI_TYPE_NAME_C_FLOAT, ffiType: &C.ffi_type_float, size: unsafe.Sizeof(float32(0))}
	C_FLOAT64 = &CPrimitiveType{name: constants.FFI_TYPE_NAME_C_DOUBLE, ffiType: &C.ffi_type_double, size: unsafe.Sizeof(float64(0))}
	C_VOID_P  = &CPointerType{Pointee: nil}

	C_CHAR      *CPrimitiveType
	C_UCHAR     *CPrimitiveType
	C_SHORT     *CPrimitiveType
	C_USHORT    *CPrimitiveType
	C_LONG      *CPrimitiveType
	C_ULONG     *CPrimitiveType
	C_LONGLONG  *CPrimitiveType
	C_ULONGLONG *CPrimitiveType
	C_BOOL      *CPrimitiveType
	C_WCHAR_T   *wcharType

	C_CHAR_P  *CPointerType
	C_WCHAR_P *CPointerType
	C_PID_T   *CPrimitiveType
	C_TIME_T  *CPrimitiveType

	C_FILE_P = &CPointerType{Pointee: nil}
	C_DIR_P  = &CPointerType{Pointee: nil}
	C_HANDLE = &CPointerType{Pointee: nil}
)

type wcharType struct {
	ffiType *C.ffi_type
	size    uintptr
	name    string
}

func (w *wcharType) GetObjectAttribute(ctx object.ExecutionContext, name string) (object.Object, bool) {
	if name == constants.FFI_SIZE_METHOD_NAME {
		return &object.Builtin{
			Name: constants.FFI_PRIMITIVE_SIZE_BUILTIN_NAME,
			Fn: func(callCtx object.ExecutionContext, args ...object.Object) object.Object {
				if len(args) != 0 {
					return object.NewError(constants.TypeError, constants.FFI_SIZE_TAKES_NO_ARGS)
				}
				return &object.Integer{Value: int64(w.Size())}
			},
		}, true
	}
	return nil, false
}

func (w *wcharType) Type() object.ObjectType { return constants.FFI_PRIMITIVE_TYPE_NAME }
func (w *wcharType) Inspect() string         { return fmt.Sprintf(constants.FFI_PRIMITIVE_INSPECT, w.name) }

func (w *wcharType) GetFFIType() *C.ffi_type {
	switch w.size {
	case 2:
		return &C.ffi_type_sint16
	case 4:
		return &C.ffi_type_sint32
	default:
		panic(fmt.Sprintf(constants.FFI_UNSUPPORTED_WCHAR_SIZE, w.size))
	}
}

func (w *wcharType) Size() uintptr      { return w.size }
func (w *wcharType) Alignment() uintptr { return w.size }

func (w *wcharType) ToC(val object.Object, dest unsafe.Pointer) (func(), error) {
	switch v := val.(type) {
	case *object.Integer:
		switch w.size {
		case 2:
			*(*C.wchar_t)(dest) = C.wchar_t(v.Value)
		case 4:
			*(*C.wchar_t)(dest) = C.wchar_t(v.Value)
		default:
			return nil, fmt.Errorf(constants.FFI_WCHAR_INT_CONV_ERR, w.size)
		}
		return nil, nil
	case *object.String:
		utf16Codes := utf16.Encode([]rune(v.Value))
		numWChars := len(utf16Codes) + 1
		totalSize := C.size_t(numWChars) * C.size_t(w.size)
		cWStringPtr := C.malloc(totalSize)
		if cWStringPtr == nil {
			return nil, &FFIError{Code: ErrOutOfMemory, Message: constants.FFI_MALLOC_WCHAR_STR_ERR}
		}

		if w.size == 2 {
			wcharSlice := (*[1 << 30]C.wchar_t)(cWStringPtr)[:numWChars:numWChars]
			for i, code := range utf16Codes {
				wcharSlice[i] = C.wchar_t(code)
			}
			wcharSlice[len(utf16Codes)] = 0
		} else if w.size == 4 {
			runes := []rune(v.Value)
			wcharSlice := (*[1 << 30]C.wchar_t)(cWStringPtr)[: len(runes)+1 : len(runes)+1]
			for i, r := range runes {
				wcharSlice[i] = C.wchar_t(r)
			}
			wcharSlice[len(runes)] = 0
		} else {
			C.free(cWStringPtr)
			return nil, fmt.Errorf(constants.FFI_WCHAR_STR_CONV_ERR, w.size)
		}

		*(*unsafe.Pointer)(dest) = cWStringPtr
		return func() { C.free(cWStringPtr) }, nil
	default:
		return nil, fmt.Errorf(constants.FFI_CONVERT_WCHAR_ERR, val.Type())
	}
}

func (w *wcharType) FromC(src unsafe.Pointer) (object.Object, error) {
	var cWCharValue C.wchar_t
	switch w.size {
	case 2:
		cWCharValue = *(*C.wchar_t)(src)
	case 4:
		cWCharValue = *(*C.wchar_t)(src)
	default:
		return nil, fmt.Errorf(constants.FFI_WCHAR_READ_ERR, w.size)
	}
	return &object.String{Value: string(rune(cWCharValue))}, nil
}

func pyFreeCResource(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError(constants.TypeError, constants.FFI_FREE_C_RES_ARG_ERR)
	}
	switch ptrObj := args[0].(type) {
	case *Pointer:
		if ptrObj.Address != nil {
			C.free(ptrObj.Address)
			ptrObj.Address = nil
		}
		return object.NULL
	default:
		return object.NewError(constants.TypeError, constants.FFI_FREE_C_RES_TYPE_ERR)
	}
}

type StructField struct {
	Name   string
	Type   FFIType
	Offset uintptr
}

type CStructType struct {
	object.Object
	Name      string
	Fields    []StructField
	ffiType   *C.ffi_type
	size      uintptr
	alignment uintptr
	mu        sync.Mutex
}

func (s *CStructType) GetObjectAttribute(ctx object.ExecutionContext, name string) (object.Object, bool) {
	if name == constants.FFI_SIZE_METHOD_NAME {
		return &object.Builtin{
			Name: constants.FFI_STRUCT_SIZE_BUILTIN_NAME,
			Fn: func(callCtx object.ExecutionContext, args ...object.Object) object.Object {
				if len(args) != 0 {
					return object.NewError(constants.TypeError, constants.FFI_SIZE_TAKES_NO_ARGS)
				}
				return &object.Integer{Value: int64(s.Size())}
			},
		}, true
	}
	return nil, false
}

func (s *CStructType) Type() object.ObjectType { return constants.FFI_STRUCT_TYPE_NAME }
func (s *CStructType) Inspect() string         { return fmt.Sprintf(constants.FFI_STRUCT_INSPECT, s.Name) }

func (s *CStructType) GetFFIType() *C.ffi_type {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.ffiType != nil {
		return s.ffiType
	}

	numFields := len(s.Fields)
	cElements := (**C.ffi_type)(C.malloc(C.size_t(numFields+1) * C.size_t(unsafe.Sizeof((*C.ffi_type)(nil)))))
	if cElements == nil {
		panic(constants.FFI_MALLOC_STRUCT_ELEMS_ERR)
	}

	cElementsSlice := (*[1 << 30]*C.ffi_type)(unsafe.Pointer(cElements))[: numFields+1 : numFields+1]
	for i, field := range s.Fields {
		cElementsSlice[i] = field.Type.GetFFIType()
	}
	cElementsSlice[numFields] = nil

	ffiType := (*C.ffi_type)(C.malloc(C.size_t(unsafe.Sizeof(C.ffi_type{}))))
	if ffiType == nil {
		panic(constants.FFI_MALLOC_FFI_TYPE_STRUCT_ERR)
	}

	ffiType.size = 0
	ffiType.alignment = 0
	ffiType._type = C.FFI_TYPE_STRUCT
	ffiType.elements = cElements

	var dummyCif C.ffi_cif
	if C.ffi_prep_cif(&dummyCif, C.FFI_DEFAULT_ABI, 0, ffiType, nil) == C.FFI_OK {
		s.size = uintptr(ffiType.size)
		s.alignment = uintptr(ffiType.alignment)
	} else {
		fmt.Fprintf(os.Stderr, constants.FFI_STRUCT_LAYOUT_WARN, s.Name)
	}

	s.ffiType = ffiType
	return s.ffiType
}

func (s *CStructType) Size() uintptr {
	if s.size == 0 {
		s.GetFFIType()
	}
	return s.size
}

func (s *CStructType) Alignment() uintptr {
	if s.alignment == 0 {
		s.GetFFIType()
	}
	return s.alignment
}

func (s *CStructType) ToC(val object.Object, dest unsafe.Pointer) (func(), error) {
	dict, ok := val.(interface {
		Get(key string) (object.Object, bool)
	})
	if !ok {
		return nil, fmt.Errorf(constants.FFI_CONVERT_STRUCT_ERR, val.Type(), s.Name)
	}

	cleanupFns := make([]func(), 0)

	for _, field := range s.Fields {
		fieldVal, found := dict.Get(field.Name)
		if !found {
			continue
		}
		fieldDest := unsafe.Pointer(uintptr(dest) + field.Offset)
		cleanup, err := field.Type.ToC(fieldVal, fieldDest)
		if err != nil {
			for _, fn := range cleanupFns {
				if fn != nil {
					fn()
				}
			}
			return nil, fmt.Errorf(constants.FFI_MARSHAL_STRUCT_FIELD_ERR, field.Name, err)
		}
		if cleanup != nil {
			cleanupFns = append(cleanupFns, cleanup)
		}
	}
	return func() {
		for _, fn := range cleanupFns {
			if fn != nil {
				fn()
			}
		}
	}, nil
}

func (s *CStructType) FromC(src unsafe.Pointer) (object.Object, error) {
	fields := make(map[string]object.Object)
	for _, field := range s.Fields {
		fieldSrc := unsafe.Pointer(uintptr(src) + field.Offset)
		fieldVal, err := field.Type.FromC(fieldSrc)
		if err != nil {
			return nil, fmt.Errorf(constants.FFI_UNMARSHAL_STRUCT_FIELD_ERR, field.Name, err)
		}
		fields[field.Name] = fieldVal
	}
	return &object.Dict{Pairs: ToHashDictPairs(fields)}, nil
}

func ToHashDictPairs(m map[string]object.Object) map[object.HashKey]object.DictPair {
	out := make(map[object.HashKey]object.DictPair, len(m))
	for k, v := range m {
		keyObj := &object.String{Value: k}
		hashKey, _ := keyObj.HashKey()
		out[hashKey] = object.DictPair{Key: keyObj, Value: v}
	}
	return out
}

type UnionField struct {
	Name string
	Type FFIType
}

type CUnionType struct {
	object.Object
	Name      string
	Fields    []UnionField
	ffiType   *C.ffi_type
	size      uintptr
	alignment uintptr
	mu        sync.Mutex
}

func (u *CUnionType) GetObjectAttribute(ctx object.ExecutionContext, name string) (object.Object, bool) {
	if name == constants.FFI_SIZE_METHOD_NAME {
		return &object.Builtin{
			Name: constants.FFI_UNION_SIZE_BUILTIN_NAME,
			Fn: func(callCtx object.ExecutionContext, args ...object.Object) object.Object {
				if len(args) != 0 {
					return object.NewError(constants.TypeError, constants.FFI_SIZE_TAKES_NO_ARGS)
				}
				return &object.Integer{Value: int64(u.Size())}
			},
		}, true
	}
	return nil, false
}

func (u *CUnionType) Type() object.ObjectType { return constants.FFI_UNION_TYPE_NAME }
func (u *CUnionType) Inspect() string         { return fmt.Sprintf(constants.FFI_UNION_INSPECT, u.Name) }

func (u *CUnionType) ensureLayoutCalculated() {
	if u.size == 0 && len(u.Fields) > 0 {
		panic(constants.FFI_UNION_UNINIT_ERR)
	}
}
func (u *CUnionType) Size() uintptr {
	u.ensureLayoutCalculated()
	return u.size
}
func (u *CUnionType) Alignment() uintptr {
	u.ensureLayoutCalculated()
	return u.alignment
}

func (u *CUnionType) GetFFIType() *C.ffi_type {
	u.mu.Lock()
	defer u.mu.Unlock()

	if u.ffiType != nil {
		return u.ffiType
	}

	numFields := len(u.Fields)
	if numFields == 0 {
		return &C.ffi_type_void
	}

	cElements := (**C.ffi_type)(C.malloc(C.size_t(numFields+1) * C.size_t(unsafe.Sizeof((*C.ffi_type)(nil)))))
	if cElements == nil {
		panic(constants.FFI_MALLOC_UNION_ELEMS_ERR)
	}

	cElementsSlice := (*[1 << 30]*C.ffi_type)(unsafe.Pointer(cElements))[: numFields+1 : numFields+1]
	for i, field := range u.Fields {
		cElementsSlice[i] = field.Type.GetFFIType()
	}
	cElementsSlice[numFields] = nil

	ffiType := (*C.ffi_type)(C.malloc(C.size_t(unsafe.Sizeof(C.ffi_type{}))))
	if ffiType == nil {
		C.free(unsafe.Pointer(cElements))
		panic(constants.FFI_MALLOC_FFI_TYPE_UNION_ERR)
	}

	ffiType.size = 0
	ffiType.alignment = 0
	ffiType._type = C.FFI_TYPE_STRUCT
	ffiType.elements = cElements

	var dummyCif C.ffi_cif
	if C.ffi_prep_cif(&dummyCif, C.FFI_DEFAULT_ABI, 0, ffiType, nil) == C.FFI_OK {
	} else {
		fmt.Fprintf(os.Stderr, constants.FFI_UNION_LAYOUT_WARN, u.Name)
	}

	u.ffiType = ffiType
	return u.ffiType
}

func (u *CUnionType) ToC(val object.Object, dest unsafe.Pointer) (func(), error) {
	var dict *object.Dict
	switch v := val.(type) {
	case *object.Dict:
		dict = v
	default:
		return nil, fmt.Errorf(constants.FFI_CONVERT_UNION_ERR, val.Type(), u.Name)
	}

	if len(dict.Pairs) != 1 {
		return nil, fmt.Errorf(constants.FFI_UNION_TOC_DICT_ERR)
	}

	C.memset(dest, 0, C.size_t(u.Size()))

	for _, pair := range dict.Pairs {
		fieldName, ok := pair.Key.(*object.String)
		if !ok {
			return nil, fmt.Errorf(constants.FFI_UNION_KEY_TYPE_ERR)
		}
		for _, field := range u.Fields {
			if field.Name == fieldName.Value {
				return field.Type.ToC(pair.Value, dest)
			}
		}
		return nil, fmt.Errorf(constants.FFI_UNION_MEMBER_NOT_FOUND_ERR, u.Name, fieldName.Value)
	}
	return nil, fmt.Errorf(constants.FFI_UNION_TOC_INTERNAL_ERR)
}

func (u *CUnionType) FromC(src unsafe.Pointer) (object.Object, error) {
	if src == nil {
		return object.NULL, nil
	}

	ownedData := C.malloc(C.size_t(u.Size()))
	if ownedData == nil {
		return nil, &FFIError{Code: ErrOutOfMemory, Message: constants.FFI_MALLOC_UNION_INST_ERR}
	}
	C.memcpy(ownedData, src, C.size_t(u.Size()))

	instance := &UnionObject{
		UnionType: u,
		Address:   ownedData,
	}

	runtime.SetFinalizer(instance, func(obj *UnionObject) {
		if obj.Address != nil {
			C.free(obj.Address)
			obj.Address = nil
		}
	})

	return instance, nil
}

type UnionObject struct {
	object.Object
	UnionType *CUnionType
	Address   unsafe.Pointer
}

func (uo *UnionObject) Type() object.ObjectType { return constants.FFI_UNION_INST_TYPE_NAME }
func (uo *UnionObject) Inspect() string {
	if uo.Address == nil {
		return fmt.Sprintf(constants.FFI_UNION_FREED_INSPECT, uo.UnionType.Name)
	}
	return fmt.Sprintf(constants.FFI_UNION_INST_INSPECT, uo.UnionType.Name, uo.Address)
}

func (uo *UnionObject) GetObjectAttribute(ctx object.ExecutionContext, name string) (object.Object, bool) {
	if uo.Address == nil {
		return object.NewError(constants.ValueError, constants.FFI_UNION_FREED_ACCESS_ERR), true
	}

	for _, field := range uo.UnionType.Fields {
		if field.Name == name {
			val, err := field.Type.FromC(uo.Address)
			if err != nil {
				return object.NewError(constants.FFI_ERROR_CLASS_NAME, constants.FFI_UNION_READ_MEMBER_ERR, name, err), true
			}
			return val, true
		}
	}

	if name == constants.FFI_UNION_ADDRESS_ATTR {
		return &Pointer{Address: uo.Address, PtrType: C_VOID_P}, true
	}

	return nil, false
}

func (uo *UnionObject) SetObjectAttribute(ctx object.ExecutionContext, name string, value object.Object) (object.Object, bool) {
	if uo.Address == nil {
		return object.NewError(constants.ValueError, constants.FFI_UNION_FREED_ACCESS_ERR), true
	}

	for _, field := range uo.UnionType.Fields {
		if field.Name == name {
			C.memset(uo.Address, 0, C.size_t(uo.UnionType.Size()))
			_, err := field.Type.ToC(value, uo.Address)
			if err != nil {
				return object.NewError(constants.FFI_ERROR_CLASS_NAME, constants.FFI_UNION_WRITE_MEMBER_ERR, name, err), true
			}
			return value, true
		}
	}

	return nil, false
}

func pyCreateUnionType(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 2 {
		return object.NewError(constants.TypeError, constants.FFI_CREATE_UNION_ARG_ERR)
	}
	nameObj, ok := args[0].(*object.String)
	if !ok {
		return object.NewError(constants.TypeError, constants.FFI_ARG1_NAME_STR_ERR)
	}
	fieldsListObj, ok := args[1].(*object.List)
	if !ok {
		return object.NewError(constants.TypeError, constants.FFI_ARG2_FIELDS_LIST_ERR)
	}

	var fields []UnionField
	var maxSize, maxAlignment uintptr = 0, 0

	for i, fieldItem := range fieldsListObj.Elements {
		var elements []object.Object
		if fieldList, ok := fieldItem.(*object.List); ok {
			elements = fieldList.Elements
		} else if fieldTuple, ok := fieldItem.(*object.Tuple); ok {
			elements = fieldTuple.Elements
		} else {
			return object.NewError(constants.TypeError, constants.FFI_FIELD_TUPLE_ERR, i)
		}

		if len(elements) != 2 {
			return object.NewError(constants.TypeError, constants.FFI_FIELD_LEN_ERR, i)
		}

		fieldNameObj, ok := elements[0].(*object.String)
		if !ok {
			return object.NewError(constants.TypeError, constants.FFI_FIELD_NAME_STR_ERR, i)
		}
		fieldTypeObj, ok := elements[1].(FFIType)
		if !ok {
			return object.NewError(constants.TypeError, constants.FFI_FIELD_TYPE_ERR, i)
		}

		fields = append(fields, UnionField{
			Name: fieldNameObj.Value,
			Type: fieldTypeObj,
		})

		if size := fieldTypeObj.Size(); size > maxSize {
			maxSize = size
		}
		if align := fieldTypeObj.Alignment(); align > maxAlignment {
			maxAlignment = align
		}
	}

	if len(fields) == 0 {
		return object.NewError(constants.ValueError, constants.FFI_UNION_NO_FIELDS_ERR)
	}

	unionType := &CUnionType{
		Name:      nameObj.Value,
		Fields:    fields,
		size:      maxSize,
		alignment: maxAlignment,
	}

	return unionType
}

type Library struct {
	object.Object
	Name   string
	Path   string
	handle platform.LibraryHandle
	funcs  map[string]*Function
	mu     sync.RWMutex
}

func (l *Library) Type() object.ObjectType { return constants.FFI_LIBRARY_TYPE_NAME }
func (l *Library) Inspect() string         { return fmt.Sprintf(constants.FFI_LIBRARY_INSPECT, l.Name, l.Path) }

func (l *Library) GetObjectAttribute(ctx object.ExecutionContext, name string) (object.Object, bool) {
	l.mu.RLock()
	fn, ok := l.funcs[name]
	l.mu.RUnlock()
	if ok {
		return fn, true
	}
	return nil, false
}

type Function struct {
	object.Object
	Name          string
	Lib           *Library
	ptr           platform.FuncPtr
	cif           C.ffi_cif
	cArgTypesPtr  **C.ffi_type
	ReturnType    FFIType
	ArgTypes      []FFIType
	IsVariadic    bool
	FixedArgCount int
}

func (f *Function) Type() object.ObjectType { return constants.FFI_FUNCTION_TYPE_NAME }
func (f *Function) Inspect() string {
	return fmt.Sprintf(constants.FFI_FUNCTION_INSPECT, f.Name, f.Lib.Name)
}

func generateSignatureKey(name string, retType FFIType, argTypes []FFIType) string {
	var sb strings.Builder
	sb.WriteString(name)
	sb.WriteString(":")
	var typeToString func(t FFIType) string
	typeToString = func(t FFIType) string {
		if t == nil {
			return constants.FFI_TYPE_VOID
		}
		switch tt := t.(type) {
		case *CPrimitiveType:
			return tt.name
		case *CPointerType:
			if tt.Pointee == nil {
				return constants.FFI_TYPE_C_VOID_P
			}
			return typeToString(tt.Pointee) + "*"
		default:
			return tt.Inspect()
		}
	}
	sb.WriteString(typeToString(retType))
	sb.WriteString("(")
	for i, at := range argTypes {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(typeToString(at))
	}
	sb.WriteString(")")
	return sb.String()
}

func (l *Library) DefineFunction(name string, retType FFIType, argTypes []FFIType, isVariadic bool) (*Function, error) {
	if !isVariadic {
		signatureKey := generateSignatureKey(name, retType, argTypes)
		l.mu.RLock()
		if fn, ok := l.funcs[signatureKey]; ok {
			l.mu.RUnlock()
			return fn, nil
		}
		l.mu.RUnlock()
	}

	procPtr, err := platform.GetManager().GetProcAddress(l.handle, name)
	if err != nil {
		return nil, &FFIError{Code: ErrFuncNotFound, Message: err.Error()}
	}

	if isVariadic {
		fn := &Function{
			Name:          name,
			Lib:           l,
			ptr:           procPtr,
			ReturnType:    retType,
			ArgTypes:      argTypes,
			IsVariadic:    true,
			FixedArgCount: len(argTypes),
		}
		return fn, nil
	}

	numArgs := len(argTypes)
	var cArgTypesPtr **C.ffi_type
	if numArgs > 0 {
		sizeOfPtrArray := C.size_t(numArgs) * C.size_t(unsafe.Sizeof((*C.ffi_type)(nil)))
		cArgTypesPtr = (**C.ffi_type)(C.malloc(sizeOfPtrArray))
		if cArgTypesPtr == nil {
			return nil, &FFIError{Code: ErrOutOfMemory, Message: constants.FFI_MALLOC_ARG_TYPES_ERR}
		}
		cArgTypesSlice := (*[1 << 30]*C.ffi_type)(unsafe.Pointer(cArgTypesPtr))[:numArgs:numArgs]
		for i, argType := range argTypes {
			cArgTypesSlice[i] = argType.GetFFIType()
		}
	}

	var cRetType *C.ffi_type
	if retType != nil {
		cRetType = retType.GetFFIType()
	} else {
		cRetType = &C.ffi_type_void
	}

	var cif C.ffi_cif
	ffiStatus := C.ffi_prep_cif(&cif, C.FFI_DEFAULT_ABI, C.uint(numArgs), cRetType, cArgTypesPtr)
	if ffiStatus != C.FFI_OK {
		if cArgTypesPtr != nil {
			C.free(unsafe.Pointer(cArgTypesPtr))
		}
		return nil, &FFIError{Code: ErrBadSignature, Message: fmt.Sprintf(constants.FFI_PREP_CIF_ERR, ffiStatus)}
	}

	fn := &Function{
		Name: name, Lib: l, ptr: procPtr, cif: cif, cArgTypesPtr: cArgTypesPtr,
		ReturnType: retType, ArgTypes: argTypes, IsVariadic: false, FixedArgCount: numArgs,
	}

	signatureKey := generateSignatureKey(name, retType, argTypes)
	l.mu.Lock()
	l.funcs[signatureKey] = fn
	l.mu.Unlock()
	return fn, nil
}

func (f *Function) Call(pylearnArgs ...object.Object) (object.Object, error) {
	if f.IsVariadic {
		return f.callVariadic(pylearnArgs...)
	}
	return f.callFixed(pylearnArgs...)
}

func (f *Function) callFixed(pylearnArgs ...object.Object) (object.Object, error) {
	if len(pylearnArgs) != len(f.ArgTypes) {
		return nil, &FFIError{Code: ErrArgCount, Message: fmt.Sprintf(constants.FFI_ARITY_MISMATCH_ERR, f.Name, len(f.ArgTypes), len(pylearnArgs))}
	}

	numArgs := len(f.ArgTypes)

	cArgsValues := make([]unsafe.Pointer, numArgs)
	cleanupFns := make([]func(), 0)

	defer func() {
		for _, fn := range cleanupFns {
			if fn != nil {
				fn()
			}
		}
		for _, ptr := range cArgsValues {
			if ptr != nil {
				C.free(ptr)
			}
		}
	}()

	var cArgsPtrsStart unsafe.Pointer
	if numArgs > 0 {
		sizeOfPtrArray := C.size_t(numArgs) * C.size_t(unsafe.Sizeof(uintptr(0)))
		cArgsPtrsStart = C.malloc(sizeOfPtrArray)
		if cArgsPtrsStart == nil {
			return nil, &FFIError{Code: ErrOutOfMemory, Message: constants.FFI_MALLOC_ARG_PTRS_ERR}
		}
		defer C.free(cArgsPtrsStart)

		cArgsPtrsArray := (*[1 << 30]unsafe.Pointer)(cArgsPtrsStart)
		for i, argType := range f.ArgTypes {
			argMemory := C.malloc(C.size_t(argType.Size()))
			if argMemory == nil {
				return nil, &FFIError{Code: ErrOutOfMemory, Message: constants.FFI_MALLOC_ARG_ERR}
			}
			cArgsValues[i] = argMemory

			cleanup, err := argType.ToC(pylearnArgs[i], argMemory)
			if err != nil {
				return nil, &FFIError{Code: ErrArgMarshal, Message: fmt.Sprintf(constants.FFI_CONVERT_ARG_ERR, i, err)}
			}
			if cleanup != nil {
				cleanupFns = append(cleanupFns, cleanup)
			}
			cArgsPtrsArray[i] = argMemory
		}
	}

	var retValSize uintptr
	if f.ReturnType != nil {
		retValSize = f.ReturnType.Size()
	} else {
		retValSize = 1
	}
	cRetValPtr := C.malloc(C.size_t(retValSize))
	if cRetValPtr == nil {
		return nil, &FFIError{Code: ErrOutOfMemory, Message: constants.FFI_MALLOC_RET_ERR}
	}
	defer C.free(cRetValPtr)

	cFuncPtr := (C.void_fn)(unsafe.Pointer(f.ptr))
	C.pylearn_ffi_call_shim(&f.cif, cFuncPtr, cRetValPtr, cArgsPtrsStart)

	if f.ReturnType == nil {
		return object.NULL, nil
	}
	pylearnResult, err := f.ReturnType.FromC(cRetValPtr)
	if err != nil {
		return nil, &FFIError{Code: ErrRetUnmarshal, Message: fmt.Sprintf(constants.FFI_CONVERT_RET_ERR, err)}
	}

	return pylearnResult, nil
}

func (f *Function) callVariadic(pylearnArgs ...object.Object) (object.Object, error) {
	if len(pylearnArgs) < f.FixedArgCount {
		return nil, &FFIError{Code: ErrArgCount, Message: fmt.Sprintf(constants.FFI_VARIADIC_ARITY_ERR, f.Name, f.FixedArgCount, len(pylearnArgs))}
	}

	totalArgs := len(pylearnArgs)
	var localCif C.ffi_cif

	var cRetType *C.ffi_type
	if f.ReturnType != nil {
		cRetType = f.ReturnType.GetFFIType()
	} else {
		cRetType = &C.ffi_type_void
	}

	allArgTypes := make([]FFIType, totalArgs)
	copy(allArgTypes, f.ArgTypes)
	for i := f.FixedArgCount; i < totalArgs; i++ {
		switch arg := pylearnArgs[i].(type) {
		case *object.Integer:
			allArgTypes[i] = C_INT64
		case *object.Float:
			allArgTypes[i] = C_FLOAT64
		case *object.String:
			allArgTypes[i] = C_CHAR_P
		case *object.Bytes:
			allArgTypes[i] = C_CHAR_P
		case *Pointer:
			allArgTypes[i] = C_VOID_P
		case *Callback:
			allArgTypes[i] = C_VOID_P
		default:
			return nil, &FFIError{Code: ErrArgMarshal, Message: fmt.Sprintf(constants.FFI_VARIADIC_INFER_ERR, i, arg.Type())}
		}
	}

	sizeOfPtrArray := C.size_t(totalArgs) * C.size_t(unsafe.Sizeof((*C.ffi_type)(nil)))
	cAllArgTypesPtr := (**C.ffi_type)(C.malloc(sizeOfPtrArray))
	if cAllArgTypesPtr == nil {
		return nil, &FFIError{Code: ErrOutOfMemory, Message: constants.FFI_MALLOC_VARIADIC_TYPES_ERR}
	}
	defer C.free(unsafe.Pointer(cAllArgTypesPtr))

	cArgTypesSlice := (*[1 << 30]*C.ffi_type)(unsafe.Pointer(cAllArgTypesPtr))[:totalArgs:totalArgs]
	for i, argType := range allArgTypes {
		cArgTypesSlice[i] = argType.GetFFIType()
	}

	status := C.ffi_prep_cif_var(
		&localCif, C.FFI_DEFAULT_ABI, C.uint(f.FixedArgCount),
		C.uint(totalArgs), cRetType, cAllArgTypesPtr,
	)
	if status != C.FFI_OK {
		return nil, &FFIError{Code: ErrBadSignature, Message: fmt.Sprintf(constants.FFI_PREP_CIF_VAR_ERR, status)}
	}

	cArgsValues := make([]unsafe.Pointer, totalArgs)
	cleanupFns := make([]func(), 0)

	defer func() {
		for _, fn := range cleanupFns {
			if fn != nil {
				fn()
			}
		}
		for _, ptr := range cArgsValues {
			if ptr != nil {
				C.free(ptr)
			}
		}
	}()

	var cArgsPtrsStart unsafe.Pointer
	if totalArgs > 0 {
		sizeOfPtrArray := C.size_t(totalArgs) * C.size_t(unsafe.Sizeof(uintptr(0)))
		cArgsPtrsStart = C.malloc(sizeOfPtrArray)
		if cArgsPtrsStart == nil {
			return nil, &FFIError{Code: ErrOutOfMemory, Message: constants.FFI_MALLOC_ARG_PTRS_ERR}
		}
		defer C.free(cArgsPtrsStart)

		cArgsPtrsArray := (*[1 << 30]unsafe.Pointer)(cArgsPtrsStart)
		for i, argType := range allArgTypes {
			argMemory := C.malloc(C.size_t(argType.Size()))
			if argMemory == nil {
				return nil, &FFIError{Code: ErrOutOfMemory, Message: constants.FFI_MALLOC_ARG_ERR}
			}
			cArgsValues[i] = argMemory

			cleanup, err := argType.ToC(pylearnArgs[i], argMemory)
			if err != nil {
				return nil, &FFIError{Code: ErrArgMarshal, Message: fmt.Sprintf(constants.FFI_CONVERT_ARG_ERR, i, err)}
			}
			if cleanup != nil {
				cleanupFns = append(cleanupFns, cleanup)
			}
			cArgsPtrsArray[i] = argMemory
		}
	}

	var retValSize uintptr
	if f.ReturnType != nil {
		retValSize = f.ReturnType.Size()
	} else {
		retValSize = 1
	}
	cRetValPtr := C.malloc(C.size_t(retValSize))
	if cRetValPtr == nil {
		return nil, &FFIError{Code: ErrOutOfMemory, Message: constants.FFI_MALLOC_RET_ERR}
	}
	defer C.free(cRetValPtr)

	cFuncPtr := (C.void_fn)(unsafe.Pointer(f.ptr))
	C.pylearn_ffi_call_shim(&localCif, cFuncPtr, cRetValPtr, cArgsPtrsStart)

	if f.ReturnType == nil {
		return object.NULL, nil
	}
	pylearnResult, err := f.ReturnType.FromC(cRetValPtr)
	if err != nil {
		return nil, &FFIError{Code: ErrRetUnmarshal, Message: fmt.Sprintf(constants.FFI_CONVERT_RET_ERR, err)}
	}

	return pylearnResult, nil
}

type Pointer struct {
	object.Object
	Address unsafe.Pointer
	PtrType *CPointerType
}

func (p *Pointer) Type() object.ObjectType { return constants.FFI_POINTER_TYPE_NAME_INST }
func (p *Pointer) Inspect() string {
	if p.Address == nil {
		return constants.FFI_POINTER_NULL_INSPECT
	}
	return fmt.Sprintf(constants.FFI_POINTER_INST_INSPECT, p.Address)
}

func (p *Pointer) GetObjectAttribute(ctx object.ExecutionContext, name string) (object.Object, bool) {
	if name == constants.FFI_POINTER_ADDRESS_ATTR {
		return &object.Integer{Value: int64(uintptr(p.Address))}, true
	}
	return nil, false
}

var _ object.AttributeGetter = (*Pointer)(nil)

func pyMalloc(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError(constants.TypeError, constants.FFI_MALLOC_ARG_COUNT_ERR)
	}
	sizeObj, ok := args[0].(*object.Integer)
	if !ok {
		return object.NewError(constants.TypeError, constants.FFI_SIZE_INT_ERR)
	}
	ptr := C.malloc(C.size_t(sizeObj.Value))
	if ptr == nil {
		return object.NewError(constants.MemoryError, constants.FFI_MALLOC_FAIL_ERR)
	}
	return &Pointer{Address: ptr, PtrType: C_VOID_P}
}

func pyFree(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError(constants.TypeError, constants.FFI_FREE_ARG_COUNT_ERR)
	}
	ptrObj, ok := args[0].(*Pointer)
	if !ok {
		return object.NewError(constants.TypeError, constants.FFI_ARG_MUST_BE_PTR_ERR)
	}
	if ptrObj.Address != nil {
		C.free(ptrObj.Address)
		ptrObj.Address = nil // Prevent double-free
	}
	return object.NULL
}

func pyMemcpy(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 3 {
		return object.NewError(constants.TypeError, constants.FFI_MEMCPY_ARG_COUNT_ERR)
	}
	dest, ok1 := args[0].(*Pointer)
	src, ok2 := args[1].(*Pointer)
	size, ok3 := args[2].(*object.Integer)
	if !ok1 || !ok2 || !ok3 {
		return object.NewError(constants.TypeError, constants.FFI_MEMCPY_ARG_TYPE_ERR)
	}
	C.memcpy(dest.Address, src.Address, C.size_t(size.Value))
	return object.NULL
}

func pyAddressof(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError(constants.TypeError, constants.FFI_ADDRESSOF_ARG_COUNT_ERR)
	}
	var address unsafe.Pointer
	switch obj := args[0].(type) {
	case *object.Bytes:
		if len(obj.Value) > 0 {
			address = unsafe.Pointer(&obj.Value[0])
		}
		return &Pointer{Address: address, PtrType: &CPointerType{Pointee: C_UINT8}}
	case *Pointer:
		return obj
	default:
		return object.NewError(constants.TypeError, constants.FFI_ADDRESSOF_UNSUPPORTED_ERR, args[0].Type())
	}
}

func pyReadMemory(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 2 {
		return object.NewError(constants.TypeError, constants.FFI_READ_MEM_ARG_COUNT_ERR)
	}
	ptr, ok1 := args[0].(*Pointer)
	typ, ok2 := args[1].(FFIType)
	if !ok1 || !ok2 {
		return object.NewError(constants.TypeError, constants.FFI_PTR_TYPE_ARGS_ERR)
	}
	if ptr.Address == nil {
		return object.NewError(constants.ValueError, constants.FFI_NULL_PTR_READ_ERR)
	}
	val, err := typ.FromC(ptr.Address)
	if err != nil {
		return object.NewError(constants.FFI_ERROR_CLASS_NAME, constants.FFI_READ_FAIL_ERR, err)
	}
	return val
}

func pyWriteMemory(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 3 {
		return object.NewError(constants.TypeError, constants.FFI_WRITE_MEM_ARG_COUNT_ERR)
	}
	ptr, ok1 := args[0].(*Pointer)
	typ, ok2 := args[1].(FFIType)
	val := args[2]
	if !ok1 || !ok2 {
		return object.NewError(constants.TypeError, constants.FFI_WRITE_MEM_ARG_TYPE_ERR)
	}
	if ptr.Address == nil {
		return object.NewError(constants.ValueError, constants.FFI_NULL_PTR_WRITE_ERR)
	}

	_, err := typ.ToC(val, ptr.Address)
	if err != nil {
		return object.NewError(constants.FFI_ERROR_CLASS_NAME, constants.FFI_WRITE_FAIL_ERR, err)
	}
	return object.NULL
}

func pyWriteMemoryWithOffset(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 4 {
		return object.NewError(constants.TypeError, constants.FFI_WRITE_MEM_OFFSET_ARG_COUNT_ERR)
	}
	ptr, ok1 := args[0].(*Pointer)
	off, ok2 := args[1].(*object.Integer)
	typ, ok3 := args[2].(FFIType)
	val := args[3]
	if !ok1 || !ok2 || !ok3 {
		return object.NewError(constants.TypeError, constants.FFI_WRITE_MEM_OFFSET_ARG_TYPE_ERR)
	}
	if ptr.Address == nil {
		return object.NewError(constants.ValueError, constants.FFI_NULL_PTR_WRITE_ERR)
	}
	dest := unsafe.Pointer(uintptr(ptr.Address) + uintptr(off.Value))
	_, err := typ.ToC(val, dest)
	if err != nil {
		return object.NewError(constants.FFI_ERROR_CLASS_NAME, constants.FFI_WRITE_OFFSET_FAIL_ERR, err)
	}
	return object.NULL
}

func pyReadMemoryWithOffset(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 3 {
		return object.NewError(constants.TypeError, constants.FFI_READ_MEM_OFFSET_ARG_COUNT_ERR)
	}
	ptr, ok1 := args[0].(*Pointer)
	off, ok2 := args[1].(*object.Integer)
	typ, ok3 := args[2].(FFIType)
	if !ok1 || !ok2 || !ok3 {
		return object.NewError(constants.TypeError, constants.FFI_READ_MEM_OFFSET_ARG_TYPE_ERR)
	}
	if ptr.Address == nil {
		return object.NewError(constants.ValueError, constants.FFI_NULL_PTR_READ_ERR)
	}
	src := unsafe.Pointer(uintptr(ptr.Address) + uintptr(off.Value))
	val, err := typ.FromC(src)
	if err != nil {
		return object.NewError(constants.FFI_ERROR_CLASS_NAME, constants.FFI_READ_OFFSET_FAIL_ERR, err)
	}
	return val
}

func pyBufferToBytes(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 2 {
		return object.NewError(constants.TypeError, constants.FFI_BUF_TO_BYTES_ARG_COUNT_ERR)
	}
	ptr, ok1 := args[0].(*Pointer)
	length, ok2 := args[1].(*object.Integer)
	if !ok1 || !ok2 {
		return object.NewError(constants.TypeError, constants.FFI_PTR_TYPE_ARGS_ERR)
	}
	if ptr.Address == nil {
		return object.NewError(constants.ValueError, constants.FFI_NULL_PTR_READ_ERR)
	}
	if length.Value < 0 {
		return object.NewError(constants.ValueError, constants.FFI_LEN_NEGATIVE_ERR)
	}
	return &object.Bytes{Value: C.GoBytes(ptr.Address, C.int(length.Value))}
}

func (cb *Callback) Free() {
	registryMutex.Lock()
	defer registryMutex.Unlock()

	registryKey := uintptr(unsafe.Pointer(cb))
	if _, ok := callbackRegistry[registryKey]; !ok {
		return
	}

	C.ffi_closure_free(unsafe.Pointer(cb.closure))
	if cb.cArgTypesPtr != nil {
		C.free(unsafe.Pointer(cb.cArgTypesPtr))
	}
	C.free(cb.cUserData)

	cb.closure = nil
	cb.cArgTypesPtr = nil
	cb.cUserData = nil

	delete(callbackRegistry, registryKey)
}

func pyFreeCallback(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError(constants.TypeError, constants.FFI_FREE_CB_ARG_COUNT_ERR)
	}
	cb, ok := args[0].(*Callback)
	if !ok {
		return object.NewError(constants.TypeError, constants.FFI_ARG_MUST_BE_CB_ERR)
	}
	cb.Free()
	return object.NULL
}

var (
	callbackRegistry = make(map[uintptr]*Callback)
	registryMutex    sync.RWMutex
	globalExecCtx    object.ExecutionContext
)

func SetGlobalExecutionContext(ctx object.ExecutionContext) {
	globalExecCtx = ctx
}

type Callback struct {
	object.Object
	pylearnFunc  object.Object
	execCtx      object.ExecutionContext
	cif          C.ffi_cif
	cArgTypesPtr **C.ffi_type
	cUserData    unsafe.Pointer
	argTypes     []FFIType
	retType      FFIType
	closure      *C.ffi_closure
	codePtr      unsafe.Pointer
}

func (cb *Callback) GetObjectAttribute(ctx object.ExecutionContext, name string) (object.Object, bool) {
	if name == constants.FFI_IS_CALLBACK_ATTR {
		return object.TRUE, true
	}
	return nil, false
}

func (cb *Callback) Type() object.ObjectType { return constants.FFI_CALLBACK_TYPE_NAME }
func (cb *Callback) Inspect() string {
	return fmt.Sprintf(constants.FFI_CALLBACK_INSPECT, cb.pylearnFunc.Inspect())
}
func (cb *Callback) GetPointer() *Pointer    { return &Pointer{Address: cb.codePtr, PtrType: C_VOID_P} }
func (cb *Callback) GetFFIType() *C.ffi_type { return &C.ffi_type_pointer }
func (cb *Callback) ToC(obj object.Object, dest unsafe.Pointer) (func(), error) {
	if c, ok := obj.(*Callback); ok {
		*(*unsafe.Pointer)(dest) = c.codePtr
		return nil, nil
	}
	return nil, fmt.Errorf(constants.FFI_CONVERT_TO_CB_ERR, obj)
}
func (cb *Callback) FromC(src unsafe.Pointer) (object.Object, error) {
	return nil, fmt.Errorf(constants.FFI_CONVERT_C_PTR_TO_CB_ERR)
}

func NewCallback(pylearnFunc object.Object, retType FFIType, argTypes []FFIType, ctx object.ExecutionContext) (*Callback, error) {
	if ctx == nil {
		return nil, fmt.Errorf(constants.FFI_EXEC_CTX_NIL_CB_ERR)
	}
	cb := &Callback{
		pylearnFunc: pylearnFunc,
		argTypes:    argTypes,
		retType:     retType,
		execCtx:     ctx,
	}
	numArgs := len(argTypes)
	var cArgTypesPtr **C.ffi_type
	if numArgs > 0 {
		sizeOfPtrArray := C.size_t(numArgs) * C.size_t(unsafe.Sizeof((*C.ffi_type)(nil)))
		cArgTypesPtr = (**C.ffi_type)(C.malloc(sizeOfPtrArray))
		if cArgTypesPtr == nil {
			return nil, fmt.Errorf(constants.FFI_MALLOC_ARG_TYPES_FAIL_ERR)
		}
		cb.cArgTypesPtr = cArgTypesPtr
		cArgTypesSlice := (*[1 << 30]*C.ffi_type)(unsafe.Pointer(cArgTypesPtr))[:numArgs:numArgs]
		for i, argType := range argTypes {
			cArgTypesSlice[i] = argType.GetFFIType()
		}
	}
	var cRetType *C.ffi_type
	if retType != nil {
		cRetType = retType.GetFFIType()
	} else {
		cRetType = &C.ffi_type_void
	}
	if C.ffi_prep_cif(&cb.cif, C.FFI_DEFAULT_ABI, C.uint(numArgs), cRetType, cArgTypesPtr) != C.FFI_OK {
		if cArgTypesPtr != nil {
			C.free(unsafe.Pointer(cArgTypesPtr))
		}
		return nil, fmt.Errorf(constants.FFI_PREP_CIF_FAIL_ERR)
	}
	cb.closure = C.new_closure(&cb.codePtr)
	if cb.closure == nil {
		if cArgTypesPtr != nil {
			C.free(unsafe.Pointer(cArgTypesPtr))
		}
		return nil, fmt.Errorf(constants.FFI_CLOSURE_ALLOC_FAIL_ERR)
	}
	cb.cUserData = C.malloc(C.size_t(unsafe.Sizeof(uintptr(0))))
	if cb.cUserData == nil {
		if cArgTypesPtr != nil {
			C.free(unsafe.Pointer(cArgTypesPtr))
		}
		C.ffi_closure_free(unsafe.Pointer(cb.closure))
		return nil, fmt.Errorf(constants.FFI_MALLOC_USER_DATA_FAIL_ERR)
	}
	*(*uintptr)(cb.cUserData) = uintptr(unsafe.Pointer(cb))
	registryKey := uintptr(unsafe.Pointer(cb))
	registryMutex.Lock()
	callbackRegistry[registryKey] = cb
	registryMutex.Unlock()
	if C.ffi_prep_closure_loc(cb.closure, &cb.cif, (*[0]byte)(C.c_callback_shim), cb.cUserData, cb.codePtr) != C.FFI_OK {
		registryMutex.Lock()
		delete(callbackRegistry, registryKey)
		registryMutex.Unlock()
		if cArgTypesPtr != nil {
			C.free(unsafe.Pointer(cArgTypesPtr))
		}
		C.free(cb.cUserData)
		C.ffi_closure_free(unsafe.Pointer(cb.closure))
		return nil, fmt.Errorf(constants.FFI_PREP_CLOSURE_LOC_FAIL_ERR)
	}
	return cb, nil
}

//export goCallbackHandler
func goCallbackHandler(cif *C.ffi_cif, ret unsafe.Pointer, args unsafe.Pointer, user_data unsafe.Pointer) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, constants.FFI_FATAL_PANIC_CB_ERR, r)
		}
	}()

	cb := (*Callback)(unsafe.Pointer(*(*uintptr)(user_data)))

	if cb.execCtx == nil {
		fmt.Fprintln(os.Stderr, constants.FFI_FATAL_MISSING_CTX_ERR)
		return
	}

	numArgs := int(cif.nargs)
	pylearnArgs := make([]object.Object, numArgs)
	cArgsArray := (*[1 << 30]unsafe.Pointer)(args)
	for i := 0; i < numArgs; i++ {
		pylearnObj, err := cb.argTypes[i].FromC(cArgsArray[i])
		if err != nil {
			fmt.Fprintf(os.Stderr, constants.FFI_UNMARSHAL_CB_ARG_ERR, i, err)
			pylearnArgs[i] = object.NULL
		} else {
			pylearnArgs[i] = pylearnObj
		}
	}

	resultObj := cb.execCtx.Execute(cb.pylearnFunc, pylearnArgs...)

	if object.IsError(resultObj) {
		fmt.Fprintln(os.Stderr, constants.FFI_UNHANDLED_EXC_CB_HEADER)
		if err, ok := resultObj.(*object.Error); ok {
			funcName := constants.FFI_UNKNOWN_FUNC_NAME
			if cb.pylearnFunc != nil {
				funcName = cb.pylearnFunc.Inspect()
			}
			fmt.Fprintf(os.Stderr, constants.FFI_CB_TRACEBACK_FILE_FMT, funcName)
			fmt.Fprintf(os.Stderr, constants.FFI_CB_EXC_FMT, err.ErrorClass.Name, err.Message)
		} else {
			fmt.Fprintln(os.Stderr, resultObj.Inspect())
		}
		fmt.Fprintln(os.Stderr, constants.FFI_CB_EXC_FOOTER)
		return
	}

	if cb.retType != nil {
		_, err := cb.retType.ToC(resultObj, ret)
		if err != nil {
			fmt.Fprintf(os.Stderr, constants.FFI_MARSHAL_RET_ERR, err)
		}
	}
}

func pyLoadLibrary(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError(constants.TypeError, constants.FFI_LOAD_LIB_ARG_COUNT_ERR)
	}
	libNameObj, ok := args[0].(*object.String)
	if !ok {
		return object.NewError(constants.TypeError, constants.FFI_ARG_MUST_BE_STR_ERR)
	}
	lib, err := LoadLibrary(libNameObj.Value)
	if err != nil {
		if ffiErr, ok := err.(*FFIError); ok {
			return object.NewError(constants.FFI_ERROR_CLASS_NAME, ffiErr.Error())
		}
		return object.NewError(constants.FFI_ERROR_CLASS_NAME, err.Error())
	}
	return lib
}

func pyDefineFunction(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) < 4 || len(args) > 5 {
		return object.NewError(constants.TypeError, constants.FFI_DEF_FUNC_ARG_COUNT_ERR)
	}
	lib, ok1 := args[0].(*Library)
	name, ok2 := args[1].(*object.String)
	if !ok1 || !ok2 {
		return object.NewError(constants.TypeError, constants.FFI_DEF_FUNC_ARG_TYPE_ERR)
	}
	var retType FFIType
	if args[2] != object.NULL {
		if rt, ok := args[2].(FFIType); ok {
			retType = rt
		} else {
			return object.NewError(constants.TypeError, constants.FFI_RET_TYPE_INVALID_ERR)
		}
	}
	argTypesList, ok := args[3].(*object.List)
	if !ok {
		return object.NewError(constants.TypeError, constants.FFI_ARG_TYPES_NOT_LIST_ERR)
	}
	argTypes := make([]FFIType, len(argTypesList.Elements))
	for i, elem := range argTypesList.Elements {
		if at, ok := elem.(FFIType); ok {
			argTypes[i] = at
		} else {
			return object.NewError(constants.TypeError, constants.FFI_ARG_TYPE_INVALID_ERR)
		}
	}

	isVariadic := false
	if len(args) == 5 {
		if b, ok := args[4].(*object.Boolean); ok {
			isVariadic = b.Value
		} else {
			return object.NewError(constants.TypeError, constants.FFI_IS_VAR_NOT_BOOL_ERR)
		}
	}

	fn, err := lib.DefineFunction(name.Value, retType, argTypes, isVariadic)

	if err != nil {
		if ffiErr, ok := err.(*FFIError); ok {
			return object.NewError(constants.FFI_ERROR_CLASS_NAME, ffiErr.Error())
		}
		return object.NewError(constants.FFI_ERROR_CLASS_NAME, err.Error())
	}
	return fn
}

func pyCallFunction(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) < 1 {
		return object.NewError(constants.TypeError, constants.FFI_CALL_FUNC_REQ_FUNC_ERR)
	}
	fn, ok := args[0].(*Function)
	if !ok {
		return object.NewError(constants.TypeError, constants.FFI_ARG_MUST_BE_FFI_FUNC_ERR)
	}
	result, err := fn.Call(args[1:]...)
	if err != nil {
		if ffiErr, ok := err.(*FFIError); ok {
			return object.NewError(constants.FFI_ERROR_CLASS_NAME, ffiErr.Error())
		}
		return object.NewError(constants.FFI_ERROR_CLASS_NAME, err.Error())
	}
	return result
}

func pyCreateCallback(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 3 {
		return object.NewError(constants.TypeError, constants.FFI_CALLBACK_ARG_COUNT_ERR)
	}
	pylearnFunc := args[0]
	if !object.IsCallable(pylearnFunc) {
		return object.NewError(constants.TypeError, constants.FFI_CB_ARG1_NOT_CALLABLE_ERR)
	}
	var retType FFIType
	if args[1] != object.NULL {
		if rt, ok := args[1].(FFIType); ok {
			retType = rt
		} else {
			return object.NewError(constants.TypeError, constants.FFI_CB_RESTYPE_INVALID_ERR)
		}
	}
	argTypesList, ok := args[2].(*object.List)
	if !ok {
		return object.NewError(constants.TypeError, constants.FFI_CB_ARGTYPES_NOT_LIST_ERR)
	}
	argTypes := make([]FFIType, len(argTypesList.Elements))
	for i, elem := range argTypesList.Elements {
		if at, ok := elem.(FFIType); ok {
			argTypes[i] = at
		} else {
			return object.NewError(constants.TypeError, constants.FFI_CB_ARGTYPE_INVALID_ERR)
		}
	}
	cb, err := NewCallback(pylearnFunc, retType, argTypes, ctx)
	if err != nil {
		return object.NewError(constants.FFI_ERROR_CLASS_NAME, err.Error())
	}
	return cb
}

func pyStringAt(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) < 1 || len(args) > 3 {
		return object.NewError(constants.TypeError, constants.FFI_STR_AT_ARG_COUNT_ERR)
	}
	ptr, ok := args[0].(*Pointer)
	if !ok {
		return object.NewError(constants.TypeError, constants.FFI_ARG1_MUST_BE_PTR_ERR)
	}
	if ptr.Address == nil {
		return object.NewError(constants.ValueError, constants.FFI_NULL_PTR_READ_ERR)
	}

	targetAddr := ptr.Address
	var length int64 = -1
	var offset int64 = 0

	if len(args) >= 2 && args[1] != object.NULL {
		lenObj, ok := args[1].(*object.Integer)
		if !ok {
			return object.NewError(constants.TypeError, constants.FFI_ARG2_MUST_BE_INT_ERR)
		}
		length = lenObj.Value
		if length < 0 && length != -1 {
			return object.NewError(constants.ValueError, constants.FFI_LEN_NEGATIVE_ERR)
		}
	}

	if len(args) == 3 && args[2] != object.NULL {
		offObj, ok := args[2].(*object.Integer)
		if !ok {
			return object.NewError(constants.TypeError, constants.FFI_ARG3_MUST_BE_INT_ERR)
		}
		offset = offObj.Value
		targetAddr = unsafe.Pointer(uintptr(targetAddr) + uintptr(offset))
	}

	if length == -1 {
		return &object.String{Value: C.GoString((*C.char)(targetAddr))}
	}

	goBytes := C.GoBytes(targetAddr, C.int(length))
	strVal := string(goBytes)
	if nullIdx := strings.IndexByte(strVal, 0); nullIdx != -1 {
		strVal = strVal[:nullIdx]
	}
	return &object.String{Value: strVal}
}

func pyGetFuncAddress(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 2 {
		return object.NewError(constants.TypeError, constants.FFI_GET_FUNC_ADDR_ARG_COUNT_ERR)
	}
	lib, ok1 := args[0].(*Library)
	name, ok2 := args[1].(*object.String)
	if !ok1 || !ok2 {
		return object.NewError(constants.TypeError, constants.FFI_GET_FUNC_ADDR_ARG_TYPE_ERR)
	}
	procPtr, err := platform.GetManager().GetProcAddress(lib.handle, name.Value)
	if err != nil {
		return object.NewError(constants.FFI_ERROR_CLASS_NAME, err.Error())
	}
	return &Pointer{Address: unsafe.Pointer(procPtr), PtrType: C_VOID_P}
}

func pyCreateStructType(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 2 {
		return object.NewError(constants.TypeError, constants.FFI_CREATE_STRUCT_ARG_COUNT_ERR)
	}

	nameObj, ok := args[0].(*object.String)
	if !ok {
		return object.NewError(constants.TypeError, constants.FFI_ARG1_NAME_STR_ERR)
	}

	fieldsListObj, ok := args[1].(*object.List)
	if !ok {
		return object.NewError(constants.TypeError, constants.FFI_ARG2_FIELDS_LIST_ERR)
	}

	var fields []StructField
	for i, fieldItem := range fieldsListObj.Elements {
		var elements []object.Object
		if fieldList, ok := fieldItem.(*object.List); ok {
			elements = fieldList.Elements
		} else if fieldTuple, ok := fieldItem.(*object.Tuple); ok {
			elements = fieldTuple.Elements
		} else {
			return object.NewError(constants.TypeError, constants.FFI_FIELD_TUPLE_ERR, i)
		}

		if len(elements) != 2 {
			return object.NewError(constants.TypeError, constants.FFI_FIELD_LEN_ERR, i)
		}

		fieldNameObj, ok := elements[0].(*object.String)
		if !ok {
			return object.NewError(constants.TypeError, constants.FFI_FIELD_NAME_STR_ERR, i)
		}

		fieldTypeObj, ok := elements[1].(FFIType)
		if !ok {
			return object.NewError(constants.TypeError, constants.FFI_FIELD_TYPE_ERR, i)
		}

		fields = append(fields, StructField{
			Name:   fieldNameObj.Value,
			Type:   fieldTypeObj,
			Offset: 0,
		})
	}

	totalSize, totalAlignment, offsets, err := calculateLayout(fields)
	if err != nil {
		return object.NewError(constants.FFI_ERROR_CLASS_NAME, constants.FFI_CALC_STRUCT_LAYOUT_ERR, err)
	}

	for i := range fields {
		fields[i].Offset = offsets[i]
	}

	structType := &CStructType{
		Name:      nameObj.Value,
		Fields:    fields,
		size:      totalSize,
		alignment: totalAlignment,
	}

	return structType
}

func pyGetOrCreatePointerType(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError(constants.TypeError, constants.FFI_GET_CREATE_PTR_ARG_COUNT_ERR)
	}

	if _, ok := args[0].(*object.Class); ok {
		return &CPointerType{
			Pointee:   nil,
			ArraySize: 0,
		}
	}

	pointee, ok := args[0].(FFIType)
	if !ok {
		return object.NewError(constants.TypeError, constants.FFI_ARG_VALID_FFI_TYPE_ERR)
	}

	return &CPointerType{
		Pointee:   pointee,
		ArraySize: 0,
	}
}

func pyCreatePointerType(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 2 {
		return object.NewError(constants.TypeError, constants.FFI_CREATE_PTR_ARG_COUNT_ERR)
	}

	var pointee FFIType
	if args[0] != object.NULL {
		var ok bool
		pointee, ok = args[0].(FFIType)
		if !ok {
			return object.NewError(constants.TypeError, constants.FFI_CREATE_PTR_ARG1_ERR)
		}
	}

	sizeObj, ok := args[1].(*object.Integer)
	if !ok {
		return object.NewError(constants.TypeError, constants.FFI_CREATE_PTR_ARG2_ERR)
	}
	arraySize := int(sizeObj.Value)
	if arraySize < 0 {
		return object.NewError(constants.ValueError, constants.FFI_ARRAY_SIZE_NEG_ERR)
	}

	return &CPointerType{
		Pointee:   pointee,
		ArraySize: arraySize,
	}
}

func init() {
	longSize := unsafe.Sizeof(C.long(0))
	wcharSize := unsafe.Sizeof(C.wchar_t(0))
	boolSize := unsafe.Sizeof(C._Bool(false))
	charSize := unsafe.Sizeof(C.char(0))
	shortSize := unsafe.Sizeof(C.short(0))

	if charSize != 1 {
		panic(fmt.Sprintf(constants.FFI_UNSUPPORTED_C_CHAR_SIZE, charSize))
	}
	C_CHAR = &CPrimitiveType{name: constants.FFI_TYPE_NAME_C_CHAR, ffiType: &C.ffi_type_sint8, size: charSize}
	C_UCHAR = &CPrimitiveType{name: constants.FFI_TYPE_NAME_C_UCHAR, ffiType: &C.ffi_type_uint8, size: charSize}

	if shortSize != 2 {
		panic(fmt.Sprintf(constants.FFI_UNSUPPORTED_C_SHORT_SIZE, shortSize))
	}
	C_SHORT = &CPrimitiveType{name: constants.FFI_TYPE_NAME_C_SHORT, ffiType: &C.ffi_type_sint16, size: shortSize}
	C_USHORT = &CPrimitiveType{name: constants.FFI_TYPE_NAME_C_USHORT, ffiType: &C.ffi_type_uint16, size: shortSize}

	var longFFIType, ulongFFIType *C.ffi_type
	switch longSize {
	case 4:
		longFFIType, ulongFFIType = &C.ffi_type_sint32, &C.ffi_type_uint32
	case 8:
		longFFIType, ulongFFIType = &C.ffi_type_sint64, &C.ffi_type_uint64
	default:
		panic(fmt.Sprintf(constants.FFI_UNSUPPORTED_C_LONG_SIZE, longSize))
	}
	C_LONG = &CPrimitiveType{name: constants.FFI_TYPE_NAME_C_LONG, ffiType: longFFIType, size: longSize}
	C_ULONG = &CPrimitiveType{name: constants.FFI_TYPE_NAME_C_ULONG, ffiType: ulongFFIType, size: longSize}

	C_LONGLONG = &CPrimitiveType{name: constants.FFI_TYPE_NAME_C_LONGLONG, ffiType: &C.ffi_type_sint64, size: 8}
	C_ULONGLONG = &CPrimitiveType{name: constants.FFI_TYPE_NAME_C_ULONGLONG, ffiType: &C.ffi_type_uint64, size: 8}

	if boolSize != 1 {
		panic(fmt.Sprintf(constants.FFI_UNSUPPORTED_C_BOOL_SIZE, boolSize))
	}
	C_BOOL = &CPrimitiveType{name: constants.FFI_TYPE_NAME_C_BOOL, ffiType: &C.ffi_type_sint8, size: boolSize}

	C_WCHAR_T = &wcharType{name: constants.FFI_TYPE_NAME_C_WCHAR_T, size: wcharSize}

	C_CHAR_P = &CPointerType{Pointee: C_CHAR}
	C_WCHAR_P = &CPointerType{Pointee: C_WCHAR_T}

	C_PID_T = C_INT32
	C_TIME_T = C_INT64

	C_FILE_P = &CPointerType{Pointee: nil}
	C_DIR_P = &CPointerType{Pointee: nil}
	C_HANDLE = &CPointerType{Pointee: nil}

	env := object.NewEnvironment()

	// Bind FFI Functions
	env.Set(constants.FFI_ENV_LOAD_LIBRARY, &object.Builtin{Name: constants.FFI_BUILTIN_LOAD_LIBRARY, Fn: pyLoadLibrary})
	env.Set(constants.FFI_ENV_DEFINE_FUNCTION, &object.Builtin{Name: constants.FFI_BUILTIN_DEFINE_FUNCTION, Fn: pyDefineFunction})
	env.Set(constants.FFI_ENV_CALL_FUNCTION, &object.Builtin{Name: constants.FFI_BUILTIN_CALL_FUNCTION, Fn: pyCallFunction})
	env.Set(constants.FFI_ENV_MALLOC, &object.Builtin{Name: constants.FFI_BUILTIN_MALLOC, Fn: pyMalloc})
	env.Set(constants.FFI_ENV_FREE, &object.Builtin{Name: constants.FFI_BUILTIN_FREE, Fn: pyFree})
	env.Set(constants.FFI_ENV_MEMCPY, &object.Builtin{Name: constants.FFI_BUILTIN_MEMCPY, Fn: pyMemcpy})
	env.Set(constants.FFI_ENV_ADDRESSOF, &object.Builtin{Name: constants.FFI_BUILTIN_ADDRESSOF, Fn: pyAddressof})
	env.Set(constants.FFI_ENV_READ_MEMORY, &object.Builtin{Name: constants.FFI_BUILTIN_READ_MEMORY, Fn: pyReadMemory})
	env.Set(constants.FFI_ENV_WRITE_MEMORY, &object.Builtin{Name: constants.FFI_BUILTIN_WRITE_MEMORY, Fn: pyWriteMemory})
	env.Set(constants.FFI_ENV_WRITE_MEMORY_OFFSET, &object.Builtin{Name: constants.FFI_BUILTIN_WRITE_MEMORY_OFFSET, Fn: pyWriteMemoryWithOffset})
	env.Set(constants.FFI_ENV_READ_MEMORY_OFFSET, &object.Builtin{Name: constants.FFI_BUILTIN_READ_MEMORY_OFFSET, Fn: pyReadMemoryWithOffset})
	env.Set(constants.FFI_ENV_CALLBACK, &object.Builtin{Name: constants.FFI_BUILTIN_CALLBACK, Fn: pyCreateCallback})
	env.Set(constants.FFI_ENV_BUFFER_TO_BYTES, &object.Builtin{Name: constants.FFI_BUILTIN_BUFFER_TO_BYTES, Fn: pyBufferToBytes})

	// Bind Type Creation Helpers
	env.Set(constants.FFI_ENV_GET_CREATE_PTR_TYPE, &object.Builtin{Name: constants.FFI_BUILTIN_GET_CREATE_PTR_TYPE, Fn: pyGetOrCreatePointerType})
	env.Set(constants.FFI_ENV_CREATE_PTR_TYPE, &object.Builtin{Name: constants.FFI_BUILTIN_CREATE_PTR_TYPE, Fn: pyCreatePointerType})
	env.Set(constants.FFI_ENV_CREATE_STRUCT_TYPE, &object.Builtin{Name: constants.FFI_BUILTIN_CREATE_STRUCT_TYPE, Fn: pyCreateStructType})
	env.Set(constants.FFI_ENV_CREATE_UNION_TYPE, &object.Builtin{Name: constants.FFI_BUILTIN_CREATE_UNION_TYPE, Fn: pyCreateUnionType})

	// Bind Other Helpers
	env.Set(constants.FFI_ENV_FREE_CALLBACK, &object.Builtin{Name: constants.FFI_BUILTIN_FREE_CALLBACK, Fn: pyFreeCallback})
	env.Set(constants.FFI_ENV_STRING_AT, &object.Builtin{Name: constants.FFI_BUILTIN_STRING_AT, Fn: pyStringAt})
	env.Set(constants.FFI_ENV_GET_FUNC_ADDRESS, &object.Builtin{Name: constants.FFI_BUILTIN_GET_FUNC_ADDRESS, Fn: pyGetFuncAddress})
	env.Set(constants.FFI_ENV_FREE_C_RESOURCE, &object.Builtin{Name: constants.FFI_BUILTIN_FREE_C_RESOURCE, Fn: pyFreeCResource})

	registerPlatformSpecifics(env)

	// Bind C Types
	env.Set(constants.FFI_TYPE_NAME_C_INT8, C_INT8)
	env.Set(constants.FFI_TYPE_NAME_C_UINT8, C_UINT8)
	env.Set(constants.FFI_TYPE_NAME_C_INT32, C_INT32)
	env.Set(constants.FFI_TYPE_NAME_C_UINT32, C_UINT32)
	env.Set(constants.FFI_TYPE_NAME_C_INT64, C_INT64)
	env.Set(constants.FFI_TYPE_NAME_C_UINT64, C_UINT64)
	env.Set(constants.FFI_TYPE_NAME_C_FLOAT, C_FLOAT32)
	env.Set(constants.FFI_TYPE_NAME_C_DOUBLE, C_FLOAT64)
	env.Set(constants.FFI_TYPE_C_VOID_P, C_VOID_P)

	env.Set(constants.FFI_TYPE_NAME_C_CHAR, C_CHAR)
	env.Set(constants.FFI_TYPE_NAME_C_UCHAR, C_UCHAR)
	env.Set(constants.FFI_TYPE_NAME_C_SHORT, C_SHORT)
	env.Set(constants.FFI_TYPE_NAME_C_USHORT, C_USHORT)
	env.Set(constants.FFI_TYPE_NAME_C_LONG, C_LONG)
	env.Set(constants.FFI_TYPE_NAME_C_ULONG, C_ULONG)
	env.Set(constants.FFI_TYPE_NAME_C_LONGLONG, C_LONGLONG)
	env.Set(constants.FFI_TYPE_NAME_C_ULONGLONG, C_ULONGLONG)
	env.Set(constants.FFI_TYPE_NAME_C_BOOL, C_BOOL)
	env.Set(constants.FFI_TYPE_NAME_C_WCHAR_T, C_WCHAR_T)
	env.Set(constants.FFI_TYPE_NAME_C_CHAR_P, C_CHAR_P)
	env.Set(constants.FFI_TYPE_NAME_C_WCHAR_P, C_WCHAR_P)

	env.Set(constants.FFI_TYPE_NAME_C_FILE_P, C_FILE_P)
	env.Set(constants.FFI_TYPE_NAME_C_DIR_P, C_DIR_P)
	env.Set(constants.FFI_TYPE_NAME_C_PID_T, C_PID_T)
	env.Set(constants.FFI_TYPE_NAME_C_TIME_T, C_TIME_T)
	env.Set(constants.FFI_TYPE_NAME_C_HANDLE, C_HANDLE)

	// Create and bind FFIError class
	ffiErrorClass := object.CreateExceptionClass(constants.FFI_ERROR_CLASS_NAME, object.ExceptionClass)
	object.BuiltinExceptionClasses[constants.FFI_ERROR_CLASS_NAME] = ffiErrorClass
	env.Set(constants.FFI_ENV_ERROR, ffiErrorClass)

	// Register Native Module
	module := &object.Module{Name: constants.FFI_MODULE_NAME, Path: constants.FFI_MODULE_PATH, Env: env}
	object.RegisterNativeModule(constants.FFI_MODULE_NAME, module)
}
