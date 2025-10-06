from locust import HttpUser, FastHttpUser, task, between
import random
import json

# class ProductAPIUser(HttpUser):
#     """Standard HttpUser for testing Product API"""
#     wait_time = between(1, 3)
    
#     def on_start(self):
#         """Called when a user starts"""
#         self.client.verify = False  # Disable SSL verification for testing
        
#     @task(3)
#     def get_product(self):
#         """GET request - most common operation (weight: 3)"""
#         product_id = random.randint(1, 10)  # Test existing and non-existing products
#         self.client.get(f"/products/{product_id}", name="GET /products/{id}")
    
#     @task(1)
#     def add_product(self):
#         """POST request - less common operation (weight: 1)"""
#         product_id = random.randint(11, 100)  # Use higher IDs to avoid conflicts
#         product_data = {
#             "product_id": product_id,
#             "sku": f"TEST-{product_id:03d}",
#             "manufacturer": f"TestCorp-{product_id}",
#             "category_id": random.randint(1, 5),
#             "weight": random.randint(100, 2000),
#             "some_other_id": product_id + 1000
#         }
#         self.client.post(
#             f"/products/{product_id}/details", 
#             json=product_data,
#             name="POST /products/{id}/details"
#         )
    
#     @task(1)
#     def health_check(self):
#         """Health check endpoint"""
#         self.client.get("/health", name="GET /health")

class FastProductAPIUser(FastHttpUser):
    """FastHttpUser for comparison testing"""
    wait_time = between(1, 3)
    
    def on_start(self):
        """Called when a user starts"""
        self.client.verify = False  # Disable SSL verification for testing
        
    @task(3)
    def get_product(self):
        """GET request - most common operation (weight: 3)"""
        product_id = random.randint(1, 10)
        self.client.get(f"/products/{product_id}", name="GET /products/{id}")
    
    @task(1)
    def add_product(self):
        """POST request - less common operation (weight: 1)"""
        product_id = random.randint(11, 100)
        product_data = {
            "product_id": product_id,
            "sku": f"TEST-{product_id:03d}",
            "manufacturer": f"TestCorp-{product_id}",
            "category_id": random.randint(1, 5),
            "weight": random.randint(100, 2000),
            "some_other_id": product_id + 1000
        }
        self.client.post(
            f"/products/{product_id}/details", 
            json=product_data,
            name="POST /products/{id}/details"
        )
    
    @task(1)
    def health_check(self):
        """Health check endpoint"""
        self.client.get("/health", name="GET /health")

class ReadHeavyUser(HttpUser):
    """User class for read-heavy scenario (90% GET, 10% POST)"""
    wait_time = between(0.5, 2)
    
    def on_start(self):
        self.client.verify = False
        
    @task(9)
    def get_product(self):
        """Heavy read operations"""
        product_id = random.randint(1, 20)
        self.client.get(f"/products/{product_id}", name="GET /products/{id}")
    
    @task(1)
    def add_product(self):
        """Light write operations"""
        product_id = random.randint(101, 200)
        product_data = {
            "product_id": product_id,
            "sku": f"READ-{product_id:03d}",
            "manufacturer": f"ReadCorp-{product_id}",
            "category_id": random.randint(1, 3),
            "weight": random.randint(200, 1500),
            "some_other_id": product_id + 2000
        }
        self.client.post(
            f"/products/{product_id}/details", 
            json=product_data,
            name="POST /products/{id}/details"
        )

class WriteHeavyUser(HttpUser):
    """User class for write-heavy scenario (30% GET, 70% POST)"""
    wait_time = between(1, 4)
    
    def on_start(self):
        self.client.verify = False
        
    @task(3)
    def get_product(self):
        """Light read operations"""
        product_id = random.randint(1, 50)
        self.client.get(f"/products/{product_id}", name="GET /products/{id}")
    
    @task(7)
    def add_product(self):
        """Heavy write operations"""
        product_id = random.randint(201, 500)
        product_data = {
            "product_id": product_id,
            "sku": f"WRITE-{product_id:03d}",
            "manufacturer": f"WriteCorp-{product_id}",
            "category_id": random.randint(1, 10),
            "weight": random.randint(50, 3000),
            "some_other_id": product_id + 3000
        }
        self.client.post(
            f"/products/{product_id}/details", 
            json=product_data,
            name="POST /products/{id}/details"
        )
