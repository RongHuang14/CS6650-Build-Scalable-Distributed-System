# locustfile.py
"""
This script tests both vulnerable and fixed versions of a shopping cart service
to demonstrate how Circuit Breaker pattern prevents cascading failures.

Usage:
    # Test vulnerable version
    TEST_MODE=vulnerable locust -H http://localhost:8000
    
    # Test fixed version  
    TEST_MODE=fixed locust -H http://localhost:8000
    
    # Test both simultaneously (for comparison)
    TEST_MODE=both locust -H http://localhost:8000
"""

from locust import HttpUser, task, between, events
import random
import os
import logging
from datetime import datetime

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s [%(levelname)s] %(message)s'
)
logger = logging.getLogger(__name__)

class EcommerceUser(HttpUser):
    """
    Simulates realistic e-commerce user behavior.
    Tests cart operations that depend on the product service.
    """
    
    # Realistic wait time between user actions
    wait_time = between(0.5, 2.0)
    # Increase proportion of traffic from business users
    weight = 20
    
    def on_start(self):
        """
        Initialize user session.
        Determines which service version to test based on environment variable.
        """
        # Determine test mode from environment variable
        self.test_mode = os.getenv("TEST_MODE", "vulnerable").lower()
        
        # Generate unique user ID for this session
        self.user_id = f"user_{random.randint(1, 1000)}"
        
        # Available product IDs (must match product service)
        self.product_ids = ["1", "2", "3", "4", "5"]
        
        # Track last seen circuit breaker state to avoid log spam
        self.last_cb_state_fixed = None
        
        logger.info(f"User {self.user_id} started - Testing {self.test_mode} version(s)")
    
    @task(10)
    def add_item_to_cart(self):
        """
        Primary user action: Add products to shopping cart.
        Weight 10 indicates this is the most common operation.
        """
        # Select random product and quantity
        product_id = random.choice(self.product_ids)
        quantity = random.randint(1, 3)
        
        # Test based on mode
        if self.test_mode in ["vulnerable", "both"]:
            self._test_add_to_cart("vulnerable", product_id, quantity)
            
        if self.test_mode in ["fixed", "both"]:
            self._test_add_to_cart("fixed", product_id, quantity)
    
    @task(3)
    def view_cart(self):
        """
        Secondary action: View cart contents.
        Weight 3 indicates moderate frequency.
        """
        if self.test_mode in ["vulnerable", "both"]:
            self._test_view_cart("vulnerable")
            
        if self.test_mode in ["fixed", "both"]:
            self._test_view_cart("fixed")
    
    
    
    def _test_add_to_cart(self, version: str, product_id: str, quantity: int):
        """
        Helper method to test adding items to cart.
        Handles both success and failure scenarios appropriately.
        """
        # Route to the correct service version
        base_path = "/fixed" if version == "fixed" else "/vulnerable"
        endpoint = f"{base_path}/cart/{self.user_id}/add"
        
        with self.client.post(
            endpoint,
            json={"product_id": product_id, "quantity": quantity},
            name=f"/{version}/cart/add",  # Group similar requests in stats
            catch_response=True
        ) as response:
            
            if version == "vulnerable":
                # Vulnerable version: 503 is complete failure
                if response.status_code == 503:
                    response.failure("Cascading failure - service unavailable")
                elif response.status_code != 200:
                    response.failure(f"Unexpected status: {response.status_code}")
                    
            elif version == "fixed":
                # Fixed version: Check for degraded but operational state
                if response.status_code == 200:
                    try:
                        data = response.json()
                        cart_item = data.get("cart_item", {})
                        
                        # Check if operating in degraded mode
                        if cart_item.get("status") == "price_pending":
                            # Still successful but degraded
                            logger.debug(f"Fixed version: Degraded mode - price unavailable")
                        
                        # Log circuit breaker state
                        cb_state = data.get("circuit_breaker_state")
                        if cb_state and cb_state != self.last_cb_state_fixed:
                            logger.info(f"Circuit Breaker State: {cb_state}")
                            self.last_cb_state_fixed = cb_state
                            
                    except Exception as e:
                        logger.error(f"Failed to parse response: {e}")
                        
                elif response.status_code == 503:
                    response.failure("Fixed version also unavailable")
                else:
                    response.failure(f"Unexpected status: {response.status_code}")
    
    def _test_view_cart(self, version: str):
        """
        Helper method to test viewing cart contents.
        """
        # Route to the correct service version
        base_path = "/fixed" if version == "fixed" else "/vulnerable"
        endpoint = f"{base_path}/cart/{self.user_id}"
        
        with self.client.get(
            endpoint,
            name=f"/{version}/cart/view",
            catch_response=True
        ) as response:
            
            if response.status_code == 200:
                if version == "fixed":
                    try:
                        data = response.json()
                        if data.get("degraded_mode"):
                            logger.debug(f"Cart retrieved in degraded mode for user {self.user_id}")
                    except:
                        pass
                        
            elif response.status_code == 503:
                response.failure(f"{version} service: Cannot retrieve cart")
            else:
                response.failure(f"Unexpected status: {response.status_code}")
    

class AdminUser(HttpUser):
    """
    Special user type for controlling the demo.
    Can trigger failures and recovery in the product service.
    """
    
    # Admin operations are infrequent
    wait_time = between(30, 60)
    
    # Low weight to limit number of admin users
    weight = 1
    
    def on_start(self):
        """Initialize admin user"""
        logger.warning("Admin user initialized - Can control product service state")
        self.crash_triggered = False
        self.recovery_triggered = False
    
    @task
    def monitor_product_service(self):
        """
        Monitor product service health status
        """
        with self.client.get(
            "/health",
            name="/product/health",
            catch_response=True
        ) as response:
            if response.status_code == 503:
                if not self.crash_triggered:
                    logger.error("⚠️ PRODUCT SERVICE IS UNHEALTHY!")
                    self.crash_triggered = True
            elif response.status_code == 200:
                if self.crash_triggered and not self.recovery_triggered:
                    logger.info("✅ Product service has recovered")
                    self.recovery_triggered = True
    
    # Uncomment the following methods to allow automatic crash/recovery
    # WARNING: This will affect all users in the test!
    
    # @task
    # def trigger_crash(self):
    #     """
    #     Trigger product service failure
    #     Only triggers once per admin user
    #     """
    #     if not self.crash_triggered:
    #         response = self.client.post("/fail/on", name="ADMIN: Trigger Crash")
    #         if response.status_code == 200:
    #             logger.critical("🔥 PRODUCT SERVICE CRASH TRIGGERED BY ADMIN!")
    #             self.crash_triggered = True
    
    # @task
    # def trigger_recovery(self):
    #     """
    #     Recover product service
    #     Only triggers after crash has been triggered
    #     """
    #     if self.crash_triggered and not self.recovery_triggered:
    #         response = self.client.post("/fail/off", name="ADMIN: Recover Service")
    #         if response.status_code == 200:
    #             logger.warning("✅ PRODUCT SERVICE RECOVERED BY ADMIN!")
    #             self.recovery_triggered = True

# Event handlers for test lifecycle
@events.test_start.add_listener
def on_test_start(environment, **kwargs):
    """
    Called when the test starts.
    Displays test configuration and instructions.
    """
    test_mode = os.getenv("TEST_MODE", "vulnerable").upper()
    
    print("\n" + "="*70)
    print(" CIRCUIT BREAKER PATTERN - LOAD TEST ")
    print("="*70)
    print(f" Test Mode: {test_mode}")
    print(f" Start Time: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
    print(f" Target Host: {environment.host}")
    print("="*70)
    
    if test_mode == "VULNERABLE":
        print("\n Testing VULNERABLE version - No circuit breaker protection")
        print(" Expect cascading failures when product service fails")
    elif test_mode == "FIXED":
        print("\n Testing FIXED version - With circuit breaker protection")
        print(" Expect graceful degradation when product service fails")
    elif test_mode == "BOTH":
        print("\n Testing BOTH versions simultaneously for comparison")
        print(" Compare behavior during product service failure")
    
    print("\n To trigger failure: curl -X POST http://<product-host>/fail/on")
    print(" To recover service: curl -X POST http://<product-host>/fail/off")
    print("="*70 + "\n")

@events.test_stop.add_listener
def on_test_stop(environment, **kwargs):
    """
    Called when the test stops.
    Displays summary statistics.
    """
    print("\n" + "="*70)
    print(" TEST COMPLETED ")
    print("="*70)
    print(f" End Time: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
    print(f" Total Requests: {environment.stats.total.num_requests}")
    print(f" Total Failures: {environment.stats.total.num_failures}")
    
    if environment.stats.total.num_requests > 0:
        failure_rate = (environment.stats.total.num_failures / 
                       environment.stats.total.num_requests * 100)
        print(f" Failure Rate: {failure_rate:.2f}%")
        print(f" Median Response Time: {environment.stats.total.median_response_time}ms")
        print(f" 95%ile Response Time: {environment.stats.total.get_response_time_percentile(0.95)}ms")
    
    print("="*70 + "\n")
