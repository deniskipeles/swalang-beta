package pyos

import (
	"os"

	"github.com/deniskipeles/pylearn/internal/object"
)

func pyGetcwd(ctx object.ExecutionContext, args ...object.Object) object.Object {
	dir, err := os.Getwd()
	if err != nil {
		return object.NewError("OSError", "Failed to get current directory: %v", err)
	}
	return &object.String{Value: dir}
}

func pyGetenv(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) < 1 || len(args) > 2 {
		return object.NewError("TypeError", "getenv() takes 1 or 2 arguments")
	}
	keyStr, ok := args[0].(*object.String)
	if !ok {
		return object.NewError("TypeError", "getenv() key must be a string")
	}
	
	val, exists := os.LookupEnv(keyStr.Value)
	if !exists {
		if len(args) == 2 {
			return args[1] // Return default
		}
		return object.NULL
	}
	return &object.String{Value: val}
}

func pyListdir(ctx object.ExecutionContext, args ...object.Object) object.Object {
	path := "."
	if len(args) == 1 {
		pathStr, ok := args[0].(*object.String)
		if !ok {
			return object.NewError("TypeError", "listdir() path must be a string")
		}
		path = pathStr.Value
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return object.NewError("OSError", "Failed to read directory: %v", err)
	}

	list := &object.List{Elements: make([]object.Object, len(entries))}
	for i, entry := range entries {
		list.Elements[i] = &object.String{Value: entry.Name()}
	}
	return list
}

func pyRemove(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError("TypeError", "remove() takes exactly 1 argument")
	}
	pathStr, ok := args[0].(*object.String)
	if !ok {
		return object.NewError("TypeError", "remove() path must be a string")
	}
	
	err := os.Remove(pathStr.Value)
	if err != nil {
		return object.NewError("OSError", "Failed to remove file: %v", err)
	}
	return object.NULL
}

func init() {
	env := object.NewEnvironment()
	env.Set("getcwd", &object.Builtin{Name: "os.getcwd", Fn: pyGetcwd})
	env.Set("getenv", &object.Builtin{Name: "os.getenv", Fn: pyGetenv})
	env.Set("listdir", &object.Builtin{Name: "os.listdir", Fn: pyListdir})
	env.Set("remove", &object.Builtin{Name: "os.remove", Fn: pyRemove})

	module := &object.Module{
		Name: "os",
		Path: "<builtin_os>",
		Env:  env,
	}
	object.RegisterNativeModule("os", module)
}