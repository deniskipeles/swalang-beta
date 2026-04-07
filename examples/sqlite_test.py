import sqlite

# Connect to a database (creates if not exists)
conn = sqlite.connect("./examples/users.db")

# Create a cursor object
cur = conn.cursor()

# Execute SQL
cur.execute("CREATE TABLE IF NOT EXISTS users (name TEXT, score INTEGER)")
# cur.execute("INSERT INTO users VALUES (?, ?)", ("Alice", 100))
# cur.execute("INSERT INTO users VALUES (?, ?)", ("Bob", 95))
# conn.commit()

# Fetch results
cur.execute("SELECT * FROM users WHERE score > ?", (90,))
# user = cur.fetchone() # Fetches ('Alice', 100)
# print(f("User: {user[0]}, Score: {user[1]}"))
users = cur.fetchall() # Fetches ('Alice', 100)
print(users)
for user in users:
    print(format_str("User: {user[0]}, Score: {user[1]}"))


# Close the connection
conn.close()