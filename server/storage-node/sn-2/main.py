from fastapi import FastAPI, UploadFile, File, HTTPException, Request
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

app = FastAPI(title="Storage Node")

# Konfigurasi dari environment variables
NODE_ID = os.getenv("NODE_ID", "node-1")
NODE_PORT = int(os.getenv("NODE_PORT", "8001"))
NAMING_SERVICE_URL = os.getenv("NAMING_SERVICE_URL", "http://localhost:8080")

# Parse ALL_NODES dari environment atau gunakan default
def parse_all_nodes():
    env_nodes = os.getenv("ALL_NODES", "")
    if env_nodes:
        nodes = {}
        for pair in env_nodes.split(","):
            if "=" in pair:
                node_id, url = pair.split("=", 1)
                nodes[node_id.strip()] = url.strip()
        return nodes
    # Default untuk local development
    return {
        "node-1": "http://localhost:8001",
        "node-2": "http://localhost:8002",
        "node-3": "http://localhost:8003",
    }

ALL_NODES = parse_all_nodes()

# Folder penyimpanan file
BASE_DIR = Path(__file__).resolve().parent
UPLOAD_DIR = BASE_DIR / "uploads"
UPLOAD_DIR.mkdir(exist_ok=True)


def get_other_nodes():
    """Get list of other nodes (exclude self)"""
    return {k: v for k, v in ALL_NODES.items() if k != NODE_ID}


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
        p for p in UPLOAD_DIR.glob(f"{file_id}.*")
        if not p.name.endswith(".meta.json")
    ]
    if not candidates:
        direct = UPLOAD_DIR / file_id
        if direct.exists():
            candidates.append(direct)
    if not candidates:
        raise FileNotFoundError("File tidak ditemukan")
    return candidates[0]


async def replicate_to_node(node_id: str, node_url: str, file_path: Path, 
                            file_id: str, original_filename: str) -> dict:
    """Replicate file to another node"""
    try:
        async with httpx.AsyncClient(timeout=30.0) as client:
            with open(file_path, "rb") as f:
                files = {"file": (original_filename, f, "application/octet-stream")}
                response = await client.post(
                    f"{node_url}/files?file_id={file_id}&is_replica=true",
                    files=files
                )
                if response.status_code == 200:
                    return {"node_id": node_id, "success": True}
                else:
                    return {"node_id": node_id, "success": False, "error": f"HTTP {response.status_code}"}
    except Exception as e:
        return {"node_id": node_id, "success": False, "error": str(e)}


async def replicate_to_all_nodes(file_path: Path, file_id: str, original_filename: str) -> tuple:
    """Replicate to all other nodes, return (successful_nodes, failed_nodes)"""
    other_nodes = get_other_nodes()
    
    tasks = [
        replicate_to_node(node_id, node_url, file_path, file_id, original_filename)
        for node_id, node_url in other_nodes.items()
    ]
    
    results = await asyncio.gather(*tasks)
    
    successful = [r["node_id"] for r in results if r.get("success")]
    failed = [r["node_id"] for r in results if not r.get("success")]
    
    return successful, failed


async def register_to_naming_service(file_key: str, original_filename: str,
                                     size_bytes: int, checksum: str,
                                     successful_nodes: List[str], failed_nodes: List[str]):
    """Register file metadata to naming service"""
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
                print(f"[{NODE_ID}] Registered {file_key} to naming service")
                
                for node_id in successful_nodes:
                    try:
                        await client.post(
                            f"{NAMING_SERVICE_URL}/files/register-location",
                            json={"file_key": file_key, "node_id": node_id}
                        )
                    except:
                        pass
            else:
                print(f"[{NODE_ID}] Failed to register {file_key}: HTTP {response.status_code}")
    except Exception as e:
        print(f"[{NODE_ID}] Failed to register to naming service: {e}")


@app.get("/health")
def health_check():
    return {"status": "UP", "node_id": NODE_ID}


@app.post("/files")
async def upload_file(file: UploadFile = File(...), request: Request = None):
    is_replica = request.query_params.get("is_replica", "false").lower() == "true" if request else False
    override_id = request.query_params.get("file_id") if request else None
    
    file_id = override_id or str(uuid4())

    try:
        result = save_file_to_disk(file, file_id)
        save_metadata(file_id, result["stored_name"], file.filename or "")
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Gagal menyimpan file: {e}")

    successful_nodes = []
    failed_nodes = []
    
    if not is_replica:
        file_path = Path(result["file_path"])
        successful_nodes, failed_nodes = await replicate_to_all_nodes(
            file_path, file_id, file.filename or ""
        )
        
        print(f"[{NODE_ID}] Upload {file_id}: replicated to {successful_nodes}, failed: {failed_nodes}")
        
        await register_to_naming_service(
            file_id,
            file.filename or "",
            result["size"],
            result["checksum"],
            successful_nodes,
            failed_nodes
        )

    return {
        "success": True,
        "file_id": file_id,
        "stored_name": result["stored_name"],
        "original_filename": file.filename,
        "size_bytes": result["size"],
        "checksum_sha256": result["checksum"],
        "node_id": NODE_ID,
        "is_replica": is_replica,
        "replication": {
            "successful": successful_nodes,
            "failed": failed_nodes,
        } if not is_replica else None
    }


@app.get("/files/{file_id}")
async def download_file(file_id: str):
    try:
        file_path = resolve_file_path(file_id)
    except FileNotFoundError:
        raise HTTPException(status_code=404, detail="File tidak ditemukan")

    meta = load_metadata(file_id)
    download_name = meta["original_filename"] if meta and meta.get("original_filename") else file_path.name

    media_type, _ = mimetypes.guess_type(download_name)
    if media_type is None:
        media_type = "application/octet-stream"

    return FileResponse(path=file_path, media_type=media_type, filename=download_name)


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
        except:
            pass

    return {"success": True, "file_id": file_id, "node_id": NODE_ID}
