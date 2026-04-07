# pylearn/stdlib/mongoose/__init__.py

import ffi
import sys

def _load_library_with_fallbacks(base_name):
    platform = sys.platform
    candidates = []
    if platform == 'linux':
        candidates.append(format_str("bin/x86_64-linux/mongoose/lib{base_name}.so"))
        candidates.append(format_str("lib{base_name}.so"))
    elif platform == 'windows':
        candidates.append(format_str("bin/x86_64-windows-gnu/mongoose/{base_name}.dll"))
        candidates.append(format_str("{base_name}.dll"))

    for name in candidates:
        try:
            return ffi.CDLL(name)
        except ffi.FFIError:
            continue
    raise ffi.FFIError(format_str("Could not load {base_name}"))

_lib = None
try:
    _lib = _load_library_with_fallbacks("mongoose")
except ffi.FFIError:
    pass

if _lib:
    _mg_version = _lib.mg_version([], ffi.c_char_p)

def get_version():
    if not _lib: return ""
    return _mg_version()
