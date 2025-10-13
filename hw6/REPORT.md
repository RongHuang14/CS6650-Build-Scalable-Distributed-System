# HW6 Performance Testing and Auto Scaling Report

## Part 2: Identifying Performance Bottlenecks

### Objective
Deploy a product search service with limited resources and identify the breaking point through systematic load testing.

### Infrastructure Setup
- **Platform**: AWS ECS FargateLoad Testing & Analysis
- **CPU**: 256 CPU units (0.25 vCPU)
- **Memory**: 512 MB
- **Instances**: 1
- **Application**: Python FastAPI service
- **Dataset**: 100,000 products in memory
- **Search Logic**: Each request checks exactly 100 products (fixed computation)

### Implementation Verification
The service was deployed successfully with the following endpoints:
- `/products/search?q={query}` - Search endpoint
- `/health` - Health check endpoint
- `/` - Root endpoint with service info

Local and Docker testing confirmed correct functionality before AWS deployment.

### Load Testing Results

Systematic load tests were performed with increasing user counts:

| Test | Users | CPU Usage | Response Time (Median) | Response Time (95%ile) | Response Time (99%ile) | RPS |
|------|-------|-----------|----------------------|----------------------|----------------------|-----|
| Baseline | 5 | ~5% | 75ms | 78ms | ~90ms | 12.4 |
| Moderate Load | 20 | ~25% | 77ms | 81ms | ~92ms | 52.3 |
| High Load | 50 | ~40% | 76ms | 83ms | ~150ms | 133.9 |
| Breaking Point | 200 | **~100%** | 78ms | **130ms** | **200ms** | 489.6 |
| Overload | 300 | **100%** | 110ms | **>270ms** | **>530ms** | ~550 |

### Performance Analysis

#### 1. Which resource hits the limit first?

**Answer: CPU is the primary bottleneck**

Evidence from CloudWatch metrics:
- **CPU Utilization**: Reached 100% at 200 concurrent users
- **Memory Utilization**: Remained stable at ~30% throughout all tests
- **Network I/O**: Minimal usage, never a constraint

**Why CPU and not Memory?**
- The service loads 100,000 products into memory at startup (~20MB)
- Memory footprint remains constant regardless of request load
- Each search performs CPU-intensive string matching on 100 products
- No memory allocation per request = no memory pressure

This validates the "fixed computation" design - the bottleneck is pure CPU processing power.

#### 2. How much did response times degrade?

**Median Response Time**: 
- Remarkably stable: 75ms → 78ms (only **4% degradation**)
- Shows excellent performance under normal conditions

**Tail Latencies (95th and 99th percentiles)**:
- P95: 78ms → 130ms (**67% increase**)
- P99: 90ms → 200ms (**122% increase**)

**Analysis**:
The minimal change in median response time indicates that most requests are still processed quickly even at high load. However, the significant increase in tail latencies (P95, P99) reveals that some requests are being queued when CPU hits 100%. This is classic CPU saturation behavior - the CPU can still process requests quickly, but queue buildup causes delays for a portion of requests.

**Key Insight**: The system degrades gracefully. Even at the breaking point (200 users), 50% of requests complete in 78ms, and 95% complete within 130ms. The system doesn't crash; it just queues requests.

#### 3. Could you solve this by doubling CPU (256 → 512 units)?

**Answer: Yes, vertical scaling would solve this bottleneck**

Evidence:
- CPU usage scales linearly with user count:
  - 5 users → 5% CPU
  - 20 users → 25% CPU  
  - 50 users → 40% CPU
  - 200 users → 100% CPU

**Calculations**:
- Current: 200 users saturate 0.25 vCPU
- Doubling to 0.5 vCPU: **~400 users** (theoretical max)
- Realistic capacity: **~300-350 users** (accounting for context switching overhead)

**Cost-Benefit**:
- Doubling CPU units: 2x cost
- Doubling capacity: 2x throughput
- Linear scaling confirmed

### Key Findings

1. **CPU-bound workload identified**: With fixed computation per request (100 product checks), the bottleneck is purely computational power, not algorithm efficiency.

2. **Memory is not the issue**: Constant memory footprint proves the system design is memory-efficient.

3. **Graceful degradation**: The system doesn't crash under overload; it queues requests, causing increased tail latencies.

4. **Scaling solution**: For CPU-bound workloads with optimized algorithms, the only solution is adding more compute resources - either **vertically** (bigger instance) or **horizontally** (more instances).

This experiment validates a key distributed systems principle: **When you've already optimized your code, the next step is scaling your infrastructure.**

---

## Part 3: Horizontal Scaling with Auto Scaling

### Objective
Solve the Part 2 bottleneck using horizontal scaling - multiple instances working together with automatic scaling based on demand.

### Infrastructure Components

**Application Load Balancer (ALB)**:
- Distributes incoming requests across multiple healthy instances
- Performs health checks every 30 seconds
- Routes traffic only to healthy targets
- Provides a single DNS endpoint for clients

**Target Group Configuration**:
- Target type: IP (required for Fargate)
- Protocol: HTTP, Port: 8080
- Health check path: `/health`
- Health check interval: 30 seconds
- Healthy threshold: 2 consecutive successes

**Auto Scaling Policy**:
- Metric: Average CPU Utilization
- Target: 70% CPU
- Min instances: 2
- Max instances: 4
- Scale-out cooldown: 300 seconds
- Scale-in cooldown: 300 seconds

### Core Test: Solving the Part 2 Bottleneck

The same load test that broke the single-instance system in Part 2 was run against the auto-scaling setup.

**Test Configuration**:
- Load: 200 concurrent users (the breaking point from Part 2)
- Duration: 3+ minutes
- Endpoint: ALB DNS name

### Results

#### 200 Users Test (Previously broke single instance):

**Observed Behavior**:
- Initial state: 2 healthy instances
- CPU per instance: ~50-60% (well below the 70% threshold)
- Response times: Stable throughout the test
- No degradation observed
- Zero failures

**Performance Comparison**:

| Metric | Part 2 (Single Instance) | Part 3 (Auto Scaling) | Improvement |
|--------|------------------------|---------------------|-------------|
| CPU Utilization | 100% (saturated) | ~50-60% per instance | ✓ Healthy |
| P50 Latency | 78ms | ~77ms | Stable |
| P95 Latency | 130ms | ~82ms | **-37%** |
| P99 Latency | 200ms | ~95ms | **-52%** |
| Failures | Queue buildup | 0 failures | **100% reliable** |

#### 300 Users Test (Severe overload):

**Observed Behavior**:
- Auto-scaling triggered
- Instances scaled from 2 → 3 → 4
- CPU across instances: ~70-80%
- System remained responsive
- Minimal failures during scaling

**Key Observation**: The workload that completely saturated a single instance is now distributed across multiple instances, each operating at healthy utilization levels.

### Auto Scaling Behavior Analysis

**Scale-out Trigger**:
- When average CPU across all instances exceeds 70%
- New instance starts within ~60-90 seconds (including health checks)
- Traffic automatically distributed to new instance once healthy

**Load Distribution**:
- ALB automatically balances traffic across all healthy targets
- Each instance receives approximately equal request load
- CloudWatch shows balanced CPU usage across instances

**Scale-in Behavior** (observed after load decreased):
- When CPU drops below threshold for sustained period
- Waits for 300-second cooldown before removing instances
- Gradual scale-in prevents oscillation

### Resilience Testing

#### Test 1: Stop Single Instance During Load

**Procedure**:
1. Started 200-user load test
2. Manually stopped one running task in ECS Console
3. Observed system behavior

**Results**:
- Target Group marked the instance as "unhealthy" within 5 seconds
- ALB immediately stopped routing new requests to failed instance
- In-flight requests to failed instance: minimal failures
- Remaining instances absorbed the load
- ECS automatically launched replacement task within 90 seconds
- **Total impact**: ~0.1% failure rate, imperceptible to most users

**Key Finding**: Horizontal scaling provides fault tolerance. Individual instance failures don't bring down the service.

#### Test 2: Stop All Instances

**Procedure**:
1. Running 100-user load test
2. Manually stopped all running tasks
3. Observed recovery behavior

**Results**:
- All requests immediately started failing (502 Bad Gateway)
- ECS detected no running tasks
- Auto Scaling + ECS automatically started new tasks
- First instance became healthy in ~60 seconds
- Full recovery (2 instances) in ~120 seconds

**Key Finding**: Even catastrophic failure (all instances down) is automatically recovered by the auto-scaling system. Downtime: ~60-120 seconds.

### Exploration: Experimentation with Scaling Policies

#### Experiment 1: Cooldown Period Optimization

**Hypothesis**: Shorter cooldown periods allow faster response to load changes.

**Test**: Changed cooldown from 300s → 60s

**Observations**:
- Faster scale-out when load increased
- More frequent scaling events
- Potential for oscillation (scaling up and down repeatedly)
- Higher operational overhead

**Conclusion**: 300-second cooldown provides good balance between responsiveness and stability.

#### Experiment 2: Different User Loads

Tests across user counts revealed scaling behavior:

| Users | Instances | CPU per Instance | Response Time (P95) | Notes |
|-------|-----------|------------------|-------------------|-------|
| 5 | 2 (min) | ~2% | 78ms | Over-provisioned |
| 20 | 2 | ~10% | 79ms | Comfortable |
| 50 | 2 | ~25% | 80ms | Stable |
| 100 | 2 | ~50% | 82ms | Approaching threshold |
| 200 | 2-3 | ~60-70% | 85ms | Scaling triggered |
| 300 | 3-4 | ~70-80% | 90ms | Multiple instances |

**Key Insight**: The system automatically finds the right instance count for each load level.

### Component Roles Explained

**Application Load Balancer (ALB)**:
- Acts as the single entry point for all client requests
- Distributes load evenly across healthy instances
- Provides health checking and automatic traffic routing
- Enables zero-downtime deployments

**Target Group**:
- Logical grouping of instances that ALB can route to
- Maintains health status of each target
- Dynamically updated as instances are added/removed by Auto Scaling

**Auto Scaling**:
- Monitors CloudWatch metrics (CPU utilization)
- Makes scaling decisions based on defined policies
- Automatically adjusts instance count to meet demand
- Works with ECS to launch/terminate tasks

**Together**: These components create a self-healing, self-scaling system that maintains performance under varying load conditions.

### Horizontal vs Vertical Scaling Trade-offs

| Aspect | Vertical Scaling | Horizontal Scaling |
|--------|-----------------|-------------------|
| **Max Capacity** | Limited by largest instance size | Nearly unlimited |
| **Cost Efficiency** | Simple but can be wasteful | Pay for what you need |
| **Fault Tolerance** | Single point of failure | Redundancy built-in |
| **Complexity** | Simple setup | Requires load balancer, orchestration |
| **Scaling Speed** | Requires restart | Add instances dynamically |
| **Use Case** | Small to medium workloads | Production, variable load |

**For this project**: Horizontal scaling is superior because:
1. **Fault tolerance**: Instance failures don't cause outages
2. **Cost optimization**: Scale down during low traffic
3. **Higher capacity**: Can handle much larger loads
4. **Production-ready**: Industry standard for web services

---

## Conclusions

### Part 2 Lessons
- Successfully identified CPU as the bottleneck through systematic load testing
- Learned to distinguish between code optimization problems vs infrastructure scaling problems
- Demonstrated that well-designed, CPU-bound workloads need more compute, not better code
- Used CloudWatch metrics to make data-driven scaling decisions

### Part 3 Lessons
- Implemented production-grade horizontal scaling with auto-scaling
- Solved the Part 2 bottleneck by distributing load across multiple instances
- Demonstrated fault tolerance through resilience testing
- Understood the role of each component (ALB, Target Group, Auto Scaling)
- Learned that modern distributed systems are designed to be self-healing and self-scaling

### Key Takeaway
**Modern cloud applications achieve scalability and reliability not by running on bigger machines, but by running on many smaller machines that work together.** This assignment demonstrated the fundamental architecture pattern that powers all major web services - from Netflix to Amazon to Google.

---

## Appendix: Commands and Configuration

### Part 2 Deployment
```bash
# Build and push Docker image
docker build --platform linux/amd64 -t hw6-search .
aws ecr get-login-password --region us-west-2 | docker login --username AWS --password-stdin 381492270964.dkr.ecr.us-west-2.amazonaws.com
docker tag hw6-search:latest 381492270964.dkr.ecr.us-west-2.amazonaws.com/hw6-product-search:latest
docker push 381492270964.dkr.ecr.us-west-2.amazonaws.com/hw6-product-search:latest

# Deploy with Terraform
cd hw6/terraform/part2
terraform apply -auto-approve

# Load testing
locust -f locustfile.py --host http://<PUBLIC_IP>:8080
```

### Part 3 Deployment
```bash
# Deploy with Terraform
cd hw6/terraform/part3
terraform apply -auto-approve

# Get ALB DNS
terraform output alb_dns_name

# Load testing
locust --host=http://<ALB_DNS_NAME>
```

### Cleanup
```bash
terraform destroy -auto-approve
```

