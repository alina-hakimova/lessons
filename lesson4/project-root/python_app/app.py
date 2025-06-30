from flask import Flask
import os

app = Flask(__name__)

@app.route('/')
def hello():
    return 'Hello from Python!'

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=int(os.environ.get("PYTHON_PORT", 8001)))
