// pylearn/internal/stdlib/pyos/os_fs_windows.go
//go:build windows

package pyos

import (
	"os"
)

// This is the Windows-specific implementation. It provides dummy values
// as these stats are not available in the same way on Windows.
func getSysStats(info os.FileInfo) (ino, dev, nlink, uid, gid int64, atime, ctime float64) {
	// Windows does not have the same concept of inodes, links, uid, or gid.
	// We return 0 as a sensible default, which mimics how some compatibility
	// layers (like Python itself) handle this.
	return 0, 0, 1, 0, 0, 0.0, 0.0
}