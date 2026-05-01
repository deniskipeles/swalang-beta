# pylearn/stdlib/pcre2/__init__.py

"""
A complete Pylearn wrapper for the PCRE2 library using the FFI.
Provides a Python-compatible regular expression interface.
"""

import ffi
import sys

def _load_library():
    platform = sys.platform
    candidates = []
    if platform == 'linux':
        candidates = ["bin/x86_64-linux/pcre2/libpcre2-8.so", "libpcre2-8.so", "libpcre2-8.so.0"]
    elif platform == 'windows':
        candidates = ["bin/x86_64-windows-gnu/pcre2/pcre2-8.dll", "pcre2-8.dll", "libpcre2-8.dll"]
    elif platform == 'darwin':
        candidates = ["libpcre2-8.dylib"]
    
    for name in candidates:
        try:
            return ffi.CDLL(name)
        except ffi.FFIError:
            pass
    raise ffi.FFIError("Could not load pcre2-8 shared library")

_lib = _load_library()

# --- PCRE2 Constants ---
PCRE2_ANCHORED = 0x80000000
PCRE2_CASELESS = 0x00000008 # i
PCRE2_MULTILINE = 0x00000400 # m
PCRE2_DOTALL = 0x00000020 # s

# Aliases to match Python's 're' module
IGNORECASE = PCRE2_CASELESS
I = PCRE2_CASELESS
MULTILINE = PCRE2_MULTILINE
M = PCRE2_MULTILINE
DOTALL = PCRE2_DOTALL
S = PCRE2_DOTALL

class error(Exception):
    """Exception raised for PCRE2-related errors."""
    def __init__(self, code, pattern=""):
        self.code = code
        self.pattern = pattern
        
        err_buf = ffi.malloc(256)
        try:
            _pcre2_get_error_message(code, err_buf, 256)
            self.message = ffi.string_at(err_buf)
        finally:
            ffi.free(err_buf)

        full_msg = self.message
        if pattern:
            full_msg = format_str("{self.message} in pattern: '{self.pattern}'")
        super().__init__(full_msg)

# --- C Function Signatures ---
_pcre2_compile = _lib.pcre2_compile_8([ffi.c_char_p, ffi.c_uint64, ffi.c_uint32, ffi.POINTER(ffi.c_int32), ffi.POINTER(ffi.c_uint64), ffi.c_void_p], ffi.c_void_p)
_pcre2_code_free = _lib.pcre2_code_free_8([ffi.c_void_p], None)

_pcre2_match_data_create_from_pattern = _lib.pcre2_match_data_create_from_pattern_8([ffi.c_void_p, ffi.c_void_p], ffi.c_void_p)
_pcre2_match_data_free = _lib.pcre2_match_data_free_8([ffi.c_void_p], None)

_pcre2_match = _lib.pcre2_match_8([ffi.c_void_p, ffi.c_char_p, ffi.c_uint64, ffi.c_uint64, ffi.c_uint32, ffi.c_void_p, ffi.c_void_p], ffi.c_int32)

_pcre2_get_ovector_pointer = _lib.pcre2_get_ovector_pointer_8([ffi.c_void_p], ffi.c_void_p)
_pcre2_get_ovector_count = _lib.pcre2_get_ovector_count_8([ffi.c_void_p], ffi.c_uint32)
_pcre2_get_error_message = _lib.pcre2_get_error_message_8([ffi.c_int32, ffi.c_void_p, ffi.c_uint64], ffi.c_int32)

class Match:
    def __init__(self, pattern_obj, subject, match_data_ptr):
        self._pattern = pattern_obj
        self.string = subject
        self._match_data = match_data_ptr
        self._ovector = _pcre2_get_ovector_pointer(self._match_data)
        self._ovector_count = _pcre2_get_ovector_count(self._match_data)

    def _get_offsets(self, index=0):
        if index >= self._ovector_count:
            raise IndexError("No such group")
        offset = index * 2 * 8 # ffi.c_uint64.Size() is 8
        start = ffi.read_memory_with_offset(self._ovector, offset, ffi.c_uint64)
        end = ffi.read_memory_with_offset(self._ovector, offset + 8, ffi.c_uint64)
        
        # Check if the group didn't participate in the match (PCRE2_UNSET is ~0)
        # Note: We compare against a very large number close to max uint64
        if start > 4294967295: 
            return -1, -1
        return int(start), int(end)

    def group(self, index=0):
        start, end = self._get_offsets(index)
        if start == -1:
            return None
        return self.string[start:end]

    def groups(self):
        grps = []
        i = 1
        while i < self._ovector_count:
            grps.append(self.group(i))
            i = i + 1
        return tuple(grps)

    def start(self, index=0):
        s, _ = self._get_offsets(index)
        return s

    def end(self, index=0):
        _, e = self._get_offsets(index)
        return e

    def span(self, index=0):
        return self._get_offsets(index)

    def free(self):
        if self._match_data:
            _pcre2_match_data_free(self._match_data)
            self._match_data = None

    def __repr__(self):
        s, e = self.span()
        return format_str("<pcre2.Match span=({s}, {e}), match='{self.group(0)}'>")

class Pattern:
    def __init__(self, pattern, flags=0):
        self.pattern = pattern
        self.flags = flags
        
        err_code_ptr = ffi.malloc(4) # c_int32
        err_offset_ptr = ffi.malloc(8) # c_uint64
        
        try:
            # Pylearn automatically converts python strings to C strings here
            self._code = _pcre2_compile(pattern, len(pattern), flags, err_code_ptr, err_offset_ptr, None)
            
            if self._code.Address == 0:
                err_code = ffi.read_memory(err_code_ptr, ffi.c_int32)
                raise error(err_code, pattern)
        finally:
            ffi.free(err_code_ptr)
            ffi.free(err_offset_ptr)

    def _do_match(self, subject, start_offset, options):
        if not isinstance(subject, str):
            raise TypeError("Subject must be a string")

        match_data = _pcre2_match_data_create_from_pattern(self._code, None)
        if match_data.Address == 0:
            raise MemoryError("Could not create match_data block")
        
        rc = _pcre2_match(self._code, subject, len(subject), start_offset, options, match_data, None)

        if rc < 0: # An error or no match
            _pcre2_match_data_free(match_data)
            return None
        
        return Match(self, subject, match_data)

    def search(self, subject, pos=0):
        return self._do_match(subject, pos, 0)

    def match(self, subject, pos=0):
        return self._do_match(subject, pos, PCRE2_ANCHORED)

    def findall(self, subject, pos=0):
        results = []
        while pos <= len(subject):
            m = self.search(subject, pos)
            if not m:
                break
            
            num_groups = m._ovector_count - 1
            if num_groups > 0:
                grps = m.groups()
                results.append(grps if num_groups > 1 else grps[0])
            else:
                results.append(m.group(0))

            start, end = m.span()
            if start == end:
                pos = pos + 1
            else:
                pos = end
            m.free()
        
        return results

    def sub(self, repl, subject, count=0):
        """Substitute occurrences of the pattern with repl."""
        result = ""
        pos = 0
        n = 0
        while pos <= len(subject):
            if count > 0 and n >= count:
                break
            m = self.search(subject, pos)
            if not m:
                break
            
            start, end = m.span()
            result = result + subject[pos:start]
            result = result + repl
            
            if start == end:
                pos = pos + 1
            else:
                pos = end
            n = n + 1
            m.free()
            
        result = result + subject[pos:]
        return result

    def split(self, subject, maxsplit=0):
        """Split the string by the occurrences of the pattern."""
        result = []
        pos = 0
        n = 0
        while pos <= len(subject):
            if maxsplit > 0 and n >= maxsplit:
                break
            m = self.search(subject, pos)
            if not m:
                break
            start, end = m.span()
            result.append(subject[pos:start])
            
            if start == end:
                pos = pos + 1
            else:
                pos = end
            n = n + 1
            m.free()
            
        result.append(subject[pos:])
        return result

    def free(self):
        if self._code:
            _pcre2_code_free(self._code)
            self._code = None

_cache = {}

def compile(pattern, flags=0):
    cache_key = format_str("{pattern}:{flags}")
    if cache_key in _cache:
        return _cache[cache_key]
    
    p = Pattern(pattern, flags)
    _cache[cache_key] = p
    return p

def search(pattern, string, flags=0):
    return compile(pattern, flags).search(string)

def match(pattern, string, flags=0):
    return compile(pattern, flags).match(string)

def findall(pattern, string, flags=0):
    return compile(pattern, flags).findall(string)

def sub(pattern, repl, string, count=0, flags=0):
    return compile(pattern, flags).sub(repl, string, count)

def split(pattern, string, maxsplit=0, flags=0):
    return compile(pattern, flags).split(string, maxsplit)

