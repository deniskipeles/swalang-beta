package interpreter

import (
	"os"
	"path/filepath"

	"github.com/deniskipeles/pylearn/internal/constants"
)

// GetStandardLibraryPath returns the path to the standard library
func GetStandardLibraryPath() string {
	if stdPath := os.Getenv(constants.SWALANG_STDLIB_PATH_ENV); stdPath != constants.EmptyString {
		return stdPath
	}

	var possiblePaths []string

	// 1. Production Layout: Resolve relative to the Swalang executable
	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)       // e.g., root-folder/bin
		rootDir := filepath.Dir(exeDir)       // e.g., root-folder
		
		// Target: root-folder/stdlib
		possiblePaths = append(possiblePaths, filepath.Join(rootDir, constants.STDLIB_DIR_NAME))
		// Target: root-folder/bin/stdlib (in case it's bundled directly next to the binary)
		possiblePaths = append(possiblePaths, filepath.Join(exeDir, constants.STDLIB_DIR_NAME))
	}

	// 2. Development Layout Fallbacks (Relative to CWD)
	possiblePaths = append(possiblePaths,
		constants.STDLIB_DIR_NAME,
		constants.STDLIB_LIB_FALLBACK_PATH,
		constants.STDLIB_INTERNAL_FALLBACK_PATH,
	)

	for _, path := range possiblePaths {
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			if absPath, err := filepath.Abs(path); err == nil {
				return absPath
			}
		}
	}

	return constants.EmptyString
}