"""
pcre2/__init__.py — Production-ready PCRE2 regular expression wrapper for Swalang.

New in this version:
  - finditer()       — iterator of Match objects (essential for template parsing)
  - named groups     — match.group('name'), match.groupdict()
  - callable sub()   — repl can be a function(match) → str
  - subn()           — like sub() but returns (new_string, count)
  - fullmatch()      — anchors at both ends
  - escape()         — escape a string for use as a literal pattern
  - Pattern.flags    — introspect compiled flags
  - Match.re         — back-reference to the Pattern
  - Match.__bool__   — truthy test
  - VERBOSE / X flag — whitespace-insensitive patterns
  - UNICODE / U flag
  - EXTENDED flag
"""

import ffi
import sys

# ==============================================================================
#  Library Loading
# ==============================================================================

def _load_library():
    platform = sys.platform
    candidates = []
    if platform == 'linux':
        candidates = [
            "bin/x86_64-linux/pcre2/libpcre2-8.so",
            "libpcre2-8.so",
            "libpcre2-8.so.0",
        ]
    elif platform == 'windows':
        candidates = [
            "bin/x86_64-windows-gnu/pcre2/pcre2-8.dll",
            "pcre2-8.dll",
            "libpcre2-8.dll",
        ]
    elif platform == 'darwin':
        candidates = ["libpcre2-8.dylib"]

    for name in candidates:
        try:
            return ffi.CDLL(name)
        except ffi.FFIError:
            pass
    raise ffi.FFIError("Could not load pcre2-8 shared library")

_lib = _load_library()

# ==============================================================================
#  Constants
# ==============================================================================

# Compile flags
PCRE2_CASELESS   = 0x00000008
PCRE2_MULTILINE  = 0x00000400
PCRE2_DOTALL     = 0x00000020
PCRE2_EXTENDED   = 0x00000080
PCRE2_ANCHORED   = 0x80000000
PCRE2_DOLLAR_ENDONLY = 0x00000010
PCRE2_UNGREEDY   = 0x00000200
PCRE2_UTF        = 0x00080000
PCRE2_UCP        = 0x00020000

# Python-style aliases
IGNORECASE = PCRE2_CASELESS
I          = PCRE2_CASELESS
MULTILINE  = PCRE2_MULTILINE
M          = PCRE2_MULTILINE
DOTALL     = PCRE2_DOTALL
S          = PCRE2_DOTALL
VERBOSE    = PCRE2_EXTENDED
X          = PCRE2_EXTENDED
UNICODE    = PCRE2_UTF
U          = PCRE2_UTF

# Match flags (passed to pcre2_match, not compile)
PCRE2_NOTEMPTY         = 0x00000004
PCRE2_NOTEMPTY_ATSTART = 0x00000008

# Internal sentinel for unset groups
_PCRE2_UNSET = 18446744073709551615   # (size_t)~0 on 64-bit

# ==============================================================================
#  Error class
# ==============================================================================

class error(Exception):
    """Exception raised for PCRE2 errors."""
    def __init__(self, code, pattern="", offset=0):
        self.code    = code
        self.pattern = pattern
        self.offset  = offset

        err_buf = ffi.malloc(256)
        try:
            _pcre2_get_error_message(code, err_buf, 256)
            self.message = ffi.string_at(err_buf)
        except Exception:
            self.message = format_str("PCRE2 error {code}")
        finally:
            ffi.free(err_buf)

        msg = self.message
        if pattern and offset:
            msg = format_str("{self.message} at offset {offset} in pattern '{pattern}'")
        elif pattern:
            msg = format_str("{self.message} in pattern '{pattern}'")
        super().__init__(msg)

# ==============================================================================
#  C Function Bindings
# ==============================================================================

_pcre2_compile = _lib.pcre2_compile_8([ffi.c_char_p, ffi.c_uint64, ffi.c_uint32, ffi.POINTER(ffi.c_int32), ffi.POINTER(ffi.c_uint64), ffi.c_void_p], ffi.c_void_p)

_pcre2_code_free = _lib.pcre2_code_free_8([ffi.c_void_p], None)

_pcre2_match_data_create_from_pattern = _lib.pcre2_match_data_create_from_pattern_8([ffi.c_void_p, ffi.c_void_p], ffi.c_void_p)

_pcre2_match_data_free = _lib.pcre2_match_data_free_8([ffi.c_void_p], None)

_pcre2_match = _lib.pcre2_match_8([ffi.c_void_p, ffi.c_char_p, ffi.c_uint64, ffi.c_uint64, ffi.c_uint32, ffi.c_void_p, ffi.c_void_p], ffi.c_int32)

_pcre2_get_ovector_pointer = _lib.pcre2_get_ovector_pointer_8([ffi.c_void_p], ffi.c_void_p)

_pcre2_get_ovector_count = _lib.pcre2_get_ovector_count_8([ffi.c_void_p], ffi.c_uint32)

_pcre2_get_error_message = _lib.pcre2_get_error_message_8([ffi.c_int32, ffi.c_void_p, ffi.c_uint64], ffi.c_int32)

# Named capture group lookup
_pcre2_substring_number_from_name = _lib.pcre2_substring_number_from_name_8([ffi.c_void_p, ffi.c_char_p], ffi.c_int32)

# Pattern info
_pcre2_pattern_info = _lib.pcre2_pattern_info_8([ffi.c_void_p, ffi.c_uint32, ffi.c_void_p], ffi.c_int32)

# PCRE2_INFO constants
_PCRE2_INFO_CAPTURECOUNT    = 4
_PCRE2_INFO_NAMECOUNT       = 16
_PCRE2_INFO_NAMEENTRYSIZE   = 17
_PCRE2_INFO_NAMETABLE       = 18

# ==============================================================================
#  Match
# ==============================================================================

class Match:
    """
    Result of a successful regex match.

    Attributes:
      re      — the Pattern that produced this match
      string  — the subject string
      pos     — the start position passed to the matching function
      endpos  — the end position (len(string) unless restricted)
      lastindex — index of last matched capturing group, or None
    """

    def __init__(self, pattern_obj, subject, match_data_ptr, pos=0, endpos=-1):
        self._pattern    = pattern_obj
        self.re          = pattern_obj
        self.string      = subject
        self._subj_bytes = subject.encode('utf-8')
        self.pos         = pos
        self.endpos      = endpos if endpos >= 0 else len(subject)
        self._match_data = match_data_ptr
        self._ovector    = _pcre2_get_ovector_pointer(self._match_data)
        self._ov_count   = _pcre2_get_ovector_count(self._match_data)

        # Compute lastindex
        self.lastindex = None
        for i in range(self._ov_count - 1, 0, -1):
            s, e = self._get_offsets(i)
            if s != -1:
                self.lastindex = i
                break

    # ---- offset access ------------------------------------------------------

    def _get_offsets(self, index):
        if index >= self._ov_count:
            raise IndexError(format_str("No such group: {index}"))
        off   = index * 2 * 8
        start = ffi.read_memory_with_offset(self._ovector, off,     ffi.c_uint64)
        end   = ffi.read_memory_with_offset(self._ovector, off + 8, ffi.c_uint64)
        if start == _PCRE2_UNSET:
            return (-1, -1)
        return (int(start), int(end))

    # ---- group access -------------------------------------------------------

    def group(self, *indices):
        """
        Return one or more subgroups.
        group(0) or group() returns the entire match.
        group('name') looks up a named capturing group.
        Returns None for groups that did not participate.
        """
        if len(indices) == 0:
            indices = [0]

        results = []
        for idx in indices:
            if isinstance(idx, str):
                idx = self._pattern._group_index(idx)
            s, e = self._get_offsets(idx)
            if s == -1:
                results.append(None)
            else:
                b_slice = self._subj_bytes[s:e]
                results.append(b_slice.decode('utf-8'))

        return results[0] if len(results) == 1 else tuple(results)

    def groups(self, default=None):
        """Return a tuple of all capturing group strings."""
        out = []
        for i in range(1, self._ov_count):
            s, e = self._get_offsets(i)
            if s == -1:
                out.append(default)
            else:
                b_slice = self._subj_bytes[s:e]
                out.append(b_slice.decode('utf-8'))
        return tuple(out)

    def groupdict(self, default=None):
        """Return a dict mapping named group names to matched strings."""
        result = {}
        names  = self._pattern._named_groups()
        for name in names:
            idx = names[name]
            s, e = self._get_offsets(idx)
            if s == -1:
                result[name] = default
            else:
                b_slice = self._subj_bytes[s:e]
                result[name] = b_slice.decode('utf-8')
        return result

    def start(self, group=0):
        if isinstance(group, str):
            group = self._pattern._group_index(group)
        s, _ = self._get_offsets(group)
        return s

    def end(self, group=0):
        if isinstance(group, str):
            group = self._pattern._group_index(group)
        _, e = self._get_offsets(group)
        return e

    def span(self, group=0):
        if isinstance(group, str):
            group = self._pattern._group_index(group)
        return self._get_offsets(group)

    def expand(self, template):
        """
        Return the string with backreference substitutions applied.
        Supports \\1, \\g<1>, \\g<name>.
        """
        return _expand_template(template, self)

    # ---- lifecycle ----------------------------------------------------------

    def free(self):
        if self._match_data:
            _pcre2_match_data_free(self._match_data)
            self._match_data = None

    # ---- dunder -------------------------------------------------------------

    def __bool__(self):
        return True

    def __repr__(self):
        s, e = self.span()
        return format_str("<pcre2.Match span=({s},{e}) match='{self.group(0)}'>")

# ==============================================================================
#  Template expansion helper
# ==============================================================================

def _expand_template(template, match):
    """Expand a replacement template string against a Match object."""
    result = ""
    i = 0
    while i < len(template):
        if template[i] == '\\' and i + 1 < len(template):
            next_c = template[i + 1]
            if next_c.isdigit():
                idx = int(next_c)
                val = match.group(idx)
                result = result + (val if val is not None else "")
                i = i + 2
                continue
            if next_c == 'g' and i + 2 < len(template) and template[i + 2] == '<':
                end = template.find('>', i + 3)
                if end != -1:
                    name_or_idx = template[i + 3:end]
                    if name_or_idx.isdigit():
                        val = match.group(int(name_or_idx))
                    else:
                        val = match.group(name_or_idx)
                    result = result + (val if val is not None else "")
                    i = end + 1
                    continue
        result = result + template[i]
        i = i + 1
    return result

# ==============================================================================
#  Pattern
# ==============================================================================

class Pattern:
    """
    A compiled PCRE2 regular expression.

    Usage:
        p = Pattern(r'(?P<year>\d{4})-(?P<month>\d{2})-(?P<day>\d{2})')
        m = p.search('Today is 2025-07-01.')
        print(m.groupdict())   # {'year': '2025', 'month': '07', 'day': '01'}
    """

    def __init__(self, pattern, flags=0):
        self.pattern = pattern
        self.flags   = flags

        err_code_buf   = ffi.malloc(4)   # int32
        err_offset_buf = ffi.malloc(8)   # uint64

        try:
            self._code = _pcre2_compile(pattern.encode('utf-8'), len(pattern.encode('utf-8')), flags, err_code_buf, err_offset_buf, None)

            if not self._code or not getattr(self._code, "Address", None) or self._code.Address == 0:
                code   = ffi.read_memory(err_code_buf,   ffi.c_int32)
                offset = ffi.read_memory(err_offset_buf, ffi.c_uint64)
                raise error(code, pattern, int(offset))
        except Exception:
            pass
        finally:
            ffi.free(err_code_buf)
            ffi.free(err_offset_buf)

        self._named_groups_cache = None

    # ---- named group helpers ------------------------------------------------

    def _group_index(self, name):
        """Translate a named group name to its numeric index."""
        idx = _pcre2_substring_number_from_name(self._code, name.encode('utf-8'))
        if idx < 0:
            raise IndexError(format_str("No such group: '{name}'"))
        return idx

    def _named_groups(self):
        """Return a dict of name → group_index for all named groups."""
        if self._named_groups_cache is not None:
            return self._named_groups_cache

        result = {}

        count_buf     = ffi.malloc(4)
        entry_size_buf= ffi.malloc(4)
        table_buf     = ffi.malloc(8)

        try:
            _pcre2_pattern_info(self._code, _PCRE2_INFO_NAMECOUNT,     count_buf)
            _pcre2_pattern_info(self._code, _PCRE2_INFO_NAMEENTRYSIZE, entry_size_buf)
            _pcre2_pattern_info(self._code, _PCRE2_INFO_NAMETABLE,     table_buf)

            count      = ffi.read_memory(count_buf,      ffi.c_uint32)
            entry_size = ffi.read_memory(entry_size_buf, ffi.c_uint32)
            table_ptr  = ffi.read_memory(table_buf,      ffi.c_void_p)

            for i in range(count):
                entry_off = i * entry_size
                b0 = ffi.read_memory_with_offset(table_ptr, entry_off,     ffi.c_uint8)
                b1 = ffi.read_memory_with_offset(table_ptr, entry_off + 1, ffi.c_uint8)
                group_num = (b0 << 8) | b1
                name = ffi.string_at(table_ptr, -1, entry_off + 2)
                result[name] = group_num
        except Exception:
            pass
        finally:
            ffi.free(count_buf)
            ffi.free(entry_size_buf)
            ffi.free(table_buf)

        self._named_groups_cache = result
        return result

    def groupindex(self):
        """Return a dict mapping group names to group numbers."""
        return self._named_groups()

    @property
    def groups(self):
        """Number of capturing groups in the pattern."""
        buf = ffi.malloc(4)
        try:
            _pcre2_pattern_info(self._code, _PCRE2_INFO_CAPTURECOUNT, buf)
            return ffi.read_memory(buf, ffi.c_uint32)
        except Exception:
            pass
        finally:
            ffi.free(buf)

    # ---- core matching ------------------------------------------------------

    def _do_match(self, subject, pos, endpos, options):
        """
        Run pcre2_match and return a Match or None.
        subject must be a str.
        """
        if not isinstance(subject, str):
            raise TypeError("subject must be a str")

        subj_bytes = subject.encode('utf-8')
        length     = len(subj_bytes)

        if endpos < 0:
            endpos = length
        if endpos > length:
            endpos = length
        if pos > endpos:
            return None

        match_data = _pcre2_match_data_create_from_pattern(self._code, None)
        if not match_data or not getattr(match_data, "Address", None) or match_data.Address == 0:
            raise MemoryError("Could not create PCRE2 match_data")

        rc = _pcre2_match(self._code, subj_bytes, endpos, pos, options, match_data, None)

        if rc < 0:
            _pcre2_match_data_free(match_data)
            return None

        return Match(self, subject, match_data, pos, endpos)

    # ---- public API ---------------------------------------------------------

    def search(self, string, pos=0, endpos=-1):
        """
        Scan through string looking for a match.
        Returns a Match or None.
        """
        return self._do_match(string, pos, endpos, 0)

    def match(self, string, pos=0, endpos=-1):
        """
        Match at the beginning of the string (or at pos).
        Returns a Match or None.
        """
        return self._do_match(string, pos, endpos, PCRE2_ANCHORED)

    def fullmatch(self, string, pos=0, endpos=-1):
        """
        Match the whole string (or the pos..endpos slice).
        Returns a Match or None.
        """
        m = self._do_match(string, pos, endpos, PCRE2_ANCHORED)
        if m is None:
            return None
        _, end = m.span()
        real_end = endpos if endpos >= 0 else len(string.encode('utf-8'))
        if end != real_end:
            m.free()
            return None
        return m

    def findall(self, string, pos=0, endpos=-1):
        """
        Return a list of all non-overlapping matches.
        If there are groups, returns a list of group tuples.
        If there is one group, returns a list of strings.
        If no groups, returns a list of full-match strings.
        """
        results = []
        subj_bytes = string.encode('utf-8')
        real_end = endpos if endpos >= 0 else len(subj_bytes)
        cur     = pos

        while cur <= real_end:
            m = self._do_match(string, cur, real_end, 0)
            if m is None:
                break

            num_groups = self.groups
            if num_groups > 1:
                results.append(m.groups())
            elif num_groups == 1:
                g = m.group(1)
                results.append(g if g is not None else "")
            else:
                results.append(m.group(0))

            s, e = m.span()
            m.free()
            if s == e:
                cur = cur + 1
            else:
                cur = e

        return results

    def finditer(self, string, pos=0, endpos=-1):
        """
        Return an iterator yielding Match objects for all non-overlapping matches.
        The caller is responsible for calling match.free() on each result.
        """
        subj_bytes = string.encode('utf-8')
        real_end = endpos if endpos >= 0 else len(subj_bytes)
        cur      = pos

        while cur <= real_end:
            m = self._do_match(string, cur, real_end, 0)
            if m is None:
                break
            
            s, e = m.span()

            yield m

            if s == e:
                cur = cur + 1
            else:
                cur = e

    def sub(self, repl, string, count=0):
        """
        Replace occurrences of the pattern.
        repl may be a str (backreferences: \\1, \\g<name>) or a callable(match)→str.
        """
        result, _ = self.subn(repl, string, count)
        return result

    def subn(self, repl, string, count=0):
        """
        Like sub() but returns (new_string, number_of_substitutions).
        """
        parts   = []
        n_subs  = 0
        is_func = callable(repl)

        subj_bytes = string.encode('utf-8')
        cur     = 0

        while cur <= len(subj_bytes):
            if count > 0 and n_subs >= count:
                break
            m = self._do_match(string, cur, -1, 0)
            if m is None:
                break

            s, e = m.span()
            parts.append(subj_bytes[cur:s].decode('utf-8'))

            if is_func:
                replacement = repl(m)
            else:
                replacement = _expand_template(repl, m)

            parts.append(replacement)
            m.free()
            n_subs = n_subs + 1

            if s == e:
                if cur < len(subj_bytes):
                    parts.append(subj_bytes[cur:cur+1].decode('utf-8'))
                cur = cur + 1
            else:
                cur = e

        parts.append(subj_bytes[cur:].decode('utf-8'))
        return ("".join(parts), n_subs)

    def split(self, string, maxsplit=0):
        """
        Split string by occurrences of the pattern.
        Capturing groups are included in the result list.
        """
        result  = []
        n_split = 0

        subj_bytes = string.encode('utf-8')
        cur     = 0

        while cur <= len(subj_bytes):
            if maxsplit > 0 and n_split >= maxsplit:
                break
            m = self._do_match(string, cur, -1, 0)
            if m is None:
                break

            s, e = m.span()
            result.append(subj_bytes[cur:s].decode('utf-8'))

            # Include capturing groups in split result (Python-compatible)
            for i in range(1, self.groups + 1):
                result.append(m.group(i))

            m.free()
            n_split = n_split + 1
            if s == e:
                cur = cur + 1
            else:
                cur = e

        result.append(subj_bytes[cur:].decode('utf-8'))
        return result

    # ---- lifecycle ----------------------------------------------------------

    def free(self):
        if self._code:
            _pcre2_code_free(self._code)
            self._code = None

    def __repr__(self):
        return format_str("<pcre2.Pattern '{self.pattern}' flags={self.flags}>")

# ==============================================================================
#  Module-level pattern cache + convenience functions
# ==============================================================================

_cache = {}
_CACHE_MAX = 256


def compile(pattern, flags=0):
    """Compile a pattern string and cache the result."""
    key = format_str("{flags}:{pattern}")
    if key in _cache:
        return _cache[key]
    p = Pattern(pattern, flags)
    if len(_cache) >= _CACHE_MAX:
        for oldest in _cache:
            del _cache[oldest]
            break
    _cache[key] = p
    return p


def search(pattern, string, flags=0):
    """Scan string for the first location where pattern matches."""
    return compile(pattern, flags).search(string)


def match(pattern, string, flags=0):
    """Match at the beginning of string."""
    return compile(pattern, flags).match(string)


def fullmatch(pattern, string, flags=0):
    """Match the whole string against pattern."""
    return compile(pattern, flags).fullmatch(string)


def findall(pattern, string, flags=0, pos=0):
    """Return a list of all non-overlapping matches."""
    return compile(pattern, flags).findall(string, pos)


def finditer(pattern, string, flags=0, pos=0):
    """Return an iterator of Match objects for all non-overlapping matches."""
    return compile(pattern, flags).finditer(string, pos)


def sub(pattern, repl, string, count=0, flags=0):
    """
    Replace occurrences of pattern in string with repl.
    repl may be a str or a callable(match)→str.
    """
    return compile(pattern, flags).sub(repl, string, count)


def subn(pattern, repl, string, count=0, flags=0):
    """Like sub() but returns (new_string, count)."""
    return compile(pattern, flags).subn(repl, string, count)


def split(pattern, string, maxsplit=0, flags=0):
    """Split string by pattern."""
    return compile(pattern, flags).split(string, maxsplit)


def escape(pattern):
    """
    Return `pattern` with all non-alphanumeric characters backslash-escaped.
    Use this to match a literal string inside a regex.
    """
    _SPECIAL = set(list(r'\.^$*+?{}[]|()')+ ['!', '"', '#', '%', '&', "'", ',', '-', '/', ':', ';', '<', '=', '>', '@', '`', '~'])
    result = ""
    for ch in pattern:
        if not ch.isalpha() and not ch.isdigit() and ch != '_':
            result = result + '\\' + ch
        else:
            result = result + ch
    return result


def purge():
    """Clear the pattern cache."""
    global _cache
    _cache = {}