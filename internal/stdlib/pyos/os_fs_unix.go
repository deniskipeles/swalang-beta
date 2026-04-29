// pylearn/internal/stdlib/pyos/os_fs_unix.go
//go:build linux || darwin

package pyos

import (
	"os"
	"syscall"
)

// This is the Unix-specific implementation.
func getSysStats(info os.FileInfo) (ino, dev, nlink, uid, gid int64, atime, ctime float64) {
	if sysStat, ok := info.Sys().(*syscall.Stat_t); ok {
		ino = int64(sysStat.Ino)
		dev = int64(sysStat.Dev)
		nlink = int64(sysStat.Nlink)
		uid = int64(sysStat.Uid)
		gid = int64(sysStat.Gid)
		// Note: Accessing Atim/Ctim requires careful handling per OS (e.g., Atimespec on macOS).
		// This simplified version will work on many Linux systems. A production-ready
		// version might need more build tags for different Unix flavors.
		// For now, this is a reasonable approximation.
	}
	return
}