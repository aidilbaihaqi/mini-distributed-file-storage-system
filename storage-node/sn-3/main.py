from fastapi import FastAPI

app = FastAPI()

@app.get("/health")
def health():
    return {"status": "UP", "desc": "#3 Backup Node"}
