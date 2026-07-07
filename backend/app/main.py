from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from app.api import incidents


app = FastAPI(
    title="AegisOps AI",
    description="Autonomous AI Operations Center",
    version="1.0.0"
)

app.include_router(
    incidents.router
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)


@app.get("/")
def home():
    return {
        "project": "AegisOps AI",
        "status": "online",
        "message": "Autonomous IT Operations Center running"
    }


@app.get("/health")
def health():
    return {
        "status": "healthy"
    }