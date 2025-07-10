from flask import Flask, request, jsonify
import psycopg2
from pymongo import MongoClient
from datetime import datetime
import os
from dotenv import load_dotenv

load_dotenv()

app = Flask(__name__)

# Подключение к PostgreSQL
def get_pg_connection():
    return psycopg2.connect(
        host=os.getenv("POSTGRES_HOST"),
        database=os.getenv("POSTGRES_DB"),
        user=os.getenv("POSTGRES_USER"),
        password=os.getenv("POSTGRES_PASSWORD")
    )

# Подключение к MongoDB
mongo_client = MongoClient(os.getenv("MONGO_URI"))
mongo_db = mongo_client[os.getenv("MONGO_DB")]
tasks_collection = mongo_db.tasks

@app.route('/notes', methods=['POST'])
def create_note():
    data = request.get_json()
    with get_pg_connection() as conn:
        with conn.cursor() as cur:
            cur.execute(
                "INSERT INTO notes (title, content) VALUES (%s, %s) RETURNING id, created_at",
                (data['title'], data['content'])
            )
            row = cur.fetchone()
    return jsonify({
        "id": row[0],
        "title": data['title'],
        "content": data['content'],
        "created_at": row[1].isoformat()
    })

@app.route('/notes', methods=['GET'])
def get_notes():
    with get_pg_connection() as conn:
        with conn.cursor() as cur:
            cur.execute("SELECT id, title, content, created_at FROM notes ORDER BY id")
            rows = cur.fetchall()
    return jsonify([
        {
            "id": r[0],
            "title": r[1],
            "content": r[2],
            "created_at": r[3].isoformat()
        } for r in rows
    ])

@app.route('/tasks', methods=['POST'])
def create_task():
    data = request.get_json()
    task = {
        "title": data['title'],
        "description": data['description'],
        "created_at": datetime.utcnow()
    }
    result = tasks_collection.insert_one(task)
    task['_id'] = str(result.inserted_id)
    return jsonify(task)

@app.route('/tasks', methods=['GET'])
def get_tasks():
    tasks = []
    for doc in tasks_collection.find():
        tasks.append({
            "id": str(doc["_id"]),
            "title": doc["title"],
            "description": doc["description"],
            "created_at": doc["created_at"].isoformat()
        })
    return jsonify(tasks)

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5001)
