import ffi
import sys

# Robustly find the actual C standard library binary
def _load_libc():
    if sys.platform == 'windows':
        return ffi.CDLL("msvcrt.dll")
    elif sys.platform == 'darwin':
        return ffi.CDLL("libc.dylib")
    else:
        for name in ["libc.so.6", "libc.so", "c"]:
            try:
                return ffi.CDLL(name)
            except ffi.FFIError:
                pass
        raise ffi.FFIError("Could not load libc")

libc = _load_libc()

if sys.platform == 'windows':
    kernel32 = ffi.windll("kernel32.dll")
    _sleep = kernel32.Sleep([ffi.c_uint32], None)
    
    # msvcrt _time64 is 1-second precision.
    _time = libc._time64([ffi.c_void_p], ffi.c_int64)

    def sleep(seconds):
        _sleep(int(seconds * 1000))

    def time():
        return float(_time(None))

else:
    # Linux / macOS High Precision API
    _clock_gettime = libc.clock_gettime([ffi.c_int32, ffi.c_void_p], ffi.c_int32)
    _nanosleep = libc.nanosleep([ffi.c_void_p, ffi.c_void_p], ffi.c_int32)
    
    CLOCK_REALTIME = 0

    def time():
        """Returns the time in seconds since the epoch as a floating point number."""
        # struct timespec { time_t tv_sec; long tv_nsec; };
        # On 64-bit systems, this is 16 bytes (8 bytes sec, 8 bytes nsec)
        ts_ptr = ffi.malloc(16)
        try:
            res = _clock_gettime(CLOCK_REALTIME, ts_ptr)
            if res != 0:
                raise Exception("clock_gettime failed")
            
            sec = ffi.read_memory_with_offset(ts_ptr, 0, ffi.c_int64)
            nsec = ffi.read_memory_with_offset(ts_ptr, 8, ffi.c_int64)
            
            return float(sec) + (float(nsec) / 1000000000.0)
        finally:
            ffi.free(ts_ptr)

    def sleep(seconds):
        """Delay execution for a given number of seconds. The argument may be a floating point number."""
        if seconds < 0:
            raise ValueError("sleep length must be non-negative")
            
        sec = int(seconds)
        nsec = int((seconds - sec) * 1000000000.0)
        
        ts_ptr = ffi.malloc(16)
        try:
            ffi.write_memory_with_offset(ts_ptr, 0, ffi.c_int64, sec)
            ffi.write_memory_with_offset(ts_ptr, 8, ffi.c_int64, nsec)
            
            _nanosleep(ts_ptr, None)
        finally:
            ffi.free(ts_ptr)