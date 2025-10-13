# Part 2: Identifying Performance Bottlenecks

## 🎯 Objective

Deploy a single-instance product search service and use load testing to discover its breaking point. Learn to recognize when a system needs more resources vs better code.

## 📋 Infrastructure Overview

### Single Instance Setup (No Horizontal Scaling)
- **ECS Fargate**: 1 instance
- **CPU**: 256 units (0.25 vCPU)
- **Memory**: 512 MB
- **No Load Balancer**: Direct access via public IP
- **No Auto Scaling**: Fixed single instance

### The Question
When your service slows down, how do you know if you need:
- 🔧 Better code (optimization)
- 🚀 More servers (scaling)

## 🚀 Deployment Steps

### 1. Prerequisites

Make sure you have:
- [ ] Docker image pushed to ECR (same image as Part 3)
- [ ] AWS CLI configured
- [ ] Terraform installed

### 2. Configure Variables

```bash
cd terraform/part2
cp terraform.tfvars.example terraform.tfvars
```

Edit `terraform.tfvars`:
```hcl
container_image = "123456789012.dkr.ecr.us-west-2.amazonaws.com/hw6-product-search:latest"
```

### 3. Deploy Infrastructure

```bash
terraform init
terraform plan
terraform apply
```

### 4. Get the Public IP

**Option 1: AWS Console**
1. Go to ECS Console
2. Click on your cluster: `hw6-part2-cluster`
3. Click on service: `hw6-part2-service`
4. Click on the running task
5. Find "Public IP" in the Network section

**Option 2: AWS CLI**
```bash
# Get task ARN
aws ecs list-tasks \
  --cluster hw6-part2-cluster \
  --service-name hw6-part2-service

# Get task details (replace TASK_ARN)
aws ecs describe-tasks \
  --cluster hw6-part2-cluster \
  --tasks <TASK_ARN> \
  --query 'tasks[0].attachments[0].details[?name==`networkInterfaceId`].value' \
  --output text

# Get public IP from ENI
aws ec2 describe-network-interfaces \
  --network-interface-ids <ENI_ID> \
  --query 'NetworkInterfaces[0].Association.PublicIp' \
  --output text
```

### 5. Test the Service

```bash
# Health check
curl http://<PUBLIC_IP>:8080/health

# Search test
curl "http://<PUBLIC_IP>:8080/products/search?q=Electronics"
```

## 🧪 Load Testing

### Test 1: Baseline (5 users, 2 minutes)

```bash
# In hw6/ directory
locust -f locustfile.py --host http://<PUBLIC_IP>:8080 --users 5 --spawn-rate 1 --run-time 2m --headless
```

**Expected:**
- ✅ Moderate CPU (~60%)
- ✅ Fast response times
- ✅ No failures

### Test 2: Breaking Point (20 users, 3 minutes)

```bash
locust -f locustfile.py --host http://<PUBLIC_IP>:8080 --users 20 --spawn-rate 2 --run-time 3m --headless
```

**Expected:**
- ⚠️ High CPU (near 100%)
- ⚠️ Degraded response times
- ⚠️ Possible timeouts/failures

### Web UI Testing (Recommended)

```bash
# Start Locust web UI
locust -f locustfile.py --host http://<PUBLIC_IP>:8080

# Open browser: http://localhost:8089
```

## 📊 CloudWatch Monitoring

### Key Metrics to Watch

1. **CPU Utilization**
   - Path: ECS Console > Cluster > Service > Metrics
   - Expected: 60% @ 5 users, ~100% @ 20 users

2. **Memory Utilization**
   - Should remain steady (~40-50%)
   - Products loaded at startup

3. **CloudWatch Logs**
   - Log Group: `/ecs/hw6-part2`
   - Check for errors, timeouts

### Create CloudWatch Dashboard (Optional)

```bash
# Use AWS Console to create dashboard with:
# - ECS CPU Utilization
# - ECS Memory Utilization
# - Log Insights queries
```

## 🔍 Analysis Questions

### During Testing, Observe:

1. **Which resource hits the limit first?**
   - [ ] CPU
   - [ ] Memory
   - [ ] Network

2. **How much did response times degrade?**
   - 5 users: _____ ms (avg)
   - 20 users: _____ ms (avg)
   - Degradation: _____ %

3. **Could you solve this by doubling CPU (256 → 512)?**
   - [ ] Yes - CPU is the bottleneck
   - [ ] No - Need horizontal scaling

4. **What's the breaking point?**
   - Users: _____
   - RPS: _____
   - CPU: _____

## 📸 Screenshots to Capture

For your report:

1. **AWS Console**
   - [ ] ECS Service overview (1 task running)
   - [ ] CloudWatch CPU metrics (both tests)
   - [ ] CloudWatch Memory metrics
   - [ ] ECS Task details (showing public IP)

2. **Load Testing**
   - [ ] Locust results - 5 users
   - [ ] Locust results - 20 users
   - [ ] Response time charts
   - [ ] Failure rates (if any)

3. **Comparison**
   - [ ] Side-by-side: 5 users vs 20 users

## 🎓 The Lesson

### Key Insight
When doing **inherently expensive work** (like checking 100 products per request), the solution is often:
- ✅ **More compute power** (horizontal scaling)
- ❌ **NOT** code optimization (algorithm is already bounded)

### Why This Matters
- Each request does **fixed work** (100 products)
- Code is already efficient (bounded iteration)
- **Bottleneck is CPU capacity**, not code quality
- **Solution**: Part 3 - Horizontal Scaling!

## 🧹 Cleanup

When done testing:

```bash
terraform destroy
```

**Note**: Keep screenshots and data before destroying!

## 📁 File Structure

```
terraform/part2/
├── main.tf                    # Infrastructure definition
├── variables.tf               # Input variables
├── outputs.tf                 # Output values
├── terraform.tfvars.example   # Example configuration
├── terraform.tfvars          # Your actual config (gitignored)
└── README.md                 # This file
```

## 🆘 Troubleshooting

### Task not starting
```bash
# Check task logs
aws logs tail /ecs/hw6-part2 --follow

# Check task stopped reason
aws ecs describe-tasks --cluster hw6-part2-cluster --tasks <TASK_ARN>
```

### Can't access public IP
- Check security group allows port 8080 from your IP
- Verify task has public IP assigned
- Check container health (may take 60s to start)

### High CPU with 5 users
- Verify search checks **exactly 100 products** (not all 100,000)
- Check application logs for errors
- Use CloudWatch Logs Insights

## 📚 Next Steps

Once you've identified the bottleneck:
→ **Part 3**: Solve it with horizontal scaling + auto scaling!

## 🎯 Success Criteria

- [ ] Single instance deployed successfully
- [ ] Can access service via public IP
- [ ] Baseline test (5 users) shows healthy performance
- [ ] Load test (20 users) shows clear bottleneck
- [ ] Identified which resource is the bottleneck
- [ ] Captured screenshots for report
- [ ] Understand why scaling (not optimization) is needed

