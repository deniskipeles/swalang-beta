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
        candidates =["bin/x86_64-linux/yyjson/libyyjson.so", "libyyjson.so"]
    elif platform == 'windows':
        candidates =["bin/x86_64-windows-gnu/yyjson/yyjson.dll", "yyjson.dll"]
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
# We must use yyjson_read_opts because yyjson_read is static inline in the C header
_yyjson_read_opts = _lib.yyjson_read_opts([ffi.c_char_p, ffi.c_uint64, ffi.c_uint32, ffi.c_void_p, ffi.c_void_p], ffi.c_void_p)

# We CANNOT bind to yyjson_doc_free because it is static inline.
# However, yyjson_read_opts allocates a single contiguous block using the
# default allocator (libc malloc) when we pass NULL for the allocator.
# Therefore, we can safely use ffi.free() to clean it up!

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
        arr =[]
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
    
    # Call yyjson_read_opts instead of yyjson_read
    doc_ptr = _yyjson_read_opts(json_bytes, len(json_bytes), 0, None, None)
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
        # Since we use default allocator, ffi.free is perfectly safe and equivalent to yyjson_doc_free
        ffi.free(doc_ptr)

# --- Tree Building Logic (dumps) ---
# Since yyjson mutable API is largely static inline, we use a pure-Python string builder
# which is fast enough for basic testing, avoiding inline C functions.
def dumps(obj, indent=False, _level=0):
    if obj is None:
        return "null"
    elif isinstance(obj, bool):
        return "true" if obj else "false"
    elif isinstance(obj, (int, float)):
        return str(obj)
    elif isinstance(obj, str):
        s = obj.replace('\\', '\\\\').replace('"', '\\"').replace('\n', '\\n').replace('\r', '\\r').replace('\t', '\\t')
        return format_str('"{s}"')
    elif isinstance(obj, list) or isinstance(obj, tuple):
        if len(obj) == 0: return "[]"
        items =[]
        for item in obj:
            items.append(dumps(item, indent, _level+1))
        if indent:
            pad = "  " * (_level + 1)
            end_pad = "  " * _level
            return "[\n" + pad + (",\n" + pad).join(items) + "\n" + end_pad + "]"
        return "[" + ", ".join(items) + "]"
    elif isinstance(obj, dict):
        if len(obj) == 0: return "{}"
        items =[]
        for k, v in obj.items():
            if not isinstance(k, str):
                k = str(k)
            k_esc = k.replace('\\', '\\\\').replace('"', '\\"').replace('\n', '\\n').replace('\r', '\\r').replace('\t', '\\t')
            v_str = dumps(v, indent, _level+1)
            if indent:
                items.append(format_str('  "{k_esc}": {v_str}'))
            else:
                items.append(format_str('"{k_esc}": {v_str}'))
        if indent:
            pad = "  " * (_level + 1)
            end_pad = "  " * _level
            return "{\n" + pad + (",\n" + pad).join(items) + "\n" + end_pad + "}"
        return "{" + ", ".join(items) + "}"
    else:
        raise TypeError(format_str("Object of type {type(obj)} is not JSON serializable"))
