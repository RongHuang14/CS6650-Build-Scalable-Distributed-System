"""
Data layer: Product models and in-memory storage.

This module provides:
- Product data model definition
- In-memory database with 100,000 products
- Thread-safe data access
"""
from pydantic import BaseModel, Field
from typing import List
import threading
import logging

# Configure module logger
logger = logging.getLogger(__name__)

# Configuration constants
TOTAL_PRODUCTS = 100000  # Total products to generate
MAX_SEARCH_RESULTS = 20  # Maximum results per search
PRODUCTS_TO_CHECK = 100

class Product(BaseModel):
    """Product model for search service."""
    id: int = Field(..., ge=1)
    name: str = Field(...)  # searchable
    category: str = Field(...)  # searchable
    description: str = Field(...)
    brand: str = Field(...)
    
class SearchResponse(BaseModel):
    """
    Response model for product search API.
    
    Attributes:
        products: List of matched products (limited to MAX_SEARCH_RESULTS)
        total_found: Total number of matches found in the checked subset
        search_time: Time taken to perform the search
    """
    products: List[Product] = Field(..., description="Matched products")
    total_found: int = Field(..., description="Total matches in checked products", ge=0)
    search_time: str = Field(..., description="Search execution time")

class ProductDatabase:
    """In-memory product database."""
    
    def __init__(self):
        self.lock = threading.RLock()
        self.products = self._generate_products()
    
    def _generate_products(self) -> List[Product]:
        """Generate test products."""
        logger.info(f"Generating {TOTAL_PRODUCTS} products...")
        
        brands = ["Alpha", "Beta", "Gamma", "Delta", "Epsilon"]
        categories = ["Electronics", "Books", "Home", "Clothing", "Sports"]
        
        products = []
        for i in range(TOTAL_PRODUCTS):
            products.append(Product(
                id=i + 1,
                name=f"Product {brands[i % len(brands)]} {i + 1}",
                category=categories[i % len(categories)],
                description=f"Product description {i + 1}",
                brand=brands[i % len(brands)]
            ))
        
        logger.info(f"Generated {len(products)} products")
        return products
    
    def get_products(self) -> List[Product]:
        """Get all products."""
        with self.lock:
            return self.products  
    
    def get_count(self) -> int:
        """Get product count."""
        with self.lock:
            return len(self.products)
        
# Create singleton instance
db = ProductDatabase()
