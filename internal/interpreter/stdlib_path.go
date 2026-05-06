package interpreter

import (
	"os"
	"path/filepath"
)

// GetStandardLibraryPath returns the path to the standard library
func GetStandardLibraryPath() string {
	if stdPath := os.Getenv("PYLEARN_STDLIB_PATH"); stdPath != "" {
		return stdPath
	}

	var possiblePaths[]string

	// 1. Production Layout: Resolve relative to the Swalang executable
	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)       // e.g., root-folder/bin
		rootDir := filepath.Dir(exeDir)       // e.g., root-folder
		
		// Target: root-folder/stdlib
		possiblePaths = append(possiblePaths, filepath.Join(rootDir, "stdlib"))
		// Target: root-folder/bin/stdlib (in case it's bundled directly next to the binary)
		possiblePaths = append(possiblePaths, filepath.Join(exeDir, "stdlib"))
	}

	// 2. Development Layout Fallbacks (Relative to CWD)
	possiblePaths = append(possiblePaths,
		"stdlib",
		"lib/stdlib",
		"internal/stdlib",
	)

	for _, path := range possiblePaths {
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			if absPath, err := filepath.Abs(path); err == nil {
				return absPath
			}
		}
	}

	return ""
}