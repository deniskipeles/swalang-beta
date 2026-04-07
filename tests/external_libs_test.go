package tests

import (
	"fmt"
	"os"
	"runtime"
	"testing"

	"github.com/deniskipeles/pylearn/internal/testhelpers"
)

func getLibPath(target, lib, baseName string) string {
	ext := ".so"
	prefix := "lib"
	platform := "x86_64-linux"

	if runtime.GOOS == "windows" {
		ext = ".dll"
		prefix = ""
		platform = "x86_64-windows-gnu"
	}

	return fmt.Sprintf("bin/%s/%s/%s%s%s", platform, lib, prefix, baseName, ext)
}

func TestExternalLibraries(t *testing.T) {
	// Skip if libraries are not built
	libPath := getLibPath("", "mongoose", "mongoose")
	if _, err := os.Stat(libPath); os.IsNotExist(err) {
		t.Skip("External libraries not built, skipping TestExternalLibraries")
	}

	t.Run("Mongoose Load", func(t *testing.T) {
		path := getLibPath("", "mongoose", "mongoose")
		input := fmt.Sprintf(`
import ffi
lib_path = %q
try:
    lib = ffi.CDLL(lib_path)
    print("Mongoose loaded")
    version_fn = lib.mg_version([], ffi.c_char_p)
    v = version_fn()
    print(f"Mongoose version: {v}")
except Exception as e:
    print(f"Error: {e}")
    raise e
`, path)
		testhelpers.Eval(t, input)
	})

	t.Run("YYJSON Load", func(t *testing.T) {
		path := getLibPath("", "yyjson", "yyjson")
		input := fmt.Sprintf(`
import ffi
lib_path = %q
try:
    lib = ffi.CDLL(lib_path)
    print("YYJSON loaded")
    version_fn = lib.yyjson_get_version([], ffi.c_uint32)
    v = version_fn()
    print(f"YYJSON version: {v}")
except Exception as e:
    print(f"Error: {e}")
    raise e
`, path)
		testhelpers.Eval(t, input)
	})

	t.Run("PCRE2 Load", func(t *testing.T) {
		path := getLibPath("", "pcre2", "pcre2-8")
		input := fmt.Sprintf(`
import ffi
lib_path = %q
try:
    lib = ffi.CDLL(lib_path)
    print("PCRE2 loaded")
except Exception as e:
    print(f"Error: {e}")
    raise e
`, path)
		testhelpers.Eval(t, input)
	})

	t.Run("MbedTLS Load", func(t *testing.T) {
		path := getLibPath("", "mbedtls", "mbedcrypto")
		input := fmt.Sprintf(`
import ffi
lib_path = %q
try:
    lib = ffi.CDLL(lib_path)
    print("MbedTLS Crypto loaded")
except Exception as e:
    print(f"Error: {e}")
    raise e
`, path)
		testhelpers.Eval(t, input)
	})

	t.Run("Libuv Load", func(t *testing.T) {
		path := getLibPath("", "libuv", "uv")
		if runtime.GOOS != "windows" {
			path = getLibPath("", "libuv", "uv") // should be libuv.so
		}

		input := fmt.Sprintf(`
import ffi
lib_path = %q
try:
    lib = ffi.CDLL(lib_path)
    print("Libuv loaded")
    version_fn = lib.uv_version_string([], ffi.c_char_p)
    v = version_fn()
    print(f"Libuv version: {v}")
except Exception as e:
    print(f"Error: {e}")
    raise e
`, path)
		testhelpers.Eval(t, input)
	})
}
