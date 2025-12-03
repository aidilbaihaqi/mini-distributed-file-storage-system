from fastapi import FastAPI, UploadFile, File, HTTPException
from fastapi.responses import FileResponse
from uuid import uuid4
from pathlib import Path
import hashlib
import mimetypes
import os
import json
import httpx
import asyncio
from typing import List

app = FastAPI(title="Storage Node 1 - Main")

# Daftar replica nodes
REPLICA_NODES = [
    "http://localhost:8002",  # sn-2
    "http://localhost:8003",  # sn-3
]

# Naming service URL
NAMING_SERVICE_URL = "http://localhost:8080"
NODE_ID = "node-1"

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
    ext = Path(original_name).suffix  # contoh: ".png", ".docx"
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
    """
    Cari file 'data' untuk file_id, mengabaikan file metadata (.meta.json).
    """
    # kandidat yang BUKAN metadata
    candidates = [
        p
        for p in UPLOAD_DIR.glob(f"{file_id}.*")
        if not p.name.endswith(".meta.json")
    ]

    # kalau tidak ada, fallback ke nama tanpa ekstensi
    if not candidates:
        direct = UPLOAD_DIR / file_id
        if direct.exists():
            candidates.append(direct)

    if not candidates:
        raise FileNotFoundError("File tidak ditemukan")

    # ambil yang pertama (sekarang pasti bukan .meta.json)
    return candidates[0]


async def replicate_to_node(node_url: str, file_path: Path, file_id: str, original_filename: str) -> dict:
    """
    Replikasi file ke node lain
    """
    try:
        async with httpx.AsyncClient(timeout=30.0) as client:
            with open(file_path, "rb") as f:
                files = {"file": (original_filename, f, "application/octet-stream")}
                # Gunakan file_id yang sama untuk konsistensi
                response = await client.post(
                    f"{node_url}/files",
                    files=files
                )
                
                if response.status_code == 200:
                    return {
                        "node": node_url,
                        "success": True,
                        "response": response.json()
                    }
                else:
                    return {
                        "node": node_url,
                        "success": False,
                        "error": f"HTTP {response.status_code}"
                    }
    except Exception as e:
        return {
            "node": node_url,
            "success": False,
            "error": str(e)
        }


async def replicate_to_all_nodes(file_path: Path, file_id: str, original_filename: str) -> List[dict]:
    """
    Replikasi file ke semua replica nodes secara parallel
    """
    tasks = [
        replicate_to_node(node_url, file_path, file_id, original_filename)
        for node_url in REPLICA_NODES
    ]
    results = await asyncio.gather(*tasks)
    
    # Tambahkan node identifier
    for i, result in enumerate(results):
        if i == 0:
            result["node"] = "node-2"
        elif i == 1:
            result["node"] = "node-3"
    
    return list(results)


async def register_file_to_naming_service(file_key: str, original_filename: str, 
                                          size_bytes: int, checksum: str, 
                                          failed_nodes: List[str]):
    """
    Register file metadata ke naming service
    """
    try:
        async with httpx.AsyncClient(timeout=5.0) as client:
            payload = {
                "file_key": file_key,
                "original_filename": original_filename,
                "size_bytes": size_bytes,
                "checksum_sha256": checksum,
                "node_id": NODE_ID,
                "failed_nodes": failed_nodes
            }
            
            response = await client.post(
                f"{NAMING_SERVICE_URL}/files/register",
                json=payload
            )
            
            if response.status_code == 200:
                print(f"✅ Registered {file_key} to naming service")
            else:
                print(f"⚠️ Failed to register {file_key}: HTTP {response.status_code}")
    except Exception as e:
        print(f"⚠️ Failed to register {file_key} to naming service: {e}")


# ==============================
# API Endpoints
# ==============================

@app.get("/health")
def health_check():
    return {"status": "UP", "node": "main"}


@app.post("/files")
async def upload_file(file: UploadFile = File(...)):
    file_id = str(uuid4())

    try:
        result = save_file_to_disk(file, file_id)
        save_metadata(file_id, result["stored_name"], file.filename or "")
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Gagal menyimpan file: {e}")

    # Replikasi ke node lain secara asynchronous
    file_path = Path(result["file_path"])
    replication_results = await replicate_to_all_nodes(
        file_path, 
        file_id, 
        file.filename or ""
    )

    # Hitung berapa node yang berhasil
    successful_replicas = [r for r in replication_results if r["success"]]
    failed_replicas = [r for r in replication_results if not r["success"]]

    # Register ke naming service
    failed_node_ids = []
    if "node-2" not in [r.get("node") for r in successful_replicas if r.get("success")]:
        failed_node_ids.append("node-2")
    if "node-3" not in [r.get("node") for r in successful_replicas if r.get("success")]:
        failed_node_ids.append("node-3")

    await register_file_to_naming_service(
        file_id,
        file.filename or "",
        result["size"],
        result["checksum"],
        failed_node_ids
    )

    return {
        "success": True,
        "file_id": file_id,
        "stored_name": result["stored_name"],
        "original_filename": file.filename,
        "size_bytes": result["size"],
        "checksum_sha256": result["checksum"],
        "node": "main",
        "replication": {
            "total_nodes": len(REPLICA_NODES),
            "successful": len(successful_replicas),
            "failed": len(failed_replicas),
            "details": replication_results
        }
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
    # hapus file data
    try:
        file_path = resolve_file_path(file_id)
    except FileNotFoundError:
        raise HTTPException(status_code=404, detail="File tidak ditemukan")

    try:
        os.remove(file_path)
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Gagal menghapus file: {e}")

    # hapus metadata
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

