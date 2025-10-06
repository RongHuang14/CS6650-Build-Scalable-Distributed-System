"""Product API Service Implementation for CS6650"""
from fastapi import FastAPI, HTTPException, Path, Body
from pydantic import BaseModel, Field
from typing import Dict
import logging

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# Initialize FastAPI app
app = FastAPI(title="Product API", version="1.0.0")

# Data Model
class Product(BaseModel):
    """Product model per OpenAPI specification"""
    product_id: int = Field(..., ge=1)
    sku: str = Field(..., min_length=1, max_length=100)
    manufacturer: str = Field(..., min_length=1, max_length=200)
    category_id: int = Field(..., ge=1)
    weight: int = Field(..., ge=0)
    some_other_id: int = Field(..., ge=1)

# In-memory storage (HashMap for O(1) operations)
products_db: Dict[int, Product] = {}

# Initialize with test data
def init_data():
    """Initialize database with sample products"""
    test_products = [
        Product(product_id=1, sku="LAPTOP-001", manufacturer="Dell", 
                category_id=1, weight=2500, some_other_id=101),
        Product(product_id=2, sku="PHONE-001", manufacturer="Apple",
                category_id=2, weight=200, some_other_id=102)
    ]
    for product in test_products:
        products_db[product.product_id] = product
    logger.info(f"Initialized with {len(products_db)} products")

init_data()

# API Endpoints
@app.get("/products/{product_id}", response_model=Product)
async def get_product(product_id: int = Path(..., ge=1)):
    """
    GET endpoint: Retrieve product by ID
    Returns: 200 (success), 404 (not found)
    """
    logger.info(f"GET request for product {product_id}")
    
    if product_id not in products_db:
        logger.warning(f"Product {product_id} not found")
        raise HTTPException(
            status_code=404,
            detail={
                "error": "PRODUCT_NOT_FOUND",
                "message": f"Product with ID {product_id} not found"
            }
        )
    
    return products_db[product_id]

@app.post("/products/{product_id}/details", status_code=204)
async def add_product_details(
    product_id: int = Path(..., ge=1),
    product: Product = Body(...)
):
    """
    POST endpoint: Add or update product details
    Returns: 204 (success), 400 (bad request)
    """
    logger.info(f"POST request for product {product_id}")
    
    # Validate ID consistency
    if product_id != product.product_id:
        logger.warning(f"ID mismatch: path={product_id}, body={product.product_id}")
        raise HTTPException(
            status_code=400,
            detail={
                "error": "ID_MISMATCH",
                "message": "Product ID in path does not match request body",
                "details": f"Path ID: {product_id}, Body ID: {product.product_id}"
            }
        )
    
    # Store/update product
    products_db[product_id] = product
    logger.info(f"Product {product_id} stored successfully")

@app.get("/health")
async def health_check():
    """Health check endpoint for monitoring"""
    return {
        "status": "healthy",
        "products_count": len(products_db)
    }

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8080)