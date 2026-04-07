//go:build linux || darwin || windows
// pylearn/internal/stdlib/ffi/c_definitions.go
package ffi3

/*
#include "ffi_helpers.h"

// Forward declaration for the CGo-generated Go handler.
extern void goCallbackHandler(ffi_cif* cif, void* ret, void* args, void* user_data);

// Definition for c_callback_shim.
void c_callback_shim(ffi_cif* cif, void* ret, void** args, void* user_data) {
    goCallbackHandler(cif, ret, (void*)args, user_data);
}

// Definition for new_closure.
ffi_closure* new_closure(void** code_loc) {
    return (ffi_closure*)ffi_closure_alloc(sizeof(ffi_closure), code_loc);
}

// Definition for our ffi_call wrapper. It now accepts a void* and casts it internally.
void pylearn_ffi_call_shim(ffi_cif* cif, void (*fn)(void), void* rvalue, void* avalue) {
    ffi_call(cif, fn, rvalue, (void**)avalue);
}
*/
import "C"

// This file is used to centralize C definitions for the ffi package
// to prevent "multiple definition" linker errors. The Go part is empty.