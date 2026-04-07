# pylearn/stdlib/regex.py

"""
A basic implementation of Python's 're' module for Pylearn, using the
native Go backend.

This module provides regular expression matching operations similar to those
found in Python. It supports compiling patterns for efficiency and provides
top-level functions for common operations.
"""

# Import the native Go implementation, which is registered under the same name "re".
# To avoid a name clash, we import it as '_re_native'.
from pylearn_importlib import load_module_from_path
import re as _re_native
# _re_native = re #load_module_from_path("re")

# --- Public API ---

# Module-level cache for compiled patterns
_cache = {}
_MAX_CACHE_SIZE = 512

class error(Exception):
    """
    Exception raised for errors during regex compilation or matching.
    This is an alias for the native re.error.
    """
    def __init__(self, msg, pattern=None, pos=None):
        self.msg = msg
        self.pattern = pattern
        self.pos = pos
        # Call the base Exception initializer
        super().__init__(msg)

# --- Flags (for API compatibility; not yet implemented in backend) ---
IGNORECASE = 1  # i
MULTILINE = 2   # m
DOTALL = 4      # s

# --- Public Functions ---

def compile(pattern, flags=0):
    """
    Compile a regular expression pattern, returning a pattern object.
    Compiled patterns are cached for performance.
    """
    # Note: flags are currently ignored by the backend.
    cache_key = (pattern, flags)
    
    cached_pattern = _cache.get(cache_key)
    if cached_pattern is not None:
        return cached_pattern

    # If cache is full, clear it. A more sophisticated LRU cache could be used later.
    if len(_cache) > _MAX_CACHE_SIZE:
        _cache.clear()

    # Call the native Go implementation to compile the pattern.
    p = _re_native.compile(pattern)
    
    # Check if the native call returned an error (our native funcs return Error objects)
    if isinstance(p, error): # Assuming native errors are compatible
        # Re-raise it as a Pylearn re.error
        raise p

    _cache[cache_key] = p
    return p

def search(pattern, string, flags=0):
    """
    Scan through a string looking for the first location where the regular
    expression pattern produces a match, and return a corresponding match object.
    Returns None if no position in the string matches the pattern.
    """
    p = compile(pattern, flags)
    return p.search(string) # Delegate to the compiled pattern's search method

def match(pattern, string, flags=0):
    """
    If zero or more characters at the beginning of string match the regular
    expression pattern, return a corresponding match object.
    Returns None if the string does not match the pattern.
    """
    p = compile(pattern, flags)
    return p.match(string) # Delegate to the compiled pattern's match method

def findall(pattern, string, flags=0):
    """
    Return a list of all non-overlapping matches in the string.
    
    If one or more capturing groups are present in the pattern, returns a
    list of tuples. If the pattern has only one group, returns a list of
    strings matching that group.
    """
    p = compile(pattern, flags)
    return p.findall(string) # Delegate to the compiled pattern's findall method

def escape(pattern):
    """
    Escape special characters in a string for use in a regular expression.
    """
    return _re_native.escape(pattern)

def testRE():
    print("regex works fine")
