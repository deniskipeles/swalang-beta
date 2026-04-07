// pylearn/internal/stdlib/pyos/os_fs.go
package pyos

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"

	"github.com/deniskipeles/pylearn/internal/constants"
	"github.com/deniskipeles/pylearn/internal/object"
)

// ... (pyListDirFn, pyPathExistsFn, and other functions remain the same) ...
func pyListDirFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	path := "."
	if len(args) > 1 {
		return object.NewError(constants.TypeError, "os.listdir() takes at most 1 argument (%d given)", len(args))
	}
	if len(args) == 1 {
		switch pathArg := args[0].(type) {
		case *object.String:
			path = pathArg.Value
		case *object.Bytes:
			path = string(pathArg.Value)
		default:
			return object.NewError(constants.TypeError, "os.listdir() argument must be str or bytes, not %s", args[0].Type())
		}
	}
	fileInfos, err := ioutil.ReadDir(path)
	if err != nil {
		return object.NewError(constants.OSError, "[Errno %d] %s: '%s'", getErrno(err), err.Error(), path)
	}
	elements := make([]object.Object, len(fileInfos))
	for i, fi := range fileInfos {
		elements[i] = &object.String{Value: fi.Name()}
	}
	return &object.List{Elements: elements}
}
func pyPathExistsFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError(constants.TypeError, "os.path.exists() takes exactly 1 argument (%d given)", len(args))
	}
	pathStr, ok := args[0].(*object.String)
	if !ok {
		return object.NewError(constants.TypeError, "os.path.exists() argument must be a string, not %s", args[0].Type())
	}
	_, err := os.Stat(pathStr.Value)
	if err == nil {
		return object.TRUE
	}
	if os.IsNotExist(err) {
		return object.FALSE
	}
	return object.FALSE
}
func pyPathIsDirFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError(constants.TypeError, "os.path.isdir() takes exactly 1 argument (%d given)", len(args))
	}
	pathStr, ok := args[0].(*object.String)
	if !ok {
		return object.NewError(constants.TypeError, "os.path.isdir() argument must be a string, not %s", args[0].Type())
	}
	info, err := os.Stat(pathStr.Value)
	if err != nil {
		return object.FALSE
	}
	return object.NativeBoolToBooleanObject(info.IsDir())
}
func pyPathIsFileFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError(constants.TypeError, "os.path.isfile() takes exactly 1 argument (%d given)", len(args))
	}
	pathStr, ok := args[0].(*object.String)
	if !ok {
		return object.NewError(constants.TypeError, "os.path.isfile() argument must be a string, not %s", args[0].Type())
	}
	info, err := os.Stat(pathStr.Value)
	if err != nil {
		return object.FALSE
	}
	return object.NativeBoolToBooleanObject(!info.IsDir())
}
func pyPathJoinFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) == 0 {
		return &object.String{Value: ""}
	}
	pathElements := make([]string, len(args))
	for i, arg := range args {
		pathStr, ok := arg.(*object.String)
		if !ok {
			return object.NewError(constants.TypeError, "os.path.join() arguments must be strings, got %s at pos %d", arg.Type(), i)
		}
		pathElements[i] = pathStr.Value
	}
	return &object.String{Value: filepath.Join(pathElements...)}
}
func pyPathDirnameFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError(constants.TypeError, "os.path.dirname() takes exactly 1 argument (%d given)", len(args))
	}
	pathStr, ok := args[0].(*object.String)
	if !ok {
		return object.NewError(constants.TypeError, "os.path.dirname() argument must be a string, not %s", args[0].Type())
	}
	return &object.String{Value: filepath.Dir(pathStr.Value)}
}
func pyPathBasenameFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError(constants.TypeError, "os.path.basename() takes exactly 1 argument (%d given)", len(args))
	}
	pathStr, ok := args[0].(*object.String)
	if !ok {
		return object.NewError(constants.TypeError, "os.path.basename() argument must be a string, not %s", args[0].Type())
	}
	return &object.String{Value: filepath.Base(pathStr.Value)}
}
func pyPathAbsPathFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError(constants.TypeError, "os.path.abspath() takes exactly 1 argument (%d given)", len(args))
	}
	pathStr, ok := args[0].(*object.String)
	if !ok {
		return object.NewError(constants.TypeError, "os.path.abspath() argument must be a string, not %s", args[0].Type())
	}
	absPath, err := filepath.Abs(pathStr.Value)
	if err != nil {
		return object.NewError(constants.OSError, "could not determine absolute path for '%s': %v", pathStr.Value, err)
	}
	return &object.String{Value: absPath}
}
func pyPathRelPathFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) < 1 || len(args) > 2 {
		return object.NewError(constants.TypeError, "os.path.relpath() takes 1 or 2 arguments (%d given)", len(args))
	}
	pathStr, ok := args[0].(*object.String)
	if !ok {
		return object.NewError(constants.TypeError, "os.path.relpath() argument 'path' must be a string, not %s", args[0].Type())
	}
	startPath := "."
	if len(args) == 2 {
		if args[1] != object.NULL {
			startStr, okStart := args[1].(*object.String)
			if !okStart {
				return object.NewError(constants.TypeError, "os.path.relpath() argument 'start' must be a string or None, not %s", args[1].Type())
			}
			startPath = startStr.Value
		}
	}
	relPath, err := filepath.Rel(startPath, pathStr.Value)
	if err != nil {
		return object.NewError(constants.ValueError, "could not compute relative path from '%s' to '%s': %v", startPath, pathStr.Value, err)
	}
	return &object.String{Value: relPath}
}
func pyOsWalkFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError(constants.TypeError, "os.walk() takes exactly 1 argument (top) (%d given)", len(args))
	}
	topPathStr, ok := args[0].(*object.String)
	if !ok {
		return object.NewError(constants.TypeError, "os.walk() argument 'top' must be a string, not %s", args[0].Type())
	}
	top := topPathStr.Value
	resultList := &object.List{Elements: []object.Object{}}
	err := filepath.Walk(top, func(path string, info os.FileInfo, errWalk error) error {
		if errWalk != nil {
			fmt.Printf("Warning: error during os.walk at '%s': %v. Skipping.\n", path, errWalk)
			if info != nil && info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if !info.IsDir() {
			return nil
		}
		dirpathStr := path
		dirnames := []object.Object{}
		filenames := []object.Object{}
		entries, errReadDir := ioutil.ReadDir(path)
		if errReadDir != nil {
			fmt.Printf("Warning: error reading directory '%s' during os.walk: %v. Skipping entries.\n", path, errReadDir)
		} else {
			for _, entry := range entries {
				if entry.IsDir() {
					dirnames = append(dirnames, &object.String{Value: entry.Name()})
				} else {
					filenames = append(filenames, &object.String{Value: entry.Name()})
				}
			}
		}
		yieldTuple := &object.Tuple{Elements: []object.Object{
			&object.String{Value: dirpathStr},
			&object.List{Elements: dirnames},
			&object.List{Elements: filenames},
		}}
		resultList.Elements = append(resultList.Elements, yieldTuple)
		return nil
	})
	if err != nil {
		return object.NewError(constants.OSError, "error during os.walk traversal: %v", err)
	}
	return resultList
}
func pyOsMkdirFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) < 1 || len(args) > 2 {
		return object.NewError(constants.TypeError, "os.mkdir() takes 1 or 2 arguments (%d given)", len(args))
	}
	pathStr, ok := args[0].(*object.String)
	if !ok {
		return object.NewError(constants.TypeError, "os.mkdir() argument 'path' must be a string, not %s", args[0].Type())
	}
	mode := os.FileMode(0777)
	if len(args) == 2 {
		if args[1] != object.NULL {
			modeInt, okMode := args[1].(*object.Integer)
			if !okMode {
				return object.NewError(constants.TypeError, "os.mkdir() argument 'mode' must be an integer or None, not %s", args[1].Type())
			}
			mode = os.FileMode(modeInt.Value)
		}
	}
	err := os.Mkdir(pathStr.Value, mode)
	if err != nil {
		return object.NewError(constants.OSError, "[Errno %d] %s: '%s'", getErrno(err), err.Error(), pathStr.Value)
	}
	return object.NULL
}
func pyOsRemoveFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError(constants.TypeError, "os.remove() takes exactly 1 argument (%d given)", len(args))
	}
	pathStr, ok := args[0].(*object.String)
	if !ok {
		return object.NewError(constants.TypeError, "os.remove() argument 'path' must be a string, not %s", args[0].Type())
	}
	err := os.Remove(pathStr.Value)
	if err != nil {
		return object.NewError(constants.OSError, "[Errno %d] %s: '%s'", getErrno(err), err.Error(), pathStr.Value)
	}
	return object.NULL
}
func pyOsRenameFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 2 {
		return object.NewError(constants.TypeError, "os.rename() takes exactly 2 arguments (src, dst) (%d given)", len(args))
	}
	srcStr, okSrc := args[0].(*object.String)
	if !okSrc {
		return object.NewError(constants.TypeError, "os.rename() argument 'src' must be a string, not %s", args[0].Type())
	}
	dstStr, okDst := args[1].(*object.String)
	if !okDst {
		return object.NewError(constants.TypeError, "os.rename() argument 'dst' must be a string, not %s", args[1].Type())
	}
	err := os.Rename(srcStr.Value, dstStr.Value)
	if err != nil {
		if linkErr, ok := err.(*os.LinkError); ok {
			return object.NewError(constants.OSError, "[Errno %d] %s: '%s' -> '%s'", getErrno(err), linkErr.Error(), linkErr.Old, linkErr.New)
		}
		return object.NewError(constants.OSError, "[Errno %d] %s", getErrno(err), err.Error())
	}
	return object.NULL
}

// --- THIS IS THE MODIFIED FUNCTION ---
// pyOsStatFn now calls the platform-agnostic helper.
func pyOsStatFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError(constants.TypeError, "os.stat() takes exactly 1 argument (%d given)", len(args))
	}
	pathStr, ok := args[0].(*object.String)
	if !ok {
		return object.NewError(constants.TypeError, "os.stat() argument 'path' must be a string, not %s", args[0].Type())
	}

	info, err := os.Stat(pathStr.Value)
	if err != nil {
		return object.NewError(constants.OSError, "[Errno %d] %s: '%s'", getErrno(err), err.Error(), pathStr.Value)
	}

	statData := make(map[object.HashKey]object.DictPair)
	setStatVal := func(keyName string, val object.Object) {
		keyObj := &object.String{Value: keyName}
		hashKey, _ := keyObj.HashKey()
		statData[hashKey] = object.DictPair{Key: keyObj, Value: val}
	}

	setStatVal("st_mode", &object.Integer{Value: int64(info.Mode())})
	setStatVal("st_size", &object.Integer{Value: info.Size()})
	setStatVal("st_mtime", &object.Float{Value: float64(info.ModTime().UnixNano()) / 1e9})

	// Call the platform-independent helper to get the rest of the stats.
	stIno, stDev, stNlink, stUid, stGid, stAtime, stCtime := getPlatformSpecificStats(info)

	setStatVal("st_ino", &object.Integer{Value: stIno})
	setStatVal("st_dev", &object.Integer{Value: stDev})
	setStatVal("st_nlink", &object.Integer{Value: stNlink})
	setStatVal("st_uid", &object.Integer{Value: stUid})
	setStatVal("st_gid", &object.Integer{Value: stGid})
	setStatVal("st_atime", &object.Float{Value: stAtime})
	setStatVal("st_ctime", &object.Float{Value: stCtime})

	return &object.Dict{Pairs: statData}
}
// --- END OF MODIFICATION ---


var (
	ListDir      = &object.Builtin{Name: "os.listdir", Fn: pyListDirFn}
	OsWalk       = &object.Builtin{Name: "os.walk", Fn: pyOsWalkFn}
	OsMkdir      = &object.Builtin{Name: "os.mkdir", Fn: pyOsMkdirFn}
	OsRemove     = &object.Builtin{Name: "os.remove", Fn: pyOsRemoveFn}
	OsRename     = &object.Builtin{Name: "os.rename", Fn: pyOsRenameFn}
	OsStat       = &object.Builtin{Name: "os.stat", Fn: pyOsStatFn}
	PathExists   = &object.Builtin{Name: "os.path.exists", Fn: pyPathExistsFn}
	PathIsDir    = &object.Builtin{Name: "os.path.isdir", Fn: pyPathIsDirFn}
	PathIsFile   = &object.Builtin{Name: "os.path.isfile", Fn: pyPathIsFileFn}
	PathJoin     = &object.Builtin{Name: "os.path.join", Fn: pyPathJoinFn}
	PathDirname  = &object.Builtin{Name: "os.path.dirname", Fn: pyPathDirnameFn}
	PathBasename = &object.Builtin{Name: "os.path.basename", Fn: pyPathBasenameFn}
	PathAbsPath  = &object.Builtin{Name: "os.path.abspath", Fn: pyPathAbsPathFn}
	PathRelPath  = &object.Builtin{Name: "os.path.relpath", Fn: pyPathRelPathFn}
)

func getErrno(err error) int {
	if e, ok := err.(*os.PathError); ok {
		if errno, ok := e.Err.(syscall.Errno); ok {
			return int(errno)
		}
	}
	if e, ok := err.(*os.LinkError); ok {
		if errno, ok := e.Err.(syscall.Errno); ok {
			return int(errno)
		}
	}
	if errno, ok := err.(syscall.Errno); ok {
		return int(errno)
	}
	if os.IsNotExist(err) { return int(syscall.ENOENT) }
	if os.IsPermission(err) { return int(syscall.EACCES) }
	if os.IsExist(err) { return int(syscall.EEXIST) }
	return 1
}