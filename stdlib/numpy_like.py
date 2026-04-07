# pylearn/stdlib/numpy_like.py

"""
A comprehensive, NumPy-like array computing library for Pylearn, providing
a powerful N-dimensional array object and routines for high-performance
mathematical operations backed by OpenBLAS.
"""

import ffi
import math

# --- Helper to load OpenBLAS safely ---
def _load_openblas():
    try:
        return ffi.CDLL("libopenblas.so.0")
    except ffi.FFIError:
        try:
            return ffi.CDLL("libopenblas.so")
        except ffi.FFIError:
            print("Warning: OpenBLAS not found. Linear algebra functions will be slower.")
            return None

_lib = _load_openblas()


# --- FFI Function Definitions (only if library was loaded) ---
_cblas_ddot = None
_cblas_daxpy = None
_cblas_dscal = None
_cblas_dcopy = None
_cblas_dgemm = None

if _lib is not None:
    # Vector-Vector dot product
    _cblas_ddot = _lib.cblas_ddot(
        [ffi.c_int32, ffi.c_void_p, ffi.c_int32, ffi.c_void_p, ffi.c_int32],
        ffi.c_double
    )
    # Vector copy: Y = X
    _cblas_dcopy = _lib.cblas_dcopy(
        [ffi.c_int32, ffi.c_void_p, ffi.c_int32, ffi.c_void_p, ffi.c_int32], 
        None
    )
    # Vector-scalar multiply and add: Y = alpha*X + Y
    _cblas_daxpy = _lib.cblas_daxpy(
        [ffi.c_int32, ffi.c_double, ffi.c_void_p, ffi.c_int32, ffi.c_void_p, ffi.c_int32], 
        None
    )
    # Vector-scalar multiply: X = alpha*X
    _cblas_dscal = _lib.cblas_dscal(
        [ffi.c_int32, ffi.c_double, ffi.c_void_p, ffi.c_int32], 
        None
    )
    # Matrix-Matrix multiplication: C = alpha*A*B + beta*C
    _cblas_dgemm = _lib.cblas_dgemm(
        [ffi.c_int32, ffi.c_int32, ffi.c_int32, ffi.c_int32, ffi.c_int32, ffi.c_int32,
         ffi.c_double, ffi.c_void_p, ffi.c_int32, ffi.c_void_p, ffi.c_int32,
         ffi.c_double, ffi.c_void_p, ffi.c_int32],
        None
    )


class NDArray:
    """A multi-dimensional, homogeneous array of fixed-size items."""
    
    def __init__(self, data, shape=None, dtype=float, _internal_empty=False):
        self.dtype = dtype
        
        if _internal_empty:
            self.shape = shape
            self.size = 1
            for dim in shape:
                self.size = self.size * dim
            self.ndim = len(shape)
            self._pylearn_data = [0.0] * self.size
        else:
            flat_data = []
            def _flatten(item):
                if isinstance(item, (list, tuple)):
                    for sub_item in item:
                        _flatten(sub_item)
                else:
                    flat_data.append(self.dtype(item))
            _flatten(data)

            self.shape = shape
            if self.shape is None:
                def _infer(item):
                    if not isinstance(item, (list, tuple)):
                        return ()
                    if len(item) == 0:
                        return (0,)
                    return (len(item),) + _infer(item[0])
                self.shape = _infer(data)
            
            self.size = 1
            for dim in self.shape:
                self.size = self.size * dim

            if self.size != len(flat_data):
                raise ValueError("shape is invalid for input data size")
            
            self.ndim = len(self.shape)
            self._pylearn_data = flat_data

        self._c_double_size = ffi.c_double.Size()
        self._c_buffer_size = self.size * self._c_double_size
        self._c_pointer = ffi.malloc(self._c_buffer_size)
        if self._c_pointer is None:
            raise MemoryError("Failed to allocate memory")
        
        if not _internal_empty:
            i = 0
            while i < self.size:
                offset = i * self._c_double_size
                ffi.write_memory_with_offset(self._c_pointer, offset, ffi.c_double, self._pylearn_data[i])
                i = i + 1
        
        self.strides = []
        stride = 1
        i = len(self.shape) - 1
        while i >= 0:
            self.strides.insert(0, stride)
            stride = stride * self.shape[i]
            i = i - 1

    def __str__(self):
        return self._format_recursive(self.shape, 0)
    
    def __repr__(self):
        return format_str("array({self.__str__()})")

    def _format_recursive(self, shape, offset):
        if len(shape) == 1:
            items = []
            i = 0
            while i < shape[0]:
                items.append(str(self._pylearn_data[offset + i]))
                i = i + 1
            return "[" + ", ".join(items) + "]"
        
        result = "["
        step = self.strides[self.ndim - len(shape)]
        
        i = 0
        while i < shape[0]:
            if i > 0:
                result = result + ",\n" + (" " * (self.ndim - len(shape) + 1))
            
            result = result + self._format_recursive(shape[1:], offset + i * step)
            i = i + 1
        
        result = result + "]"
        return result

    def __getitem__(self, index):
        if not isinstance(index, int):
            raise TypeError("array indices must be integers")
        
        if index < 0:
            index = self.size + index
            
        if index < 0 or index >= self.size:
            raise IndexError("array index out of range")
            
        offset = index * self._c_double_size
        return ffi.read_memory_with_offset(self._c_pointer, offset, ffi.c_double)

    def __add__(self, other):
        if not isinstance(other, NDArray) or self.shape != other.shape:
            return NotImplemented
        if _lib is None:
            raise RuntimeError("OpenBLAS not available")

        result_array = self.copy()
        _cblas_daxpy(self.size, 1.0, other._c_pointer, 1, result_array._c_pointer, 1)
        return result_array

    def __mul__(self, scalar):
        if not isinstance(scalar, (int, float)):
            return NotImplemented
        if _lib is None:
            raise RuntimeError("OpenBLAS not available")
            
        result_array = self.copy()
        _cblas_dscal(self.size, float(scalar), result_array._c_pointer, 1)
        return result_array

    def __rmul__(self, scalar):
        return self.__mul__(scalar)

    def copy(self):
        return NDArray(self._pylearn_data, shape=self.shape)

    def free(self):
        if self._c_pointer is not None:
            ffi.free(self._c_pointer)
            self._c_pointer = None

# --- Factory Functions ---
def array(data, dtype=float):
    return NDArray(data, dtype=dtype)

def zeros(shape, dtype=float):
    if isinstance(shape, int):
        shape = (shape,)

    new_array = NDArray(None, shape=shape, dtype=dtype, _internal_empty=True)
    
    if new_array.size > 0:
        ffi.memset(new_array._c_pointer, 0, new_array._c_buffer_size)

    return new_array

def ones(shape, dtype=float):
    if isinstance(shape, int):
        shape = (shape,)
    size = 1
    for dim in shape:
        size = size * dim
    
    return NDArray([1.0] * size, shape=shape, dtype=dtype)

def arange(start, stop=None, step=1):
    if stop is None:
        stop = start
        start = 0
    data = []
    current = float(start)
    while current < float(stop):
        data.append(current)
        current = current + float(step)
    return NDArray(data)
    
# --- Linear Algebra ---
def dot(a, b):
    if not isinstance(a, NDArray) or not isinstance(b, NDArray):
        raise TypeError("Args must be NDArray")
    if _lib is None:
        raise RuntimeError("OpenBLAS not available")

    if a.ndim == 1 and b.ndim == 1:
        if a.size != b.size:
            raise ValueError("shapes are not aligned")
        return _cblas_ddot(a.size, a._c_pointer, 1, b._c_pointer, 1)

    if a.ndim == 2 and b.ndim == 2:
        if a.shape[1] != b.shape[0]:
            raise ValueError(format_str("shapes {a.shape} and {b.shape} not aligned"))
        
        # m, n, k = a.shape[0], a.shape[1], b.shape[1]
        m = a.shape[0]
        n = a.shape[1]
        k = b.shape[1]
        c = zeros((m, k))
        
        # CblasRowMajor=101, CblasNoTrans=111
        _cblas_dgemm(101, 111, 111, m, k, n, 1.0, a._c_pointer, n, b._c_pointer, k, 0.0, c._c_pointer, k)
        return c
        
    raise ValueError("dot() not implemented for these dimensions")

# --- Universal Functions (UFuncs) ---
def sin(x):
    if isinstance(x, NDArray):
        new_data = [math.sin(val) for val in x._pylearn_data]
        return NDArray(new_data, shape=x.shape)
    return math.sin(x)

def cos(x):
    if isinstance(x, NDArray):
        new_data = [math.cos(val) for val in x._pylearn_data]
        return NDArray(new_data, shape=x.shape)
    return math.cos(x)

def add(a, b):
    return a + b

def multiply(a, scalar):
    return a * scalar