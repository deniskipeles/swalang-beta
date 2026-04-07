//go:build linux || darwin || windows
// pylearn/internal/stdlib/ffi/ffi_helpers.h
#ifndef FFI_HELPERS_H
#define FFI_HELPERS_H

#include <ffi.h>
#include <stdlib.h>
#include <string.h>

// A helper function in C to create the libffi closure.
ffi_closure* new_closure(void** code_loc);

// The C shim function that bridges libffi's required signature to CGo's generated signature.
void c_callback_shim(ffi_cif* cif, void* ret, void** args, void* user_data);

// A C wrapper around ffi_call to simplify the signature for CGo.
void pylearn_ffi_call_shim(ffi_cif* cif, void (*fn)(void), void* rvalue, void* avalue);

#endif // FFI_HELPERS_H