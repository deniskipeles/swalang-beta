package sqlite

import (
	"database/sql"

	"github.com/deniskipeles/pylearn/internal/constants"
	"github.com/deniskipeles/pylearn/internal/object"
	_ "modernc.org/sqlite" // Register sqlite driver
)

// Pylearn: sqlite.connect(database_path)
func pySqliteConnect(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError(constants.TypeError, "connect() takes exactly 1 argument (database path)")
	}
	pathObj, ok := args[0].(*object.String)
	if !ok {
		return object.NewError(constants.TypeError, "database path must be a string")
	}
	path := pathObj.Value

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return object.NewError(constants.RuntimeError, "Failed to open database: %v", err)
	}

	// Ping to verify the connection is alive
	if err := db.Ping(); err != nil {
		db.Close()
		return object.NewError(constants.RuntimeError, "Failed to connect to database: %v", err)
	}

	return &SQLiteConnection{Db: db, Path: path}
}

// --- Module Initialization ---
func init() {
	env := object.NewEnvironment()
	
	// Expose the connect function
	env.Set("connect", &object.Builtin{
		Name: "sqlite.connect",
		Fn:   pySqliteConnect,
	})
	
	// You could also expose error classes here if you define them
	// env.Set("Error", ...)
	// env.Set("IntegrityError", ...)

	sqliteModule := &object.Module{
		Name: "sqlite",
		Path: "<builtin>",
		Env:  env,
	}
	object.RegisterNativeModule("sqlite", sqliteModule)
}