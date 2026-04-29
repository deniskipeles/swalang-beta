// pylearn/internal/stdlib/pyos/os_fs_stat.go
package pyos

import "os"

// This file defines the platform-independent function signature.
// The implementation will be provided by os_fs_unix.go and os_fs_windows.go.

// getPlatformSpecificStats populates the platform-dependent fields of a stat result.
func getPlatformSpecificStats(info os.FileInfo) (ino, dev, nlink, uid, gid int64, atime, ctime float64) {
	return getSysStats(info)
}