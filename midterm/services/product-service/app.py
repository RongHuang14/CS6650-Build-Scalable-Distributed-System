# services/product-service/app.py
from fastapi import FastAPI, HTTPException, Query
from pydantic import BaseModel
import time
import os

app = FastAPI(title="Product Service")

# Simulated product database
PRODUCTS = {
    "1": {"id": "1", "name": "Laptop", "price": 999.99, "stock": 10},
    "2": {"id": "2", "name": "Mouse", "price": 29.99, "stock": 100},
    "3": {"id": "3", "name": "Keyboard", "price": 79.99, "stock": 50},
    "4": {"id": "4", "name": "Monitor", "price": 299.99, "stock": 20},
    "5": {"id": "5", "name": "Headphones", "price": 149.99, "stock": 30}
}

# Failure simulation: simple on/off with optional latency
FAILURE_MODE = os.getenv("FAILURE_MODE", "false").lower() == "true"
LATENCY_MS = int(os.getenv("LATENCY_MS", "0"))

# Basic metrics
REQUEST_COUNT: int = 0
SUCCESS_COUNT: int = 0
FAILURE_COUNT: int = 0

class Product(BaseModel):
    id: str
    name: str
    price: float
    stock: int

@app.get("/health")
async def health_check():
    """Health check endpoint"""
    if FAILURE_MODE:
        raise HTTPException(status_code=503, detail={
            "status": "unhealthy",
            "failure_mode": True,
            "latency_ms": LATENCY_MS,
        })
    return {"status": "healthy", "latency_ms": LATENCY_MS}

@app.get("/products/{product_id}")
async def get_product(product_id: str):
    """Get product details"""
    global REQUEST_COUNT, SUCCESS_COUNT, FAILURE_COUNT
    REQUEST_COUNT += 1

    # Simulate latency
    if LATENCY_MS > 0:
        time.sleep(LATENCY_MS / 1000)
    
    if FAILURE_MODE:
        FAILURE_COUNT += 1
        raise HTTPException(status_code=503, detail="Service temporarily unavailable")
    
    # Normal operation
    if product_id not in PRODUCTS:
        raise HTTPException(status_code=404, detail="Product not found")
    
    SUCCESS_COUNT += 1
    return PRODUCTS[product_id]

@app.get("/products")
async def list_products():
    """List all products"""
    global REQUEST_COUNT, SUCCESS_COUNT, FAILURE_COUNT
    REQUEST_COUNT += 1

    if LATENCY_MS > 0:
        time.sleep(LATENCY_MS / 1000)

    if FAILURE_MODE:
        FAILURE_COUNT += 1
        raise HTTPException(status_code=503, detail="Service temporarily unavailable")

    SUCCESS_COUNT += 1
    return list(PRODUCTS.values())

@app.post("/fail/on")
@app.get("/fail/on")
async def fail_on():
    global FAILURE_MODE
    FAILURE_MODE = True
    return {"message": "Failure enabled"}

@app.post("/fail/off")
@app.get("/fail/off")
async def fail_off():
    global FAILURE_MODE
    FAILURE_MODE = False
    return {"message": "Failure disabled"}

# Compatibility endpoints expected by ALB listener rules
@app.post("/crash")
@app.get("/crash")
async def crash():
    """Enable failure mode (alias for /fail/on)"""
    global FAILURE_MODE
    FAILURE_MODE = True
    return {"message": "Failure enabled"}

@app.post("/recover")
@app.get("/recover")
async def recover():
    """Disable failure mode (alias for /fail/off)"""
    global FAILURE_MODE
    FAILURE_MODE = False
    return {"message": "Failure disabled"}

@app.get("/metrics")
async def get_metrics():
    """Get service metrics"""
    return {
        "total_products": len(PRODUCTS),
        "failure_mode": FAILURE_MODE,
        "latency_ms": LATENCY_MS,
        "requests": REQUEST_COUNT,
        "successes": SUCCESS_COUNT,
        "failures": FAILURE_COUNT,
        "products": list(PRODUCTS.keys())
    }

# --- Control endpoints for clearer demos ---

@app.post("/latency")
@app.get("/latency")
async def set_latency(ms: int = Query(..., ge=0, le=60000)):
    global LATENCY_MS
    LATENCY_MS = ms
    return {"message": "Latency updated", "latency_ms": LATENCY_MS}

@app.post("/reset")
@app.get("/reset")
async def reset_all():
    global FAILURE_MODE, LATENCY_MS
    global REQUEST_COUNT, SUCCESS_COUNT, FAILURE_COUNT
    FAILURE_MODE = False
    LATENCY_MS = 0
    REQUEST_COUNT = 0
    SUCCESS_COUNT = 0
    FAILURE_COUNT = 0
    return {"message": "All settings reset"}