# services/cart-service/vulnerable/app.py
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
import httpx
import os
from typing import Dict, List
import asyncio

app = FastAPI(title="Shopping Cart Service - Vulnerable")

# Product service URL
PRODUCT_SERVICE_URL = os.getenv("PRODUCT_SERVICE_URL", "http://product-service:8000")

# In-memory cart storage (user_id -> cart items)
carts: Dict[str, List[Dict]] = {}

class CartItem(BaseModel):
    product_id: str
    quantity: int

class Cart(BaseModel):
    user_id: str
    items: List[Dict]
    total_price: float

@app.get("/health")
@app.get("/vulnerable/health")
async def health_check():
    return {"status": "healthy", "mode": "vulnerable"}

@app.post("/cart/{user_id}/add")
@app.post("/vulnerable/cart/{user_id}/add")
async def add_to_cart(user_id: str, item: CartItem):
    """Add item to cart - VULNERABLE VERSION"""
    
    # Initialize cart if doesn't exist
    if user_id not in carts:
        carts[user_id] = []
    
    # Always try to fetch product info from product service
    # This is vulnerable - no circuit breaker, no timeout handling
    try:
        async with httpx.AsyncClient() as client:
            # No timeout set - can hang indefinitely
            response = await client.get(
                f"{PRODUCT_SERVICE_URL}/products/{item.product_id}"
            )
            
            if response.status_code != 200:
                raise HTTPException(
                    status_code=response.status_code,
                    detail=f"Product service error: {response.text}"
                )
            
            product = response.json()
            
    except httpx.RequestError as e:
        # Propagate the error - no graceful handling
        raise HTTPException(
            status_code=503,
            detail=f"Failed to connect to product service: {str(e)}"
        )
    
    # Check stock
    if product["stock"] < item.quantity:
        raise HTTPException(status_code=400, detail="Insufficient stock")
    
    # Add to cart
    cart_item = {
        "product_id": item.product_id,
        "product_name": product["name"],
        "quantity": item.quantity,
        "price": product["price"],
        "subtotal": product["price"] * item.quantity
    }
    
    # Update existing item or add new
    existing = next((x for x in carts[user_id] if x["product_id"] == item.product_id), None)
    if existing:
        existing["quantity"] += item.quantity
        existing["subtotal"] = existing["price"] * existing["quantity"]
    else:
        carts[user_id].append(cart_item)
    
    return {"message": "Item added to cart", "cart_item": cart_item}

@app.get("/cart/{user_id}")
@app.get("/vulnerable/cart/{user_id}")
async def get_cart(user_id: str):
    """Get user's cart with fresh product info - VULNERABLE VERSION"""
    
    if user_id not in carts or not carts[user_id]:
        return Cart(user_id=user_id, items=[], total_price=0.0)
    
    # Try to refresh all product info - can cause cascading failures
    updated_items = []
    total = 0.0
    
    async with httpx.AsyncClient() as client:
        for item in carts[user_id]:
            try:
                # Each call can fail and cause the entire cart to fail
                response = await client.get(
                    f"{PRODUCT_SERVICE_URL}/products/{item['product_id']}"
                )
                
                if response.status_code == 200:
                    product = response.json()
                    item["price"] = product["price"]
                    item["subtotal"] = product["price"] * item["quantity"]
                    updated_items.append(item)
                    total += item["subtotal"]
                else:
                    # If one product fails, entire cart fails
                    raise HTTPException(
                        status_code=503,
                        detail=f"Failed to fetch product {item['product_id']}"
                    )
                    
            except httpx.RequestError as e:
                # No graceful degradation
                raise HTTPException(
                    status_code=503,
                    detail=f"Product service unavailable: {str(e)}"
                )
    
    return Cart(user_id=user_id, items=updated_items, total_price=total)

@app.delete("/cart/{user_id}")
async def clear_cart(user_id: str):
    """Clear user's cart"""
    if user_id in carts:
        carts[user_id] = []
    return {"message": "Cart cleared"}

@app.get("/metrics")
@app.get("/vulnerable/metrics")
async def get_metrics():
    """Get service metrics"""
    return {
        "total_carts": len(carts),
        "total_items": sum(len(items) for items in carts.values()),
        "mode": "vulnerable",
        "circuit_breaker": "disabled"
    }