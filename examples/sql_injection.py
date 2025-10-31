# Example vulnerable Python code - SQL Injection
# CWE-89: Improper Neutralization of Special Elements used in an SQL Command

import sqlite3

def get_user_by_name(username):
    conn = sqlite3.connect('users.db')
    cursor = conn.cursor()

    # VULNERABLE: Direct string concatenation in SQL query
    # Allows SQL injection attacks
    query = f"SELECT * FROM users WHERE username = '{username}'"

    cursor.execute(query)
    result = cursor.fetchone()

    conn.close()
    return result

# Attacker could use: username = "admin' OR '1'='1"
user = get_user_by_name(input("Enter username: "))
print(user)
