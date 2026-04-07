//go:build windows
// internal/stdlib/ffi3/cgo_windows.go
package ffi3

/*
#cgo pkg-config: libffi
// This tells CGo where to find the ffi.h header for the Windows target.
#cgo CFLAGS: -IC:/msys64/mingw64/include

// This tells CGo where to find the libffi.a library for linking for the Windows target.
#cgo LDFLAGS: -LC:/msys64/mingw64/lib -lffi
*/
import "C"
