from flask import Flask
import os

app = Flask(__name__)

@app.route('/')
def hello():
    secret = os.getenv('SECRET_MESSAGE', 'no secret provided')
    return f"Secret message: {secret}"

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5000)

