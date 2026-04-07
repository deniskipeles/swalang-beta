package sqlite

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/deniskipeles/pylearn/internal/constants"
	"github.com/deniskipeles/pylearn/internal/object"
	_ "modernc.org/sqlite" // Register sqlite driver
)

const (
	SQLITE_CONNECTION_OBJ object.ObjectType = constants.SQLITE_CONNECTION_OBJ_TYPE
	SQLITE_CURSOR_OBJ     object.ObjectType = constants.SQLITE_CURSOR_OBJ_TYPE
)

// --- Connection Object ---
type SQLiteConnection struct {
	mu     sync.Mutex
	Db     *sql.DB
	Path   string
	closed bool
}

func (c *SQLiteConnection) Type() object.ObjectType { return SQLITE_CONNECTION_OBJ }
func (c *SQLiteConnection) Inspect() string {
	status := constants.SQLITE_CONN_STATUS_OPEN
	if c.closed {
		status = constants.SQLITE_CONN_STATUS_CLOSED
	}
	return fmt.Sprintf(constants.SQLITE_CONN_INSPECT_FORMAT, c, c.Path, status)
}
var _ object.Object = (*SQLiteConnection)(nil)
var _ object.AttributeGetter = (*SQLiteConnection)(nil)

// --- Cursor Object ---
type SQLiteCursor struct {
	mu         sync.Mutex
	Conn       *SQLiteConnection // Reference to parent connection
	Rows       *sql.Rows         // Holds the result of a query
	closed     bool
	lastInsertId int64
	rowsAffected int64
}

func (c *SQLiteCursor) Type() object.ObjectType { return SQLITE_CURSOR_OBJ }
func (c *SQLiteCursor) Inspect() string {
	status := constants.SQLITE_CURSOR_STATUS_ACTIVE
	if c.closed {
		status = constants.SQLITE_CURSOR_STATUS_CLOSED
	}
	return fmt.Sprintf(constants.SQLITE_CURSOR_INSPECT_FORMAT, c, status)
}
var _ object.Object = (*SQLiteCursor)(nil)
var _ object.AttributeGetter = (*SQLiteCursor)(nil)

// --- Connection Methods ---

// Pylearn: conn.cursor() -> Cursor
func (c *SQLiteConnection) PyCursor(ctx object.ExecutionContext, args ...object.Object) object.Object {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return object.NewError(constants.RuntimeError, constants.SQLITE_CLOSED_CONN_ERROR)
	}
	// Cursors share the same underlying connection pool from sql.DB
	return &SQLiteCursor{Conn: c}
}

// Pylearn: conn.commit()
func (c *SQLiteConnection) PyCommit(ctx object.ExecutionContext, args ...object.Object) object.Object {
	// The `modernc.org/sqlite` driver doesn't support transactions in the traditional
	// Begin/Commit/Rollback way over the database/sql interface. Commits are implicit.
	// We provide this method for API compatibility with Python's sqlite3. It's a no-op.
	return object.NULL
}

// Pylearn: conn.close()
func (c *SQLiteConnection) PyClose(ctx object.ExecutionContext, args ...object.Object) object.Object {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return object.NULL // Already closed
	}
	err := c.Db.Close()
	if err != nil {
		return object.NewError(constants.RuntimeError, constants.SQLITE_CLOSE_CONN_ERROR, err)
	}
	c.closed = true
	return object.NULL
}

// --- Cursor Methods ---

// Pylearn: cur.execute(sql, parameters=None)
func (c *SQLiteCursor) PyExecute(ctx object.ExecutionContext, args ...object.Object) object.Object {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return object.NewError(constants.RuntimeError, constants.SQLITE_CLOSED_CURSOR_ERROR)
	}
	if len(args) < 2 || len(args) > 3 {
		return object.NewError(constants.TypeError, constants.SQLITE_EXECUTE_ARG_COUNT_ERROR)
	}

	sqlQueryObj, ok := args[1].(*object.String)
	if !ok {
		return object.NewError(constants.TypeError, constants.SQLITE_EXECUTE_SQL_TYPE_ERROR)
	}
	sqlQuery := sqlQueryObj.Value

	var params []interface{}
	if len(args) == 3 {
		// Convert Pylearn tuple/list of params to Go `[]interface{}`
		paramsTuple, ok := args[2].(*object.Tuple)
		if !ok {
			return object.NewError(constants.TypeError, constants.SQLITE_EXECUTE_PARAMS_TYPE_ERROR)
		}
		params = make([]interface{}, len(paramsTuple.Elements))
		for i, elem := range paramsTuple.Elements {
			// Convert Pylearn objects to basic Go types for the driver
			switch v := elem.(type) {
			case *object.String:
				params[i] = v.Value
			case *object.Integer:
				params[i] = v.Value
			case *object.Float:
				params[i] = v.Value
			case *object.Boolean:
				params[i] = v.Value
			case *object.Bytes:
				params[i] = v.Value
			case *object.Null:
				params[i] = nil
			default:
				return object.NewError(constants.TypeError, constants.SQLITE_EXECUTE_UNSUPPORTED_PARAM_TYPE, elem.Type())
			}
		}
	}

	// If a previous query was run, close its result set.
	if c.Rows != nil {
		c.Rows.Close()
		c.Rows = nil
	}

	// Check if it's a SELECT or an action query
	if len(sqlQuery) >= 6 && (strings.ToUpper(sqlQuery[:6]) == constants.SQLITE_SELECT_KEYWORD_UPPER || strings.ToLower(sqlQuery[:6]) == constants.SQLITE_SELECT_KEYWORD_LOWER) {
		rows, err := c.Conn.Db.Query(sqlQuery, params...)
		if err != nil {
			return object.NewError(constants.RuntimeError, constants.SQLITE_EXECUTE_QUERY_ERROR, err)
		}
		c.Rows = rows
		c.lastInsertId = 0
		c.rowsAffected = 0
	} else {
		// For INSERT, UPDATE, DELETE, CREATE, etc.
		result, err := c.Conn.Db.Exec(sqlQuery, params...)
		if err != nil {
			return object.NewError(constants.RuntimeError, constants.SQLITE_EXECUTE_STATEMENT_ERROR, err)
		}
		c.lastInsertId, _ = result.LastInsertId()
		c.rowsAffected, _ = result.RowsAffected()
	}

	return c // Return self for chaining
}

// Pylearn: cur.fetchone() -> object.Tuple or None
func (c *SQLiteCursor) PyFetchOne(ctx object.ExecutionContext, args ...object.Object) object.Object {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.Rows == nil {
		return object.NewError(constants.RuntimeError, constants.SQLITE_FETCHONE_NO_QUERY_ERROR)
	}

	if !c.Rows.Next() {
		// No more rows or an error occurred during iteration
		if err := c.Rows.Err(); err != nil {
			return object.NewError(constants.RuntimeError, constants.SQLITE_FETCHONE_ROW_ITER_ERROR, err)
		}
		c.Rows.Close() // Exhausted, so close it.
		c.Rows = nil
		return object.NULL // No more rows
	}

	// Scan the row into Pylearn objects
	row, err := scanRow(c.Rows)
	if err != nil {
		return object.NewError(constants.RuntimeError, constants.SQLITE_FETCHONE_SCAN_ROW_ERROR, err)
	}
	return row
}

// Pylearn: cur.fetchall() -> object.List of Tuples
func (c *SQLiteCursor) PyFetchAll(ctx object.ExecutionContext, args ...object.Object) object.Object {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.Rows == nil {
		return object.NewError(constants.RuntimeError, constants.SQLITE_FETCHALL_NO_QUERY_ERROR)
	}
	defer func() {
		c.Rows.Close()
		c.Rows = nil
	}()

	var results []object.Object
	for c.Rows.Next() {
		row, err := scanRow(c.Rows)
		if err != nil {
			return object.NewError(constants.RuntimeError, constants.SQLITE_FETCHALL_SCAN_ROW_ERROR, err)
		}
		results = append(results, row)
	}

	if err := c.Rows.Err(); err != nil {
		return object.NewError(constants.RuntimeError, constants.SQLITE_FETCHALL_ROW_ITER_ERROR, err)
	}
	return &object.List{Elements: results}
}

// Pylearn: cur.close()
func (c *SQLiteCursor) PyClose(ctx object.ExecutionContext, args ...object.Object) object.Object {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return object.NULL
	}
	if c.Rows != nil {
		c.Rows.Close()
		c.Rows = nil
	}
	c.closed = true
	return object.NULL
}

// --- Attribute Getters ---

func (c *SQLiteConnection) GetObjectAttribute(ctx object.ExecutionContext, name string) (object.Object, bool) {
	makeMethod := func(goFn object.BuiltinFunction) *object.Builtin {
		return &object.Builtin{Fn: func(callCtx object.ExecutionContext, scriptArgs ...object.Object) object.Object {
			return goFn(callCtx, append([]object.Object{c}, scriptArgs...)...)
		}}
	}
	switch name {
	case constants.SQLITE_CURSOR_METHOD_NAME:
		return makeMethod(c.PyCursor), true
	case constants.SQLITE_COMMIT_METHOD_NAME:
		return makeMethod(c.PyCommit), true
	case constants.SQLITE_CLOSE_METHOD_NAME:
		return makeMethod(c.PyClose), true
	}
	return nil, false
}

func (c *SQLiteCursor) GetObjectAttribute(ctx object.ExecutionContext, name string) (object.Object, bool) {
	makeMethod := func(goFn object.BuiltinFunction) *object.Builtin {
		return &object.Builtin{Fn: func(callCtx object.ExecutionContext, scriptArgs ...object.Object) object.Object {
			return goFn(callCtx, append([]object.Object{c}, scriptArgs...)...)
		}}
	}
	switch name {
	case constants.SQLITE_EXECUTE_METHOD_NAME:
		return makeMethod(c.PyExecute), true
	case constants.SQLITE_FETCHONE_METHOD_NAME:
		return makeMethod(c.PyFetchOne), true
	case constants.SQLITE_FETCHALL_METHOD_NAME:
		return makeMethod(c.PyFetchAll), true
	case constants.SQLITE_CLOSE_METHOD_NAME:
		return makeMethod(c.PyClose), true
	case constants.SQLITE_LASTROWID_ATTR_NAME:
		return &object.Integer{Value: c.lastInsertId}, true
	case constants.SQLITE_ROWCOUNT_ATTR_NAME:
		return &object.Integer{Value: c.rowsAffected}, true
	}
	return nil, false
}

// --- Helper Functions ---

// scanRow reads the columns of the current row and converts them to Pylearn objects.
func scanRow(rows *sql.Rows) (*object.Tuple, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	
	values := make([]interface{}, len(columns))
	scanArgs := make([]interface{}, len(columns))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	err = rows.Scan(scanArgs...)
	if err != nil {
		return nil, err
	}

	tupleElements := make([]object.Object, len(columns))
	for i, col := range values {
		if col == nil {
			tupleElements[i] = object.NULL
			continue
		}
		switch val := col.(type) {
		case int64:
			tupleElements[i] = &object.Integer{Value: val}
		case float64:
			tupleElements[i] = &object.Float{Value: val}
		case bool:
			tupleElements[i] = object.NativeBoolToBooleanObject(val)
		case []byte:
			tupleElements[i] = &object.Bytes{Value: val}
		case string:
			tupleElements[i] = &object.String{Value: val}
		case time.Time: // SQLite dates often come back as strings or time.Time
			tupleElements[i] = &object.String{Value: val.Format(constants.SQLITE_TIME_FORMAT)}
		default:
			return nil, fmt.Errorf(constants.SQLITE_UNSUPPORTED_DB_TYPE, col)
		}
	}

	return &object.Tuple{Elements: tupleElements}, nil
}