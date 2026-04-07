# pylearn/stdlib/pcre2.py

"""
A Pylearn wrapper for the PCRE2 (Perl Compatible Regular Expressions) library,
built using the Pylearn FFI. This module provides a powerful and Pythonic
interface for advanced regular expression operations.
"""

import ffi
import sys

# ==============================================================================
#  Platform-Aware Library Loading
# ==============================================================================

def _load_library_with_fallbacks(base_name):
    """Tries to load libpcre2-8 using common platform-specific names."""
    platform = sys.platform
    
    candidates = []
    if platform == 'windows':
        # PCRE2 for Windows often has this -0 suffix
        candidates = ['libpcre2-8-0.dll', 'libpcre2-8.dll']
    elif platform == 'darwin':
        candidates = ['libpcre2-8.0.dylib', 'libpcre2-8.dylib']
    else: # Linux and other Unix-like
        candidates = ['libpcre2-8.so.0', 'libpcre2-8.so']

    last_error = None
    for name in candidates:
        try:
            return ffi.CDLL(name)
        except ffi.FFIError as e:
            last_error = e
    
    raise last_error

_lib = None
PCRE2_AVAILABLE = False
try:
    _lib = _load_library_with_fallbacks("pcre2-8")
    PCRE2_AVAILABLE = True
except ffi.FFIError as e:
    print(format_str("Warning: Failed to load libpcre2-8: {e}"))
    print("PCRE2 regular expression functionality will not be available.")

# ==============================================================================
#  Exception and Constants
# ==============================================================================

class PCREError(Exception):
    """Raised for PCRE2-related errors during compilation or matching."""
    def __init__(self, code, pattern=""):
        self.code = code
        self.pattern = pattern
        
        # Allocate a buffer for the error message
        error_buffer = ffi.malloc(256)
        try:
            # Get the error message from the C library
            _pcre2_get_error_message(code, error_buffer, 256)
            self.message = ffi.string_at(error_buffer)
        finally:
            ffi.free(error_buffer)

        full_message = self.message
        if pattern:
            full_message = format_str("{self.message} in pattern: '{self.pattern}'")
        super().__init__(full_message)

# PCRE2 Options
PCRE2_ANCHORED = 0x80000000
PCRE2_CASELESS = 0x00000008 # i
PCRE2_MULTILINE = 0x00000400 # m
PCRE2_DOTALL = 0x00000020 # s

# ==============================================================================
#  Define C Function Signatures
# ==============================================================================

if PCRE2_AVAILABLE:
    _pcre2_compile = _lib.pcre2_compile_8(
        [ffi.c_void_p, ffi.c_uint64, ffi.c_uint32, ffi.POINTER(ffi.c_int32), ffi.POINTER(ffi.c_uint64), ffi.c_void_p],
        ffi.c_void_p
    )
    _pcre2_code_free = _lib.pcre2_code_free_8([ffi.c_void_p], None)
    
    _pcre2_match_data_create_from_pattern = _lib.pcre2_match_data_create_from_pattern_8(
        [ffi.c_void_p, ffi.c_void_p],
        ffi.c_void_p
    )
    _pcre2_match_data_free = _lib.pcre2_match_data_free_8([ffi.c_void_p], None)
    
    _pcre2_match = _lib.pcre2_match_8(
        [ffi.c_void_p, ffi.c_void_p, ffi.c_uint64, ffi.c_uint64, ffi.c_uint32, ffi.c_void_p, ffi.c_void_p],
        ffi.c_int32
    )
    
    _pcre2_get_ovector_pointer = _lib.pcre2_get_ovector_pointer_8([ffi.c_void_p], ffi.c_void_p)
    _pcre2_get_ovector_count = _lib.pcre2_get_ovector_count_8([ffi.c_void_p], ffi.c_uint32)
    
    _pcre2_get_error_message = _lib.pcre2_get_error_message_8(
        [ffi.c_int32, ffi.c_void_p, ffi.c_uint64],
        ffi.c_int32
    )

# ==============================================================================
#  High-Level Pylearn API
# ==============================================================================

_cache = {}
_MAX_CACHE_SIZE = 512

class Match:
    """Represents a successful match."""
    def __init__(self, pattern_obj, subject, match_data_ptr):
        self._pattern = pattern_obj
        self._subject = subject # Keep the original subject bytes
        self._match_data = match_data_ptr
        self._ovector = _pcre2_get_ovector_pointer(self._match_data)
        self._ovector_count = _pcre2_get_ovector_count(self._match_data)

    def group(self, index=0):
        """Returns the captured group by index."""
        if index >= self._ovector_count:
            raise IndexError("no such group")

        offset = index * 2 * ffi.c_uint64.Size()
        start = ffi.read_memory_with_offset(self._ovector, offset, ffi.c_uint64)
        end = ffi.read_memory_with_offset(self._ovector, offset + ffi.c_uint64.Size(), ffi.c_uint64)
        
        return self._subject[start:end].decode('utf-8')

    def groups(self):
        """Returns a tuple of all captured groups."""
        grps = []
        i = 1
        while i < self._ovector_count:
            grps.append(self.group(i))
            i = i + 1
        return tuple(grps)

    def free(self):
        """Frees the C memory associated with the match object."""
        if self._match_data:
            _pcre2_match_data_free(self._match_data)
            self._match_data = None

    def __repr__(self):
        return format_str("<pcre2.Match span=({self.group(0)}), match='{self.group(0)}'>")

class Pattern:
    """Represents a compiled regular expression."""
    def __init__(self, pattern, flags=0):
        self.pattern = pattern
        self.flags = flags
        
        error_code_ptr = ffi.malloc(ffi.c_int32.Size())
        error_offset_ptr = ffi.malloc(ffi.c_uint64.Size())
        
        try:
            pattern_bytes = pattern.encode('utf-8')
            self._code = _pcre2_compile(
                pattern_bytes,
                len(pattern_bytes),
                flags,
                error_code_ptr,
                error_offset_ptr,
                None # compile_context
            )
            
            if self._code.Address == 0:
                error_code = ffi.read_memory(error_code_ptr, ffi.c_int32)
                raise PCREError(error_code, pattern)
                
        finally:
            ffi.free(error_code_ptr)
            ffi.free(error_offset_ptr)

    def _do_match(self, subject, start_offset, options):
        if isinstance(subject, str):
            subject_bytes = subject.encode('utf-8')
        elif isinstance(subject, bytes):
            subject_bytes = subject
        else:
            raise TypeError("subject must be a string or bytes")

        match_data = _pcre2_match_data_create_from_pattern(self._code, None)
        if match_data.Address == 0:
            raise MemoryError("could not create match_data block")
        
        rc = _pcre2_match(
            self._code,
            subject_bytes,
            len(subject_bytes),
            start_offset,
            options,
            match_data,
            None # match_context
        )

        if rc < 0: # An error or no match
            _pcre2_match_data_free(match_data)
            return None
        
        return Match(self, subject_bytes, match_data)

    def search(self, subject, pos=0):
        """Scan through string looking for a match."""
        return self._do_match(subject, pos, 0)

    def match(self, subject, pos=0):
        """Match a pattern at the beginning of a string."""
        return self._do_match(subject, pos, PCRE2_ANCHORED)

    def findall(self, subject, pos=0):
        """Find all non-overlapping matches in the string."""
        results = []
        while pos <= len(subject):
            m = self.search(subject, pos)
            if m is None:
                break
            
            num_groups = m._ovector_count - 1
            if num_groups > 0:
                results.append(m.groups() if num_groups > 1 else m.group(1))
            else:
                results.append(m.group(0))

            # Advance past the current match
            # Handle zero-length matches to avoid infinite loops
            match_end = ffi.read_memory_with_offset(m._ovector, ffi.c_uint64.Size(), ffi.c_uint64)
            if match_end == pos:
                pos = pos + 1
            else:
                pos = match_end
            
            m.free()
        
        return results

    def free(self):
        """Frees the C memory for the compiled pattern."""
        if self._code:
            _pcre2_code_free(self._code)
            self._code = None

def compile(pattern, flags=0):
    """Compile a regular expression pattern, returning a Pattern object."""
    if not PCRE2_AVAILABLE:
        raise PCREError(-1, "libpcre2-8 library not available")
    
    cache_key = (pattern, flags)
    if cache_key in _cache:
        return _cache[cache_key]

    if len(_cache) > _MAX_CACHE_SIZE:
        _cache.clear()
        
    p = Pattern(pattern, flags)
    _cache[cache_key] = p
    return p

def search(pattern, string, flags=0):
    """Scan string for a match, returning a Match object or None."""
    p = compile(pattern, flags)
    return p.search(string)

def match(pattern, string, flags=0):
    """Match pattern at start of string, returning a Match object or None."""
    p = compile(pattern, flags)
    return p.match(string)

def findall(pattern, string, flags=0):
    """Find all non-overlapping matches, returning a list of strings/tuples."""
    p = compile(pattern, flags)
    return p.findall(string)