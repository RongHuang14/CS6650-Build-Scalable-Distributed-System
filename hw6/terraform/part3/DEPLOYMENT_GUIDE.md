# Quick Deployment Guide for HW6 Part III

## 🚀 Quick Start (Step-by-Step)

### Prerequisites Check

```bash
# Verify AWS CLI is configured
aws sts get-caller-identity

# Verify Terraform is installed
terraform version

# Verify Docker is running
docker info
```

### Step 1: Initialize Terraform (2 minutes)

```bash
cd /Users/ronghuang/MyCScode/NEU/CS6650/hw6/part3
terraform init
```

Expected output: "Terraform has been successfully initialized!"

### Step 2: Create ECR Repository (1 minute)

```bash
# Create just the ECR repository first
terraform apply -target=aws_ecr_repository.main

# Type 'yes' when prompted
```

### Step 3: Build and Push Docker Image (3-5 minutes)

```bash
# Get ECR URL
ECR_URL=$(terraform output -raw ecr_repository_url)
echo "ECR Repository: $ECR_URL"

# Login to ECR
aws ecr get-login-password --region us-east-1 | \
  docker login --username AWS --password-stdin $ECR_URL

# Build image (from hw6 directory)
cd /Users/ronghuang/MyCScode/NEU/CS6650/hw6
docker build --platform linux/amd64 -t hw6-search .

# Tag and push
docker tag hw6-search:latest $ECR_URL:latest
docker push $ECR_URL:latest

# Verify image is pushed
aws ecr describe-images --repository-name hw6-search-repo --region us-east-1
```

### Step 4: Deploy Full Infrastructure (5-10 minutes)

```bash
# Return to part3 directory
cd /Users/ronghuang/MyCScode/NEU/CS6650/hw6/part3

# Deploy everything
terraform apply

# Type 'yes' when prompted
```

**Wait for completion** - you'll see progress for:
- ✓ VPC and networking
- ✓ Security groups
- ✓ Load balancer
- ✓ ECS cluster and service
- ✓ Auto scaling configuration

### Step 5: Verify Deployment (2 minutes)

```bash
# Save ALB URL
ALB_URL=$(terraform output -raw alb_url)
echo "Your ALB URL: $ALB_URL"

# Test health endpoint (may take 1-2 minutes for tasks to be healthy)
curl "$ALB_URL/health"

# Expected response:
# {"status":"healthy","products_count":100000}

# Test search endpoint
curl "$ALB_URL/products/search?q=alpha" | jq '.total_found'

# Expected: number of products found
```

### Step 6: Update Locust Configuration

```bash
# Edit locustfile.py
cd /Users/ronghuang/MyCScode/NEU/CS6650/hw6
```

Update the host in `locustfile.py`:
```python
class ProductSearchUser(FastHttpUser):
    host = "http://YOUR-ALB-DNS-HERE"  # Paste ALB DNS from terraform output
```

Get the exact DNS:
```bash
cd part3
terraform output alb_dns_name
```

### Step 7: Run Load Test

```bash
cd /Users/ronghuang/MyCScode/NEU/CS6650/hw6

# Start Locust
locust -f locustfile.py

# Open browser to: http://localhost:8089
```

Configure test:
- **Number of users**: 20
- **Spawn rate**: 2
- **Host**: (already set in locustfile.py)

Click **Start swarming**

### Step 8: Monitor Scaling

Open these in separate browser tabs:

**Tab 1: ECS Service**
```
AWS Console → ECS → Clusters → hw6-search-cluster → hw6-search-service
Watch: Tasks count (should scale from 2 → 3 → 4)
```

**Tab 2: Target Group Health**
```
AWS Console → EC2 → Target Groups → hw6-search-tg → Targets
Watch: Number of healthy targets
```

**Tab 3: CloudWatch Dashboard**
```bash
# Get dashboard URL
terraform output cloudwatch_dashboard_url
# Open in browser
```

**Tab 4: Locust Dashboard**
```
http://localhost:8089
Watch: Response times, requests/second
```

## 📊 What to Observe

### During Load Test (20 users)

1. **Initial State** (0-2 minutes):
   - 2 tasks running
   - CPU climbing toward 70%
   - Response times stable

2. **Scale-Out Trigger** (2-5 minutes):
   - CPU exceeds 70% average
   - Auto scaling initiates
   - 3rd task starts provisioning

3. **Scaling** (5-8 minutes):
   - New task becomes healthy
   - Load redistributes
   - CPU per task decreases
   - May scale to 4 tasks if needed

4. **Steady State** (8+ minutes):
   - 3-4 tasks running
   - CPU per task ~50-60%
   - Response times improved vs Part II

### After Stopping Load Test

1. **Scale-In Cooldown** (5 minutes):
   - CPU drops below 70%
   - Cooldown timer counts down

2. **Scale-In** (5-10 minutes after stopping):
   - Tasks gradually removed
   - Returns to minimum 2 tasks

## 🧪 Resilience Test

While load test is running:

```bash
# List running tasks
aws ecs list-tasks \
  --cluster hw6-search-cluster \
  --service-name hw6-search-service \
  --region us-east-1

# Stop one task (copy a task ID from above)
aws ecs stop-task \
  --cluster hw6-search-cluster \
  --task arn:aws:ecs:us-east-1:ACCOUNT:task/hw6-search-cluster/TASK_ID \
  --region us-east-1
```

**Watch what happens**:
- Target group marks task as unhealthy
- ECS immediately starts replacement
- Load test continues with minimal disruption
- New task becomes healthy within ~60 seconds

## 📈 Comparison with Part II

| Metric | Part II (1 instance) | Part III (2-4 instances) |
|--------|---------------------|--------------------------|
| 5 users | CPU ~60%, fast | CPU ~30% per task, fast |
| 20 users | CPU 100%, slow | CPU ~60% per task, stable |
| Failure impact | Complete outage | Graceful degradation |
| Recovery | Manual restart | Automatic replacement |

## 🧹 Clean Up (When Done)

```bash
cd /Users/ronghuang/MyCScode/NEU/CS6650/hw6/part3

# Destroy all resources
terraform destroy

# Type 'yes' to confirm

# Verify resources are deleted
aws ecs list-clusters --region us-east-1
aws elbv2 describe-load-balancers --region us-east-1
```

**Cost savings**: Deleting resources stops all charges immediately.

## 🔧 Troubleshooting

### Problem: Health checks failing

```bash
# Check task logs
aws logs tail /ecs/hw6-search --follow --region us-east-1

# Check task status
aws ecs describe-tasks \
  --cluster hw6-search-cluster \
  --tasks $(aws ecs list-tasks --cluster hw6-search-cluster --region us-east-1 --output text --query 'taskArns[0]') \
  --region us-east-1
```

**Common causes**:
- Container not listening on port 8080
- Health endpoint not responding
- Security group blocking ALB → ECS traffic

**Fix**: Check security group `hw6-search-ecs-tasks-sg` allows port 8080 from ALB security group

### Problem: Auto scaling not working

```bash
# Check current CPU utilization
aws cloudwatch get-metric-statistics \
  --namespace AWS/ECS \
  --metric-name CPUUtilization \
  --dimensions Name=ServiceName,Value=hw6-search-service Name=ClusterName,Value=hw6-search-cluster \
  --start-time $(date -u -v-1H '+%Y-%m-%dT%H:%M:%S') \
  --end-time $(date -u '+%Y-%m-%dT%H:%M:%S') \
  --period 300 \
  --statistics Average \
  --region us-east-1
```

**Common causes**:
- Load too low to trigger scaling
- Still in cooldown period
- Target value set too high

**Fix**: Increase load or adjust `cpu_target_value` in variables.tf

### Problem: Tasks starting but immediately stopping

```bash
# Check service events
aws ecs describe-services \
  --cluster hw6-search-cluster \
  --services hw6-search-service \
  --region us-east-1 \
  --query 'services[0].events[0:5]'
```

**Common causes**:
- Image not found in ECR
- Insufficient memory/CPU
- Container crashing on startup

**Fix**: Verify image exists: `aws ecr describe-images --repository-name hw6-search-repo --region us-east-1`

## 🎯 Success Criteria

✅ ALB DNS resolves and responds to /health
✅ 2 tasks running initially
✅ Load test with 20 users runs successfully
✅ Tasks scale up when CPU > 70%
✅ Response times better than Part II
✅ Stopping one task doesn't break the service
✅ CloudWatch dashboard shows metrics

## 📝 For Your Report

Document these observations:

1. **Screenshots**:
   - CloudWatch dashboard during load test
   - ECS service showing 2 → 4 task scaling
   - Target group with healthy targets
   - Locust results (20 users)

2. **Metrics**:
   - Response time comparison (Part II vs Part III)
   - CPU utilization per task
   - Time to scale out
   - Impact of stopping one task

3. **Analysis**:
   - How horizontal scaling solved the bottleneck
   - Role of each component (ALB, Target Group, Auto Scaling)
   - Trade-offs vs vertical scaling
   - Predicted behavior for different loads

## 🚀 Next Steps

Experiment with:
- Different CPU targets (50%, 90%)
- Different max capacities (6, 8 instances)
- Different load patterns (gradual vs spike)
- Multiple simultaneous task failures

