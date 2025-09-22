# locustfile.py - Basic load test with 1:1 GET/POST ratio
from locust import HttpUser, task, between

class BasicUser(HttpUser):
    # Simulate realistic user behavior with 1-2s between requests
    wait_time = between(1, 2)
    
    @task  # Equal weight - runs 50% of time
    def get_item(self):
        # Test read operation
        self.client.get("/get?key=test", name="/get")
    
    @task  # Equal weight - runs 50% of time
    def post_item(self):
        # Test write operation  
        self.client.post("/post", 
                        json={"key": "test", "value": "data"},
                        name="/post")