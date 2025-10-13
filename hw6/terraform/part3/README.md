# HW6 Part III: Horizontal Scaling with Auto Scaling

## 🎯 Overview

This part demonstrates deploying a product search service with **horizontal scaling and auto-scaling** to handle loads that would break a single instance.

**Key Technologies**:
- AWS ECS Fargate (container orchestration)
- Application Load Balancer (traffic distribution)
- Auto Scaling (automatic capacity adjustment)
- Terraform (infrastructure as code)

---

## 🏗️ Architecture

```
Internet
   ↓
Application Load Balancer (ALB)
   ↓
Target Group (health checks)
   ↓
┌─────────────┬─────────────┬─────────────┬─────────────┐
│  ECS Task 1 │  ECS Task 2 │  ECS Task 3 │  ECS Task 4 │
│  (Fargate)  │  (Fargate)  │  (Fargate)  │  (Fargate)  │
└─────────────┴─────────────┴─────────────┴─────────────┘
         Auto Scaling (2-4 instances, 70% CPU target)
```

**Components**:
1. **ALB**: Distributes requests to healthy instances
2. **Target Group**: Tracks instance health via `/health` endpoint
3. **ECS Service**: Maintains desired task count, auto-recovers failures
4. **Auto Scaling**: Adds/removes tasks based on CPU utilization

---

## 📦 Prerequisites

- AWS CLI configured with credentials
- Terraform installed (v1.0+)
- Docker installed
- Locust installed (`pip install locust`)
- Existing ECR repository from Part II

---

## 🚀 Deployment

### 1. Configure Variables

Create `terraform.tfvars`:
```hcl
aws_region     = "us-west-2"
ecr_repository = "hw6-part3-repo"  # Your ECR repo name
image_tag      = "latest"
min_capacity   = 2  # Minimum instances
max_capacity   = 4  # Maximum instances
```

### 2. Build and Push Docker Image

```bash
# Build
docker build -t hw6-part3-repo:latest .

# Tag for ECR
aws ecr get-login-password --region us-west-2 | docker login --username AWS --password-stdin <account-id>.dkr.ecr.us-west-2.amazonaws.com
docker tag hw6-part3-repo:latest <account-id>.dkr.ecr.us-west-2.amazonaws.com/hw6-part3-repo:latest

# Push
docker push <account-id>.dkr.ecr.us-west-2.amazonaws.com/hw6-part3-repo:latest
```

### 3. Deploy Infrastructure

```bash
cd part3/
terraform init
terraform plan
terraform apply
```

**Outputs** (save these):
```
alb_dns_name = "hw6-part3-alb-XXXXXXXXXX.us-west-2.elb.amazonaws.com"
ecs_cluster_name = "hw6-part3-cluster"
ecs_service_name = "hw6-part3-service"
```

### 4. Verify Deployment

```bash
# Check service status
aws ecs describe-services \
  --cluster hw6-part3-cluster \
  --services hw6-part3-service \
  --region us-west-2

# Check target health
aws elbv2 describe-target-health \
  --target-group-arn <your-target-group-arn> \
  --region us-west-2

# Test endpoint
curl http://<alb-dns-name>/health
curl "http://<alb-dns-name>/products/search?q=laptop"
```

---

## 🧪 Load Testing

### Run the Load Test

```bash
# Start Locust
cd ../
locust -f locustfile.py --host http://<alb-dns-name>

# Open browser to http://localhost:8089
# Configure:
#   - Number of users: 200
#   - Ramp up: 1 user/second
#   - Host: (already set)
# Click "Start Swarming"
```

### What to Observe

**In Locust UI** (http://localhost:8089):
- Request success rate (should be 100%)
- Response times (charts tab)
- Requests per second
- Number of failures (should be 0)

**In AWS Console**:
1. **ECS Service**: Watch task count change
2. **Target Groups**: Monitor healthy target count
3. **CloudWatch**: View CPU utilization, request metrics
4. **Auto Scaling**: Check scaling activities

---

## 🔥 Resilience Testing

### Test 1: Stop One Instance

```bash
# List running tasks
aws ecs list-tasks \
  --cluster hw6-part3-cluster \
  --service hw6-part3-service \
  --region us-west-2

# Stop one task (while load test is running!)
aws ecs stop-task \
  --cluster hw6-part3-cluster \
  --task <task-id> \
  --reason "Resilience testing" \
  --region us-west-2
```

**Expected Behavior**:
- ✅ Load test continues with 100% success
- ✅ ALB stops routing to failed instance
- ✅ ECS automatically launches replacement task
- ✅ System recovers to desired count

### Test 2: Monitor Auto Scaling

```bash
# Watch scaling activities
aws application-autoscaling describe-scaling-activities \
  --service-namespace ecs \
  --resource-id service/hw6-part3-cluster/hw6-part3-service \
  --region us-west-2

# View CPU metrics
aws cloudwatch get-metric-statistics \
  --namespace AWS/ECS \
  --metric-name CPUUtilization \
  --dimensions Name=ServiceName,Value=hw6-part3-service Name=ClusterName,Value=hw6-part3-cluster \
  --start-time $(date -u -v-1H +%Y-%m-%dT%H:%M:%S) \
  --end-time $(date -u +%Y-%m-%dT%H:%M:%S) \
  --period 300 \
  --statistics Average Maximum \
  --region us-west-2
```

---

## 📊 CloudWatch Metrics

### Create Dashboard

1. Go to CloudWatch Console → Dashboards
2. Create new dashboard: "hw6-part3-monitoring"
3. Add widgets:

**ECS CPU Utilization**:
```
Namespace: AWS/ECS
Metric: CPUUtilization
Dimensions: ClusterName=hw6-part3-cluster, ServiceName=hw6-part3-service
Statistic: Average
```

**ALB Request Count**:
```
Namespace: AWS/ApplicationELB
Metric: RequestCount
Dimensions: LoadBalancer=app/hw6-part3-alb/*
Statistic: Sum
```

**Target Response Time**:
```
Namespace: AWS/ApplicationELB
Metric: TargetResponseTime
Dimensions: LoadBalancer=app/hw6-part3-alb/*
Statistic: Average
```

**Healthy Host Count**:
```
Namespace: AWS/ApplicationELB
Metric: HealthyHostCount
Dimensions: TargetGroup=targetgroup/hw6-part3-tg/*, LoadBalancer=app/hw6-part3-alb/*
Statistic: Average
```

---

## 🔍 Monitoring Commands

### Quick Status Check
```bash
# Service health
aws ecs describe-services \
  --cluster hw6-part3-cluster \
  --services hw6-part3-service \
  --query 'services[0].{Desired:desiredCount,Running:runningCount,Pending:pendingCount}' \
  --region us-west-2 \
  --output table

# Target health
aws elbv2 describe-target-health \
  --target-group-arn $(aws elbv2 describe-target-groups \
    --region us-west-2 \
    --query "TargetGroups[?TargetGroupName=='hw6-part3-tg'].TargetGroupArn" \
    --output text) \
  --region us-west-2 \
  --query 'TargetHealthDescriptions[*].{IP:Target.Id,State:TargetHealth.State}' \
  --output table
```

### View Logs
```bash
# Get task ARNs
TASKS=$(aws ecs list-tasks \
  --cluster hw6-part3-cluster \
  --service hw6-part3-service \
  --region us-west-2 \
  --query 'taskArns[0]' \
  --output text)

# View logs (if CloudWatch logging enabled)
aws logs tail /ecs/hw6-part3-task --follow --region us-west-2
```

---

## 🧹 Cleanup

```bash
# Destroy infrastructure
cd part3/
terraform destroy

# Confirm with: yes
```

**Note**: This will delete:
- ECS Service and Tasks
- Application Load Balancer
- Target Groups
- Auto Scaling policies
- VPC resources (if created)
- CloudWatch log groups (if created)

**NOT deleted** (manual cleanup if needed):
- ECR repository and images
- CloudWatch dashboards

---

## 📈 Test Results

See **[PART3_TEST_RESULTS.md](./PART3_TEST_RESULTS.md)** for:
- Complete load test results
- Resilience testing outcomes
- CloudWatch metrics analysis
- Performance comparisons
- Discovery question answers
- Horizontal vs vertical scaling trade-offs

---

## 🎓 Key Learnings

### Why Horizontal Scaling?

1. **High Availability**: No single point of failure
2. **Unlimited Growth**: Add instances as needed
3. **Cost Efficiency**: Auto-scale down when idle
4. **Fault Tolerance**: Individual failures don't affect service
5. **Rolling Updates**: Zero-downtime deployments

### Component Responsibilities

- **ALB**: Traffic distribution, health checking
- **Target Group**: Instance pool management
- **ECS Service**: Task lifecycle, desired state
- **Auto Scaling**: Capacity adjustment based on metrics

### Production Best Practices

1. Set appropriate health check parameters
2. Configure cooldown periods to prevent flapping
3. Choose CPU target based on workload characteristics
4. Set min_capacity high enough for baseline load
5. Set max_capacity with cost limits in mind
6. Monitor CloudWatch metrics continuously
7. Test resilience regularly

---

## 🐛 Troubleshooting

### Targets Unhealthy
```bash
# Check security groups
# - ALB must be able to reach ECS tasks on port 8080
# - Verify security group rules allow ALB → ECS communication

# Check health endpoint
curl http://<alb-dns>/health

# View task logs
aws ecs describe-tasks --cluster hw6-part3-cluster --tasks <task-arn> --region us-west-2
```

### Auto Scaling Not Triggering
```bash
# Verify metrics are being reported
aws cloudwatch get-metric-statistics \
  --namespace AWS/ECS \
  --metric-name CPUUtilization \
  --dimensions Name=ServiceName,Value=hw6-part3-service Name=ClusterName,Value=hw6-part3-cluster \
  --start-time $(date -u -v-10M +%Y-%m-%dT%H:%M:%S) \
  --end-time $(date -u +%Y-%m-%dT%H:%M:%S) \
  --period 60 \
  --statistics Average \
  --region us-west-2

# Check scaling policy
aws application-autoscaling describe-scaling-policies \
  --service-namespace ecs \
  --resource-id service/hw6-part3-cluster/hw6-part3-service \
  --region us-west-2
```

### Load Not Distributing
```bash
# Check all targets are healthy
aws elbv2 describe-target-health --target-group-arn <arn> --region us-west-2

# Verify ALB listener
aws elbv2 describe-listeners --load-balancer-arn <alb-arn> --region us-west-2
```

---

## 📚 Resources

- [AWS ECS Documentation](https://docs.aws.amazon.com/ecs/)
- [Application Load Balancer Guide](https://docs.aws.amazon.com/elasticloadbalancing/latest/application/)
- [ECS Auto Scaling](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/service-auto-scaling.html)
- [Terraform AWS Provider](https://registry.terraform.io/providers/hashicorp/aws/latest/docs)
- [Locust Documentation](https://docs.locust.io/)

---

**Created**: October 2025  
**Course**: CS6650 - Building Scalable Distributed Systems  
**Assignment**: HW6 Part III
