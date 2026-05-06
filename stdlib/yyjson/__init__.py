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
# We must use yyjson_read_opts because yyjson_read is static inline in the C header
_yyjson_read_opts = _lib.yyjson_read_opts([ffi.c_char_p, ffi.c_uint64, ffi.c_uint32, ffi.c_void_p, ffi.c_void_p], ffi.c_void_p)

# --- Memory Parsing Logic (loads) ---
def _parse_val(root_ptr, offset):
    """
    Directly reads a yyjson_val struct from memory.
    yyjson_val is exactly 16 bytes.
    Returns: (parsed_value, offset_of_next_sibling)
    """
    tag_obj = ffi.read_memory_with_offset(root_ptr, offset, ffi.c_uint64)
    tag = int(str(tag_obj))
    
    # 3 bits for type
    typ = tag & 7 
    # 5 bits for subtype
    subtype = (tag >> 3) & 31 
    
    # By default, primitive values take exactly 16 bytes
    next_off = offset + 16
    
    # For Collections (Array/Object), yyjson stores the total byte size
    # of the collection (including children) in the union field at offset + 8
    if typ == 6 or typ == 7:
        ofs_obj = ffi.read_memory_with_offset(root_ptr, offset + 8, ffi.c_uint64)
        next_off = offset + int(str(ofs_obj))
    
    if typ == 2: # NULL
        return None, next_off
    elif typ == 3: # BOOL
        return (subtype == 1), next_off
    elif typ == 4: # NUM
        if subtype == 0:   # uint
            num_obj = ffi.read_memory_with_offset(root_ptr, offset + 8, ffi.c_uint64)
            return int(str(num_obj)), next_off
        elif subtype == 1: # sint
            num_obj = ffi.read_memory_with_offset(root_ptr, offset + 8, ffi.c_int64)
            return int(str(num_obj)), next_off
        elif subtype == 2: # real
            num_obj = ffi.read_memory_with_offset(root_ptr, offset + 8, ffi.c_double)
            return float(str(num_obj)), next_off
    elif typ == 5: # STR
        str_ptr = ffi.read_memory_with_offset(root_ptr, offset + 8, ffi.c_void_p)
        # Length is stored in the upper 56 bits
        str_len = tag >> 8
        str_val = ffi.string_at(str_ptr, str_len)
        return str_val if isinstance(str_val, str) else "", next_off
    elif typ == 6: # ARR
        # Number of elements is stored in the upper 56 bits
        count = tag >> 8
        arr = []
        curr_off = offset + 16 # First element immediately follows the array header
        for _ in range(count):
            val, curr_off = _parse_val(root_ptr, curr_off)
            arr.append(val)
        return arr, next_off
    elif typ == 7: # OBJ
        count = tag >> 8
        obj = {}
        curr_off = offset + 16
        for _ in range(count):
            # Keys and values alternate sequentially in memory
            key, curr_off = _parse_val(root_ptr, curr_off)
            val, curr_off = _parse_val(root_ptr, curr_off)
            obj[key] = val
        return obj, next_off
        
    raise YYJSONError(format_str("Unknown yyjson type tag: {typ} (Full tag: {tag})"))

def loads(json_string):
    if not isinstance(json_string, str):
        raise TypeError("JSON document must be a string")
        
    json_bytes = json_string.encode('utf-8')
    
    # Call yyjson_read_opts instead of yyjson_read
    doc_ptr = _yyjson_read_opts(json_bytes, len(json_bytes), 0, None, None)
    if not doc_ptr or not getattr(doc_ptr, "Address", None):
        raise YYJSONError("Failed to parse JSON string")
        
    try:
        # The root yyjson_val pointer is the first member (offset 0) of the doc struct
        root_val_ptr = ffi.read_memory_with_offset(doc_ptr, 0, ffi.c_void_p)
        if not root_val_ptr or not getattr(root_val_ptr, "Address", None):
            return None
            
        result, _ = _parse_val(root_val_ptr, 0)
        return result
    finally:
        # Since we use default allocator, ffi.free is perfectly safe and equivalent to yyjson_doc_free
        ffi.free(doc_ptr)

# --- Tree Building Logic (dumps) ---
# Since yyjson mutable API is largely static inline, we use a pure-Python string builder
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
        items = []
        for pair in obj.items():
            k = pair[0]
            v = pair[1]
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