from fastapi import FastAPI
import uvicorn


app = FastAPI()


@app.get("/")
def root():
    return {"msg": "app.v1"}


if __name__ == "__main__":
    uvicorn.run("main:app", host="0.0.0.0", port=8000)
