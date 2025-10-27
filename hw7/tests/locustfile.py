"""
Homework 7: Load Testing for Flash Sale System
Tests synchronous vs asynchronous order processing architectures.

Usage:
    # Phase 1: Normal operations (5 users, 30s, spawn 1/sec)
    locust -f locustfile.py --host=http://YOUR-ALB \
      --user-classes SyncOrderUser --users 5 --spawn-rate 1 \
      --run-time 30s --headless --html phase1_normal.html

    # Phase 2: Flash sale sync (20 users, 60s, spawn 10/sec)
    locust -f locustfile.py --host=http://YOUR-ALB \
      --user-classes SyncOrderUser --users 20 --spawn-rate 10 \
      --run-time 60s --headless --html phase2_flashsale.html

    # Phase 3-5: Flash sale async (20 users, 60s, spawn 10/sec)
    locust -f locustfile.py --host=http://YOUR-ALB \
      --user-classes AsyncOrderUser --users 20 --spawn-rate 10 \
      --run-time 60s --headless --html phase3_async.html
"""

import random
from locust import HttpUser, task, between


class SyncOrderUser(HttpUser):
    """
    Phase 1-2: Test synchronous order processing
    - Normal operations: 5 users, 30s, spawn 1/sec
    - Flash sale: 20 users, 60s, spawn 10/sec
    Expected: High response times, timeouts, bottleneck visible
    """
    # Per assignment: "random 100-500ms between requests"
    wait_time = between(0.1, 0.5)

    def on_start(self):
        """Initialize user with random customer ID"""
        self.customer_id = random.randint(1000, 9999)

    @task
    def sync_order(self):
        """
        Synchronous order - waits 3 seconds for payment processing
        Expected: 200 OK after 3+ seconds, or 503 timeout under load
        """
        order_data = {
            "customer_id": self.customer_id,
            "items": [
                {
                    "product_id": f"product_{random.randint(1, 100)}",
                    "quantity": random.randint(1, 5),
                    "price": round(random.uniform(10.0, 100.0), 2)
                }
            ]
        }
        
        with self.client.post("/orders/sync", 
                            json=order_data, 
                            catch_response=True,
                            timeout=35) as response:
            if response.status_code == 200:
                response.success()
            elif response.status_code == 503:
                response.failure("Payment processor timeout - BOTTLENECK!")
            else:
                response.failure(f"Failed with status {response.status_code}")


class AsyncOrderUser(HttpUser):
    """
    Phase 3-5: Test asynchronous order processing
    - Flash sale: 20 users, 60s, spawn 10/sec
    - Worker scaling: test with 1, 5, 20, 100 goroutines in processor
    Expected: 100% acceptance rate (<100ms), fast responses, queue builds up
    """
    # Per assignment: "random 100-500ms between requests"
    wait_time = between(0.1, 0.5)

    def on_start(self):
        """Initialize user with random customer ID"""
        self.customer_id = random.randint(1000, 9999)

    @task
    def async_order(self):
        """
        Asynchronous order - returns immediately with 202 Accepted
        Expected: 202 Accepted in <100ms, order queued for processing
        """
        order_data = {
            "customer_id": self.customer_id,
            "items": [
                {
                    "product_id": f"product_{random.randint(1, 100)}",
                    "quantity": random.randint(1, 5),
                    "price": round(random.uniform(10.0, 100.0), 2)
                }
            ]
        }
        
        with self.client.post("/orders/async", 
                            json=order_data, 
                            catch_response=True,
                            timeout=5) as response:
            if response.status_code == 202:
                response.success()
            else:
                response.failure(f"Failed with status {response.status_code}")
