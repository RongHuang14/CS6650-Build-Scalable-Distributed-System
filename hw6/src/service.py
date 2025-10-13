# src/service.py
"""Business logic layer: Search implementation."""
import time
import logging
from database import (  
    db,
    SearchResponse,
    PRODUCTS_TO_CHECK,
    MAX_SEARCH_RESULTS
)

logger = logging.getLogger(__name__)

def search_products(query: str) -> SearchResponse:
    """
    Search for products matching the query with bounded iterations.
    Only check the first PRODUCTS_TO_CHECK products to simulate fixed computation time.
    """
    start_time = time.time()
    products = db.get_products()
    
    matched = []
    checked_count = 0
    query_lower = query.lower()
    
    for product in products:
        if checked_count >= PRODUCTS_TO_CHECK:
            break
        checked_count += 1
        if query_lower in product.name.lower() or query_lower in product.category.lower():
            matched.append(product)
    
    result = matched[:MAX_SEARCH_RESULTS]
    search_time = time.time() - start_time
    logger.info(f"Search '{query}': checked={checked_count}, "
                f"found={len(matched)}, returned={len(result)}, "
                f"time={search_time:.3f}s")
    
    return SearchResponse(
        products=result,
        total_found=len(matched),
        search_time=f"{search_time:.3f}s"
    )