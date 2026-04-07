package pyos

import (
	"github.com/deniskipeles/pylearn/internal/object"
)

// pathModule will hold the 'os.path' submodule
var pathModule *object.Module

func initOsPathModule() {
	pathEnv := object.NewEnvironment()
	pathEnv.Set("exists", PathExists)     // From os_fs.go
	pathEnv.Set("isdir", PathIsDir)       // From os_fs.go
	pathEnv.Set("isfile", PathIsFile)     // From os_fs.go
	pathEnv.Set("join", PathJoin)         // From os_fs.go
	pathEnv.Set("dirname", PathDirname)   // From os_fs.go
	pathEnv.Set("basename", PathBasename) // From os_fs.go
	pathEnv.Set("abspath", PathAbsPath)   // From os_fs.go
	pathEnv.Set("relpath", PathRelPath)   // From os_fs.go
	// Add other os.path functions like split,splitext, etc. here

	pathModule = &object.Module{
		Name: "path", // Name of the submodule
		Path: "<builtin>.os.path",
		Env:  pathEnv,
	}
}

func init() {
	initOsPathModule() // Initialize the path submodule

	env := object.NewEnvironment()

	// os module level functions
	env.Set("getenv", GetEnv)   // From os_env.go
	env.Set("listdir", ListDir) // From os_fs.go
	env.Set("getcwd", GetCWD)   // From os_proc.go
	env.Set("walk", OsWalk)     // From os_fs.go
	env.Set("mkdir", OsMkdir)   // New
	env.Set("remove", OsRemove) // New
	env.Set("unlink", OsRemove) // Alias for remove
	env.Set("rename", OsRename) // New
	env.Set("stat", OsStat)     // New

	// Add the 'path' submodule to the 'os' module's environment
	env.Set("path", pathModule)

	osModule := &object.Module{
		Name: "os",
		Path: "<builtin>",
		Env:  env,
	}
	object.RegisterNativeModule("os", osModule)
}


