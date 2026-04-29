# pylearn/stdlib/cjson.py

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
    candidates = []
    if platform == 'linux':
        candidates = ["bin/x86_64-linux/cjson/libcjson.so", "libcjson.so"]
    elif platform == 'windows':
        candidates = ["bin/x86_64-windows-gnu/cjson/cjson.dll", "cjson.dll"]
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
    _cJSON_PrintUnformatted = _lib.cJSON_PrintUnformatted([ffi.c_void_p], ffi.c_char_p)
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
    _cJSON_AddItemToArray = _lib.cJSON_AddItemToArray([ffi.c_void_p, ffi.c_void_p], ffi.c_bool)
    _cJSON_AddItemToObject = _lib.cJSON_AddItemToObject([ffi.c_void_p, ffi.c_char_p, ffi.c_void_p], ffi.c_bool)

    # --- Object Traversal and Reading ---
    _cJSON_GetObjectItem = _lib.cJSON_GetObjectItem([ffi.c_void_p, ffi.c_char_p], ffi.c_void_p)
    _cJSON_GetArraySize = _lib.cJSON_GetArraySize([ffi.c_void_p], ffi.c_int32)
    _cJSON_GetArrayItem = _lib.cJSON_GetArrayItem([ffi.c_void_p, ffi.c_int32], ffi.c_void_p)

# ==============================================================================
#  Internal Helper Functions for Conversion
# ==============================================================================

def _cjson_to_pylearn(item_ptr):
    """
    Recursively converts a cJSON C structure to a Pylearn object by reading its memory directly.
    """
    if not item_ptr or not item_ptr.Address:
        return None

    # cJSON struct field offsets for a 64-bit system.
    # struct cJSON {
    #     cJSON *next;        // offset 0
    #     cJSON *prev;        // offset 8
    #     cJSON *child;       // offset 16
    #     int type;           // offset 24
    #     char *valuestring;  // offset 32
    #     int valueint;       // offset 40 (value is in valuedouble)
    #     double valuedouble; // offset 48
    #     char *string;       // offset 56 (this is the key for object items)
    # };
    
    # Read the 'type' field, which comes back as a Pylearn Integer object.
    item_type_obj = ffi.read_memory_with_offset(item_ptr, 24, ffi.c_int32)
    # Convert the Pylearn Integer object to a primitive integer for comparison.
    item_type = int(str(item_type_obj))

    if item_type == cJSON_False:
        return False
    elif item_type == cJSON_True:
        return True
    elif item_type == cJSON_NULL:
        return None
    elif item_type == cJSON_Number:
        # Read 'valuedouble' and convert the Float object to a primitive.
        num_obj = ffi.read_memory_with_offset(item_ptr, 48, ffi.c_double)
        num = float(str(num_obj))
        if num == int(num):
            return int(num)
        return num
    elif item_type == cJSON_String:
        # Read the 'valuestring' pointer, then get the string at that address.
        str_ptr = ffi.read_memory_with_offset(item_ptr, 32, ffi.c_char_p)
        return ffi.string_at(str_ptr)
    elif item_type == cJSON_Array:
        arr = []
        size = _cJSON_GetArraySize(item_ptr)
        i = 0
        while i < size:
            elem_ptr = _cJSON_GetArrayItem(item_ptr, i)
            arr.append(_cjson_to_pylearn(elem_ptr))
            i = i + 1
        return arr
    elif item_type == cJSON_Object:
        obj = {}
        # Read the 'child' pointer to get the first item in the object's linked list.
        child_ptr = ffi.read_memory_with_offset(item_ptr, 16, ffi.c_void_p)
        
        while child_ptr and child_ptr.Address:
            # Read the key from the child's 'string' field.
            key_ptr = ffi.read_memory_with_offset(child_ptr, 56, ffi.c_char_p)
            key = ffi.string_at(key_ptr)
            
            # The value is the child item itself.
            obj[key] = _cjson_to_pylearn(child_ptr)
            
            # Move to the next item in the linked list.
            child_ptr = ffi.read_memory_with_offset(child_ptr, 0, ffi.c_void_p)
        return obj

    return None

def _pylearn_to_cjson(item):
    """Recursively converts a Pylearn object to a cJSON C structure."""
    if item is None:
        return _cJSON_CreateNull()
    elif isinstance(item, bool):
        return _cJSON_CreateTrue() if item else _cJSON_CreateFalse()
    elif isinstance(item, (int, float)):
        return _cJSON_CreateNumber(float(item))
    elif isinstance(item, str):
        return _cJSON_CreateString(item)
    elif isinstance(item, list):
        c_array = _cJSON_CreateArray()
        for elem in item:
            c_elem = _pylearn_to_cjson(elem)
            _cJSON_AddItemToArray(c_array, c_elem)
        return c_array
    elif isinstance(item, dict):
        c_obj = _cJSON_CreateObject()
        for key, value in item.items():
            if not isinstance(key, str):
                raise CJSONError("JSON object keys must be strings.")
            c_value = _pylearn_to_cjson(value)
            _cJSON_AddItemToObject(c_obj, key, c_value)
        return c_obj
    else:
        raise CJSONError(format_str("Object of type '{type(item)}' is not JSON serializable"))

# ==============================================================================
#  High-Level Pylearn API
# ==============================================================================

def loads(json_string):
    """
    Parse a JSON string, returning a Pylearn object.
    """
    if not CJSON_AVAILABLE:
        raise CJSONError("cJSON library not available")
    if not isinstance(json_string, str):
        raise TypeError("the JSON object must be a string")

    root_ptr = _cJSON_Parse(json_string)
    if not root_ptr or not root_ptr.Address:
        # A proper implementation would get the error pointer from cJSON_GetErrorPtr()
        raise CJSONError("Failed to parse JSON string.")
    
    try:
        pylearn_obj = _cjson_to_pylearn(root_ptr)
        return pylearn_obj
    finally:
        # CRITICAL: Always free the memory allocated by cJSON_Parse
        _cJSON_Delete(root_ptr)

def dumps(obj):
    """
    Serialize a Pylearn object to a JSON formatted string.
    """
    if not CJSON_AVAILABLE:
        raise CJSONError("cJSON library not available")

    cjson_obj_ptr = _pylearn_to_cjson(obj)
    if not cjson_obj_ptr or not cjson_obj_ptr.Address:
        raise CJSONError("Failed to convert Pylearn object to cJSON structure.")

    try:
        # This allocates a new string that we must free.
        char_ptr = _cJSON_PrintUnformatted(cjson_obj_ptr)
        if not char_ptr or not char_ptr.Address:
            raise CJSONError("Failed to print cJSON structure to string.")
        
        try:
            result_string = ffi.string_at(char_ptr)
            return result_string
        finally:
            # cJSON's print functions use the same malloc as the main library.
            # We can use our ffi.free, which calls the standard C free().
            ffi.free(char_ptr)
            
    finally:
        # CRITICAL: Always free the object structure we built.
        _cJSON_Delete(cjson_obj_ptr)