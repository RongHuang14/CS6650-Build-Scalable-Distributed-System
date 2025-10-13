# src/main.py
"""API layer: FastAPI application and routes."""
from fastapi import FastAPI, Query
import logging

from service import search_products
from database import db, SearchResponse

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# Create FastAPI application
app = FastAPI(
    title="HW6 Product Search API", 
    version="1.0.0",
    description="Performance testing for hw6"
)

@app.get("/products/search", response_model=SearchResponse)
async def search(q: str = Query(..., description="Search query")):
    """Search for products"""
    return search_products(q)

@app.get("/health")
async def health_check():
    """Health check endpoint for ALB monitoring."""
    return {
        "status": "healthy",
        "products_count": db.get_count()
    }
    
@app.get("/")
async def root():
    """Root endpoint with service info."""
    return {
        "service": "HW6 Product Search API",
        "endpoints": ["/products/search", "/health"]
    }

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8080)