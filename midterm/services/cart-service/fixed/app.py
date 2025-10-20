# services/cart-service/fixed/app.py
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
import httpx
import os
from typing import Dict, List, Optional
import asyncio
from circuit_breaker import CircuitBreaker
import time
from datetime import datetime

app = FastAPI(title="Shopping Cart Service - Fixed with Circuit Breaker")

# Product service URL
PRODUCT_SERVICE_URL = os.getenv("PRODUCT_SERVICE_URL", "http://product-service:8000")

# In-memory cart storage
carts: Dict[str, List[Dict]] = {}

# In-memory cache for product info (fallback mechanism)
product_cache: Dict[str, Dict] = {}
cache_timestamps: Dict[str, float] = {}
CACHE_TTL = 300  # 5 minutes

# Initialize Circuit Breaker
circuit_breaker = CircuitBreaker(
    failure_threshold=3,  # Open after 3 failures
    recovery_timeout=30,   # Try to recover after 30 seconds
    expected_exception=httpx.RequestError
)

class CartItem(BaseModel):
    product_id: str
    quantity: int

class Cart(BaseModel):
    user_id: str
    items: List[Dict]
    total_price: float
    degraded_mode: bool = False

async def fetch_product_with_circuit_breaker(product_id: str) -> Optional[Dict]:
    """Fetch product info with circuit breaker protection"""
    
    @circuit_breaker
    async def _fetch():
        async with httpx.AsyncClient(timeout=httpx.Timeout(5.0)) as client:
            response = await client.get(f"{PRODUCT_SERVICE_URL}/products/{product_id}")
            if response.status_code == 200:
                return response.json()
            raise httpx.RequestError(f"Product service returned {response.status_code}")
    
    try:
        product = await _fetch()
        # Update cache on successful fetch
        product_cache[product_id] = product
        cache_timestamps[product_id] = time.time()
        return product
    except Exception as e:
        # Circuit is open or request failed
        # Try to use cached data
        if product_id in product_cache:
            cache_age = time.time() - cache_timestamps.get(product_id, 0)
            return {
                **product_cache[product_id],
                "cached": True,
                "cache_age_seconds": int(cache_age)
            }
        return None

@app.get("/health")
@app.get("/fixed/health")
async def health_check():
    return {
        "status": "healthy",
        "mode": "protected",
        "circuit_breaker_state": circuit_breaker.state
    }

@app.post("/cart/{user_id}/add")
@app.post("/fixed/cart/{user_id}/add")
async def add_to_cart(user_id: str, item: CartItem):
    """Add item to cart - PROTECTED VERSION with Circuit Breaker"""
    
    if user_id not in carts:
        carts[user_id] = []
    
    # Try to fetch product with circuit breaker protection
    product = await fetch_product_with_circuit_breaker(item.product_id)
    
    if product is None:
        # Graceful degradation - allow adding to cart with limited info
        cart_item = {
            "product_id": item.product_id,
            "product_name": f"Product {item.product_id}",
            "quantity": item.quantity,
            "price": 0.0,  # Price unavailable
            "subtotal": 0.0,
            "status": "price_pending",
            "message": "Product service temporarily unavailable. Price will be updated later."
        }
    else:
        # Normal operation with product info
        if product.get("stock", 0) < item.quantity:
            raise HTTPException(status_code=400, detail="Insufficient stock")
        
        cart_item = {
            "product_id": item.product_id,
            "product_name": product["name"],
            "quantity": item.quantity,
            "price": product["price"],
            "subtotal": product["price"] * item.quantity,
            "status": "confirmed",
            "cached": product.get("cached", False)
        }
    
    # Update existing item or add new
    existing = next((x for x in carts[user_id] if x["product_id"] == item.product_id), None)
    if existing:
        existing["quantity"] += item.quantity
        if existing.get("price", 0) > 0:
            existing["subtotal"] = existing["price"] * existing["quantity"]
    else:
        carts[user_id].append(cart_item)
    
    return {
        "message": "Item added to cart",
        "cart_item": cart_item,
        "circuit_breaker_state": circuit_breaker.state
    }

@app.get("/cart/{user_id}")
@app.get("/fixed/cart/{user_id}")
async def get_cart(user_id: str):
    """Get user's cart - PROTECTED VERSION with graceful degradation"""
    
    if user_id not in carts or not carts[user_id]:
        return Cart(user_id=user_id, items=[], total_price=0.0, degraded_mode=False)
    
    updated_items = []
    total = 0.0
    degraded_mode = False
    
    for item in carts[user_id]:
        # Try to refresh product info, but don't fail if unavailable
        product = await fetch_product_with_circuit_breaker(item["product_id"])
        
        if product:
            # Update with fresh/cached data
            item["price"] = product["price"]
            item["subtotal"] = product["price"] * item["quantity"]
            item["product_name"] = product["name"]
            item["data_freshness"] = "cached" if product.get("cached") else "live"
            if product.get("cached"):
                item["cache_age"] = product.get("cache_age_seconds", 0)
                degraded_mode = True
        else:
            # Product service completely unavailable
            item["data_freshness"] = "unavailable"
            item["message"] = "Price information temporarily unavailable"
            degraded_mode = True
        
        updated_items.append(item)
        if item.get("price", 0) > 0:
            total += item["subtotal"]
    
    return Cart(
        user_id=user_id,
        items=updated_items,
        total_price=total,
        degraded_mode=degraded_mode
    )

@app.delete("/cart/{user_id}")
async def clear_cart(user_id: str):
    """Clear user's cart"""
    if user_id in carts:
        carts[user_id] = []
    return {"message": "Cart cleared"}

@app.get("/metrics")
@app.get("/fixed/metrics")
async def get_metrics():
    """Get service metrics including circuit breaker stats"""
    return {
        "total_carts": len(carts),
        "total_items": sum(len(items) for items in carts.values()),
        "mode": "protected",
        "circuit_breaker": {
            "state": circuit_breaker.state,
            "failure_count": circuit_breaker.failure_count,
            "last_failure_time": circuit_breaker.last_failure_time.isoformat() if circuit_breaker.last_failure_time else None,
            "success_count": circuit_breaker.success_count,
            "total_calls": circuit_breaker.success_count + circuit_breaker.failure_count
        },
        "cache": {
            "cached_products": len(product_cache),
            "cache_entries": list(product_cache.keys())
        }
    }

@app.post("/circuit-breaker/reset")
@app.post("/fixed/circuit-breaker/reset")
async def reset_circuit_breaker():
    """Manually reset the circuit breaker for demo purposes"""
    circuit_breaker.reset()
    return {
        "message": "Circuit breaker reset",
        "state": circuit_breaker.state
    }