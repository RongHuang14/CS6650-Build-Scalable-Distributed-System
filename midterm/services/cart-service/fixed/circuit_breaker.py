# services/cart-service/fixed/circuit_breaker.py
import asyncio
from datetime import datetime, timedelta
from enum import Enum
from typing import Callable, Any, Optional, Type
import functools

class CircuitState(Enum):
    CLOSED = "CLOSED"      # Normal operation
    OPEN = "OPEN"          # Failing, reject requests
    HALF_OPEN = "HALF_OPEN" # Testing if service recovered

class CircuitBreakerError(Exception):
    """Raised when circuit breaker is open"""
    pass

class CircuitBreaker:
    def __init__(
        self,
        failure_threshold: int = 5,
        recovery_timeout: int = 60,
        expected_exception: Type[Exception] = Exception,
        success_threshold: int = 2
    ):
        """
        Initialize Circuit Breaker
        
        Args:
            failure_threshold: Number of failures before opening circuit
            recovery_timeout: Seconds to wait before attempting recovery
            expected_exception: Exception types to count as failures
            success_threshold: Successes needed in HALF_OPEN to close circuit
        """
        self.failure_threshold = failure_threshold
        self.recovery_timeout = recovery_timeout
        self.expected_exception = expected_exception
        self.success_threshold = success_threshold
        
        # State management
        self._state = CircuitState.CLOSED
        self.failure_count = 0
        self.success_count = 0
        self.last_failure_time: Optional[datetime] = None
        self.half_open_success_count = 0
        
        # Metrics
        self.total_failures = 0
        self.total_successes = 0
        self.times_opened = 0
        self.last_open_time: Optional[datetime] = None
    
    @property
    def state(self) -> str:
        """Get current circuit state as string"""
        self._update_state()
        return self._state.value
    
    def _update_state(self):
        """Update circuit state based on current conditions"""
        if self._state == CircuitState.OPEN:
            # Check if recovery timeout has passed
            if self.last_failure_time and \
               datetime.now() - self.last_failure_time >= timedelta(seconds=self.recovery_timeout):
                self._state = CircuitState.HALF_OPEN
                self.half_open_success_count = 0
    
    def __call__(self, func: Callable) -> Callable:
        """Decorator for protecting functions with circuit breaker"""
        @functools.wraps(func)
        async def wrapper(*args, **kwargs):
            # Update state before checking
            self._update_state()
            
            # Check if circuit is open
            if self._state == CircuitState.OPEN:
                raise CircuitBreakerError(
                    f"Circuit breaker is OPEN. Service unavailable. "
                    f"Will retry after {self.recovery_timeout} seconds."
                )
            
            try:
                # Attempt to call the function
                if asyncio.iscoroutinefunction(func):
                    result = await func(*args, **kwargs)
                else:
                    result = func(*args, **kwargs)
                
                # Call succeeded
                self._on_success()
                return result
                
            except self.expected_exception as e:
                # Expected failure occurred
                self._on_failure()
                raise e
            except Exception as e:
                # Unexpected exception - let it through but don't count
                raise e
        
        return wrapper
    
    def _on_success(self):
        """Handle successful call"""
        self.success_count += 1
        self.total_successes += 1
        
        if self._state == CircuitState.HALF_OPEN:
            self.half_open_success_count += 1
            
            # Check if we've had enough successes to close the circuit
            if self.half_open_success_count >= self.success_threshold:
                self._state = CircuitState.CLOSED
                self.failure_count = 0
                self.half_open_success_count = 0
        
        elif self._state == CircuitState.CLOSED:
            # Reset failure count on success in closed state
            self.failure_count = 0
    
    def _on_failure(self):
        """Handle failed call"""
        self.failure_count += 1
        self.total_failures += 1
        self.last_failure_time = datetime.now()
        
        if self._state == CircuitState.CLOSED:
            # Check if we should open the circuit
            if self.failure_count >= self.failure_threshold:
                self._state = CircuitState.OPEN
                self.times_opened += 1
                self.last_open_time = datetime.now()
        
        elif self._state == CircuitState.HALF_OPEN:
            # Single failure in half-open state opens the circuit again
            self._state = CircuitState.OPEN
            self.half_open_success_count = 0
    
    def reset(self):
        """Manually reset the circuit breaker"""
        self._state = CircuitState.CLOSED
        self.failure_count = 0
        self.success_count = 0
        self.half_open_success_count = 0
        self.last_failure_time = None
    
    def get_stats(self) -> dict:
        """Get detailed statistics"""
        self._update_state()
        return {
            "state": self._state.value,
            "failure_count": self.failure_count,
            "success_count": self.success_count,
            "total_failures": self.total_failures,
            "total_successes": self.total_successes,
            "times_opened": self.times_opened,
            "failure_threshold": self.failure_threshold,
            "recovery_timeout": self.recovery_timeout,
            "last_failure_time": self.last_failure_time.isoformat() if self.last_failure_time else None,
            "last_open_time": self.last_open_time.isoformat() if self.last_open_time else None
        }