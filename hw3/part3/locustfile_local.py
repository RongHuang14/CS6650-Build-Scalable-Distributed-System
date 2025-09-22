# locustfile_local.py - Realistic load test with 3:1 GET/POST ratio
from locust import HttpUser, task, between

class LocalTestUser(HttpUser):
    # Shorter wait time for higher load
    wait_time = between(0.5, 1.5)
    
    @task(3)  # Weight 3 - runs 75% of time
    def get_item(self):
        self.client.get("/get?key=test", name="/get")
    
    @task(1)  # Weight 1 - runs 25% of time  
    def post_item(self):
        self.client.post("/post", 
                        json={"key": "test", "value": "data"},
                        name="/post")