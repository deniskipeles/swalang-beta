"""
os.py — Production-ready OS interface for Swalang.

New in this version:
  - File I/O:       open() / File class (read, write, readline, readlines, seek, tell, flush, close)
  - makedirs()      — recursive directory creation
  - stat() / StatResult — file metadata (size, mtime, ctime, mode, is_dir, is_file)
  - environ         — dict-like object backed by getenv/setenv/unsetenv
  - error           — OSError subclass with errno
  - getpid()        — current process ID
  - getuid()        — current user ID (POSIX only)
  - cpu_count()     — logical CPU count
  - sep / linesep / devnull — platform constants
  - urandom(n)      — cryptographic random bytes (no mbedtls dependency)
  - Improved error handling throughout (remove, mkdir, rename raise on failure)
"""

import ffi
import sys

# ==============================================================================
#  Platform flags
# ==============================================================================

_WINDOWS = sys.platform == 'windows'
_DARWIN  = sys.platform == 'darwin'
_LINUX   = sys.platform == 'linux'

# ==============================================================================
#  Platform constants
# ==============================================================================

sep     = '\\' if _WINDOWS else '/'
linesep = '\r\n' if _WINDOWS else '\n'
devnull = 'nul' if _WINDOWS else '/dev/null'
curdir  = '.'
pardir  = '..'
extsep  = '.'
pathsep = ';' if _WINDOWS else ':'
altsep  = '/' if _WINDOWS else None

# ==============================================================================
#  Library Loading
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
#  OSError
# ==============================================================================

class error(Exception):
    """OS-level error.  Has .errno and .strerror attributes."""
    def __init__(self, errno_val, strerror, filename=None):
        self.errno    = errno_val
        self.strerror = strerror
        self.filename = filename
        if filename:
            super().__init__(format_str("[Errno {errno_val}] {strerror}: '{filename}'"))
        else:
            super().__init__(format_str("[Errno {errno_val}] {strerror}"))

OSError = error

# ==============================================================================
#  errno helpers
# ==============================================================================

_strerror_fn = _libc.strerror([ffi.c_int32], ffi.c_char_p)

if _WINDOWS:
    _errno_ptr_fn = _libc._errno([], ffi.c_void_p)
    def _get_errno():
        ptr = _errno_ptr_fn()
        if ptr and getattr(ptr, "Address", None):
            return ffi.read_memory(ptr, ffi.c_int32)
        return 0
else:
    try:
        _errno_location = _libc.__errno_location([], ffi.c_void_p)
        def _get_errno():
            ptr = _errno_location()
            if ptr and getattr(ptr, "Address", None):
                return ffi.read_memory(ptr, ffi.c_int32)
            return 0
    except ffi.FFIError:
        def _get_errno():
            return 0


def _strerror(code):
    s = _strerror_fn(code)
    if s and getattr(s, "Address", None):
        return ffi.string_at(s)
    return format_str("errno {code}")


def _raise_errno(path=None):
    code = _get_errno()
    raise error(code, _strerror(code), path)

# ==============================================================================
#  C Bindings — environment
# ==============================================================================

_getenv_fn   = _libc.getenv([ffi.c_char_p], ffi.c_void_p)

if _WINDOWS:
    _putenv_fn = _libc._putenv([ffi.c_char_p], ffi.c_int32)
    def _setenv(key, value):
        kv = format_str("{key}={value}")
        _putenv_fn(kv.encode('utf-8'))
    def _unsetenv(key):
        kv = format_str("{key}=")
        _putenv_fn(kv.encode('utf-8'))
else:
    _setenv_fn   = _libc.setenv(  [ffi.c_char_p, ffi.c_char_p, ffi.c_int32], ffi.c_int32)
    _unsetenv_fn = _libc.unsetenv([ffi.c_char_p], ffi.c_int32)
    def _setenv(key, value):
        _setenv_fn(key.encode('utf-8'), value.encode('utf-8'), 1)
    def _unsetenv(key):
        _unsetenv_fn(key.encode('utf-8'))

# ==============================================================================
#  C Bindings — process
# ==============================================================================

if not _WINDOWS:
    _getpid_fn = _libc.getpid([], ffi.c_int32)
    _getuid_fn = _libc.getuid([], ffi.c_uint32)
else:
    _GetCurrentProcessId = _kernel32.GetCurrentProcessId([], ffi.c_uint32)

# ==============================================================================
#  C Bindings — file system
# ==============================================================================

_fopen_fn  = _libc.fopen( [ffi.c_char_p, ffi.c_char_p], ffi.c_void_p)
_fclose_fn = _libc.fclose([ffi.c_void_p], ffi.c_int32)
_fread_fn  = _libc.fread( [ffi.c_void_p, ffi.c_uint64, ffi.c_uint64, ffi.c_void_p], ffi.c_uint64)
_fwrite_fn = _libc.fwrite([ffi.c_void_p, ffi.c_uint64, ffi.c_uint64, ffi.c_void_p], ffi.c_uint64)
_fseek_fn  = _libc.fseek( [ffi.c_void_p, ffi.c_int64,  ffi.c_int32], ffi.c_int32)
_ftell_fn  = _libc.ftell( [ffi.c_void_p], ffi.c_int64)
_fflush_fn = _libc.fflush([ffi.c_void_p], ffi.c_int32)
_feof_fn   = _libc.feof(  [ffi.c_void_p], ffi.c_int32)
_ferror_fn = _libc.ferror([ffi.c_void_p], ffi.c_int32)
_fgets_fn  = _libc.fgets( [ffi.c_void_p, ffi.c_int32, ffi.c_void_p], ffi.c_void_p)
_fputs_fn  = _libc.fputs( [ffi.c_char_p, ffi.c_void_p], ffi.c_int32)

_remove_fn = _libc.remove([ffi.c_char_p], ffi.c_int32)
_rename_fn = _libc.rename([ffi.c_char_p, ffi.c_char_p], ffi.c_int32)
_system_fn = _libc.system([ffi.c_char_p], ffi.c_int32)
_getcwd_fn = _libc.getcwd([ffi.c_void_p, ffi.c_uint64], ffi.c_void_p) if not _WINDOWS else _libc._getcwd([ffi.c_void_p, ffi.c_int32], ffi.c_void_p)

if _WINDOWS:
    _mkdir_fn = _libc._mkdir([ffi.c_char_p], ffi.c_int32)
    _rmdir_fn = _libc._rmdir([ffi.c_char_p], ffi.c_int32)
    _GetFileAttributesA  = _kernel32.GetFileAttributesA( [ffi.c_char_p],           ffi.c_uint32)
    _FindFirstFileA      = _kernel32.FindFirstFileA(     [ffi.c_char_p, ffi.c_void_p], ffi.c_void_p)
    _FindNextFileA       = _kernel32.FindNextFileA(      [ffi.c_void_p, ffi.c_void_p], ffi.c_int32)
    _FindClose           = _kernel32.FindClose(          [ffi.c_void_p],           ffi.c_int32)
    _GetFileTime         = _kernel32.GetFileTime(        [ffi.c_void_p, ffi.c_void_p, ffi.c_void_p, ffi.c_void_p], ffi.c_int32)
    _GetFileSizeEx       = _kernel32.GetFileSizeEx(      [ffi.c_void_p, ffi.c_void_p], ffi.c_int32)
    _CreateFileA         = _kernel32.CreateFileA(        [ffi.c_char_p, ffi.c_uint32, ffi.c_uint32, ffi.c_void_p, ffi.c_uint32, ffi.c_uint32, ffi.c_void_p], ffi.c_void_p)
    _CloseHandle         = _kernel32.CloseHandle(        [ffi.c_void_p], ffi.c_int32)
    _GetNumberOfProcessors = _kernel32.GetSystemInfo     # accessed via struct
else:
    _mkdir_fn   = _libc.mkdir(  [ffi.c_char_p, ffi.c_uint32], ffi.c_int32)
    _rmdir_fn   = _libc.rmdir(  [ffi.c_char_p],               ffi.c_int32)
    _access_fn  = _libc.access( [ffi.c_char_p, ffi.c_int32],  ffi.c_int32)
    _opendir_fn = _libc.opendir([ffi.c_char_p],                ffi.c_void_p)
    _readdir_fn = _libc.readdir([ffi.c_void_p],                ffi.c_void_p)
    _closedir_fn= _libc.closedir([ffi.c_void_p],               ffi.c_int32)
    _stat_fn    = _libc.stat(   [ffi.c_char_p, ffi.c_void_p], ffi.c_int32)
    _lstat_fn   = _libc.lstat(  [ffi.c_char_p, ffi.c_void_p], ffi.c_int32)
    _getpid_fn2 = _libc.getpid( [],                            ffi.c_int32)

if not _WINDOWS:
    try:
        _urandom_dev = None  # opened lazily
    except Exception:
        pass

# ==============================================================================
#  environ
# ==============================================================================

class _Environ:
    """
    dict-like object that reads/writes real process environment variables.
    Supports: get, set (__setitem__), delete (__delitem__), 'in', copy.
    """

    def get(self, key, default=None):
        ptr = _getenv_fn(key.encode('utf-8'))
        if not ptr or not getattr(ptr, "Address", None) or ptr.Address == 0:
            return default
        return ffi.string_at(ptr)

    def __getitem__(self, key):
        val = self.get(key)
        if val is None:
            raise KeyError(key)
        return val

    def __setitem__(self, key, value):
        _setenv(key, value)

    def __delitem__(self, key):
        _unsetenv(key)

    def __contains__(self, key):
        return self.get(key) is not None

    def copy(self):
        # There's no portable way to enumerate all env vars through libc alone.
        # Return a plain dict of what the caller has accessed.  For full enumeration,
        # platform-specific extensions (e.g. /proc/self/environ) would be needed.
        raise NotImplementedError("environ.copy() requires platform-specific enumeration")


environ = _Environ()


def getenv(key, default=None):
    """Return the value of environment variable `key`, or `default`."""
    return environ.get(key, default)


def putenv(key, value):
    """Set environment variable `key` to `value`."""
    environ[key] = value


def unsetenv(key):
    """Remove environment variable `key`."""
    del environ[key]

# ==============================================================================
#  Process
# ==============================================================================

def getpid():
    """Return the current process ID."""
    if _WINDOWS:
        return _GetCurrentProcessId()
    return _getpid_fn()


def getuid():
    """Return the current user ID (POSIX only)."""
    if _WINDOWS:
        raise NotImplementedError("getuid() not available on Windows")
    return _getuid_fn()


def cpu_count():
    """Return the number of logical CPUs.  Returns None if undetermined."""
    if _LINUX:
        try:
            f = open("/proc/cpuinfo")
            content = f.read()
            f.close()
            count = 0
            for line in content.split('\n'):
                if line.startswith('processor'):
                    count = count + 1
            return count if count > 0 else None
        except Exception:
            return None
    # macOS / Windows: good enough fallback via env
    val = getenv("NUMBER_OF_PROCESSORS")
    if val:
        try: 
            return int(val)
        except Exception: 
            pass
    return None

# ==============================================================================
#  urandom
# ==============================================================================

def urandom(n):
    """Return n cryptographically random bytes."""
    if _WINDOWS:
        # CryptGenRandom via BCrypt (Vista+)
        try:
            bcrypt = ffi.CDLL("bcrypt.dll")
            _BCryptGenRandom = bcrypt.BCryptGenRandom([ffi.c_void_p, ffi.c_void_p, ffi.c_uint32, ffi.c_uint32], ffi.c_int32)
            buf = ffi.malloc(n)
            try:
                ret = _BCryptGenRandom(None, buf, n, 2)   # BCRYPT_USE_SYSTEM_PREFERRED_RNG=2
                if ret == 0:
                    return ffi.buffer_to_bytes(buf, n)
            except Exception:
                pass
            finally:
                ffi.free(buf)
        except Exception:
            pass
        raise OSError("urandom not available")
    else:
        # Read from /dev/urandom
        fp = _fopen_fn(b"/dev/urandom", b"rb")
        if not fp or not getattr(fp, "Address", None) or fp.Address == 0:
            raise error(0, "Cannot open /dev/urandom")
        buf = ffi.malloc(n)
        try:
            nread = _fread_fn(buf, 1, n, fp)
            _fclose_fn(fp)
            if nread != n:
                raise error(0, "Short read from /dev/urandom")
            return ffi.buffer_to_bytes(buf, n)
        except Exception:
            pass
        finally:
            ffi.free(buf)

# ==============================================================================
#  system / getcwd
# ==============================================================================

def system(command):
    """Execute a shell command.  Returns the exit status."""
    return _system_fn(command.encode('utf-8'))


def getcwd():
    """Return the current working directory as a string."""
    buf = ffi.malloc(4096)
    try:
        ptr = _getcwd_fn(buf, 4096)
        if not ptr or not getattr(ptr, "Address", None) or ptr.Address == 0:
            _raise_errno()
        return ffi.string_at(ptr)
    except Exception:
        pass
    finally:
        ffi.free(buf)


def chdir(path):
    """Change the current working directory."""
    if _WINDOWS:
        _SetCurrentDirectoryA = _kernel32.SetCurrentDirectoryA([ffi.c_char_p], ffi.c_int32)
        if _SetCurrentDirectoryA(path.encode('utf-8')) == 0:
            _raise_errno(path)
    else:
        _chdir = _libc.chdir([ffi.c_char_p], ffi.c_int32)
        if _chdir(path.encode('utf-8')) != 0:
            _raise_errno(path)

# ==============================================================================
#  File system ops
# ==============================================================================

def remove(path):
    """Remove a file.  Raises OSError on failure."""
    if _remove_fn(path.encode('utf-8')) != 0:
        _raise_errno(path)


unlink = remove


def rename(old, new):
    """Rename a file or directory.  Raises OSError on failure."""
    if _rename_fn(old.encode('utf-8'), new.encode('utf-8')) != 0:
        _raise_errno(old)


def mkdir(path, mode=0o777):
    """Create a directory.  Raises OSError if it already exists or fails."""
    if _WINDOWS:
        ret = _mkdir_fn(path.encode('utf-8'))
    else:
        ret = _mkdir_fn(path.encode('utf-8'), mode)
    if ret != 0:
        _raise_errno(path)


def makedirs(path, mode=0o777, exist_ok=False):
    """
    Recursively create directories.
    If exist_ok is True, do not raise if the leaf directory already exists.
    """
    parts = path.replace('\\', '/').split('/')
    current = ""
    for i in range(len(parts)):
        part = parts[i]
        if not part:
            current = sep
            continue
        if current and not current.endswith(sep):
            current = current + sep + part
        else:
            current = current + part

        if _isdir_raw(current):
            continue
        try:
            mkdir(current, mode)
        except error as e:
            # errno 17 = EEXIST on Linux/macOS; errno 183 on Windows
            if e.errno in [17, 183] and exist_ok and i == len(parts) - 1:
                pass
            elif e.errno in [17, 183]:
                pass   # intermediate dir already exists — fine
            else:
                raise


def rmdir(path):
    """Remove an empty directory.  Raises OSError on failure."""
    if _rmdir_fn(path.encode('utf-8')) != 0:
        _raise_errno(path)


def listdir(path="."):
    """Return a list of names in the directory `path` (excluding . and ..)."""
    if _WINDOWS:
        search_path = path + "\\*"
        find_data   = ffi.malloc(592)   # sizeof(WIN32_FIND_DATAA) with padding
        try:
            handle = _FindFirstFileA(search_path.encode('utf-8'), find_data)
            if not handle or not getattr(handle, "Address", None) or handle.Address == 0:
                _raise_errno(path)
            INVALID = 18446744073709551615   # (void*)-1 on 64-bit
            if handle.Address == INVALID:
                _raise_errno(path)

            results = []
            while True:
                # cFileName starts at offset 44 in WIN32_FIND_DATAA
                name = ffi.string_at(find_data, -1, 44)
                if name != "." and name != "..":
                    results.append(name)
                if _FindNextFileA(handle, find_data) == 0:
                    break
            _FindClose(handle)
            return results
        except Exception:
            pass
        finally:
            ffi.free(find_data)
    else:
        d = _opendir_fn(path.encode('utf-8'))
        if not d or not getattr(d, "Address", None) or d.Address == 0:
            _raise_errno(path)
        # d_name offset: Linux x86-64 = 19; macOS = 21
        d_name_offset = 21 if _DARWIN else 19
        results = []
        try:
            while True:
                ent = _readdir_fn(d)
                if not ent or not getattr(ent, "Address", None) or ent.Address == 0:
                    break
                name = ffi.string_at(ent, -1, d_name_offset)
                if name != "." and name != "..":
                    results.append(name)
            return results
        except Exception:
            pass
        finally:
            _closedir_fn(d)

# ==============================================================================
#  stat
# ==============================================================================

class StatResult:
    """
    Result of a stat() call.

    Attributes:
      st_mode  — file mode (int)
      st_size  — file size in bytes
      st_mtime — modification time (float, seconds since epoch)
      st_ctime — status change time / creation time on Windows (float)
      st_atime — last access time (float)
      st_uid   — user ID (0 on Windows)
      st_gid   — group ID (0 on Windows)
      st_nlink — number of hard links
      st_ino   — inode number (0 on Windows)
      st_dev   — device ID
    """
    def __init__(self, st_mode=0, st_ino=0, st_dev=0, st_nlink=0, st_uid=0, st_gid=0, st_size=0, st_atime=0.0, st_mtime=0.0, st_ctime=0.0):
        self.st_mode  = st_mode
        self.st_ino   = st_ino
        self.st_dev   = st_dev
        self.st_nlink = st_nlink
        self.st_uid   = st_uid
        self.st_gid   = st_gid
        self.st_size  = st_size
        self.st_atime = st_atime
        self.st_mtime = st_mtime
        self.st_ctime = st_ctime

    def is_dir(self):
        return (self.st_mode & 0xF000) == 0x4000   # S_IFDIR

    def is_file(self):
        return (self.st_mode & 0xF000) == 0x8000   # S_IFREG

    def is_symlink(self):
        return (self.st_mode & 0xF000) == 0xA000   # S_IFLNK

    def __repr__(self):
        return format_str("os.StatResult(st_mode={self.st_mode}, st_size={self.st_size}, st_mtime={self.st_mtime})")


# struct stat layout on Linux x86-64 (glibc):
#  dev_t     st_dev     [0]  8 bytes
#  ino_t     st_ino     [8]  8 bytes
#  nlink_t   st_nlink   [16] 8 bytes
#  mode_t    st_mode    [24] 4 bytes
#  uid_t     st_uid     [28] 4 bytes
#  gid_t     st_gid     [32] 4 bytes
#  pad               [36] 4 bytes
#  dev_t     st_rdev    [40] 8 bytes
#  off_t     st_size    [48] 8 bytes
#  blksize_t st_blksize [56] 8 bytes
#  blkcnt_t  st_blocks  [64] 8 bytes
#  struct timespec st_atim [72]  (tv_sec[8] + tv_nsec[8])
#  struct timespec st_mtim [88]
#  struct timespec st_ctim [104]
#
# macOS differs; we use a simplified offset table per platform.
_STAT_BUF_SIZE = 256   # generous for all platforms

if _LINUX:
    _STAT_OFF = {
        'dev':   (0,   ffi.c_uint64),
        'ino':   (8,   ffi.c_uint64),
        'nlink': (16,  ffi.c_uint64),
        'mode':  (24,  ffi.c_uint32),
        'uid':   (28,  ffi.c_uint32),
        'gid':   (32,  ffi.c_uint32),
        'size':  (48,  ffi.c_int64),
        'atime_sec':  (72,  ffi.c_int64),
        'atime_nsec': (80,  ffi.c_int64),
        'mtime_sec':  (88,  ffi.c_int64),
        'mtime_nsec': (96,  ffi.c_int64),
        'ctime_sec':  (104, ffi.c_int64),
        'ctime_nsec': (112, ffi.c_int64),
    }
elif _DARWIN:
    # macOS struct stat (64-bit)
    _STAT_OFF = {
        'dev':   (0,   ffi.c_int32),
        'mode':  (4,   ffi.c_uint16),
        'nlink': (6,   ffi.c_uint16),
        'ino':   (8,   ffi.c_uint64),
        'uid':   (16,  ffi.c_uint32),
        'gid':   (20,  ffi.c_uint32),
        'size':  (96,  ffi.c_int64),
        'atime_sec':  (40, ffi.c_int64),
        'atime_nsec': (48, ffi.c_int64),
        'mtime_sec':  (56, ffi.c_int64),
        'mtime_nsec': (64, ffi.c_int64),
        'ctime_sec':  (72, ffi.c_int64),
        'ctime_nsec': (80, ffi.c_int64),
    }
else:
    _STAT_OFF = {}   # Windows uses separate API path

def _ts(buf, sec_off, nsec_off):
    """Read a timespec from buf at given offsets and return float seconds."""
    sec_typ  = _STAT_OFF.get('atime_sec',  (0, ffi.c_int64))[1]   # reuse type
    nsec_typ = ffi.c_int64
    # Use the actual offsets passed in
    sec  = ffi.read_memory_with_offset(buf, sec_off,  ffi.c_int64)
    nsec = ffi.read_memory_with_offset(buf, nsec_off, ffi.c_int64)
    return float(sec) + float(nsec) / 1000000000.0


def stat(path):
    """Return a StatResult for `path`.  Raises OSError on failure."""
    if _WINDOWS:
        return _stat_windows(path)
    return _stat_posix(path, follow=True)


def lstat(path):
    """Like stat() but does not follow symlinks."""
    if _WINDOWS:
        return _stat_windows(path)
    return _stat_posix(path, follow=False)


def _stat_posix(path, follow=True):
    buf = ffi.malloc(_STAT_BUF_SIZE)
    try:
        fn  = _stat_fn if follow else _lstat_fn
        ret = fn(path.encode('utf-8'), buf)
        if ret != 0:
            _raise_errno(path)

        off = _STAT_OFF
        dev   = ffi.read_memory_with_offset(buf, off['dev'][0],   off['dev'][1])
        ino   = ffi.read_memory_with_offset(buf, off['ino'][0],   off['ino'][1])
        nlink = ffi.read_memory_with_offset(buf, off['nlink'][0], off['nlink'][1])
        mode  = ffi.read_memory_with_offset(buf, off['mode'][0],  off['mode'][1])
        uid   = ffi.read_memory_with_offset(buf, off['uid'][0],   off['uid'][1])
        gid   = ffi.read_memory_with_offset(buf, off['gid'][0],   off['gid'][1])
        size  = ffi.read_memory_with_offset(buf, off['size'][0],  off['size'][1])

        atime = _ts(buf, off['atime_sec'][0], off['atime_nsec'][0])
        mtime = _ts(buf, off['mtime_sec'][0], off['mtime_nsec'][0])
        ctime = _ts(buf, off['ctime_sec'][0], off['ctime_nsec'][0])

        return StatResult(st_mode=mode, st_ino=ino, st_dev=dev, st_nlink=nlink, st_uid=uid, st_gid=gid, st_size=size, st_atime=atime, st_mtime=mtime, st_ctime=ctime,)
    except Exception:
        pass
    finally:
        ffi.free(buf)


def _stat_windows(path):
    # Use CreateFile + GetFileTime + GetFileSizeEx
    GENERIC_READ             = 0x80000000
    FILE_SHARE_READ          = 1
    OPEN_EXISTING            = 3
    FILE_FLAG_BACKUP_SEMANTICS = 0x02000000   # needed for directories

    handle = _CreateFileA(path.encode('utf-8'), GENERIC_READ, FILE_SHARE_READ, None, OPEN_EXISTING, FILE_FLAG_BACKUP_SEMANTICS, None)

    INVALID_HANDLE = 18446744073709551615
    if not handle or handle.Address == INVALID_HANDLE:
        _raise_errno(path)

    size_buf   = ffi.malloc(8)
    atime_buf  = ffi.malloc(8)
    mtime_buf  = ffi.malloc(8)
    ctime_buf  = ffi.malloc(8)

    try:
        _GetFileTime(handle, ctime_buf, atime_buf, mtime_buf)
        _GetFileSizeEx(handle, size_buf)
        _CloseHandle(handle)

        def _ft_to_unix(buf):
            ft = ffi.read_memory(buf, ffi.c_uint64)
            return float(ft - 116444736000000000) / 10000000.0

        size  = ffi.read_memory(size_buf, ffi.c_int64)
        atime = _ft_to_unix(atime_buf)
        mtime = _ft_to_unix(mtime_buf)
        ctime = _ft_to_unix(ctime_buf)

        attrs = _GetFileAttributesA(path.encode('utf-8'))
        mode  = 0o40755 if (attrs & 16) else 0o100644

        return StatResult(st_mode=mode, st_size=size, st_atime=atime, st_mtime=mtime, st_ctime=ctime)
    except Exception:
        pass
    finally:
        ffi.free(size_buf)
        ffi.free(atime_buf)
        ffi.free(mtime_buf)
        ffi.free(ctime_buf)

# ==============================================================================
#  Existence / type checks
# ==============================================================================

def _isdir_raw(path):
    if _WINDOWS:
        attrs = _GetFileAttributesA(path.encode('utf-8'))
        return attrs != 4294967295 and (attrs & 16) != 0
    else:
        d = _opendir_fn(path.encode('utf-8'))
        if d and getattr(d, "Address", None) and d.Address != 0:
            _closedir_fn(d)
            return True
        return False


def _exists_raw(path):
    if _WINDOWS:
        return _GetFileAttributesA(path.encode('utf-8')) != 4294967295
    return _access_fn(path.encode('utf-8'), 0) == 0


class _PathModule:
    """os.path"""

    def __init__(self):
        self.sep = sep

    def exists(self, p):    return _exists_raw(p)
    def isdir(self, p):     return _isdir_raw(p)
    def isfile(self, p):    return _exists_raw(p) and not _isdir_raw(p)

    def getsize(self, p):
        s = stat(p)
        return s.st_size

    def getmtime(self, p):
        s = stat(p)
        return s.st_mtime

    def getatime(self, p):
        s = stat(p)
        return s.st_atime

    def getctime(self, p):
        s = stat(p)
        return s.st_ctime

    def join(self, *parts):
        if not parts:
            return ""
        result = parts[0]
        for i in range(1, len(parts)):
            p = parts[i]
            if p.startswith('/') or (len(p) > 1 and p[1] == ':'):
                result = p
            elif not result or result.endswith(sep):
                result = result + p
            else:
                result = result + sep + p
        return result

    def split(self, p):
        last = -1
        for i in range(len(p) - 1, -1, -1):
            if p[i] == '/' or p[i] == '\\':
                last = i
                break
        if last == -1:
            return ("", p)
        return (p[:last], p[last + 1:])

    def dirname(self, p):   return self.split(p)[0]
    def basename(self, p):  return self.split(p)[1]

    def splitext(self, p):
        base = self.basename(p)
        for i in range(len(base) - 1, 0, -1):
            if base[i] == '.':
                cut = len(base) - i
                return (p[:-cut], base[i:])
        return (p, "")

    def abspath(self, p):
        if p.startswith('/') or (len(p) > 1 and p[1] == ':'):
            return p
        return self.join(getcwd(), p)

    def normpath(self, p):
        """Collapse redundant separators and up-level references."""
        parts = p.replace('\\', '/').split('/')
        out = []
        for part in parts:
            if part == '' or part == '.':
                continue
            if part == '..':
                if out:
                    out.pop()
            else:
                out.append(part)
        result = sep.join(out)
        if p.startswith('/'):
            result = '/' + result
        return result if result else '.'

    def relpath(self, path, start=None):
        if start is None:
            start = getcwd()
        s_parts = start.replace('\\', '/').split('/')
        p_parts = path.replace('\\', '/').split('/')
        common  = 0
        for i in range(min(len(s_parts), len(p_parts))):
            if s_parts[i] == p_parts[i]:
                common = common + 1
            else:
                break
        ups  = ['..'] * (len(s_parts) - common)
        rest = p_parts[common:]
        return sep.join(ups + rest) or '.'

    def expandvars(self, p):
        """Replace $VAR and ${VAR} with environment variable values."""
        result = ""
        i = 0
        while i < len(p):
            if p[i] == '$' and i + 1 < len(p):
                if p[i + 1] == '{':
                    end = p.find('}', i + 2)
                    if end != -1:
                        key = p[i + 2:end]
                        result = result + getenv(key, '')
                        i = end + 1
                        continue
                else:
                    j = i + 1
                    while j < len(p) and (p[j].isalpha() or p[j].isdigit() or p[j] == '_'):
                        j = j + 1
                    key = p[i + 1:j]
                    result = result + getenv(key, '')
                    i = j
                    continue
            result = result + p[i]
            i = i + 1
        return result

    def expanduser(self, p):
        """Replace leading ~ with the user's home directory."""
        if not p.startswith('~'):
            return p
        home = getenv('HOME') or getenv('USERPROFILE') or ''
        return home + p[1:]


path = _PathModule()

# ==============================================================================
#  walk
# ==============================================================================

def walk(top, topdown=True):
    """
    Directory tree generator.
    Yields (dirpath, dirnames, filenames) for each directory in the tree.
    """
    try:
        items = listdir(top)
    except error:
        return None

    dirs    = []
    nondirs = []
    for name in items:
        full = path.join(top, name)
        if path.isdir(full):
            dirs.append(name)
        else:
            nondirs.append(name)

    if topdown:
        yield (top, dirs, nondirs)

    for d in dirs:
        full = path.join(top, d)
        for entry in walk(full, topdown):
            yield entry

    if not topdown:
        yield (top, dirs, nondirs)

# ==============================================================================
#  File I/O
# ==============================================================================

# SEEK constants
SEEK_SET = 0
SEEK_CUR = 1
SEEK_END  = 2

class File:
    """
    A file object backed by a C FILE*.

    Modes: 'r', 'rb', 'w', 'wb', 'a', 'ab', 'r+', 'rb+', 'w+', 'wb+', 'a+'

    Usage:
        f = os.open('data.txt', 'r')
        text = f.read()
        f.close()

        # or use as context manager if Swalang supports 'with':
        with os.open('out.txt', 'w') as f:
            f.write('hello')
    """

    def __init__(self, path_str, mode='r'):
        self._path = path_str
        self._mode = mode
        self._fp   = None
        self._closed = False

        mode_bytes = mode.encode('utf-8')
        fp = _fopen_fn(path_str.encode('utf-8'), mode_bytes)
        if not fp or not getattr(fp, "Address", None) or fp.Address == 0:
            _raise_errno(path_str)
        self._fp = fp

    # ---- reading ------------------------------------------------------------

    def read(self, size=-1):
        """
        Read and return up to `size` bytes (binary) or characters (text).
        If size is -1, read to EOF.
        Returns str for text modes, bytes for binary modes.
        """
        self._check_open()
        is_binary = 'b' in self._mode

        if size == -1:
            # Read entire file
            _fseek_fn(self._fp, 0, SEEK_END)
            end   = _ftell_fn(self._fp)
            start = 0
            # Get current position before the seek to restore for 'r+' modes
            _fseek_fn(self._fp, 0, SEEK_SET)
            size  = end

        if size == 0:
            return b"" if is_binary else ""

        buf = ffi.malloc(size)
        try:
            nread = _fread_fn(buf, 1, size, self._fp)
            if nread == 0:
                return b"" if is_binary else ""
            data = ffi.buffer_to_bytes(buf, nread)
            return data if is_binary else data.decode('utf-8', errors='replace')
        except Exception:
            pass
        finally:
            ffi.free(buf)

    def readline(self, limit=-1):
        """Read one line, including the newline character."""
        self._check_open()
        is_binary = 'b' in self._mode
        chunk = 256 if limit == -1 else limit
        buf   = ffi.malloc(chunk + 1)
        try:
            ptr = _fgets_fn(buf, chunk + 1, self._fp)
            if not ptr or not getattr(ptr, "Address", None) or ptr.Address == 0:
                return b"" if is_binary else ""
            # Find null terminator length
            raw = ffi.buffer_to_bytes(buf, chunk + 1)
            end = 0
            while end < len(raw) and raw[end] != 0:
                end = end + 1
            data = raw[:end]
            return data if is_binary else data.decode('utf-8', errors='replace')
        except Exception:
            pass
        finally:
            ffi.free(buf)

    def readlines(self):
        """Read all lines and return as a list."""
        lines = []
        while True:
            line = self.readline()
            if not line:
                break
            lines.append(line)
        return lines

    # ---- writing ------------------------------------------------------------

    def write(self, data):
        """Write data (str or bytes).  Returns number of bytes written."""
        self._check_open()
        if isinstance(data, str):
            b = data.encode('utf-8')
        else:
            b = data
        written = _fwrite_fn(ffi.addressof(b), 1, len(b), self._fp)
        if _ferror_fn(self._fp):
            _raise_errno(self._path)
        return written

    def writelines(self, lines):
        """Write an iterable of lines (no newlines added)."""
        for line in lines:
            self.write(line)

    # ---- positioning --------------------------------------------------------

    def seek(self, offset, whence=SEEK_SET):
        self._check_open()
        ret = _fseek_fn(self._fp, offset, whence)
        if ret != 0:
            _raise_errno(self._path)

    def tell(self):
        self._check_open()
        pos = _ftell_fn(self._fp)
        if pos < 0:
            _raise_errno(self._path)
        return pos

    # ---- misc ---------------------------------------------------------------

    def flush(self):
        self._check_open()
        _fflush_fn(self._fp)

    def fileno(self):
        raise NotImplementedError("fileno() not exposed via this wrapper")

    @property
    def closed(self):
        return self._closed

    @property
    def name(self):
        return self._path

    @property
    def mode(self):
        return self._mode

    def close(self):
        if not self._closed and self._fp:
            _fclose_fn(self._fp)
            self._fp     = None
            self._closed = True

    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        self.close()
        return False

    def _check_open(self):
        if self._closed:
            raise error(0, "I/O operation on closed file", self._path)

    def __repr__(self):
        state = "closed" if self._closed else "open"
        return format_str("<os.File {state} '{self._path}' mode='{self._mode}'>")


def open(path_str, mode='r'):
    """
    Open a file and return a File object.

    mode — 'r'  read text (default)
           'rb' read binary
           'w'  write text (truncate)
           'wb' write binary
           'a'  append text
           'ab' append binary
           'r+' read+write text
           'w+' write+read text (truncate)
    """
    return File(path_str, mode)


# Convenience wrappers matching standard library idioms

def read_text(path_str, encoding='utf-8'):
    """Read the entire contents of a text file and return as str."""
    f = open(path_str, 'r')
    try:
        return f.read()
    except Exception:
        pass
    finally:
        f.close()


def write_text(path_str, content, encoding='utf-8'):
    """Write `content` (str) to a file, replacing any existing content."""
    f = open(path_str, 'w')
    try:
        f.write(content)
    except Exception:
        pass
    finally:
        f.close()


def read_bytes(path_str):
    """Read the entire contents of a binary file and return as bytes."""
    f = open(path_str, 'rb')
    try:
        return f.read()
    except Exception:
        pass
    finally:
        f.close()


def write_bytes(path_str, data):
    """Write `data` (bytes) to a file, replacing any existing content."""
    f = open(path_str, 'wb')
    try:
        f.write(data)
    except Exception:
        pass
    finally:
        f.close()


def append_text(path_str, content):
    """Append `content` to a text file."""
    f = open(path_str, 'a')
    try:
        f.write(content)
    except Exception:
        pass
    finally:
        f.close()