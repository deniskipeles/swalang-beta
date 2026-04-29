import sqlite
import os

print("===========================================")
print("🧪 Testing SQLite FFI Wrapper")
print("===========================================")

db_path = "test_ffi.db"
# Clean up before testing
if db_path in os.listdir():
    os.remove(db_path)

# Connect to the database (creates it if it doesn't exist)
conn = sqlite.connect(db_path)
print("✅ Database opened successfully.")

cur = conn.cursor()

# 1. Create a table
cur.execute("CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT, age REAL, is_admin INTEGER)")
print("✅ Table 'users' created.")

# 2. Insert data using Parameter Binding (?)
cur.execute("INSERT INTO users (name, age, is_admin) VALUES (?, ?, ?)", ("Zeus", 29.5, 1))
cur.execute("INSERT INTO users (name, age, is_admin) VALUES (?, ?, ?)", ("Apollo", 35.2, 0))
cur.execute("INSERT INTO users (name, age, is_admin) VALUES (?, ?, ?)", ("Athena", None, 1))

print(format_str("✅ Rows inserted. Last Row ID: {cur.lastrowid}"))
assert cur.lastrowid == 3

# 3. Query the data
cur.execute("SELECT * FROM users WHERE is_admin = ?", (1,))

# Fetchone test
first_admin = cur.fetchone()
print(format_str("👉 First admin row fetched: {first_admin}"))
assert first_admin[1] == "Zeus"

# Fetchall test
remaining_admins = cur.fetchall()
print(format_str("👉 Remaining admin rows fetched: {remaining_admins}"))
assert len(remaining_admins) == 1
assert remaining_admins[0][1] == "Athena"
assert remaining_admins[0][2] is None

# 4. Cleanup
conn.close()
print("✅ Connection closed.")

if db_path in os.listdir():
    os.remove(db_path)
    print("✅ Cleanup complete.")

print("\n🎉 SQLite FFI Wrapper tests passed successfully!")