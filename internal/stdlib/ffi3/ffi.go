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
	return fmt.Sprintf("FFI Error: %s", e.Message)
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
	if name == "Size" {
		return &object.Builtin{
			Name: "FFIType.Size",
			Fn: func(callCtx object.ExecutionContext, args ...object.Object) object.Object {
				if len(args) != 0 {
					return object.NewError("TypeError", "Size() takes no arguments")
				}
				return &object.Integer{Value: int64(p.Size())}
			},
		}, true
	}
	return nil, false
}

func (p *CPrimitiveType) Type() object.ObjectType { return "FFI_PRIMITIVE_TYPE" }
func (p *CPrimitiveType) Inspect() string         { return fmt.Sprintf("<ffi_type %s>", p.name) }
func (p *CPrimitiveType) GetFFIType() *C.ffi_type { return p.ffiType }
func (p *CPrimitiveType) Size() uintptr           { return p.size }
func (p *CPrimitiveType) Alignment() uintptr      { return p.size }

func (p *CPrimitiveType) ToC(val object.Object, dest unsafe.Pointer) (func(), error) {
	switch p.ffiType {
	case &C.ffi_type_sint8:
		i, ok := val.(*object.Integer)
		if !ok {
			return nil, fmt.Errorf("expected int, got %s", val.Type())
		}
		*(*C.schar)(dest) = C.schar(i.Value)
	case &C.ffi_type_uint8:
		i, ok := val.(*object.Integer)
		if !ok {
			return nil, fmt.Errorf("expected int, got %s", val.Type())
		}
		*(*C.uchar)(dest) = C.uchar(i.Value)
	case &C.ffi_type_sint32:
		i, ok := val.(*object.Integer)
		if !ok {
			return nil, fmt.Errorf("expected int, got %s", val.Type())
		}
		*(*C.int)(dest) = C.int(i.Value)
	case &C.ffi_type_uint32:
		i, ok := val.(*object.Integer)
		if !ok {
			return nil, fmt.Errorf("expected int, got %s", val.Type())
		}
		*(*C.uint)(dest) = C.uint(i.Value)
	case &C.ffi_type_sint64:
		i, ok := val.(*object.Integer)
		if !ok {
			return nil, fmt.Errorf("expected int, got %s", val.Type())
		}
		*(*C.longlong)(dest) = C.longlong(i.Value)
	case &C.ffi_type_uint64:
		i, ok := val.(*object.Integer)
		if !ok {
			return nil, fmt.Errorf("expected int, got %s", val.Type())
		}
		*(*C.ulonglong)(dest) = C.ulonglong(i.Value)
	case &C.ffi_type_float:
		f, ok := val.(*object.Float)
		if !ok {
			return nil, fmt.Errorf("expected float, got %s", val.Type())
		}
		*(*C.float)(dest) = C.float(f.Value)
	case &C.ffi_type_double:
		f, ok := val.(*object.Float)
		if !ok {
			return nil, fmt.Errorf("expected float, got %s", val.Type())
		}
		*(*C.double)(dest) = C.double(f.Value)
	default:
		return nil, fmt.Errorf("unsupported primitive type for marshalling: %s", p.name)
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
		return nil, fmt.Errorf("unsupported primitive type for unmarshalling: %s", p.name)
	}
}

type CPointerType struct {
	object.Object
	Pointee   FFIType
	ArraySize int
}

func (p *CPointerType) GetObjectAttribute(ctx object.ExecutionContext, name string) (object.Object, bool) {
	if name == "Size" {
		return &object.Builtin{
			Name: "FFIPointerType.Size",
			Fn: func(callCtx object.ExecutionContext, args ...object.Object) object.Object {
				if len(args) != 0 {
					return object.NewError("TypeError", "Size() takes no arguments")
				}
				return &object.Integer{Value: int64(p.Size())}
			},
		}, true
	}
	return nil, false
}

func (p *CPointerType) Type() object.ObjectType { return "FFI_POINTER_TYPE" }
func (p *CPointerType) Inspect() string {
	if p.Pointee != nil {
		pointeeStr := p.Pointee.Inspect()
		if p.ArraySize > 0 {
			return fmt.Sprintf("<ffi_type POINTER TO %s[%d]>", pointeeStr, p.ArraySize)
		}
		return fmt.Sprintf("<ffi_type POINTER TO %s>", pointeeStr)
	}
	return "<ffi_type c_void_p>"
}

func (p *CPointerType) GetFFIType() *C.ffi_type { return &C.ffi_type_pointer }
func (p *CPointerType) Size() uintptr           { return unsafe.Sizeof(uintptr(0)) }
func (p *CPointerType) Alignment() uintptr      { return unsafe.Sizeof(uintptr(0)) }

func (p *CPointerType) ToC(val object.Object, dest unsafe.Pointer) (func(), error) {
	if p.ArraySize > 0 {
		switch v := val.(type) {
		case *object.List:
			if len(v.Elements) != p.ArraySize {
				return nil, fmt.Errorf("list length %d does not match array size %d", len(v.Elements), p.ArraySize)
			}
			totalSize := C.size_t(p.ArraySize) * C.size_t(p.Pointee.Size())
			arrayPtr := C.malloc(totalSize)
			if arrayPtr == nil {
				return nil, &FFIError{Code: ErrOutOfMemory, Message: "failed to malloc for fixed array"}
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
					return nil, fmt.Errorf("failed to marshal array element: %v", err)
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
			return nil, fmt.Errorf("cannot convert Pylearn type %s to C array[%d]", val.Type(), p.ArraySize)
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
				return nil, fmt.Errorf("cannot automatically convert Pylearn bytes to pointer of type %s", p.Pointee.Inspect())
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
						return nil, &FFIError{Code: ErrOutOfMemory, Message: "failed to malloc for wchar_t string"}
					}
					wcharSlice := (*[1 << 30]C.wchar_t)(cWStringPtr)[:numWChars:numWChars]
					for i, code := range utf16Codes {
						wcharSlice[i] = C.wchar_t(code)
					}
					wcharSlice[len(utf16Codes)] = 0
					*(*unsafe.Pointer)(dest) = cWStringPtr
					return func() { C.free(cWStringPtr) }, nil
				} else {
					return nil, fmt.Errorf("wchar_t* ToC not fully implemented for size %d", C_WCHAR_T.Size())
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
			return nil, fmt.Errorf("cannot convert Pylearn type %s to C pointer", val.Type())
		}
	}
}

func (p *CPointerType) FromC(src unsafe.Pointer) (object.Object, error) {
	cPtr := *(*unsafe.Pointer)(src)

	if p.ArraySize > 0 {
		if cPtr == nil {
			return object.NewError("ValueError", "cannot read from NULL pointer for array"), nil
		}
		elements := make([]object.Object, p.ArraySize)
		elementPtr := cPtr
		for i := 0; i < p.ArraySize; i++ {
			elem, err := p.Pointee.FromC(elementPtr)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal array element [%d]: %v", i, err)
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
			return nil, fmt.Errorf("wchar_t* FromC not fully implemented for size %d", C_WCHAR_T.Size())
		}
	}

	return &Pointer{Address: cPtr, PtrType: p}, nil
}

var (
	C_INT8    = &CPrimitiveType{name: "c_int8", ffiType: &C.ffi_type_sint8, size: unsafe.Sizeof(int8(0))}
	C_UINT8   = &CPrimitiveType{name: "c_uint8", ffiType: &C.ffi_type_uint8, size: unsafe.Sizeof(uint8(0))}
	C_INT32   = &CPrimitiveType{name: "c_int32", ffiType: &C.ffi_type_sint32, size: unsafe.Sizeof(int32(0))}
	C_UINT32  = &CPrimitiveType{name: "c_uint32", ffiType: &C.ffi_type_uint32, size: unsafe.Sizeof(uint32(0))}
	C_INT64   = &CPrimitiveType{name: "c_int64", ffiType: &C.ffi_type_sint64, size: unsafe.Sizeof(int64(0))}
	C_UINT64  = &CPrimitiveType{name: "c_uint64", ffiType: &C.ffi_type_uint64, size: unsafe.Sizeof(int64(0))}
	C_FLOAT32 = &CPrimitiveType{name: "c_float", ffiType: &C.ffi_type_float, size: unsafe.Sizeof(float32(0))}
	C_FLOAT64 = &CPrimitiveType{name: "c_double", ffiType: &C.ffi_type_double, size: unsafe.Sizeof(float64(0))}
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
	if name == "Size" {
		return &object.Builtin{
			Name: "FFIType.Size",
			Fn: func(callCtx object.ExecutionContext, args ...object.Object) object.Object {
				if len(args) != 0 {
					return object.NewError("TypeError", "Size() takes no arguments")
				}
				return &object.Integer{Value: int64(w.Size())}
			},
		}, true
	}
	return nil, false
}

func (w *wcharType) Type() object.ObjectType { return "FFI_PRIMITIVE_TYPE" }
func (w *wcharType) Inspect() string         { return fmt.Sprintf("<ffi_type %s>", w.name) }

func (w *wcharType) GetFFIType() *C.ffi_type {
	switch w.size {
	case 2:
		return &C.ffi_type_sint16
	case 4:
		return &C.ffi_type_sint32
	default:
		panic(fmt.Sprintf("Unsupported wchar_t size: %d", w.size))
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
			return nil, fmt.Errorf("unsupported wchar_t size for integer conversion: %d", w.size)
		}
		return nil, nil
	case *object.String:
		utf16Codes := utf16.Encode([]rune(v.Value))
		numWChars := len(utf16Codes) + 1
		totalSize := C.size_t(numWChars) * C.size_t(w.size)
		cWStringPtr := C.malloc(totalSize)
		if cWStringPtr == nil {
			return nil, &FFIError{Code: ErrOutOfMemory, Message: "failed to malloc for wchar_t string"}
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
			return nil, fmt.Errorf("unsupported wchar_t size for string conversion: %d", w.size)
		}

		*(*unsafe.Pointer)(dest) = cWStringPtr
		return func() { C.free(cWStringPtr) }, nil
	default:
		return nil, fmt.Errorf("cannot convert Pylearn type %s to C wchar_t", val.Type())
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
		return nil, fmt.Errorf("unsupported wchar_t size for reading: %d", w.size)
	}
	return &object.String{Value: string(rune(cWCharValue))}, nil
}

func pyFreeCResource(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError("TypeError", "free_c_resource() takes 1 argument")
	}
	switch ptrObj := args[0].(type) {
	case *Pointer:
		if ptrObj.Address != nil {
			C.free(ptrObj.Address)
			ptrObj.Address = nil
		}
		return object.NULL
	default:
		return object.NewError("TypeError", "arg must be a Pointer holding a C resource")
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
	if name == "Size" {
		return &object.Builtin{
			Name: "FFIStructType.Size",
			Fn: func(callCtx object.ExecutionContext, args ...object.Object) object.Object {
				if len(args) != 0 {
					return object.NewError("TypeError", "Size() takes no arguments")
				}
				return &object.Integer{Value: int64(s.Size())}
			},
		}, true
	}
	return nil, false
}

func (s *CStructType) Type() object.ObjectType { return "FFI_STRUCT_TYPE" }
func (s *CStructType) Inspect() string         { return fmt.Sprintf("<ffi_type struct %s>", s.Name) }

func (s *CStructType) GetFFIType() *C.ffi_type {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.ffiType != nil {
		return s.ffiType
	}

	numFields := len(s.Fields)
	cElements := (**C.ffi_type)(C.malloc(C.size_t(numFields+1) * C.size_t(unsafe.Sizeof((*C.ffi_type)(nil)))))
	if cElements == nil {
		panic("FFI: failed to malloc for struct elements")
	}

	cElementsSlice := (*[1 << 30]*C.ffi_type)(unsafe.Pointer(cElements))[: numFields+1 : numFields+1]
	for i, field := range s.Fields {
		cElementsSlice[i] = field.Type.GetFFIType()
	}
	cElementsSlice[numFields] = nil

	ffiType := (*C.ffi_type)(C.malloc(C.size_t(unsafe.Sizeof(C.ffi_type{}))))
	if ffiType == nil {
		panic("FFI: failed to malloc for ffi_type struct")
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
		fmt.Fprintf(os.Stderr, "FFI Warning: could not pre-calculate layout for struct %s\n", s.Name)
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
		return nil, fmt.Errorf("cannot convert Pylearn type %s to C struct %s", val.Type(), s.Name)
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
			return nil, fmt.Errorf("failed to marshal struct field '%s': %v", field.Name, err)
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
			return nil, fmt.Errorf("failed to unmarshal struct field '%s': %v", field.Name, err)
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
	if name == "Size" {
		return &object.Builtin{
			Name: "FFIUnionType.Size",
			Fn: func(callCtx object.ExecutionContext, args ...object.Object) object.Object {
				if len(args) != 0 {
					return object.NewError("TypeError", "Size() takes no arguments")
				}
				return &object.Integer{Value: int64(u.Size())}
			},
		}, true
	}
	return nil, false
}

func (u *CUnionType) Type() object.ObjectType { return "FFI_UNION_TYPE" }
func (u *CUnionType) Inspect() string         { return fmt.Sprintf("<ffi_type union %s>", u.Name) }

func (u *CUnionType) ensureLayoutCalculated() {
	if u.size == 0 && len(u.Fields) > 0 {
		panic("CUnionType used without being properly initialized via create_union_type")
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
		panic("FFI: failed to malloc for union elements")
	}

	cElementsSlice := (*[1 << 30]*C.ffi_type)(unsafe.Pointer(cElements))[: numFields+1 : numFields+1]
	for i, field := range u.Fields {
		cElementsSlice[i] = field.Type.GetFFIType()
	}
	cElementsSlice[numFields] = nil

	ffiType := (*C.ffi_type)(C.malloc(C.size_t(unsafe.Sizeof(C.ffi_type{}))))
	if ffiType == nil {
		C.free(unsafe.Pointer(cElements))
		panic("FFI: failed to malloc for ffi_type union")
	}

	ffiType.size = 0
	ffiType.alignment = 0
	ffiType._type = C.FFI_TYPE_STRUCT
	ffiType.elements = cElements

	var dummyCif C.ffi_cif
	if C.ffi_prep_cif(&dummyCif, C.FFI_DEFAULT_ABI, 0, ffiType, nil) == C.FFI_OK {
	} else {
		fmt.Fprintf(os.Stderr, "FFI Warning: could not pre-calculate layout for union %s\n", u.Name)
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
		return nil, fmt.Errorf("cannot convert Pylearn type %s to C union %s; expected Dict", val.Type(), u.Name)
	}

	if len(dict.Pairs) != 1 {
		return nil, fmt.Errorf("union ToC expects a Dict with exactly one key-value pair to specify the active member")
	}

	C.memset(dest, 0, C.size_t(u.Size()))

	for _, pair := range dict.Pairs {
		fieldName, ok := pair.Key.(*object.String)
		if !ok {
			return nil, fmt.Errorf("union key must be a string representing a member name")
		}
		for _, field := range u.Fields {
			if field.Name == fieldName.Value {
				return field.Type.ToC(pair.Value, dest)
			}
		}
		return nil, fmt.Errorf("union '%s' has no member named '%s'", u.Name, fieldName.Value)
	}
	return nil, fmt.Errorf("internal error during union ToC")
}

func (u *CUnionType) FromC(src unsafe.Pointer) (object.Object, error) {
	if src == nil {
		return object.NULL, nil
	}

	ownedData := C.malloc(C.size_t(u.Size()))
	if ownedData == nil {
		return nil, &FFIError{Code: ErrOutOfMemory, Message: "failed to malloc for union instance"}
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

func (uo *UnionObject) Type() object.ObjectType { return "FFI_UNION_INSTANCE" }
func (uo *UnionObject) Inspect() string {
	if uo.Address == nil {
		return fmt.Sprintf("<freed union %s>", uo.UnionType.Name)
	}
	return fmt.Sprintf("<union %s instance at %p>", uo.UnionType.Name, uo.Address)
}

func (uo *UnionObject) GetObjectAttribute(ctx object.ExecutionContext, name string) (object.Object, bool) {
	if uo.Address == nil {
		return object.NewError("ValueError", "cannot access members of a freed union instance"), true
	}

	for _, field := range uo.UnionType.Fields {
		if field.Name == name {
			val, err := field.Type.FromC(uo.Address)
			if err != nil {
				return object.NewError("FFIError", "failed to read union member '%s': %v", name, err), true
			}
			return val, true
		}
	}

	if name == "address" {
		return &Pointer{Address: uo.Address, PtrType: C_VOID_P}, true
	}

	return nil, false
}

func (uo *UnionObject) SetObjectAttribute(ctx object.ExecutionContext, name string, value object.Object) (object.Object, bool) {
	if uo.Address == nil {
		return object.NewError("ValueError", "cannot access members of a freed union instance"), true
	}

	for _, field := range uo.UnionType.Fields {
		if field.Name == name {
			C.memset(uo.Address, 0, C.size_t(uo.UnionType.Size()))
			_, err := field.Type.ToC(value, uo.Address)
			if err != nil {
				return object.NewError("FFIError", "failed to write to union member '%s': %v", name, err), true
			}
			return value, true
		}
	}

	return nil, false
}

func pyCreateUnionType(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 2 {
		return object.NewError("TypeError", "create_union_type() takes 2 arguments (name, fields_list)")
	}
	nameObj, ok := args[0].(*object.String)
	if !ok {
		return object.NewError("TypeError", "argument 1 (name) must be a string")
	}
	fieldsListObj, ok := args[1].(*object.List)
	if !ok {
		return object.NewError("TypeError", "argument 2 (fields) must be a list")
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
			return object.NewError("TypeError", "field %d must be a list or tuple of (name, type)", i)
		}

		if len(elements) != 2 {
			return object.NewError("TypeError", "field %d must have exactly 2 elements: (name, type)", i)
		}

		fieldNameObj, ok := elements[0].(*object.String)
		if !ok {
			return object.NewError("TypeError", "field %d name must be a string", i)
		}
		fieldTypeObj, ok := elements[1].(FFIType)
		if !ok {
			return object.NewError("TypeError", "field %d type must be a valid FFI type", i)
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
		return object.NewError("ValueError", "cannot create a union with no fields")
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

func (l *Library) Type() object.ObjectType { return "FFI_LIBRARY" }
func (l *Library) Inspect() string         { return fmt.Sprintf("<ffi.Library '%s' from %s>", l.Name, l.Path) }

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

func (f *Function) Type() object.ObjectType { return "FFI_FUNCTION" }
func (f *Function) Inspect() string {
	return fmt.Sprintf("<ffi.Function %s from %s>", f.Name, f.Lib.Name)
}

func generateSignatureKey(name string, retType FFIType, argTypes []FFIType) string {
	var sb strings.Builder
	sb.WriteString(name)
	sb.WriteString(":")
	var typeToString func(t FFIType) string
	typeToString = func(t FFIType) string {
		if t == nil {
			return "void"
		}
		switch tt := t.(type) {
		case *CPrimitiveType:
			return tt.name
		case *CPointerType:
			if tt.Pointee == nil {
				return "c_void_p"
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
			return nil, &FFIError{Code: ErrOutOfMemory, Message: "failed to malloc for arg types array"}
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
		return nil, &FFIError{Code: ErrBadSignature, Message: fmt.Sprintf("libffi ffi_prep_cif failed: %d", ffiStatus)}
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
		return nil, &FFIError{Code: ErrArgCount, Message: fmt.Sprintf("arity mismatch: %s expects %d, got %d", f.Name, len(f.ArgTypes), len(pylearnArgs))}
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
			return nil, &FFIError{Code: ErrOutOfMemory, Message: "failed to malloc arg pointers array"}
		}
		defer C.free(cArgsPtrsStart)

		cArgsPtrsArray := (*[1 << 30]unsafe.Pointer)(cArgsPtrsStart)
		for i, argType := range f.ArgTypes {
			argMemory := C.malloc(C.size_t(argType.Size()))
			if argMemory == nil {
				return nil, &FFIError{Code: ErrOutOfMemory, Message: "failed to malloc for argument"}
			}
			cArgsValues[i] = argMemory

			cleanup, err := argType.ToC(pylearnArgs[i], argMemory)
			if err != nil {
				return nil, &FFIError{Code: ErrArgMarshal, Message: fmt.Sprintf("failed to convert arg %d: %v", i, err)}
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
		return nil, &FFIError{Code: ErrOutOfMemory, Message: "failed to malloc for return value"}
	}
	defer C.free(cRetValPtr)

	cFuncPtr := (C.void_fn)(unsafe.Pointer(f.ptr))
	C.pylearn_ffi_call_shim(&f.cif, cFuncPtr, cRetValPtr, cArgsPtrsStart)

	if f.ReturnType == nil {
		return object.NULL, nil
	}
	pylearnResult, err := f.ReturnType.FromC(cRetValPtr)
	if err != nil {
		return nil, &FFIError{Code: ErrRetUnmarshal, Message: fmt.Sprintf("failed to convert return value: %v", err)}
	}

	return pylearnResult, nil
}

func (f *Function) callVariadic(pylearnArgs ...object.Object) (object.Object, error) {
	if len(pylearnArgs) < f.FixedArgCount {
		return nil, &FFIError{Code: ErrArgCount, Message: fmt.Sprintf("variadic function %s expects at least %d fixed args, got %d", f.Name, f.FixedArgCount, len(pylearnArgs))}
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
			return nil, &FFIError{Code: ErrArgMarshal, Message: fmt.Sprintf("cannot infer FFI type for variadic arg %d of type %s", i, arg.Type())}
		}
	}

	sizeOfPtrArray := C.size_t(totalArgs) * C.size_t(unsafe.Sizeof((*C.ffi_type)(nil)))
	cAllArgTypesPtr := (**C.ffi_type)(C.malloc(sizeOfPtrArray))
	if cAllArgTypesPtr == nil {
		return nil, &FFIError{Code: ErrOutOfMemory, Message: "failed to malloc for variadic arg types"}
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
		return nil, &FFIError{Code: ErrBadSignature, Message: fmt.Sprintf("ffi_prep_cif_var failed: %d", status)}
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
			return nil, &FFIError{Code: ErrOutOfMemory, Message: "failed to malloc arg pointers array"}
		}
		defer C.free(cArgsPtrsStart)

		cArgsPtrsArray := (*[1 << 30]unsafe.Pointer)(cArgsPtrsStart)
		for i, argType := range allArgTypes {
			argMemory := C.malloc(C.size_t(argType.Size()))
			if argMemory == nil {
				return nil, &FFIError{Code: ErrOutOfMemory, Message: "failed to malloc for argument"}
			}
			cArgsValues[i] = argMemory

			cleanup, err := argType.ToC(pylearnArgs[i], argMemory)
			if err != nil {
				return nil, &FFIError{Code: ErrArgMarshal, Message: fmt.Sprintf("failed to convert arg %d: %v", i, err)}
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
		return nil, &FFIError{Code: ErrOutOfMemory, Message: "failed to malloc for return value"}
	}
	defer C.free(cRetValPtr)

	cFuncPtr := (C.void_fn)(unsafe.Pointer(f.ptr))
	C.pylearn_ffi_call_shim(&localCif, cFuncPtr, cRetValPtr, cArgsPtrsStart)

	if f.ReturnType == nil {
		return object.NULL, nil
	}
	pylearnResult, err := f.ReturnType.FromC(cRetValPtr)
	if err != nil {
		return nil, &FFIError{Code: ErrRetUnmarshal, Message: fmt.Sprintf("failed to convert return value: %v", err)}
	}

	return pylearnResult, nil
}

type Pointer struct {
	object.Object
	Address unsafe.Pointer
	PtrType *CPointerType
}

func (p *Pointer) Type() object.ObjectType { return "FFI_POINTER" }
func (p *Pointer) Inspect() string {
	if p.Address == nil {
		return "<ffi.Pointer NULL>"
	}
	return fmt.Sprintf("<ffi.Pointer at %p>", p.Address)
}

func (p *Pointer) GetObjectAttribute(ctx object.ExecutionContext, name string) (object.Object, bool) {
	if name == "Address" {
		return &object.Integer{Value: int64(uintptr(p.Address))}, true
	}
	return nil, false
}

var _ object.AttributeGetter = (*Pointer)(nil)

func pyMalloc(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError("TypeError", "malloc() takes 1 argument")
	}
	sizeObj, ok := args[0].(*object.Integer)
	if !ok {
		return object.NewError("TypeError", "size must be an integer")
	}
	ptr := C.malloc(C.size_t(sizeObj.Value))
	if ptr == nil {
		return object.NewError("MemoryError", "malloc failed")
	}
	return &Pointer{Address: ptr, PtrType: C_VOID_P}
}

func pyFree(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError("TypeError", "free() takes 1 argument")
	}
	ptrObj, ok := args[0].(*Pointer)
	if !ok {
		return object.NewError("TypeError", "arg must be a Pointer")
	}
	if ptrObj.Address != nil {
		C.free(ptrObj.Address)
		ptrObj.Address = nil // Prevent double-free
	}
	return object.NULL
}

func pyMemcpy(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 3 {
		return object.NewError("TypeError", "memcpy() takes 3 arguments")
	}
	dest, ok1 := args[0].(*Pointer)
	src, ok2 := args[1].(*Pointer)
	size, ok3 := args[2].(*object.Integer)
	if !ok1 || !ok2 || !ok3 {
		return object.NewError("TypeError", "args must be (Pointer, Pointer, Integer)")
	}
	C.memcpy(dest.Address, src.Address, C.size_t(size.Value))
	return object.NULL
}

func pyAddressof(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError("TypeError", "addressof() takes 1 argument")
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
		return object.NewError("TypeError", "addressof() unsupported for type %s", args[0].Type())
	}
}

func pyReadMemory(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 2 {
		return object.NewError("TypeError", "read_memory() takes 2 arguments")
	}
	ptr, ok1 := args[0].(*Pointer)
	typ, ok2 := args[1].(FFIType)
	if !ok1 || !ok2 {
		return object.NewError("TypeError", "args must be (Pointer, FFIType)")
	}
	if ptr.Address == nil {
		return object.NewError("ValueError", "cannot read from NULL pointer")
	}
	val, err := typ.FromC(ptr.Address)
	if err != nil {
		return object.NewError("FFIError", "read failed: %v", err)
	}
	return val
}

func pyWriteMemory(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 3 {
		return object.NewError("TypeError", "write_memory() takes 3 arguments")
	}
	ptr, ok1 := args[0].(*Pointer)
	typ, ok2 := args[1].(FFIType)
	val := args[2]
	if !ok1 || !ok2 {
		return object.NewError("TypeError", "args must be (Pointer, FFIType, value)")
	}
	if ptr.Address == nil {
		return object.NewError("ValueError", "cannot write to NULL pointer")
	}

	_, err := typ.ToC(val, ptr.Address)
	if err != nil {
		return object.NewError("FFIError", "write failed: %v", err)
	}
	return object.NULL
}

func pyWriteMemoryWithOffset(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 4 {
		return object.NewError("TypeError", "write_memory_with_offset() takes 4 arguments")
	}
	ptr, ok1 := args[0].(*Pointer)
	off, ok2 := args[1].(*object.Integer)
	typ, ok3 := args[2].(FFIType)
	val := args[3]
	if !ok1 || !ok2 || !ok3 {
		return object.NewError("TypeError", "args must be (Pointer, Integer, FFIType, value)")
	}
	if ptr.Address == nil {
		return object.NewError("ValueError", "cannot write to NULL pointer")
	}
	dest := unsafe.Pointer(uintptr(ptr.Address) + uintptr(off.Value))
	_, err := typ.ToC(val, dest)
	if err != nil {
		return object.NewError("FFIError", "write with offset failed: %v", err)
	}
	return object.NULL
}

func pyReadMemoryWithOffset(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 3 {
		return object.NewError("TypeError", "read_memory_with_offset() takes 3 arguments")
	}
	ptr, ok1 := args[0].(*Pointer)
	off, ok2 := args[1].(*object.Integer)
	typ, ok3 := args[2].(FFIType)
	if !ok1 || !ok2 || !ok3 {
		return object.NewError("TypeError", "args must be (Pointer, Integer, FFIType)")
	}
	if ptr.Address == nil {
		return object.NewError("ValueError", "cannot read from NULL pointer")
	}
	src := unsafe.Pointer(uintptr(ptr.Address) + uintptr(off.Value))
	val, err := typ.FromC(src)
	if err != nil {
		return object.NewError("FFIError", "read with offset failed: %v", err)
	}
	return val
}

func pyBufferToBytes(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 2 {
		return object.NewError("TypeError", "buffer_to_bytes() takes 2 arguments")
	}
	ptr, ok1 := args[0].(*Pointer)
	length, ok2 := args[1].(*object.Integer)
	if !ok1 || !ok2 {
		return object.NewError("TypeError", "args must be (Pointer, Integer)")
	}
	if ptr.Address == nil {
		return object.NewError("ValueError", "cannot read from NULL pointer")
	}
	if length.Value < 0 {
		return object.NewError("ValueError", "length cannot be negative")
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
		return object.NewError("TypeError", "free_callback() takes 1 argument")
	}
	cb, ok := args[0].(*Callback)
	if !ok {
		return object.NewError("TypeError", "argument must be a callback object")
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
	if name == "is_callback" {
		return object.TRUE, true
	}
	return nil, false
}

func (cb *Callback) Type() object.ObjectType { return "FFI_CALLBACK" }
func (cb *Callback) Inspect() string {
	return fmt.Sprintf("<ffi.Callback for %s>", cb.pylearnFunc.Inspect())
}
func (cb *Callback) GetPointer() *Pointer    { return &Pointer{Address: cb.codePtr, PtrType: C_VOID_P} }
func (cb *Callback) GetFFIType() *C.ffi_type { return &C.ffi_type_pointer }
func (cb *Callback) ToC(obj object.Object, dest unsafe.Pointer) (func(), error) {
	if c, ok := obj.(*Callback); ok {
		*(*unsafe.Pointer)(dest) = c.codePtr
		return nil, nil
	}
	return nil, fmt.Errorf("cannot convert %T to callback", obj)
}
func (cb *Callback) FromC(src unsafe.Pointer) (object.Object, error) {
	return nil, fmt.Errorf("cannot convert C pointer to callback object")
}

func NewCallback(pylearnFunc object.Object, retType FFIType, argTypes []FFIType, ctx object.ExecutionContext) (*Callback, error) {
	if ctx == nil {
		return nil, fmt.Errorf("execution context cannot be nil for callback")
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
			return nil, fmt.Errorf("malloc failed for arg types")
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
		return nil, fmt.Errorf("ffi_prep_cif failed")
	}
	cb.closure = C.new_closure(&cb.codePtr)
	if cb.closure == nil {
		if cArgTypesPtr != nil {
			C.free(unsafe.Pointer(cArgTypesPtr))
		}
		return nil, fmt.Errorf("ffi_closure_alloc failed")
	}
	cb.cUserData = C.malloc(C.size_t(unsafe.Sizeof(uintptr(0))))
	if cb.cUserData == nil {
		if cArgTypesPtr != nil {
			C.free(unsafe.Pointer(cArgTypesPtr))
		}
		C.ffi_closure_free(unsafe.Pointer(cb.closure))
		return nil, fmt.Errorf("malloc for user_data failed")
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
		return nil, fmt.Errorf("ffi_prep_closure_loc failed")
	}
	return cb, nil
}

//export goCallbackHandler
func goCallbackHandler(cif *C.ffi_cif, ret unsafe.Pointer, args unsafe.Pointer, user_data unsafe.Pointer) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "\n--- FFI FATAL: Panic in callback function ---\n%v\n", r)
		}
	}()

	cb := (*Callback)(unsafe.Pointer(*(*uintptr)(user_data)))

	if cb.execCtx == nil {
		fmt.Fprintln(os.Stderr, "FFI FATAL: Callback is missing its ExecutionContext")
		return
	}

	numArgs := int(cif.nargs)
	pylearnArgs := make([]object.Object, numArgs)
	cArgsArray := (*[1 << 30]unsafe.Pointer)(args)
	for i := 0; i < numArgs; i++ {
		pylearnObj, err := cb.argTypes[i].FromC(cArgsArray[i])
		if err != nil {
			fmt.Fprintf(os.Stderr, "FFI ERROR: Failed to unmarshal arg %d: %v\n", i, err)
			pylearnArgs[i] = object.NULL
		} else {
			pylearnArgs[i] = pylearnObj
		}
	}

	resultObj := cb.execCtx.Execute(cb.pylearnFunc, pylearnArgs...)

	if object.IsError(resultObj) {
		fmt.Fprintln(os.Stderr, "\n--- Unhandled exception in FFI callback ---")
		if err, ok := resultObj.(*object.Error); ok {
			funcName := "<unknown>"
			if cb.pylearnFunc != nil {
				funcName = cb.pylearnFunc.Inspect()
			}
			fmt.Fprintf(os.Stderr, "  File \"<c_callback>\", in %s\n", funcName)
			fmt.Fprintf(os.Stderr, "%s: %s\n", err.ErrorClass.Name, err.Message)
		} else {
			fmt.Fprintln(os.Stderr, resultObj.Inspect())
		}
		fmt.Fprintln(os.Stderr, "--- End of FFI callback exception ---")
		return
	}

	if cb.retType != nil {
		_, err := cb.retType.ToC(resultObj, ret)
		if err != nil {
			fmt.Fprintf(os.Stderr, "FFI ERROR: Failed to marshal return value: %v\n", err)
		}
	}
}

func pyLoadLibrary(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError("TypeError", "load_library() takes 1 argument")
	}
	libNameObj, ok := args[0].(*object.String)
	if !ok {
		return object.NewError("TypeError", "arg must be a string")
	}
	lib, err := LoadLibrary(libNameObj.Value)
	if err != nil {
		if ffiErr, ok := err.(*FFIError); ok {
			return object.NewError("FFIError", ffiErr.Error())
		}
		return object.NewError("FFIError", err.Error())
	}
	return lib
}

func pyDefineFunction(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) < 4 || len(args) > 5 {
		return object.NewError("TypeError", "define_function() takes 4 or 5 arguments")
	}
	lib, ok1 := args[0].(*Library)
	name, ok2 := args[1].(*object.String)
	if !ok1 || !ok2 {
		return object.NewError("TypeError", "args must be (Library, string, ...)")
	}
	var retType FFIType
	if args[2] != object.NULL {
		if rt, ok := args[2].(FFIType); ok {
			retType = rt
		} else {
			return object.NewError("TypeError", "return_type is not a valid FFI type")
		}
	}
	argTypesList, ok := args[3].(*object.List)
	if !ok {
		return object.NewError("TypeError", "arg_types must be a list")
	}
	argTypes := make([]FFIType, len(argTypesList.Elements))
	for i, elem := range argTypesList.Elements {
		if at, ok := elem.(FFIType); ok {
			argTypes[i] = at
		} else {
			return object.NewError("TypeError", "item in arg_types is not a valid FFI type")
		}
	}

	isVariadic := false
	if len(args) == 5 {
		if b, ok := args[4].(*object.Boolean); ok {
			isVariadic = b.Value
		} else {
			return object.NewError("TypeError", "arg 5 (is_variadic) must be a boolean")
		}
	}

	fn, err := lib.DefineFunction(name.Value, retType, argTypes, isVariadic)

	if err != nil {
		if ffiErr, ok := err.(*FFIError); ok {
			return object.NewError("FFIError", ffiErr.Error())
		}
		return object.NewError("FFIError", err.Error())
	}
	return fn
}

func pyCallFunction(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) < 1 {
		return object.NewError("TypeError", "call_function() requires a function argument")
	}
	fn, ok := args[0].(*Function)
	if !ok {
		return object.NewError("TypeError", "arg must be an FFI Function")
	}
	result, err := fn.Call(args[1:]...)
	if err != nil {
		if ffiErr, ok := err.(*FFIError); ok {
			return object.NewError("FFIError", ffiErr.Error())
		}
		return object.NewError("FFIError", err.Error())
	}
	return result
}

func pyCreateCallback(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 3 {
		return object.NewError("TypeError", "callback() takes 3 arguments")
	}
	pylearnFunc := args[0]
	if !object.IsCallable(pylearnFunc) {
		return object.NewError("TypeError", "arg 1 must be callable")
	}
	var retType FFIType
	if args[1] != object.NULL {
		if rt, ok := args[1].(FFIType); ok {
			retType = rt
		} else {
			return object.NewError("TypeError", "restype is not valid FFI type")
		}
	}
	argTypesList, ok := args[2].(*object.List)
	if !ok {
		return object.NewError("TypeError", "argtypes must be a list")
	}
	argTypes := make([]FFIType, len(argTypesList.Elements))
	for i, elem := range argTypesList.Elements {
		if at, ok := elem.(FFIType); ok {
			argTypes[i] = at
		} else {
			return object.NewError("TypeError", "item in argtypes not a valid FFI type")
		}
	}
	cb, err := NewCallback(pylearnFunc, retType, argTypes, ctx)
	if err != nil {
		return object.NewError("FFIError", err.Error())
	}
	return cb
}

func pyStringAt(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) < 1 || len(args) > 3 {
		return object.NewError("TypeError", "string_at() takes 1 to 3 arguments")
	}
	ptr, ok := args[0].(*Pointer)
	if !ok {
		return object.NewError("TypeError", "arg 1 must be a Pointer")
	}
	if ptr.Address == nil {
		return object.NewError("ValueError", "cannot read from NULL pointer")
	}

	targetAddr := ptr.Address
	var length int64 = -1
	var offset int64 = 0

	if len(args) >= 2 && args[1] != object.NULL {
		lenObj, ok := args[1].(*object.Integer)
		if !ok {
			return object.NewError("TypeError", "arg 2 (length) must be an Integer")
		}
		length = lenObj.Value
		if length < 0 && length != -1 {
			return object.NewError("ValueError", "length cannot be negative")
		}
	}

	if len(args) == 3 && args[2] != object.NULL {
		offObj, ok := args[2].(*object.Integer)
		if !ok {
			return object.NewError("TypeError", "arg 3 (offset) must be an Integer")
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

func pyCreateStructType(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 2 {
		return object.NewError("TypeError", "create_struct_type() takes 2 arguments (name, fields_list)")
	}

	nameObj, ok := args[0].(*object.String)
	if !ok {
		return object.NewError("TypeError", "argument 1 (name) must be a string")
	}

	fieldsListObj, ok := args[1].(*object.List)
	if !ok {
		return object.NewError("TypeError", "argument 2 (fields) must be a list")
	}

	var fields []StructField
	for i, fieldItem := range fieldsListObj.Elements {
		var elements []object.Object
		if fieldList, ok := fieldItem.(*object.List); ok {
			elements = fieldList.Elements
		} else if fieldTuple, ok := fieldItem.(*object.Tuple); ok {
			elements = fieldTuple.Elements
		} else {
			return object.NewError("TypeError", "field %d must be a list or tuple of (name, type)", i)
		}

		if len(elements) != 2 {
			return object.NewError("TypeError", "field %d must have exactly 2 elements: (name, type)", i)
		}

		fieldNameObj, ok := elements[0].(*object.String)
		if !ok {
			return object.NewError("TypeError", "field %d name must be a string", i)
		}

		fieldTypeObj, ok := elements[1].(FFIType)
		if !ok {
			return object.NewError("TypeError", "field %d type must be a valid FFI type", i)
		}

		fields = append(fields, StructField{
			Name:   fieldNameObj.Value,
			Type:   fieldTypeObj,
			Offset: 0,
		})
	}

	totalSize, totalAlignment, offsets, err := calculateLayout(fields)
	if err != nil {
		return object.NewError("FFIError", "failed to calculate struct layout: %v", err)
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
		return object.NewError("TypeError", "_get_or_create_pointer_type() takes 1 argument (pointee_type)")
	}

	if _, ok := args[0].(*object.Class); ok {
		return &CPointerType{
			Pointee:   nil,
			ArraySize: 0,
		}
	}

	pointee, ok := args[0].(FFIType)
	if !ok {
		return object.NewError("TypeError", "argument must be a valid FFI type")
	}

	return &CPointerType{
		Pointee:   pointee,
		ArraySize: 0,
	}
}

func pyCreatePointerType(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 2 {
		return object.NewError("TypeError", "_create_pointer_type() takes 2 arguments (pointee_type, array_size)")
	}

	var pointee FFIType
	if args[0] != object.NULL {
		var ok bool
		pointee, ok = args[0].(FFIType)
		if !ok {
			return object.NewError("TypeError", "argument 1 (pointee_type) must be an FFI type or NULL")
		}
	}

	sizeObj, ok := args[1].(*object.Integer)
	if !ok {
		return object.NewError("TypeError", "argument 2 (array_size) must be an integer")
	}
	arraySize := int(sizeObj.Value)
	if arraySize < 0 {
		return object.NewError("ValueError", "array_size must be non-negative")
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
		panic(fmt.Sprintf("Unsupported C char size: %d", charSize))
	}
	C_CHAR = &CPrimitiveType{name: "c_char", ffiType: &C.ffi_type_sint8, size: charSize}
	C_UCHAR = &CPrimitiveType{name: "c_uchar", ffiType: &C.ffi_type_uint8, size: charSize}

	if shortSize != 2 {
		panic(fmt.Sprintf("Unsupported C short size: %d", shortSize))
	}
	C_SHORT = &CPrimitiveType{name: "c_short", ffiType: &C.ffi_type_sint16, size: shortSize}
	C_USHORT = &CPrimitiveType{name: "c_ushort", ffiType: &C.ffi_type_uint16, size: shortSize}

	var longFFIType, ulongFFIType *C.ffi_type
	switch longSize {
	case 4:
		longFFIType, ulongFFIType = &C.ffi_type_sint32, &C.ffi_type_uint32
	case 8:
		longFFIType, ulongFFIType = &C.ffi_type_sint64, &C.ffi_type_uint64
	default:
		panic(fmt.Sprintf("Unsupported C long size: %d", longSize))
	}
	C_LONG = &CPrimitiveType{name: "c_long", ffiType: longFFIType, size: longSize}
	C_ULONG = &CPrimitiveType{name: "c_ulong", ffiType: ulongFFIType, size: longSize}

	C_LONGLONG = &CPrimitiveType{name: "c_longlong", ffiType: &C.ffi_type_sint64, size: 8}
	C_ULONGLONG = &CPrimitiveType{name: "c_ulonglong", ffiType: &C.ffi_type_uint64, size: 8}

	if boolSize != 1 {
		panic(fmt.Sprintf("Unsupported C _Bool size: %d", boolSize))
	}
	C_BOOL = &CPrimitiveType{name: "c_bool", ffiType: &C.ffi_type_sint8, size: boolSize}

	C_WCHAR_T = &wcharType{name: "c_wchar_t", size: wcharSize}

	C_CHAR_P = &CPointerType{Pointee: C_CHAR}
	C_WCHAR_P = &CPointerType{Pointee: C_WCHAR_T}

	C_PID_T = C_INT32
	C_TIME_T = C_INT64

	C_FILE_P = &CPointerType{Pointee: nil}
	C_DIR_P = &CPointerType{Pointee: nil}
	C_HANDLE = &CPointerType{Pointee: nil}

	env := object.NewEnvironment()
	env.Set("load_library", &object.Builtin{Name: "_ffi.load_library", Fn: pyLoadLibrary})
	env.Set("define_function", &object.Builtin{Name: "_ffi.define_function", Fn: pyDefineFunction})
	env.Set("call_function", &object.Builtin{Name: "_ffi.call_function", Fn: pyCallFunction})
	env.Set("malloc", &object.Builtin{Name: "_ffi.malloc", Fn: pyMalloc})
	env.Set("free", &object.Builtin{Name: "_ffi.free", Fn: pyFree})
	env.Set("memcpy", &object.Builtin{Name: "_ffi.memcpy", Fn: pyMemcpy})
	env.Set("addressof", &object.Builtin{Name: "_ffi.addressof", Fn: pyAddressof})
	env.Set("read_memory", &object.Builtin{Name: "_ffi.read_memory", Fn: pyReadMemory})
	env.Set("write_memory", &object.Builtin{Name: "_ffi.write_memory", Fn: pyWriteMemory})
	env.Set("write_memory_with_offset", &object.Builtin{Name: "_ffi.write_memory_with_offset", Fn: pyWriteMemoryWithOffset})
	env.Set("read_memory_with_offset", &object.Builtin{Name: "_ffi.read_memory_with_offset", Fn: pyReadMemoryWithOffset})
	env.Set("callback", &object.Builtin{Name: "_ffi.callback", Fn: pyCreateCallback})
	env.Set("buffer_to_bytes", &object.Builtin{Name: "_ffi.buffer_to_bytes", Fn: pyBufferToBytes})

	env.Set("_get_or_create_pointer_type", &object.Builtin{Name: "_ffi._get_or_create_pointer_type", Fn: pyGetOrCreatePointerType})
	env.Set("_create_pointer_type", &object.Builtin{Name: "_ffi._create_pointer_type", Fn: pyCreatePointerType})

	env.Set("create_struct_type", &object.Builtin{Name: "_ffi.create_struct_type", Fn: pyCreateStructType})
	env.Set("create_union_type", &object.Builtin{Name: "_ffi.create_union_type", Fn: pyCreateUnionType})
	env.Set("free_callback", &object.Builtin{Name: "_ffi.free_callback", Fn: pyFreeCallback})
	env.Set("string_at", &object.Builtin{Name: "_ffi.string_at", Fn: pyStringAt})

	registerPlatformSpecifics(env)

	env.Set("free_c_resource", &object.Builtin{Name: "_ffi.free_c_resource", Fn: pyFreeCResource})

	env.Set("c_int8", C_INT8)
	env.Set("c_uint8", C_UINT8)
	env.Set("c_int32", C_INT32)
	env.Set("c_uint32", C_UINT32)
	env.Set("c_int64", C_INT64)
	env.Set("c_uint64", C_UINT64)
	env.Set("c_float", C_FLOAT32)
	env.Set("c_double", C_FLOAT64)
	env.Set("c_void_p", C_VOID_P)

	env.Set("c_char", C_CHAR)
	env.Set("c_uchar", C_UCHAR)
	env.Set("c_short", C_SHORT)
	env.Set("c_ushort", C_USHORT)
	env.Set("c_long", C_LONG)
	env.Set("c_ulong", C_ULONG)
	env.Set("c_longlong", C_LONGLONG)
	env.Set("c_ulonglong", C_ULONGLONG)
	env.Set("c_bool", C_BOOL)
	env.Set("c_wchar_t", C_WCHAR_T)
	env.Set("c_char_p", C_CHAR_P)
	env.Set("c_wchar_p", C_WCHAR_P)

	env.Set("c_file_p", C_FILE_P)
	env.Set("c_dir_p", C_DIR_P)
	env.Set("c_pid_t", C_PID_T)
	env.Set("c_time_t", C_TIME_T)
	env.Set("c_handle", C_HANDLE)

	ffiErrorClass := object.CreateExceptionClass("FFIError", object.ExceptionClass)
	object.BuiltinExceptionClasses["FFIError"] = ffiErrorClass
	env.Set("error", ffiErrorClass)

	module := &object.Module{Name: "_ffi_native", Path: "<builtin_ffi>", Env: env}
	object.RegisterNativeModule("_ffi_native", module)
}
