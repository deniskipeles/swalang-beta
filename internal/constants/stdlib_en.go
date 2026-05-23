//go:build en
package constants

// ============================================================================
// Sys Module Constants (from stdlib/pysys)
// ============================================================================
// These constants define the metadata, lookup keys, and error messages for the 
// built-in 'sys' module. 
// Translation Note: Do not translate module names or environment keys (like "sys", 
// "argv", "platform"), as they are hardcoded into the import/runtime engine. 
// When translating errors, ensure that format verbs (e.g., %d, %s) are kept as-is.
const (
	SYS_MODULE_NAME          = "sys"
	SYS_MODULE_BUILTIN_PATH  = "<builtin>"
	SYS_ARGV_KEY             = "argv"
	SYS_EXIT_KEY             = "exit"
	SYS_PLATFORM_KEY         = "platform"
	SYS_EXIT_ARG_COUNT_ERROR = "sys.exit() takes at most 1 argument (%d given)"
	SYS_EXIT_ARG_TYPE_ERROR  = "sys.exit() argument must be an integer or None, not %s"
)

// ============================================================================
// Platform Resolution Constants (from stdlib/platform)
// ============================================================================
// These constants support dynamic loading (.so / .dll) across different OS platforms.
// Translation Note: Translate the error strings to match user-facing error standards,
// while preserving the Go string format verbs (e.g., %s, %w).
const (
	PLATFORM_DLOPEN_FAILED_ERROR            = "dlopen failed: %s"
	PLATFORM_DLCLOSE_FAILED_ERROR           = "dlclose failed: %s"
	PLATFORM_DLSYM_FAILED_ERROR             = "dlsym failed for '%s': %s"
	PLATFORM_UNIX_LIB_EXTENSION             = ".so"
	PLATFORM_WINDOWS_KERNEL32_DLL           = "kernel32.dll"
	PLATFORM_WINDOWS_LOAD_KERNEL32_ERROR    = "could not load kernel32.dll: %w"
	PLATFORM_WINDOWS_SET_DLL_DIR_PROC       = "SetDllDirectoryW"
	PLATFORM_WINDOWS_FIND_PROC_ERROR        = "could not find SetDllDirectoryW in kernel32.dll: %w"
	PLATFORM_WINDOWS_UTF16_CONVERSION_ERROR = "could not convert path '%s' to UTF16: %w"
	PLATFORM_WINDOWS_SET_DLL_DIR_FAIL_ERROR = "call to SetDllDirectoryW for path '%s' failed: %w"
	PLATFORM_WINDOWS_LOAD_LIBRARY_ERROR     = "LoadLibrary failed for '%s': %w"
	PLATFORM_WINDOWS_FREE_LIBRARY_ERROR     = "FreeLibrary failed: %w"
	PLATFORM_WINDOWS_GET_PROC_ADDRESS_ERROR = "GetProcAddress failed for '%s': %w"
	PLATFORM_WINDOWS_LIB_EXTENSION          = ".dll"
)

// ============================================================================
// Importlib Module Constants (from stdlib/pyimportlib)
// ============================================================================
// Supports the dynamic import mechanism (pylearn_importlib) for loading modules.
// Translation Note: Do not alter names or paths like "pylearn_importlib", as the 
// virtual machine's loader is tightly coupled to these exact strings.
const (
	IMPORTLIB_LOAD_MODULE_BUILTIN_NAME   = "pylearn_importlib.load_module_from_path"
	IMPORTLIB_LOAD_MODULE_NOT_INIT_ERROR = "pylearn_importlib.load_module_from_path not properly initialized by the interpreter"
	IMPORTLIB_LOAD_MODULE_METHOD_NAME    = "load_module_from_path"
	IMPORTLIB_MODULE_NAME                = "pylearn_importlib"
)

// ============================================================================
// FFI3 Platform Resolution Constants (from stdlib/ffi3/platform_*)
// ============================================================================
// Directory paths, environment variable names, and resolution details for FFI operations.
// Translation Note: Path strings, env keys, and the ELF header validation substring 
// are internal system contracts and must remain untranslated.
const (
	FFI_GO_MOD_FILE                        = "go.mod"
	FFI_BIN_DIR                            = "bin"
	FFI_LIB_DIR                            = "lib"
	FFI_LIB_PREFIX                         = "lib"
	FFI_UNIX_LIB_PATH                      = "/lib"
	FFI_UNIX_USR_LIB_PATH                  = "/usr/lib"
	FFI_UNIX_USR_LOCAL_LIB_PATH            = "/usr/local/lib"
	FFI_UNIX_LIB_X86_PATH                  = "/lib/x86_64-linux-gnu"
	FFI_UNIX_USR_LIB_X86_PATH              = "/usr/lib/x86_64-linux-gnu"
	FFI_ELF_HEADER_ERROR_SUBSTR            = "invalid ELF header"
	FFI_LOAD_ELF_ERR_FORMAT                = "could not load library '%s': %v (ELF error: %v)"
	FFI_LOAD_LIB_ERR_FORMAT                = "could not load library '%s': %v \n(Original path error: %v)"
	FFI_WINDOWS_GET_LAST_ERR_PROC          = "GetLastError"
	FFI_WINDOWS_GET_LAST_ERR_ARG_ERROR     = "get_last_error() takes no arguments"
	FFI_WINDOWS_GET_LAST_ERR_METHOD        = "get_last_error"
	FFI_WINDOWS_GET_LAST_ERR_BUILTIN       = "_ffi.get_last_error"
	FFI_WINDOWS_WINDIR_ENV                 = "WINDIR"
	FFI_WINDOWS_SYSTEM32_DIR               = "System32"
	FFI_WINDOWS_SYSWOW64_DIR               = "SysWOW64"
	FFI_LAYOUT_MALLOC_FAILED               = "malloc failed for type_info array"
	FFI_LAYOUT_CALCULATION_FAILED          = "calculate_struct_layout failed, likely out of memory"
)

// ============================================================================
// FFI Types, Environment Keys, and Messages (from stdlib/ffi3/ffi.go)
// ============================================================================
// These constants define the core vocabulary of the FFI native module, mapping Go
// primitives to their C equivalents, defining inspection formats, and presenting errors.
// Translation Note:
// 1. "Type Names" and "Environment Keys" are internal identifiers — do not translate them.
// 2. "Inspection Formats" represent internal debug visualizers; change with caution.
// 3. "Errors & Messages" are user-visible during runtime failures and can be translated,
//    always preserving formattable variables like %s, %d, %v, or %T.
const (
	// --- Type Names ---
	FFI_PRIMITIVE_TYPE_NAME    = "FFI_PRIMITIVE_TYPE"
	FFI_POINTER_TYPE_NAME      = "FFI_POINTER_TYPE"
	FFI_POINTER_TYPE_NAME_INST = "FFI_POINTER"
	FFI_STRUCT_TYPE_NAME       = "FFI_STRUCT_TYPE"
	FFI_UNION_TYPE_NAME        = "FFI_UNION_TYPE"
	FFI_UNION_INST_TYPE_NAME   = "FFI_UNION_INSTANCE"
	FFI_LIBRARY_TYPE_NAME      = "FFI_LIBRARY"
	FFI_FUNCTION_TYPE_NAME     = "FFI_FUNCTION"
	FFI_CALLBACK_TYPE_NAME     = "FFI_CALLBACK"

	// --- Inspection Formats ---
	FFI_PRIMITIVE_INSPECT     = "<ffi_type %s>"
	FFI_POINTER_INSPECT       = "<ffi_type POINTER TO %s>"
	FFI_POINTER_ARRAY_INSPECT = "<ffi_type POINTER TO %s[%d]>"
	FFI_VOID_P_INSPECT        = "<ffi_type c_void_p>"
	FFI_STRUCT_INSPECT        = "<ffi_type struct %s>"
	FFI_UNION_INSPECT         = "<ffi_type union %s>"
	FFI_UNION_FREED_INSPECT   = "<freed union %s>"
	FFI_UNION_INST_INSPECT    = "<union %s instance at %p>"
	FFI_LIBRARY_INSPECT       = "<ffi.Library '%s' from %s>"
	FFI_FUNCTION_INSPECT      = "<ffi.Function %s from %s>"
	FFI_POINTER_NULL_INSPECT  = "<ffi.Pointer NULL>"
	FFI_POINTER_INST_INSPECT  = "<ffi.Pointer at %p>"
	FFI_CALLBACK_INSPECT      = "<ffi.Callback for %s>"

	// --- Attributes and Method Names ---
	FFI_SIZE_METHOD_NAME     = "Size"
	FFI_POINTER_ADDRESS_ATTR = "Address"
	FFI_UNION_ADDRESS_ATTR   = "address"
	FFI_IS_CALLBACK_ATTR     = "is_callback"

	// --- Builtin Method Display Names ---
	FFI_PRIMITIVE_SIZE_BUILTIN_NAME = "FFIType.Size"
	FFI_POINTER_SIZE_BUILTIN_NAME   = "FFIPointerType.Size"
	FFI_POINTER_CALL_BUILTIN_NAME   = "FFIPointerType.__call__"
	FFI_STRUCT_SIZE_BUILTIN_NAME    = "FFIStructType.Size"
	FFI_UNION_SIZE_BUILTIN_NAME     = "FFIUnionType.Size"

	// --- Environment Keys (Module Scope) ---
	FFI_ENV_LOAD_LIBRARY        = "load_library"
	FFI_ENV_DEFINE_FUNCTION     = "define_function"
	FFI_ENV_CALL_FUNCTION       = "call_function"
	FFI_ENV_MALLOC              = "malloc"
	FFI_ENV_FREE                = "free"
	FFI_ENV_MEMCPY              = "memcpy"
	FFI_ENV_ADDRESSOF           = "addressof"
	FFI_ENV_READ_MEMORY         = "read_memory"
	FFI_ENV_WRITE_MEMORY        = "write_memory"
	FFI_ENV_WRITE_MEMORY_OFFSET = "write_memory_with_offset"
	FFI_ENV_READ_MEMORY_OFFSET  = "read_memory_with_offset"
	FFI_ENV_CALLBACK            = "callback"
	FFI_ENV_BUFFER_TO_BYTES     = "buffer_to_bytes"
	FFI_ENV_GET_CREATE_PTR_TYPE = "_get_or_create_pointer_type"
	FFI_ENV_CREATE_PTR_TYPE     = "_create_pointer_type"
	FFI_ENV_CREATE_STRUCT_TYPE  = "create_struct_type"
	FFI_ENV_CREATE_UNION_TYPE   = "create_union_type"
	FFI_ENV_FREE_CALLBACK       = "free_callback"
	FFI_ENV_STRING_AT           = "string_at"
	FFI_ENV_GET_FUNC_ADDRESS    = "get_func_address"
	FFI_ENV_FREE_C_RESOURCE     = "free_c_resource"
	FFI_ENV_ERROR               = "error"

	// --- Builtin Names (Registered in Env) ---
	FFI_BUILTIN_LOAD_LIBRARY        = "_ffi.load_library"
	FFI_BUILTIN_DEFINE_FUNCTION     = "_ffi.define_function"
	FFI_BUILTIN_CALL_FUNCTION       = "_ffi.call_function"
	FFI_BUILTIN_MALLOC              = "_ffi.malloc"
	FFI_BUILTIN_FREE                = "_ffi.free"
	FFI_BUILTIN_MEMCPY              = "_ffi.memcpy"
	FFI_BUILTIN_ADDRESSOF           = "_ffi.addressof"
	FFI_BUILTIN_READ_MEMORY         = "_ffi.read_memory"
	FFI_BUILTIN_WRITE_MEMORY        = "_ffi.write_memory"
	FFI_BUILTIN_WRITE_MEMORY_OFFSET = "_ffi.write_memory_with_offset"
	FFI_BUILTIN_READ_MEMORY_OFFSET  = "_ffi.read_memory_with_offset"
	FFI_BUILTIN_CALLBACK            = "_ffi.callback"
	FFI_BUILTIN_BUFFER_TO_BYTES     = "_ffi.buffer_to_bytes"
	FFI_BUILTIN_GET_CREATE_PTR_TYPE = "_ffi._get_or_create_pointer_type"
	FFI_BUILTIN_CREATE_PTR_TYPE     = "_ffi._create_pointer_type"
	FFI_BUILTIN_CREATE_STRUCT_TYPE  = "_ffi.create_struct_type"
	FFI_BUILTIN_CREATE_UNION_TYPE   = "_ffi.create_union_type"
	FFI_BUILTIN_FREE_CALLBACK       = "_ffi.free_callback"
	FFI_BUILTIN_STRING_AT           = "_ffi.string_at"
	FFI_BUILTIN_GET_FUNC_ADDRESS    = "_ffi.get_func_address"
	FFI_BUILTIN_FREE_C_RESOURCE     = "_ffi.free_c_resource"

	// --- C Types Environment Keys & Strings ---
	FFI_TYPE_NAME_C_INT8      = "c_int8"
	FFI_TYPE_NAME_C_UINT8     = "c_uint8"
	FFI_TYPE_NAME_C_INT32     = "c_int32"
	FFI_TYPE_NAME_C_UINT32    = "c_uint32"
	FFI_TYPE_NAME_C_INT64     = "c_int64"
	FFI_TYPE_NAME_C_UINT64    = "c_uint64"
	FFI_TYPE_NAME_C_FLOAT     = "c_float"
	FFI_TYPE_NAME_C_DOUBLE    = "c_double"
	FFI_TYPE_NAME_C_CHAR      = "c_char"
	FFI_TYPE_NAME_C_UCHAR     = "c_uchar"
	FFI_TYPE_NAME_C_SHORT     = "c_short"
	FFI_TYPE_NAME_C_USHORT    = "c_ushort"
	FFI_TYPE_NAME_C_LONG      = "c_long"
	FFI_TYPE_NAME_C_ULONG     = "c_ulong"
	FFI_TYPE_NAME_C_LONGLONG  = "c_longlong"
	FFI_TYPE_NAME_C_ULONGLONG = "c_ulonglong"
	FFI_TYPE_NAME_C_BOOL      = "c_bool"
	FFI_TYPE_NAME_C_WCHAR_T   = "c_wchar_t"
	FFI_TYPE_NAME_C_CHAR_P    = "c_char_p"
	FFI_TYPE_NAME_C_WCHAR_P   = "c_wchar_p"
	FFI_TYPE_NAME_C_PID_T     = "c_pid_t"
	FFI_TYPE_NAME_C_TIME_T    = "c_time_t"
	FFI_TYPE_NAME_C_FILE_P    = "c_file_p"
	FFI_TYPE_NAME_C_DIR_P     = "c_dir_p"
	FFI_TYPE_NAME_C_HANDLE    = "c_handle"
	FFI_TYPE_C_VOID_P         = "c_void_p"
	FFI_TYPE_VOID             = "void"
	FFI_PTR_SUFFIX            = "*"

	// --- Module Info ---
	FFI_MODULE_NAME       = "_ffi_native"
	FFI_MODULE_PATH       = "<builtin_ffi>"
	FFI_ERROR_CLASS_NAME  = "FFIError"
	FFI_UNKNOWN_FUNC_NAME = "<unknown>"

	// --- Basic Errors & Messages ---
	FFI_ERR_PREFIX                      = "FFI Error: %s"
	FFI_SIZE_TAKES_NO_ARGS              = "Size() takes no arguments"
	FFI_EXPECTED_INT                    = "expected int, got %s"
	FFI_EXPECTED_FLOAT                  = "expected float, got %s"
	FFI_UNSUPPORTED_PRIMITIVE_MARSHAL   = "unsupported primitive type for marshalling: %s"
	FFI_UNSUPPORTED_PRIMITIVE_UNMARSHAL = "unsupported primitive type for unmarshalling: %s"
	FFI_POINTER_INSTANTIATION_ERR       = "Pointer type can only be instantiated with 0 or None"
	
	// --- Array/Pointer Errors ---
	FFI_ARRAY_LEN_MISMATCH       = "list length %d does not match array size %d"
	FFI_MALLOC_FIXED_ARRAY_ERR   = "failed to malloc for fixed array"
	FFI_MARSHAL_ARRAY_ELEM_ERR   = "failed to marshal array element: %v"
	FFI_CONVERT_ARRAY_ERR        = "cannot convert Pylearn type %s to C array[%d]"
	FFI_CONVERT_BYTES_PTR_ERR    = "cannot automatically convert Pylearn bytes to pointer of type %s"
	FFI_MALLOC_WCHAR_STR_ERR     = "failed to malloc for wchar_t string"
	FFI_WCHAR_TOC_NOT_IMPL       = "wchar_t* ToC not fully implemented for size %d"
	FFI_CONVERT_PTR_ERR          = "cannot convert Pylearn type %s to C pointer"
	FFI_NULL_PTR_ARRAY_READ_ERR  = "cannot read from NULL pointer for array"
	FFI_UNMARSHAL_ARRAY_ELEM_ERR = "failed to unmarshal array element [%d]: %v"
	FFI_WCHAR_FROMC_NOT_IMPL     = "wchar_t* FromC not fully implemented for size %d"
	FFI_UNSUPPORTED_WCHAR_SIZE   = "Unsupported wchar_t size: %d"
	FFI_WCHAR_INT_CONV_ERR       = "unsupported wchar_t size for integer conversion: %d"
	FFI_WCHAR_STR_CONV_ERR       = "unsupported wchar_t size for string conversion: %d"
	FFI_CONVERT_WCHAR_ERR        = "cannot convert Pylearn type %s to C wchar_t"
	FFI_WCHAR_READ_ERR           = "unsupported wchar_t size for reading: %d"

	// --- C-Resource / Struct / Union Errors ---
	FFI_FREE_C_RES_ARG_ERR         = "free_c_resource() takes 1 argument"
	FFI_FREE_C_RES_TYPE_ERR        = "arg must be a Pointer holding a C resource"
	FFI_MALLOC_STRUCT_ELEMS_ERR    = "FFI: failed to malloc for struct elements"
	FFI_MALLOC_FFI_TYPE_STRUCT_ERR = "FFI: failed to malloc for ffi_type struct"
	FFI_STRUCT_LAYOUT_WARN         = "FFI Warning: could not pre-calculate layout for struct %s\n"
	FFI_CONVERT_STRUCT_ERR         = "cannot convert Pylearn type %s to C struct %s"
	FFI_MARSHAL_STRUCT_FIELD_ERR   = "failed to marshal struct field '%s': %v"
	FFI_UNMARSHAL_STRUCT_FIELD_ERR = "failed to unmarshal struct field '%s': %v"
	FFI_UNION_UNINIT_ERR           = "CUnionType used without being properly initialized via create_union_type"
	FFI_MALLOC_UNION_ELEMS_ERR     = "FFI: failed to malloc for union elements"
	FFI_MALLOC_FFI_TYPE_UNION_ERR  = "FFI: failed to malloc for ffi_type union"
	FFI_UNION_LAYOUT_WARN          = "FFI Warning: could not pre-calculate layout for union %s\n"
	FFI_CONVERT_UNION_ERR          = "cannot convert Pylearn type %s to C union %s; expected Dict"
	FFI_UNION_TOC_DICT_ERR         = "union ToC expects a Dict with exactly one key-value pair to specify the active member"
	FFI_UNION_KEY_TYPE_ERR         = "union key must be a string representing a member name"
	FFI_UNION_MEMBER_NOT_FOUND_ERR = "union '%s' has no member named '%s'"
	FFI_UNION_TOC_INTERNAL_ERR     = "internal error during union ToC"
	FFI_MALLOC_UNION_INST_ERR      = "failed to malloc for union instance"
	FFI_UNION_FREED_ACCESS_ERR     = "cannot access members of a freed union instance"
	FFI_UNION_READ_MEMBER_ERR      = "failed to read union member '%s': %v"
	FFI_UNION_WRITE_MEMBER_ERR     = "failed to write to union member '%s': %v"

	// --- Library and Function Definition Errors ---
	FFI_MALLOC_ARG_TYPES_ERR      = "failed to malloc for arg types array"
	FFI_PREP_CIF_ERR              = "libffi ffi_prep_cif failed: %d"
	FFI_ARITY_MISMATCH_ERR        = "arity mismatch: %s expects %d, got %d"
	FFI_MALLOC_ARG_PTRS_ERR       = "failed to malloc arg pointers array"
	FFI_MALLOC_ARG_ERR            = "failed to malloc for argument"
	FFI_CONVERT_ARG_ERR           = "failed to convert arg %d: %v"
	FFI_MALLOC_RET_ERR            = "failed to malloc for return value"
	FFI_CONVERT_RET_ERR           = "failed to convert return value: %v"
	FFI_VARIADIC_ARITY_ERR        = "variadic function %s expects at least %d fixed args, got %d"
	FFI_VARIADIC_INFER_ERR        = "cannot infer FFI type for variadic arg %d of type %s"
	FFI_MALLOC_VARIADIC_TYPES_ERR = "failed to malloc for variadic arg types"
	FFI_PREP_CIF_VAR_ERR          = "ffi_prep_cif_var failed: %d"

	// --- Builtin Function Errors ---
	FFI_MALLOC_ARG_COUNT_ERR           = "malloc() takes 1 argument"
	FFI_SIZE_INT_ERR                   = "size must be an integer"
	FFI_MALLOC_FAIL_ERR                = "malloc failed"
	FFI_FREE_ARG_COUNT_ERR             = "free() takes 1 argument"
	FFI_ARG_MUST_BE_PTR_ERR            = "arg must be a Pointer"
	FFI_MEMCPY_ARG_COUNT_ERR           = "memcpy() takes 3 arguments"
	FFI_MEMCPY_ARG_TYPE_ERR            = "args must be (Pointer, Pointer, Integer)"
	FFI_ADDRESSOF_ARG_COUNT_ERR        = "addressof() takes 1 argument"
	FFI_ADDRESSOF_UNSUPPORTED_ERR      = "addressof() unsupported for type %s"
	FFI_READ_MEM_ARG_COUNT_ERR         = "read_memory() takes 2 arguments"
	FFI_PTR_TYPE_ARGS_ERR              = "args must be (Pointer, FFIType)"
	FFI_NULL_PTR_READ_ERR              = "cannot read from NULL pointer"
	FFI_READ_FAIL_ERR                  = "read failed: %v"
	FFI_WRITE_MEM_ARG_COUNT_ERR        = "write_memory() takes 3 arguments"
	FFI_WRITE_MEM_ARG_TYPE_ERR         = "args must be (Pointer, FFIType, value)"
	FFI_NULL_PTR_WRITE_ERR             = "cannot write to NULL pointer"
	FFI_WRITE_FAIL_ERR                 = "write failed: %v"
	FFI_WRITE_MEM_OFFSET_ARG_COUNT_ERR = "write_memory_with_offset() takes 4 arguments"
	FFI_WRITE_MEM_OFFSET_ARG_TYPE_ERR  = "args must be (Pointer, Integer, FFIType, value)"
	FFI_WRITE_OFFSET_FAIL_ERR          = "write with offset failed: %v"
	FFI_READ_MEM_OFFSET_ARG_COUNT_ERR  = "read_memory_with_offset() takes 3 arguments"
	FFI_READ_MEM_OFFSET_ARG_TYPE_ERR   = "args must be (Pointer, Integer, FFIType)"
	FFI_READ_OFFSET_FAIL_ERR           = "read with offset failed: %v"
	FFI_BUF_TO_BYTES_ARG_COUNT_ERR     = "buffer_to_bytes() takes 2 arguments"
	FFI_LEN_NEGATIVE_ERR               = "length cannot be negative"
	FFI_FREE_CB_ARG_COUNT_ERR          = "free_callback() takes 1 argument"
	FFI_ARG_MUST_BE_CB_ERR             = "argument must be a callback object"

	// --- Callback Errors ---
	FFI_CONVERT_TO_CB_ERR         = "cannot convert %T to callback"
	FFI_CONVERT_C_PTR_TO_CB_ERR   = "cannot convert C pointer to callback object"
	FFI_EXEC_CTX_NIL_CB_ERR       = "execution context cannot be nil for callback"
	FFI_MALLOC_ARG_TYPES_FAIL_ERR = "malloc failed for arg types"
	FFI_PREP_CIF_FAIL_ERR         = "ffi_prep_cif failed"
	FFI_CLOSURE_ALLOC_FAIL_ERR    = "ffi_closure_alloc failed"
	FFI_MALLOC_USER_DATA_FAIL_ERR = "malloc for user_data failed"
	FFI_PREP_CLOSURE_LOC_FAIL_ERR = "ffi_prep_closure_loc failed"
	FFI_FATAL_PANIC_CB_ERR        = "\n--- FFI FATAL: Panic in callback function ---\n%v\n"
	FFI_FATAL_MISSING_CTX_ERR     = "FFI FATAL: Callback is missing its ExecutionContext"
	FFI_UNMARSHAL_CB_ARG_ERR      = "FFI ERROR: Failed to unmarshal arg %d: %v\n"
	FFI_UNHANDLED_EXC_CB_HEADER   = "\n--- Unhandled exception in FFI callback ---"
	FFI_CB_TRACEBACK_FILE_FMT     = "  File \"<c_callback>\", in %s\n"
	FFI_CB_EXC_FMT                = "%s: %s\n"
	FFI_CB_EXC_FOOTER             = "--- End of FFI callback exception ---"
	FFI_MARSHAL_RET_ERR           = "FFI ERROR: Failed to marshal return value: %v\n"

	// --- High-Level FFI Parsing Errors ---
	FFI_LOAD_LIB_ARG_COUNT_ERR       = "load_library() takes 1 argument"
	FFI_ARG_MUST_BE_STR_ERR          = "arg must be a string"
	FFI_DEF_FUNC_ARG_COUNT_ERR       = "define_function() takes 4 or 5 arguments"
	FFI_DEF_FUNC_ARG_TYPE_ERR        = "args must be (Library, string, ...)"
	FFI_RET_TYPE_INVALID_ERR         = "return_type is not a valid FFI type"
	FFI_ARG_TYPES_NOT_LIST_ERR       = "arg_types must be a list"
	FFI_ARG_TYPE_INVALID_ERR         = "item in arg_types is not a valid FFI type"
	FFI_IS_VAR_NOT_BOOL_ERR          = "arg 5 (is_variadic) must be a boolean"
	FFI_CALL_FUNC_REQ_FUNC_ERR       = "call_function() requires a function argument"
	FFI_ARG_MUST_BE_FFI_FUNC_ERR     = "arg must be an FFI Function"
	FFI_CALLBACK_ARG_COUNT_ERR       = "callback() takes 3 arguments"
	FFI_CB_ARG1_NOT_CALLABLE_ERR     = "arg 1 must be callable"
	FFI_CB_RESTYPE_INVALID_ERR       = "restype is not valid FFI type"
	FFI_CB_ARGTYPES_NOT_LIST_ERR     = "argtypes must be a list"
	FFI_CB_ARGTYPE_INVALID_ERR       = "item in argtypes not a valid FFI type"
	FFI_STR_AT_ARG_COUNT_ERR         = "string_at() takes 1 to 3 arguments"
	FFI_ARG1_MUST_BE_PTR_ERR         = "arg 1 must be a Pointer"
	FFI_ARG2_MUST_BE_INT_ERR         = "arg 2 (length) must be an Integer"
	FFI_ARG3_MUST_BE_INT_ERR         = "arg 3 (offset) must be an Integer"
	FFI_GET_FUNC_ADDR_ARG_COUNT_ERR  = "get_func_address() takes 2 arguments"
	FFI_GET_FUNC_ADDR_ARG_TYPE_ERR   = "args must be (Library, string)"
	FFI_CREATE_STRUCT_ARG_COUNT_ERR  = "create_struct_type() takes 2 arguments (name, fields_list)"
	FFI_CALC_STRUCT_LAYOUT_ERR       = "failed to calculate struct layout: %v"
	FFI_GET_CREATE_PTR_ARG_COUNT_ERR = "_get_or_create_pointer_type() takes 1 argument (pointee_type)"
	FFI_ARG_VALID_FFI_TYPE_ERR       = "argument must be a valid FFI type"
	FFI_CREATE_PTR_ARG_COUNT_ERR     = "_create_pointer_type() takes 2 arguments (pointee_type, array_size)"
	FFI_CREATE_PTR_ARG1_ERR          = "argument 1 (pointee_type) must be an FFI type or NULL"
	FFI_CREATE_PTR_ARG2_ERR          = "argument 2 (array_size) must be an integer"
	FFI_ARRAY_SIZE_NEG_ERR           = "array_size must be non-negative"
	FFI_CREATE_UNION_ARG_ERR         = "create_union_type() takes 2 arguments (name, fields_list)"
	FFI_ARG1_NAME_STR_ERR            = "argument 1 (name) must be a string"
	FFI_ARG2_FIELDS_LIST_ERR         = "argument 2 (fields) must be a list"
	FFI_FIELD_TUPLE_ERR              = "field %d must be a list or tuple of (name, type)"
	FFI_FIELD_LEN_ERR                = "field %d must have exactly 2 elements: (name, type)"
	FFI_FIELD_NAME_STR_ERR           = "field %d name must be a string"
	FFI_FIELD_TYPE_ERR               = "field %d type must be a valid FFI type"
	FFI_UNION_NO_FIELDS_ERR          = "cannot create a union with no fields"
	FFI_UNSUPPORTED_C_CHAR_SIZE      = "Unsupported C char size: %d"
	FFI_UNSUPPORTED_C_SHORT_SIZE     = "Unsupported C short size: %d"
	FFI_UNSUPPORTED_C_LONG_SIZE      = "Unsupported C long size: %d"
	FFI_UNSUPPORTED_C_BOOL_SIZE      = "Unsupported C _Bool size: %d"
)