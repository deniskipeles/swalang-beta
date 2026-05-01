# pylearn/stdlib/yyjson/__init__.py

"""
A blazing fast Pylearn wrapper for the yyjson library using the FFI.
Reads JSON structs directly from memory offsets for maximum performance.
"""

import ffi
import sys

def _load_library():
    platform = sys.platform
    candidates = []
    if platform == 'linux':
        candidates = ["bin/x86_64-linux/yyjson/libyyjson.so", "libyyjson.so"]
    elif platform == 'windows':
        candidates = ["bin/x86_64-windows-gnu/yyjson/yyjson.dll", "yyjson.dll"]
    elif platform == 'darwin':
        candidates = ["libyyjson.dylib"]
    
    for name in candidates:
        try:
            return ffi.CDLL(name)
        except ffi.FFIError:
            pass
    raise ffi.FFIError("Could not load yyjson shared library")

_lib = _load_library()

class YYJSONError(Exception):
    pass

# --- C Function Signatures ---
# Read API (Immutable)
_yyjson_read = _lib.yyjson_read([ffi.c_char_p, ffi.c_uint64, ffi.c_uint32], ffi.c_void_p)
_yyjson_doc_free = _lib.yyjson_doc_free([ffi.c_void_p], None)

# Mutate API (Mutable, used for writing/dumps)
_yyjson_mut_doc_new = _lib.yyjson_mut_doc_new([ffi.c_void_p], ffi.c_void_p)
_yyjson_mut_doc_free = _lib.yyjson_mut_doc_free([ffi.c_void_p], None)

_yyjson_mut_null = _lib.yyjson_mut_null([ffi.c_void_p], ffi.c_void_p)
_yyjson_mut_bool = _lib.yyjson_mut_bool([ffi.c_void_p, ffi.c_bool], ffi.c_void_p)
_yyjson_mut_int = _lib.yyjson_mut_int([ffi.c_void_p, ffi.c_int64], ffi.c_void_p)
_yyjson_mut_real = _lib.yyjson_mut_real([ffi.c_void_p, ffi.c_double], ffi.c_void_p)
_yyjson_mut_strncpy = _lib.yyjson_mut_strncpy([ffi.c_void_p, ffi.c_char_p, ffi.c_uint64], ffi.c_void_p)

_yyjson_mut_arr = _lib.yyjson_mut_arr([ffi.c_void_p], ffi.c_void_p)
_yyjson_mut_obj = _lib.yyjson_mut_obj([ffi.c_void_p], ffi.c_void_p)

_yyjson_mut_arr_append = _lib.yyjson_mut_arr_append([ffi.c_void_p, ffi.c_void_p], ffi.c_bool)
_yyjson_mut_obj_add = _lib.yyjson_mut_obj_add([ffi.c_void_p, ffi.c_void_p, ffi.c_void_p], ffi.c_bool)

_yyjson_mut_doc_set_root = _lib.yyjson_mut_doc_set_root([ffi.c_void_p, ffi.c_void_p], None)
_yyjson_mut_write = _lib.yyjson_mut_write([ffi.c_void_p, ffi.c_uint32, ffi.POINTER(ffi.c_uint64)], ffi.c_void_p)


# --- Memory Parsing Logic (loads) ---
def _parse_val(root_ptr, offset):
    """
    Directly reads a yyjson_val struct from memory.
    yyjson_val is exactly 16 bytes.
    Offset 0: tag (uint64)
    Offset 8: union data (8 bytes)
    """
    tag = ffi.read_memory_with_offset(root_ptr, offset, ffi.c_uint64)
    typ = tag & 0xFF
    
    if typ == 2: # NULL
        return None, 1
    elif typ == 3: # BOOL
        subtype = (tag >> 8) & 0xFF
        return (subtype == 1), 1
    elif typ == 4: # NUM
        subtype = (tag >> 8) & 0xFF
        if subtype == 0:   # uint
            return ffi.read_memory_with_offset(root_ptr, offset + 8, ffi.c_uint64), 1
        elif subtype == 1: # sint
            return ffi.read_memory_with_offset(root_ptr, offset + 8, ffi.c_int64), 1
        elif subtype == 2: # real
            return ffi.read_memory_with_offset(root_ptr, offset + 8, ffi.c_double), 1
    elif typ == 5: # STR
        str_ptr = ffi.read_memory_with_offset(root_ptr, offset + 8, ffi.c_void_p)
        str_len = (tag >> 8) & 0xFFFFFF # Length is 24 bits
        return ffi.string_at(str_ptr, str_len), 1
    elif typ == 6: # ARR
        count = (tag >> 8) & 0xFFFFFF
        total_nodes = tag >> 32 # Number of struct blocks this collection occupies
        arr = []
        curr_off = offset + 16 # First element immediately follows the array header
        for _ in range(count):
            val, nodes = _parse_val(root_ptr, curr_off)
            arr.append(val)
            curr_off = curr_off + (nodes * 16) # Skip past this element and all its children
        return arr, total_nodes
    elif typ == 7: # OBJ
        count = (tag >> 8) & 0xFFFFFF
        total_nodes = tag >> 32
        obj = {}
        curr_off = offset + 16
        for _ in range(count):
            # Keys and values alternate sequentially in memory
            key, knodes = _parse_val(root_ptr, curr_off)
            curr_off = curr_off + (knodes * 16)
            
            val, vnodes = _parse_val(root_ptr, curr_off)
            curr_off = curr_off + (vnodes * 16)
            
            obj[key] = val
        return obj, total_nodes
        
    raise YYJSONError(format_str("Unknown yyjson type tag: {typ}"))

def loads(json_string):
    if not isinstance(json_string, str):
        raise TypeError("JSON document must be a string")
        
    json_bytes = json_string.encode('utf-8')
    
    doc_ptr = _yyjson_read(json_bytes, len(json_bytes), 0)
    if not doc_ptr or doc_ptr.Address == 0:
        raise YYJSONError("Failed to parse JSON string")
        
    try:
        # The root yyjson_val pointer is the first member (offset 0) of the doc struct
        root_val_ptr = ffi.read_memory_with_offset(doc_ptr, 0, ffi.c_void_p)
        if not root_val_ptr or root_val_ptr.Address == 0:
            return None
            
        result, _ = _parse_val(root_val_ptr, 0)
        return result
    finally:
        _yyjson_doc_free(doc_ptr)

# --- Tree Building Logic (dumps) ---
def _build_val(doc_ptr, obj):
    if obj is None:
        return _yyjson_mut_null(doc_ptr)
    elif isinstance(obj, bool):
        return _yyjson_mut_bool(doc_ptr, obj)
    elif isinstance(obj, int):
        return _yyjson_mut_int(doc_ptr, obj)
    elif isinstance(obj, float):
        return _yyjson_mut_real(doc_ptr, obj)
    elif isinstance(obj, str):
        b = obj.encode('utf-8')
        return _yyjson_mut_strncpy(doc_ptr, b, len(b))
    elif isinstance(obj, list) or isinstance(obj, tuple):
        arr_ptr = _yyjson_mut_arr(doc_ptr)
        for item in obj:
            val_ptr = _build_val(doc_ptr, item)
            _yyjson_mut_arr_append(arr_ptr, val_ptr)
        return arr_ptr
    elif isinstance(obj, dict):
        obj_ptr = _yyjson_mut_obj(doc_ptr)
        for k, v in obj.items():
            if not isinstance(k, str):
                k = str(k)
            kb = k.encode('utf-8')
            key_ptr = _yyjson_mut_strncpy(doc_ptr, kb, len(kb))
            val_ptr = _build_val(doc_ptr, v)
            _yyjson_mut_obj_add(obj_ptr, key_ptr, val_ptr)
        return obj_ptr
    else:
        raise TypeError(format_str("Object of type {type(obj)} is not JSON serializable"))

def dumps(obj, indent=False):
    doc_ptr = _yyjson_mut_doc_new(None)
    if not doc_ptr or doc_ptr.Address == 0:
        raise MemoryError("Failed to allocate yyjson mutable document")
        
    try:
        root_val_ptr = _build_val(doc_ptr, obj)
        _yyjson_mut_doc_set_root(doc_ptr, root_val_ptr)
        
        len_ptr = ffi.malloc(8) # c_uint64
        try:
            # YYJSON_WRITE_PRETTY is flag 1. 0 is minified.
            flag = 1 if indent else 0
            str_ptr = _yyjson_mut_write(doc_ptr, flag, len_ptr)
            
            if not str_ptr or str_ptr.Address == 0:
                raise YYJSONError("Failed to serialize to JSON string")
                
            length = ffi.read_memory(len_ptr, ffi.c_uint64)
            result_string = ffi.string_at(str_ptr, length)
            
            # The string allocated by yyjson must be freed using libc free
            ffi.free(str_ptr)
            return result_string
        finally:
            ffi.free(len_ptr)
    finally:
        _yyjson_mut_doc_free(doc_ptr)

