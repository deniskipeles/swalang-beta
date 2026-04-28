# pylearn/stdlib/sqlite.py

"""
A Pylearn wrapper for the SQLite3 C library, built using the Pylearn FFI.
This provides a Python-compatible DB-API 2.0 interface.
"""

import ffi
import sys

def _load_library():
    platform = sys.platform
    candidates = []
    if platform == 'linux':
        candidates = ["bin/x86_64-linux/sqlite3/libsqlite3.so", "libsqlite3.so"]
    elif platform == 'windows':
        candidates = ["bin/x86_64-windows-gnu/sqlite3/sqlite3.dll", "sqlite3.dll"]
    elif platform == 'darwin':
        candidates = ["libsqlite3.dylib"]
    
    for name in candidates:
        try:
            return ffi.CDLL(name)
        except ffi.FFIError:
            pass
    raise ffi.FFIError("Could not load sqlite3 shared library")

_lib = _load_library()

# --- Constants ---
SQLITE_OK = 0
SQLITE_ROW = 100
SQLITE_DONE = 101

SQLITE_INTEGER = 1
SQLITE_FLOAT = 2
SQLITE_TEXT = 3
SQLITE_BLOB = 4
SQLITE_NULL = 5

# Used to tell SQLite to make its own copy of a string being bound
SQLITE_TRANSIENT = -1

# --- C Function Signatures ---
_sqlite3_open = _lib.sqlite3_open([ffi.c_char_p, ffi.POINTER(ffi.c_void_p)], ffi.c_int32)
_sqlite3_close = _lib.sqlite3_close([ffi.c_void_p], ffi.c_int32)
_sqlite3_errmsg = _lib.sqlite3_errmsg([ffi.c_void_p], ffi.c_char_p)

_sqlite3_prepare_v2 = _lib.sqlite3_prepare_v2([ffi.c_void_p, ffi.c_char_p, ffi.c_int32, ffi.POINTER(ffi.c_void_p), ffi.POINTER(ffi.c_char_p)], ffi.c_int32)
_sqlite3_step = _lib.sqlite3_step([ffi.c_void_p], ffi.c_int32)
_sqlite3_finalize = _lib.sqlite3_finalize([ffi.c_void_p], ffi.c_int32)

_sqlite3_column_count = _lib.sqlite3_column_count([ffi.c_void_p], ffi.c_int32)
_sqlite3_column_type = _lib.sqlite3_column_type([ffi.c_void_p, ffi.c_int32], ffi.c_int32)
_sqlite3_column_int64 = _lib.sqlite3_column_int64([ffi.c_void_p, ffi.c_int32], ffi.c_int64)
_sqlite3_column_double = _lib.sqlite3_column_double([ffi.c_void_p, ffi.c_int32], ffi.c_double)
_sqlite3_column_text = _lib.sqlite3_column_text([ffi.c_void_p, ffi.c_int32], ffi.c_char_p)

_sqlite3_bind_int64 = _lib.sqlite3_bind_int64([ffi.c_void_p, ffi.c_int32, ffi.c_int64], ffi.c_int32)
_sqlite3_bind_double = _lib.sqlite3_bind_double([ffi.c_void_p, ffi.c_int32, ffi.c_double], ffi.c_int32)
_sqlite3_bind_text = _lib.sqlite3_bind_text([ffi.c_void_p, ffi.c_int32, ffi.c_char_p, ffi.c_int32, ffi.c_int64], ffi.c_int32)
_sqlite3_bind_null = _lib.sqlite3_bind_null([ffi.c_void_p, ffi.c_int32], ffi.c_int32)

_sqlite3_changes = _lib.sqlite3_changes([ffi.c_void_p], ffi.c_int32)
_sqlite3_last_insert_rowid = _lib.sqlite3_last_insert_rowid([ffi.c_void_p], ffi.c_int64)

class Error(Exception):
    pass

class Cursor:
    def __init__(self, db_ptr, conn):
        self.db_ptr = db_ptr
        self.conn = conn
        self.stmt_ptr = None
        self.rowcount = -1
        self.lastrowid = None
        self._first_row_ready = False

    def _check(self, res):
        if res != SQLITE_OK and res != SQLITE_ROW and res != SQLITE_DONE:
            # FIX: errmsg returns a native string directly
            err_msg = _sqlite3_errmsg(self.db_ptr)
            raise Error(format_str("SQLite error {res}: {err_msg}"))

    def execute(self, sql, parameters=None):
        if self.stmt_ptr:
            _sqlite3_finalize(self.stmt_ptr)
            self.stmt_ptr = None
            self._first_row_ready = False

        stmt_ptr_ptr = ffi.malloc(8)
        try:
            res = _sqlite3_prepare_v2(self.db_ptr, sql, -1, stmt_ptr_ptr, None)
            self._check(res)
            self.stmt_ptr = ffi.read_memory(stmt_ptr_ptr, ffi.c_void_p)
        finally:
            ffi.free(stmt_ptr_ptr)
        
        if not self.stmt_ptr or self.stmt_ptr.Address == 0:
            raise Error("Failed to prepare statement. Syntax error in SQL?")

        if parameters:
            if not isinstance(parameters, tuple):
                raise TypeError("Parameters must be a tuple")
            for i in range(len(parameters)):
                val = parameters[i]
                idx = i + 1
                if val is None:
                    self._check(_sqlite3_bind_null(self.stmt_ptr, idx))
                elif isinstance(val, int):
                    self._check(_sqlite3_bind_int64(self.stmt_ptr, idx, val))
                elif isinstance(val, float):
                    self._check(_sqlite3_bind_double(self.stmt_ptr, idx, val))
                elif isinstance(val, str):
                    self._check(_sqlite3_bind_text(self.stmt_ptr, idx, val, -1, SQLITE_TRANSIENT))
                else:
                    raise TypeError(format_str("Unsupported parameter type: {type(val)}"))

        res = _sqlite3_step(self.stmt_ptr)
        self._check(res)
        
        self.rowcount = _sqlite3_changes(self.db_ptr)
        self.lastrowid = _sqlite3_last_insert_rowid(self.db_ptr)
        
        if res == SQLITE_ROW:
            self._first_row_ready = True
        elif res == SQLITE_DONE:
            _sqlite3_finalize(self.stmt_ptr)
            self.stmt_ptr = None
            
        return self

    def _fetch_row(self):
        if not self.stmt_ptr:
            return None
        
        count = _sqlite3_column_count(self.stmt_ptr)
        row = []
        for i in range(count):
            col_type = _sqlite3_column_type(self.stmt_ptr, i)
            if col_type == SQLITE_INTEGER:
                row.append(_sqlite3_column_int64(self.stmt_ptr, i))
            elif col_type == SQLITE_FLOAT:
                row.append(_sqlite3_column_double(self.stmt_ptr, i))
            elif col_type == SQLITE_TEXT:
                # FIX: column_text returns a native string directly
                row.append(_sqlite3_column_text(self.stmt_ptr, i))
            elif col_type == SQLITE_NULL:
                row.append(None)
            else:
                row.append(None)
        return tuple(row)

    def fetchone(self):
        if not self.stmt_ptr:
            return None

        if self._first_row_ready:
            row = self._fetch_row()
            self._first_row_ready = False
            return row

        res = _sqlite3_step(self.stmt_ptr)
        if res == SQLITE_DONE:
            _sqlite3_finalize(self.stmt_ptr)
            self.stmt_ptr = None
            return None
            
        self._check(res)
        return self._fetch_row()

    def fetchall(self):
        if not self.stmt_ptr:
            return []
            
        rows = []
        if self._first_row_ready:
            rows.append(self._fetch_row())
            self._first_row_ready = False

        while True:
            res = _sqlite3_step(self.stmt_ptr)
            if res == SQLITE_DONE:
                break
            self._check(res)
            rows.append(self._fetch_row())
            
        _sqlite3_finalize(self.stmt_ptr)
        self.stmt_ptr = None
        return rows

    def close(self):
        if self.stmt_ptr:
            _sqlite3_finalize(self.stmt_ptr)
            self.stmt_ptr = None


class Connection:
    def __init__(self, database):
        self.db_ptr = None
        db_ptr_ptr = ffi.malloc(8)
        try:
            res = _sqlite3_open(database, db_ptr_ptr)
            if res != SQLITE_OK:
                raise Error(format_str("Failed to open database: {res}"))
            self.db_ptr = ffi.read_memory(db_ptr_ptr, ffi.c_void_p)
        finally:
            ffi.free(db_ptr_ptr)

    def cursor(self):
        if not self.db_ptr:
            raise Error("Connection closed")
        return Cursor(self.db_ptr, self)

    def execute(self, sql, parameters=None):
        c = self.cursor()
        return c.execute(sql, parameters)

    def commit(self):
        pass

    def close(self):
        if self.db_ptr:
            _sqlite3_close(self.db_ptr)
            self.db_ptr = None

def connect(database):
    return Connection(database)