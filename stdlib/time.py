"""
time.py — Production-ready time module for Swalang.

Covers:
  time()          — wall-clock seconds since epoch  (float)
  monotonic()     — monotonically increasing clock  (float, no epoch meaning)
  perf_counter()  — highest-resolution timer for benchmarking (float)
  sleep(s)        — suspend execution for s seconds (float ok)
  gmtime([secs])  — struct_time in UTC
  localtime([s])  — struct_time in local time
  mktime(st)      — struct_time → epoch float (local time)
  strftime(fmt, t)— format a struct_time as a string
  timezone        — offset of local (non-DST) zone from UTC, in seconds west
  daylight        — non-zero if a DST zone is defined
  tzname          — tuple of (std_zone_name, dst_zone_name)

Platform notes:
  Windows  — time()  via QueryPerformanceCounter (QPC) or GetSystemTimeAsFileTime
             perf_counter() via QPC
             monotonic()    via QPC
  Linux    — time()  via clock_gettime(CLOCK_REALTIME)
             perf_counter() via clock_gettime(CLOCK_MONOTONIC_RAW)
             monotonic()    via clock_gettime(CLOCK_MONOTONIC)
  macOS    — same as Linux (CLOCK_MONOTONIC_RAW may not exist; fallback to
             CLOCK_MONOTONIC)
"""

import ffi
import sys

# ==============================================================================
#  Internal: platform detection
# ==============================================================================

_WINDOWS = sys.platform == 'windows'
_DARWIN  = sys.platform == 'darwin'
_LINUX   = sys.platform == 'linux'

# ==============================================================================
#  Internal: libc / kernel32
# ==============================================================================

def _load_libc():
    if _WINDOWS:
        return ffi.CDLL("msvcrt.dll")
    elif _DARWIN:
        return ffi.CDLL("libc.dylib")
    else:
        for name in ["libc.so.6", "libc.so", "c"]:
            try:
                return ffi.CDLL(name)
            except ffi.FFIError:
                pass
        raise ffi.FFIError("Could not load libc")

_libc = _load_libc()

if _WINDOWS:
    _kernel32 = ffi.CDLL("kernel32.dll")

# ==============================================================================
#  Internal: clock constants (POSIX)
# ==============================================================================

# These match the values in <time.h> on Linux and macOS.
_CLOCK_REALTIME         = 0
_CLOCK_MONOTONIC        = 1
_CLOCK_MONOTONIC_RAW    = 4   # Linux only; not present on macOS
_CLOCK_PROCESS_CPUTIME  = 2
_CLOCK_THREAD_CPUTIME   = 3

# ==============================================================================
#  Internal: C function bindings
# ==============================================================================

if _WINDOWS:
    _QueryPerformanceCounter   = _kernel32.QueryPerformanceCounter(  [ffi.c_void_p], ffi.c_int32)
    _QueryPerformanceFrequency = _kernel32.QueryPerformanceFrequency([ffi.c_void_p], ffi.c_int32)
    _Sleep                     = _kernel32.Sleep(                    [ffi.c_uint32], None)
    _GetSystemTimeAsFileTime   = _kernel32.GetSystemTimeAsFileTime(  [ffi.c_void_p], None)

    # C runtime helpers for calendar time
    _gmtime_s    = _libc.gmtime_s(   [ffi.c_void_p, ffi.c_void_p], ffi.c_int32)
    _localtime_s = _libc.localtime_s([ffi.c_void_p, ffi.c_void_p], ffi.c_int32)
    _mktime      = _libc.mktime(     [ffi.c_void_p], ffi.c_int64)
    _strftime    = _libc.strftime(   [ffi.c_char_p, ffi.c_uint64, ffi.c_char_p, ffi.c_void_p], ffi.c_uint64)

    try:
        _tzname_ptr = _libc.get_address("_tzname")
        _timezone_ptr = _libc.get_address("_timezone")
        _daylight_ptr = _libc.get_address("_daylight")
    except ffi.FFIError:
        _tzname_ptr = None
        _timezone_ptr = None
        _daylight_ptr = None

    # QPC frequency (sampled once at startup — it is constant)
    _qpf_buf = ffi.malloc(8)
    _QueryPerformanceFrequency(_qpf_buf)
    _QPC_FREQ = ffi.read_memory(_qpf_buf, ffi.c_int64)
    ffi.free(_qpf_buf)

else:
    _clock_gettime = _libc.clock_gettime([ffi.c_int32, ffi.c_void_p], ffi.c_int32)
    _nanosleep     = _libc.nanosleep(    [ffi.c_void_p, ffi.c_void_p], ffi.c_int32)
    _gmtime_r      = _libc.gmtime_r(    [ffi.c_void_p, ffi.c_void_p], ffi.c_void_p)
    _localtime_r   = _libc.localtime_r( [ffi.c_void_p, ffi.c_void_p], ffi.c_void_p)
    _mktime        = _libc.mktime(      [ffi.c_void_p], ffi.c_int64)
    _strftime      = _libc.strftime(    [ffi.c_char_p, ffi.c_uint64, ffi.c_char_p, ffi.c_void_p], ffi.c_uint64)

    # Detect whether CLOCK_MONOTONIC_RAW is available
    _has_mono_raw = False
    if _LINUX:
        _ts_probe = ffi.malloc(16)
        if _clock_gettime(_CLOCK_MONOTONIC_RAW, _ts_probe) == 0:
            _has_mono_raw = True
        ffi.free(_ts_probe)

    try:
        _tzname_ptr   = _libc.get_address("tzname")
        _timezone_ptr = _libc.get_address("timezone")
        _daylight_ptr = _libc.get_address("daylight")
    except ffi.FFIError:
        _tzname_ptr   = None
        _timezone_ptr = None
        _daylight_ptr = None

# ==============================================================================
#  Internal: helpers
# ==============================================================================

def _read_qpc():
    """Read the Windows QPC counter value."""
    buf = ffi.malloc(8)
    _QueryPerformanceCounter(buf)
    val = ffi.read_memory(buf, ffi.c_int64)
    ffi.free(buf)
    return val


def _clock_gettime_val(clock_id):
    """Read a POSIX clock and return (sec, nsec)."""
    ts = ffi.malloc(16)
    try:
        ret = _clock_gettime(clock_id, ts)
        if ret != 0:
            raise OSError(format_str("clock_gettime({clock_id}) failed"))
        sec  = ffi.read_memory_with_offset(ts, 0, ffi.c_int64)
        nsec = ffi.read_memory_with_offset(ts, 8, ffi.c_int64)
        return (sec, nsec)
    except Exception:
        pass
    finally:
        ffi.free(ts)


# struct tm layout (matches glibc / MSVC on 64-bit):
#  int tm_sec   [0]  — seconds [0, 60]
#  int tm_min   [4]  — minutes [0, 59]
#  int tm_hour  [8]  — hours [0, 23]
#  int tm_mday  [12] — day of month [1, 31]
#  int tm_mon   [16] — months since January [0, 11]
#  int tm_year  [20] — years since 1900
#  int tm_wday  [24] — days since Sunday [0, 6]
#  int tm_yday  [28] — days since January 1 [0, 365]
#  int tm_isdst [32] — DST flag
#
# On MSVC (Windows) the layout is identical for the first nine int fields.
# Total struct size we allocate: 64 bytes (generous padding for any extras).
_TM_SIZE = 64

def _read_tm(tm_ptr):
    """Read a struct tm* and return a struct_time tuple."""
    sec   = ffi.read_memory_with_offset(tm_ptr,  0, ffi.c_int32)
    min_  = ffi.read_memory_with_offset(tm_ptr,  4, ffi.c_int32)
    hour  = ffi.read_memory_with_offset(tm_ptr,  8, ffi.c_int32)
    mday  = ffi.read_memory_with_offset(tm_ptr, 12, ffi.c_int32)
    mon   = ffi.read_memory_with_offset(tm_ptr, 16, ffi.c_int32)
    year  = ffi.read_memory_with_offset(tm_ptr, 20, ffi.c_int32)
    wday  = ffi.read_memory_with_offset(tm_ptr, 24, ffi.c_int32)
    yday  = ffi.read_memory_with_offset(tm_ptr, 28, ffi.c_int32)
    isdst = ffi.read_memory_with_offset(tm_ptr, 32, ffi.c_int32)
    # tm_wday  = (wday + 6) % 7,   # Python: 0=Mon; C: 0=Sun
    return struct_time(tm_year  = year + 1900, tm_mon   = mon  + 1, tm_mday  = mday, tm_hour  = hour, tm_min   = min_, tm_sec   = sec, tm_wday  = (wday + 6) % 7, tm_yday  = yday + 1, tm_isdst = isdst,)


def _write_tm(tm_ptr, st):
    """Write a struct_time back into a struct tm*."""
    ffi.write_memory_with_offset(tm_ptr,  0, ffi.c_int32, st.tm_sec)
    ffi.write_memory_with_offset(tm_ptr,  4, ffi.c_int32, st.tm_min)
    ffi.write_memory_with_offset(tm_ptr,  8, ffi.c_int32, st.tm_hour)
    ffi.write_memory_with_offset(tm_ptr, 12, ffi.c_int32, st.tm_mday)
    ffi.write_memory_with_offset(tm_ptr, 16, ffi.c_int32, st.tm_mon - 1)
    ffi.write_memory_with_offset(tm_ptr, 20, ffi.c_int32, st.tm_year - 1900)
    # wday: Python 0=Mon → C 0=Sun
    wday_c = (st.tm_wday + 1) % 7
    ffi.write_memory_with_offset(tm_ptr, 24, ffi.c_int32, wday_c)
    ffi.write_memory_with_offset(tm_ptr, 28, ffi.c_int32, st.tm_yday - 1)
    ffi.write_memory_with_offset(tm_ptr, 32, ffi.c_int32, st.tm_isdst)

# ==============================================================================
#  struct_time
# ==============================================================================

class struct_time:
    """
    Mirrors Python's time.struct_time.

    Fields (all integers):
      tm_year   — full 4-digit year, e.g. 2025
      tm_mon    — month [1, 12]
      tm_mday   — day of month [1, 31]
      tm_hour   — hour [0, 23]
      tm_min    — minute [0, 59]
      tm_sec    — second [0, 60]  (60 allows for a leap second)
      tm_wday   — weekday [0=Mon, 6=Sun]
      tm_yday   — day of year [1, 366]
      tm_isdst  — DST flag: 1=DST, 0=not, -1=unknown
    """
    def __init__(self, tm_year=1970, tm_mon=1, tm_mday=1, tm_hour=0, tm_min=0, tm_sec=0, tm_wday=0, tm_yday=1, tm_isdst=-1):
        self.tm_year  = tm_year
        self.tm_mon   = tm_mon
        self.tm_mday  = tm_mday
        self.tm_hour  = tm_hour
        self.tm_min   = tm_min
        self.tm_sec   = tm_sec
        self.tm_wday  = tm_wday
        self.tm_yday  = tm_yday
        self.tm_isdst = tm_isdst

    def __repr__(self):
        return format_str("time.struct_time(tm_year={self.tm_year}, tm_mon={self.tm_mon}, tm_mday={self.tm_mday}, tm_hour={self.tm_hour}, tm_min={self.tm_min}, tm_sec={self.tm_sec}, tm_wday={self.tm_wday}, tm_yday={self.tm_yday}, tm_isdst={self.tm_isdst})")

# ==============================================================================
#  Public: wall-clock time
# ==============================================================================

if _WINDOWS:
    # Windows FILETIME epoch is 1601-01-01; Unix epoch offset is
    # 116444736000000000 * 100ns ticks.
    _EPOCH_OFFSET_100NS = 116444736000000000

    def time():
        """Return wall-clock time as seconds since the Unix epoch (float)."""
        buf = ffi.malloc(8)
        try:
            _GetSystemTimeAsFileTime(buf)
            ft = ffi.read_memory(buf, ffi.c_uint64)
            unix_100ns = ft - _EPOCH_OFFSET_100NS
            return float(unix_100ns) / 10000000.0
        except Exception:
            pass
        finally:
            ffi.free(buf)

else:
    def time():
        """Return wall-clock time as seconds since the Unix epoch (float)."""
        sec, nsec = _clock_gettime_val(_CLOCK_REALTIME)
        return float(sec) + float(nsec) / 1000000000.0

# ==============================================================================
#  Public: monotonic clock
# ==============================================================================

if _WINDOWS:
    _MONO_START = _read_qpc()

    def monotonic():
        """
        Monotonically increasing clock (seconds, float).
        Not affected by NTP adjustments or DST.
        The reference point is undefined; only differences are meaningful.
        """
        return float(_read_qpc() - _MONO_START) / float(_QPC_FREQ)

else:
    def monotonic():
        """
        Monotonically increasing clock (seconds, float).
        Uses CLOCK_MONOTONIC; guaranteed never to go backwards.
        """
        sec, nsec = _clock_gettime_val(_CLOCK_MONOTONIC)
        return float(sec) + float(nsec) / 1000000000.0

# ==============================================================================
#  Public: high-resolution performance counter
# ==============================================================================

if _WINDOWS:
    _PERF_START = _read_qpc()

    def perf_counter():
        """
        Highest-resolution timer available (seconds, float).
        Uses QueryPerformanceCounter on Windows.
        Only differences between two calls are meaningful.
        """
        return float(_read_qpc() - _PERF_START) / float(_QPC_FREQ)

else:
    def perf_counter():
        """
        Highest-resolution timer available (seconds, float).
        Uses CLOCK_MONOTONIC_RAW on Linux (immune to NTP slewing),
        falls back to CLOCK_MONOTONIC on other POSIX platforms.
        """
        clock = _CLOCK_MONOTONIC_RAW if _has_mono_raw else _CLOCK_MONOTONIC
        sec, nsec = _clock_gettime_val(clock)
        return float(sec) + float(nsec) / 1000000000.0

# ==============================================================================
#  Public: process / thread CPU time
# ==============================================================================

if not _WINDOWS:
    def process_time():
        """CPU time consumed by the current process (seconds, float)."""
        sec, nsec = _clock_gettime_val(_CLOCK_PROCESS_CPUTIME)
        return float(sec) + float(nsec) / 1000000000.0

    def thread_time():
        """CPU time consumed by the current thread (seconds, float)."""
        sec, nsec = _clock_gettime_val(_CLOCK_THREAD_CPUTIME)
        return float(sec) + float(nsec) / 1000000000.0

# ==============================================================================
#  Public: sleep
# ==============================================================================

def sleep(seconds):
    """
    Suspend execution for `seconds` seconds.
    `seconds` may be a float for sub-second precision.
    """
    if seconds < 0:
        raise ValueError("sleep length must be non-negative")

    if _WINDOWS:
        ms = int(seconds * 1000)
        _Sleep(ms)
    else:
        sec  = int(seconds)
        nsec = int((seconds - float(sec)) * 1000000000.0)
        ts = ffi.malloc(16)
        try:
            ffi.write_memory_with_offset(ts, 0, ffi.c_int64, sec)
            ffi.write_memory_with_offset(ts, 8, ffi.c_int64, nsec)
            _nanosleep(ts, None)
        except Exception:
            pass
        finally:
            ffi.free(ts)

# ==============================================================================
#  Public: calendar time conversions
# ==============================================================================

def gmtime(secs=None):
    """
    Convert `secs` (epoch float) to a struct_time in UTC.
    If secs is None, uses the current time().
    """
    if secs is None:
        secs = time()

    epoch_sec = int(secs)
    tm_buf  = ffi.malloc(_TM_SIZE)
    sec_buf = ffi.malloc(8)
    ffi.write_memory(sec_buf, ffi.c_int64, epoch_sec)

    try:
        if _WINDOWS:
            ret = _gmtime_s(tm_buf, sec_buf)
            if ret != 0:
                raise OSError("gmtime_s failed")
        else:
            _gmtime_r(sec_buf, tm_buf)
        return _read_tm(tm_buf)
    except Exception:
            pass
    finally:
        ffi.free(tm_buf)
        ffi.free(sec_buf)


def localtime(secs=None):
    """
    Convert `secs` (epoch float) to a struct_time in local time.
    If secs is None, uses the current time().
    """
    if secs is None:
        secs = time()

    epoch_sec = int(secs)
    tm_buf  = ffi.malloc(_TM_SIZE)
    sec_buf = ffi.malloc(8)
    ffi.write_memory(sec_buf, ffi.c_int64, epoch_sec)

    try:
        if _WINDOWS:
            ret = _localtime_s(tm_buf, sec_buf)
            if ret != 0:
                raise OSError("localtime_s failed")
        else:
            _localtime_r(sec_buf, tm_buf)
        return _read_tm(tm_buf)
    except Exception:
            pass
    finally:
        ffi.free(tm_buf)
        ffi.free(sec_buf)


def mktime(t):
    """
    Convert a struct_time expressed in local time to a float epoch value.
    This is the inverse of localtime().
    """
    tm_buf = ffi.malloc(_TM_SIZE)
    try:
        _write_tm(tm_buf, t)
        result = _mktime(tm_buf)
        if result == -1:
            raise OverflowError("mktime(): argument out of range")
        return float(result)
    except Exception:
        pass
    finally:
        ffi.free(tm_buf)

# ==============================================================================
#  Public: strftime
# ==============================================================================

def strftime(format_str_arg, t=None):
    """
    Format a struct_time (or the current local time if t is None) as a string
    using strftime(3) format codes.

    Common codes:
      %Y  4-digit year        %m  month [01,12]    %d  day [01,31]
      %H  hour [00,23]        %M  minute [00,59]   %S  second [00,60]
      %A  full weekday name   %B  full month name  %Z  timezone name
      %z  UTC offset (+HHMM)  %j  day of year      %w  weekday [0=Sun]
      %%  literal %
    """
    if t is None:
        t = localtime()

    fmt_bytes = format_str_arg.encode('utf-8') + b'\x00'
    buf_size  = 256
    out_buf   = ffi.malloc(buf_size)
    tm_buf    = ffi.malloc(_TM_SIZE)
    fmt_buf   = ffi.malloc(len(fmt_bytes))
    ffi.memcpy(fmt_buf, ffi.addressof(fmt_bytes), len(fmt_bytes))

    try:
        _write_tm(tm_buf, t)
        written = _strftime(out_buf, buf_size, fmt_buf, tm_buf)
        if written == 0:
            # Buffer might be too small; try larger
            ffi.free(out_buf)
            buf_size = 1024
            out_buf  = ffi.malloc(buf_size)
            written  = _strftime(out_buf, buf_size, fmt_buf, tm_buf)

        raw = ffi.buffer_to_bytes(out_buf, written)
        return raw.decode('utf-8', errors='replace')
    except Exception:
        pass
    finally:
        ffi.free(out_buf)
        ffi.free(tm_buf)
        ffi.free(fmt_buf)

# ==============================================================================
#  Public: timezone constants
# ==============================================================================

def _init_tz():
    """Read timezone metadata from libc globals."""
    tz  = 0
    dl  = 0
    std = "UTC"
    dst = "UTC"

    if _timezone_ptr:
        try:
            tz = ffi.read_memory(_timezone_ptr, ffi.c_int64)
        except Exception:
            pass
    if _daylight_ptr:
        try:
            dl = ffi.read_memory(_daylight_ptr, ffi.c_int32)
        except Exception:
            pass
    if _tzname_ptr:
        try:
            # tzname is char*[2] — two consecutive pointer-sized entries
            p0 = ffi.read_memory_with_offset(_tzname_ptr, 0, ffi.c_void_p)
            p1 = ffi.read_memory_with_offset(_tzname_ptr, 8, ffi.c_void_p)
            if p0:
                std = ffi.string_at(p0)
            if p1:
                dst = ffi.string_at(p1)
        except Exception:
            pass

    return (tz, dl, (std, dst))


_tz, _dl, _tzname = _init_tz()

# Module-level constants (same names as Python's time module)
timezone = _tz          # seconds WEST of UTC (positive = behind UTC)
daylight = _dl          # 1 if a DST timezone is defined
tzname   = _tzname      # ("STD", "DST")

# ==============================================================================
#  Public: miscellaneous helpers
# ==============================================================================

def ctime(secs=None):
    """
    Convert `secs` to a human-readable local-time string, e.g.
    'Mon Jan  1 00:00:00 2024'
    If secs is None, uses the current time().
    """
    return strftime("%a %b %e %H:%M:%S %Y", localtime(secs))


def asctime(t=None):
    """
    Format a struct_time as a fixed-width string: 'Mon Jan  1 00:00:00 2024'.
    If t is None, uses localtime().
    """
    if t is None:
        t = localtime()
    return strftime("%a %b %e %H:%M:%S %Y", t)


class _Stopwatch:
    """
    Convenience high-resolution stopwatch using perf_counter().

    Usage:
        sw = time.Stopwatch()
        sw.start()
        ... do work ...
        elapsed = sw.elapsed()   # seconds, float
        sw.reset()
    """
    def __init__(self):
        self._start = None
        self._split = None

    def start(self):
        """Start (or restart) the stopwatch."""
        self._start = perf_counter()
        self._split = self._start
        return self

    def elapsed(self):
        """Seconds since start() was called."""
        if self._start is None:
            return 0.0
        return perf_counter() - self._start

    def split(self):
        """
        Seconds since the last split() call (or start() if first call).
        Also resets the split reference point.
        """
        now = perf_counter()
        if self._split is None:
            self._split = now
        diff = now - self._split
        self._split = now
        return diff

    def reset(self):
        """Reset all state."""
        self._start = None
        self._split = None

Stopwatch = _Stopwatch