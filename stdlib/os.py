import ffi
import sys

def _load_libc():
    if sys.platform == 'windows':
        return ffi.CDLL("msvcrt.dll")
    elif sys.platform == 'darwin':
        return ffi.CDLL("libc.dylib")
    else:
        for name in["libc.so.6", "libc.so", "c"]:
            try:
                return ffi.CDLL(name)
            except ffi.FFIError:
                pass
        raise ffi.FFIError("Could not load libc")

libc = _load_libc()

_getenv = libc.getenv([ffi.c_char_p], ffi.c_void_p)
_system = libc.system([ffi.c_char_p], ffi.c_int32)
_remove = libc.remove([ffi.c_char_p], ffi.c_int32)
_rename = libc.rename([ffi.c_char_p, ffi.c_char_p], ffi.c_int32)

_fopen = libc.fopen([ffi.c_char_p, ffi.c_char_p], ffi.c_void_p)
_fseek = libc.fseek([ffi.c_void_p, ffi.c_int64, ffi.c_int32], ffi.c_int32)
_ftell = libc.ftell([ffi.c_void_p], ffi.c_int64)
_fclose = libc.fclose([ffi.c_void_p], ffi.c_int32)

if sys.platform == 'windows':
    kernel32 = ffi.windll("kernel32.dll")
    _getcwd = libc._getcwd([ffi.c_void_p, ffi.c_int32], ffi.c_void_p)
    _mkdir_c = libc._mkdir([ffi.c_char_p], ffi.c_int32)
    _GetFileAttributesA = kernel32.GetFileAttributesA([ffi.c_char_p], ffi.c_uint32)
    _FindFirstFileA = kernel32.FindFirstFileA([ffi.c_char_p, ffi.c_void_p], ffi.c_void_p)
    _FindNextFileA = kernel32.FindNextFileA([ffi.c_void_p, ffi.c_void_p], ffi.c_int32)
    _FindClose = kernel32.FindClose([ffi.c_void_p], ffi.c_int32)
else:
    _getcwd = libc.getcwd([ffi.c_void_p, ffi.c_uint64], ffi.c_void_p)
    _mkdir_c = libc.mkdir([ffi.c_char_p, ffi.c_uint32], ffi.c_int32)
    _access = libc.access([ffi.c_char_p, ffi.c_int32], ffi.c_int32)
    _opendir = libc.opendir([ffi.c_char_p], ffi.c_void_p)
    _readdir = libc.readdir([ffi.c_void_p], ffi.c_void_p)
    _closedir = libc.closedir([ffi.c_void_p], ffi.c_int32)

def getenv(key, default_value=None):
    ptr = _getenv(key.encode('utf-8'))
    if not ptr or ptr.Address == 0:
        return default_value
    return ffi.string_at(ptr)

def system(command):
    return _system(command.encode('utf-8'))

def remove(path):
    res = _remove(path.encode('utf-8'))
    if res != 0:
        pass # Ignore failure to avoid crashing the script if file doesn't exist

unlink = remove

def rename(old, new):
    res = _rename(old.encode('utf-8'), new.encode('utf-8'))
    if res != 0:
        raise Exception(format_str("Failed to rename file from {old} to {new}"))

def mkdir(path):
    if sys.platform == 'windows':
        res = _mkdir_c(path.encode('utf-8'))
    else:
        res = _mkdir_c(path.encode('utf-8'), 511) 
    if res != 0:
        pass # Ignore error if directory exists

def getcwd():
    buf = ffi.malloc(1024)
    try:
        ptr = _getcwd(buf, 1024)
        if not ptr or ptr.Address == 0:
            raise Exception("Failed to get current working directory")
        return ffi.string_at(ptr)
    finally:
        ffi.free(buf)

def _getsize(path):
    f = _fopen(path.encode('utf-8'), b"rb")
    if not f or f.Address == 0:
        raise Exception(format_str("File not found or cannot be opened: {path}"))
    _fseek(f, 0, 2) # SEEK_END = 2
    size = _ftell(f)
    _fclose(f)
    if size > 9223372036854700000 or size == 4294967295:
        raise Exception(format_str("Failed to read file size: {path}"))
    return size

if sys.platform == 'windows':
    def _exists(path):
        attrs = _GetFileAttributesA(path.encode('utf-8'))
        return attrs != 4294967295 

    def _isdir(path):
        attrs = _GetFileAttributesA(path.encode('utf-8'))
        if attrs == 4294967295: return False
        return (attrs & 16) != 0 

    def listdir(path="."):
        search_path = path + "\\*"
        find_data = ffi.malloc(512) # Safe padding for WIN32_FIND_DATAA
        try:
            handle = _FindFirstFileA(search_path.encode('utf-8'), find_data)
            if not handle or handle.Address == 0 or handle.Address == -1:
                return []
            
            results =[]
            while True:
                # FIX: cFileName is an embedded array, not a pointer.
                # Use string_at with an offset to read directly from the struct memory!
                name_str = ffi.string_at(find_data, -1, 44)
                
                if name_str != "." and name_str != "..":
                    results.append(name_str)
                if _FindNextFileA(handle, find_data) == 0:
                    break
            _FindClose(handle)
            return results
        finally:
            ffi.free(find_data)
else:
    def _exists(path):
        return _access(path.encode('utf-8'), 0) == 0

    def _isdir(path):
        d = _opendir(path.encode('utf-8'))
        if d and d.Address != 0:
            _closedir(d)
            return True
        return False

    def listdir(path="."):
        d = _opendir(path.encode('utf-8'))
        if not d or d.Address == 0:
            return[]
            
        d_name_offset = 21 if sys.platform == 'darwin' else 19
        results =[]
        try:
            while True:
                ent_ptr = _readdir(d)
                if not ent_ptr or ent_ptr.Address == 0:
                    break
                
                name = ffi.string_at(ent_ptr, -1, d_name_offset)
                
                if name != "." and name != "..":
                    results.append(name)
            return results
        finally:
            _closedir(d)

def _isfile(path):
    return _exists(path) and not _isdir(path)


class _PathModule:
    def __init__(self):
        self.sep = '\\' if sys.platform == 'windows' else '/'
        
    def exists(self, p): return _exists(p)
    def isdir(self, p): return _isdir(p)
    def isfile(self, p): return _isfile(p)
    def getsize(self, p): return _getsize(p)
    
    def join(self, *paths):
        if len(paths) == 0: return ""
        result = paths[0]
        for i in range(1, len(paths)):
            p = paths[i]
            if p.startswith(self.sep): 
                result = p
            elif not result or result.endswith(self.sep): 
                result = result + p
            else: 
                result = result + self.sep + p
        return result
        
    def split(self, p):
        idx = -1
        for i in range(len(p)-1, -1, -1):
            if p[i] == self.sep:
                idx = i
                break
        if idx == -1: return "", p
        return p[:idx], p[idx+1:]
        
    def dirname(self, p): return self.split(p)[0]
    def basename(self, p): return self.split(p)[1]
    
    def splitext(self, p):
        base = self.basename(p)
        idx = -1
        for i in range(len(base)-1, -1, -1):
            if base[i] == '.':
                idx = i
                break
        if idx <= 0: 
            return p, ""
        return p[:-(len(base)-idx)], base[idx:]
        
    def abspath(self, p):
        if p.startswith(self.sep) or (self.sep == '\\' and len(p)>1 and p[1] == ':'):
            return p
        return self.join(getcwd(), p)

path = _PathModule()

def walk(top):
    dirs = []
    nondirs =[]
    try:
        items = listdir(top)
    except Exception:
        return None
        
    for name in items:
        full_path = path.join(top, name)
        if path.isdir(full_path):
            dirs.append(name)
        else:
            nondirs.append(name)
            
    yield top, dirs, nondirs
    
    for name in dirs:
        full_path = path.join(top, name)
        for x in walk(full_path):
            yield x