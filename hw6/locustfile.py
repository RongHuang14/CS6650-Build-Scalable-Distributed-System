from locust import FastHttpUser, task, between
import random

class ProductSearchUser(FastHttpUser):
    """Simulates users searching for products"""
    wait_time = between(0.1, 0.5)  # 100-500ms between requests
    
    search_terms = [
        "alpha", "beta", "gamma", "delta", "epsilon",
        "electronics", "books", "home", "clothing", "sports",
        "product"
    ]
    
    @task(10)  # Weight: 10
    def search_products(self):
        """Main task: search for products"""
        query = random.choice(self.search_terms)
        with self.client.get(
            f"/products/search?q={query}",
            catch_response=True
        ) as response:
            if response.status_code == 200:
                response.success()
            else:
                response.failure(f"Got status code {response.status_code}")
    
    @task(1)  # Weight: 1 (less frequent)
    def health_check(self):
        """Occasional health check"""
        self.client.get("/health")
    
    
    