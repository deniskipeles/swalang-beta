"""
A high-performance Pylearn wrapper for the cJSON library, built using the Pylearn FFI.
This module provides fast JSON encoding and decoding.
"""

import ffi
import sys

# ==============================================================================
#  Platform-Aware Library Loading
# ==============================================================================

def _load_library():
    platform = sys.platform
    candidates =[]
    if platform == 'linux':
        candidates =["bin/x86_64-linux/cjson/libcjson.so", "libcjson.so"]
    elif platform == 'windows':
        candidates =["bin/x86_64-windows-gnu/cjson/cjson.dll", "cjson.dll"]
    elif platform == 'darwin':
        candidates = ["libcjson.dylib"]
    
    for name in candidates:
        try:
            return ffi.CDLL(name)
        except ffi.FFIError:
            pass
    raise ffi.FFIError("Could not load cjson shared library")

_lib = _load_library()

CJSON_AVAILABLE = False
if _lib:
    CJSON_AVAILABLE = True
else:
    print("JSON functionality in the 'cjson' module will not be available.")

# ==============================================================================
#  Exception and Constants
# ==============================================================================

class CJSONError(Exception):
    """Exception raised for cJSON-related errors."""
    pass

# cJSON object types
cJSON_Invalid = 0
cJSON_False = (1 << 0)
cJSON_True = (1 << 1)
cJSON_NULL = (1 << 2)
cJSON_Number = (1 << 3)
cJSON_String = (1 << 4)
cJSON_Array = (1 << 5)
cJSON_Object = (1 << 6)

# ==============================================================================
#  Define C Function Signatures
# ==============================================================================

if CJSON_AVAILABLE:
    # --- Parsing and Printing ---
    _cJSON_Parse = _lib.cJSON_Parse([ffi.c_char_p], ffi.c_void_p)
    # FIX: Ensure Print returns void_p so we get the raw pointer back to free it
    _cJSON_PrintUnformatted = _lib.cJSON_PrintUnformatted([ffi.c_void_p], ffi.c_void_p)
    _cJSON_Print = _lib.cJSON_Print([ffi.c_void_p], ffi.c_void_p)
    _cJSON_Delete = _lib.cJSON_Delete([ffi.c_void_p], None)

    # --- Object Creation ---
    _cJSON_CreateNull = _lib.cJSON_CreateNull([], ffi.c_void_p)
    _cJSON_CreateTrue = _lib.cJSON_CreateTrue([], ffi.c_void_p)
    _cJSON_CreateFalse = _lib.cJSON_CreateFalse([], ffi.c_void_p)
    _cJSON_CreateNumber = _lib.cJSON_CreateNumber([ffi.c_double], ffi.c_void_p)
    _cJSON_CreateString = _lib.cJSON_CreateString([ffi.c_char_p], ffi.c_void_p)
    _cJSON_CreateArray = _lib.cJSON_CreateArray([], ffi.c_void_p)
    _cJSON_CreateObject = _lib.cJSON_CreateObject([], ffi.c_void_p)

    # --- Object Manipulation ---
    # cJSON uses 32-bit ints for booleans inside its API.
    _cJSON_AddItemToArray = _lib.cJSON_AddItemToArray([ffi.c_void_p, ffi.c_void_p], ffi.c_int32)
    _cJSON_AddItemToObject = _lib.cJSON_AddItemToObject([ffi.c_void_p, ffi.c_char_p, ffi.c_void_p], ffi.c_int32)

    # --- Object Traversal and Reading ---
    _cJSON_GetObjectItem = _lib.cJSON_GetObjectItem([ffi.c_void_p, ffi.c_char_p], ffi.c_void_p)
    _cJSON_GetArraySize = _lib.cJSON_GetArraySize([ffi.c_void_p], ffi.c_int32)
    _cJSON_GetArrayItem = _lib.cJSON_GetArrayItem([ffi.c_void_p, ffi.c_int32], ffi.c_void_p)

# ==============================================================================
#  Internal Helper Functions for Conversion
# ==============================================================================

def _cjson_to_pylearn(item_ptr):
    if not item_ptr or getattr(item_ptr, "Address", 0) == 0:
        return None

    # Mask with 0xFF to remove extra cJSON flags (like cJSON_IsReference)
    item_type_obj = ffi.read_memory_with_offset(item_ptr, 24, ffi.c_int32)
    item_type = int(str(item_type_obj)) & 0xFF

    if item_type == cJSON_False:
        return False
    elif item_type == cJSON_True:
        return True
    elif item_type == cJSON_NULL:
        return None
    elif item_type == cJSON_Number:
        num_obj = ffi.read_memory_with_offset(item_ptr, 48, ffi.c_double)
        num = float(str(num_obj))
        if num == int(num):
            return int(num)
        return num
    elif item_type == cJSON_String:
        str_val = ffi.read_memory_with_offset(item_ptr, 32, ffi.c_char_p)
        if isinstance(str_val, str):
            return str_val
        return ""
    elif item_type == cJSON_Array:
        arr =[]
        size = _cJSON_GetArraySize(item_ptr)
        i = 0
        while i < size:
            elem_ptr = _cJSON_GetArrayItem(item_ptr, i)
            arr.append(_cjson_to_pylearn(elem_ptr))
            i = i + 1
        return arr
    elif item_type == cJSON_Object:
        obj = {}
        child_ptr = ffi.read_memory_with_offset(item_ptr, 16, ffi.c_void_p)
        
        while child_ptr and getattr(child_ptr, "Address", 0) != 0:
            key_val = ffi.read_memory_with_offset(child_ptr, 56, ffi.c_char_p)
            key = key_val if isinstance(key_val, str) else ""
            
            obj[key] = _cjson_to_pylearn(child_ptr)
            child_ptr = ffi.read_memory_with_offset(child_ptr, 0, ffi.c_void_p)
        return obj

    return None

def _pylearn_to_cjson(item):
    if item is None:
        return _cJSON_CreateNull()
    elif isinstance(item, bool):
        return _cJSON_CreateTrue() if item else _cJSON_CreateFalse()
    elif isinstance(item, (int, float)):
        return _cJSON_CreateNumber(float(item))
    elif isinstance(item, str):
        return _cJSON_CreateString(item.encode('utf-8'))
    elif isinstance(item, list):
        c_array = _cJSON_CreateArray()
        for elem in item:
            c_elem = _pylearn_to_cjson(elem)
            _cJSON_AddItemToArray(c_array, c_elem)
        return c_array
    elif isinstance(item, dict):
        c_obj = _cJSON_CreateObject()
        for pair in item.items(): # Safe unpack since Go backend was fixed
            key = pair[0]
            value = pair[1]
            if not isinstance(key, str):
                key = str(key)
            c_value = _pylearn_to_cjson(value)
            _cJSON_AddItemToObject(c_obj, key.encode('utf-8'), c_value)
        return c_obj
    else:
        raise CJSONError(format_str("Object of type '{type(item)}' is not JSON serializable"))

# ==============================================================================
#  High-Level Pylearn API
# ==============================================================================

def loads(json_string):
    if not CJSON_AVAILABLE:
        raise CJSONError("cJSON library not available")
    if not isinstance(json_string, str):
        raise TypeError("the JSON object must be a string")

    root_ptr = _cJSON_Parse(json_string.encode('utf-8'))
    if not root_ptr or getattr(root_ptr, "Address", 0) == 0:
        raise CJSONError("Failed to parse JSON string.")
    
    try:
        pylearn_obj = _cjson_to_pylearn(root_ptr)
        return pylearn_obj
    finally:
        _cJSON_Delete(root_ptr)

def dumps(obj, indent=False):
    if not CJSON_AVAILABLE:
        raise CJSONError("cJSON library not available")

    cjson_obj_ptr = _pylearn_to_cjson(obj)
    if not cjson_obj_ptr or getattr(cjson_obj_ptr, "Address", 0) == 0:
        raise CJSONError("Failed to convert Pylearn object to cJSON structure.")

    try:
        if indent:
            char_ptr = _cJSON_Print(cjson_obj_ptr)
        else:
            char_ptr = _cJSON_PrintUnformatted(cjson_obj_ptr)
            
        if not char_ptr or getattr(char_ptr, "Address", 0) == 0:
            raise CJSONError("Failed to print cJSON structure to string.")
        
        try:
            result_string = ffi.string_at(char_ptr)
            return result_string
        finally:
            ffi.free(char_ptr)
            
    finally:
        _cJSON_Delete(cjson_obj_ptr)