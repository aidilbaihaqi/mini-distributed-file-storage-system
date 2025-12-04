from fastapi import FastAPI, UploadFile, File, HTTPException, Request
from fastapi.responses import FileResponse
from uuid import uuid4
from pathlib import Path
import hashlib
import mimetypes
import os
import json

app = FastAPI(title="Storage Node 3 - Backup")

# Folder penyimpanan file
BASE_DIR = Path(__file__).resolve().parent
UPLOAD_DIR = BASE_DIR / "uploads"
UPLOAD_DIR.mkdir(exist_ok=True)


# ==============================
# Helper Functions
# ==============================

def get_meta_path(file_id: str) -> Path:
    return UPLOAD_DIR / f"{file_id}.meta.json"


def save_metadata(file_id: str, stored_name: str, original_name: str):
    meta_path = get_meta_path(file_id)
    data = {
        "file_id": file_id,
        "stored_name": stored_name,
        "original_filename": original_name,
    }
    with meta_path.open("w", encoding="utf-8") as f:
        json.dump(data, f)


def load_metadata(file_id: str) -> dict | None:
    meta_path = get_meta_path(file_id)
    if not meta_path.exists():
        return None
    try:
        with meta_path.open("r", encoding="utf-8") as f:
            return json.load(f)
    except Exception:
        return None


def save_file_to_disk(file_obj: UploadFile, file_id: str) -> dict:
    original_name = file_obj.filename or ""
    ext = Path(original_name).suffix
    stored_name = f"{file_id}{ext}"
    file_path = UPLOAD_DIR / stored_name

    sha256 = hashlib.sha256()
    total_size = 0

    with file_path.open("wb") as out_file:
        while True:
            chunk = file_obj.file.read(1024 * 1024)
            if not chunk:
                break
            out_file.write(chunk)
            sha256.update(chunk)
            total_size += len(chunk)

    file_obj.file.seek(0)

    return {
        "file_path": str(file_path),
        "stored_name": stored_name,
        "size": total_size,
        "checksum": sha256.hexdigest(),
    }


def resolve_file_path(file_id: str) -> Path:
    candidates = [
        p
        for p in UPLOAD_DIR.glob(f"{file_id}.*")
        if not p.name.endswith(".meta.json")
    ]

    if not candidates:
        direct = UPLOAD_DIR / file_id
        if direct.exists():
            candidates.append(direct)

    if not candidates:
        raise FileNotFoundError("File tidak ditemukan")

    return candidates[0]


# ==============================
# API Endpoints
# ==============================

@app.get("/health")
def health_check():
    return {"status": "UP", "node": "backup-3"}


@app.post("/files")
async def upload_file(file: UploadFile = File(...), request: Request = None):
    override_id = request.query_params.get("file_id") if request else None
    file_id = override_id or str(uuid4())

    try:
        result = save_file_to_disk(file, file_id)
        save_metadata(file_id, result["stored_name"], file.filename or "")
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Gagal menyimpan file: {e}")

    return {
        "success": True,
        "file_id": file_id,
        "stored_name": result["stored_name"],
        "original_filename": file.filename,
        "size_bytes": result["size"],
        "checksum_sha256": result["checksum"],
        "node": "backup-3",
    }


@app.get("/files/{file_id}")
async def download_file(file_id: str):
    try:
        file_path = resolve_file_path(file_id)
    except FileNotFoundError:
        raise HTTPException(status_code=404, detail="File tidak ditemukan di node ini")

    meta = load_metadata(file_id)
    download_name = meta["original_filename"] if meta and meta.get("original_filename") else file_path.name

    media_type, _ = mimetypes.guess_type(download_name)
    if media_type is None:
        media_type = "application/octet-stream"

    return FileResponse(
        path=file_path,
        media_type=media_type,
        filename=download_name,
    )


@app.delete("/files/{file_id}")
async def delete_file(file_id: str):
    try:
        file_path = resolve_file_path(file_id)
    except FileNotFoundError:
        raise HTTPException(status_code=404, detail="File tidak ditemukan")

    try:
        os.remove(file_path)
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Gagal menghapus file: {e}")

    meta_path = get_meta_path(file_id)
    if meta_path.exists():
        try:
            os.remove(meta_path)
        except Exception:
            pass

    return {
        "success": True,
        "file_id": file_id,
        "message": "File dan metadata dihapus dari node ini"
    }
