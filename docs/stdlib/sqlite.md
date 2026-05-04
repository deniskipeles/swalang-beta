# `sqlite` Module Reference

The `sqlite` module provides a DB-API 2.0 compatible interface to the SQLite3 database engine.

## Functions

- `connect(database)`: Opens a connection to the SQLite database file. Returns a `Connection` object.

## Constants

- `SQLITE_OK`, `SQLITE_ROW`, `SQLITE_DONE`: Return codes.
- `SQLITE_INTEGER`, `SQLITE_FLOAT`, `SQLITE_TEXT`, `SQLITE_BLOB`, `SQLITE_NULL`: Data types.

## Classes

### `Connection`
- `cursor()`: Returns a new `Cursor` object.
- `execute(sql, parameters=None)`: Shorthand to create a cursor and execute SQL.
- `commit()`: (Placeholder) Currently a no-op as autocommit is standard in this wrapper.
- `close()`: Closes the database connection.

### `Cursor`
- `execute(sql, parameters=None)`: Executes a SQL statement. `parameters` should be a tuple.
- `fetchone()`: Returns the next row as a tuple, or `None` if no more rows are available.
- `fetchall()`: Returns a list of all remaining rows.
- `close()`: Finalizes any prepared statements and closes the cursor.
- `rowcount`: (Property) Number of rows modified by the last `execute`.
- `lastrowid`: (Property) The row ID of the last inserted row.
