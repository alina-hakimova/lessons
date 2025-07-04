from flask import Flask, request, jsonify
import psycopg2

app = Flask(__name__)

DB_HOST = "localhost"
DB_NAME = "notes_db"
DB_USER = "notes_user"
DB_PASS = "mystrongpassword"

def get_connection():
    return psycopg2.connect(
        host=DB_HOST,
        dbname=DB_NAME,
        user=DB_USER,
        password=DB_PASS
    )

@app.route('/notes', methods=['POST'])
def create_note():
    data = request.get_json()
    title = data.get('title')
    content = data.get('content')

    conn = get_connection()
    cur = conn.cursor()
    cur.execute(
        "INSERT INTO notes (title, content) VALUES (%s, %s) RETURNING id",
        (title, content)
    )
    new_id = cur.fetchone()[0]
    conn.commit()
    cur.close()
    conn.close()

    return jsonify({'id': new_id, 'title': title, 'content': content})

@app.route('/notes', methods=['GET'])
def get_notes():
    conn = get_connection()
    cur = conn.cursor()
    cur.execute("SELECT id, title, content, created_at FROM notes ORDER BY id")
    rows = cur.fetchall()
    cur.close()
    conn.close()

    notes = []
    for row in rows:
        notes.append({
            'id': row[0],
            'title': row[1],
            'content': row[2],
            'created_at': row[3].isoformat()
        })

    return jsonify(notes)

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5001)
