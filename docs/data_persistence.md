# Data Persistence and JSON Handling

Most real-world applications need to store and retrieve data. Swalang provides robust support for relational databases via **SQLite** and high-performance JSON processing.

## 1. Using SQLite

The `sqlite` module provides a DB-API 2.0 compatible interface for SQLite databases, making it familiar to anyone who has used Python's `sqlite3`.

### Connecting and Creating Tables

```python
import sqlite

# Connect to a database file (creates it if it doesn't exist)
conn = sqlite.connect("app_data.db")
cursor = conn.cursor()

# Create a table
cursor.execute("""
    CREATE TABLE IF NOT EXISTS projects (
        id INTEGER PRIMARY KEY,
        name TEXT NOT NULL,
        status TEXT
    )
""")
conn.commit()
```

### Inserting and Querying Data

Always use parameterized queries (`?` placeholders) to prevent SQL injection.

```python
# Inserting data
cursor.execute("INSERT INTO projects (name, status) VALUES (?, ?)", ("Swalang IDE", "In Progress"))
cursor.execute("INSERT INTO projects (name, status) VALUES (?, ?)", ("Native Plugin", "Completed"))
conn.commit()

# Querying data
cursor.execute("SELECT * FROM projects WHERE status = ?", ("Completed",))

# Fetch one result
project = cursor.fetchone()
if project:
    print(format_str("ID: {project[0]}, Name: {project[1]}"))

# Fetch all results
cursor.execute("SELECT name FROM projects")
all_names = cursor.fetchall()
for row in all_names:
    print(format_str("Project: {row[0]}"))

conn.close()
```

## 2. Working with JSON

Swalang offers two JSON libraries:
- `cjson`: A balanced and easy-to-use wrapper around the cJSON library.
- `yyjson`: A high-performance library optimized for speed and large datasets.

### Using `cjson`

```python
import cjson

# Encoding a dictionary to JSON string
user_data = {"id": 1, "name": "Alice", "roles": ["admin", "dev"]}
json_str = cjson.encode(user_data)
print(json_str)

# Decoding JSON string to Swalang object
decoded = cjson.decode(json_str)
print(format_str("User Name: {decoded['name']}"))
```

### High-Performance JSON with `yyjson`

For performance-critical tasks, `yyjson` is recommended.

```python
import yyjson

# Parse a JSON string
doc = yyjson.read('{"version": 1.0, "data": [10, 20, 30]}')
root = doc.root()

# Access data
version = root.get("version")
print(format_str("Version: {version}"))

# Iterate over an array
data_arr = root.get("data")
for i in range(data_arr.size()):
    print(data_arr.get_index(i))

doc.free()
```

## 3. Best Practices for Data Handling

1.  **Transactions**: Always wrap multiple related `execute()` calls in a transaction and call `conn.commit()` to ensure data integrity.
2.  **Closing Connections**: Use `conn.close()` when you are finished with the database to release file locks.
3.  **JSON Validation**: When receiving JSON from a web request, always wrap the decoding call in a `try...except` block to handle malformed input.
4.  **Schema Migrations**: For larger projects, keep a version table in your database to track and apply schema changes as your application evolves.
