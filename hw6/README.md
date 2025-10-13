# HW6: Performance Testing & Auto Scaling

## Directory Structure
```
hw6/
├── src/                    # Python application
├── Dockerfile              # Container definition
├── locustfile.py           # Load testing
└── terraform/
    ├── part2/              # Single instance
    └── part3/              # Horizontal scaling
```

## Quick Start

### 1. Build & Push Docker Image
```bash
cd hw6
docker build --platform linux/amd64 -t hw6-search .

# Create ECR repo
aws ecr create-repository --repository-name hw6-product-search --region us-west-2

# Login
aws ecr get-login-password --region us-west-2 | \
    docker login --username AWS --password-stdin <ACCOUNT_ID>.dkr.ecr.us-west-2.amazonaws.com

# Push
docker tag hw6-search:latest <ACCOUNT_ID>.dkr.ecr.us-west-2.amazonaws.com/hw6-product-search:latest
docker push <ACCOUNT_ID>.dkr.ecr.us-west-2.amazonaws.com/hw6-product-search:latest
```

### 2. Deploy Part 2
```bash
cd terraform/part2
terraform init
terraform apply

# Get public IP
aws ecs describe-tasks --cluster hw6-part2-cluster --tasks <TASK_ARN> ...

# Test
curl http://<PUBLIC_IP>:8080/health
locust -f locustfile.py --host http://<PUBLIC_IP>:8080
```

### 3. Deploy Part 3
```bash
cd terraform/part3
terraform init
terraform apply

# Get ALB DNS
terraform output alb_dns_name

# Test
curl http://<ALB_DNS>/health
locust -f locustfile.py --host http://<ALB_DNS>
```

## Architecture

### Part 2: Single Instance (Find Bottleneck)
- 1 ECS Fargate instance
- CPU: 256 units, Memory: 512 MB
- Direct public IP access
- No load balancing
- No auto scaling

### Part 3: Horizontal Scaling (Solve Bottleneck)
- 2-4 ECS Fargate instances
- Application Load Balancer
- Auto Scaling (target: 70% CPU)
- High availability

## Testing

### Load Tests
```bash
# 5 users (baseline)
locust -f locustfile.py --host <URL> --users 5 --spawn-rate 1 --run-time 2m

# 20 users (find breaking point)
locust -f locustfile.py --host <URL> --users 20 --spawn-rate 2 --run-time 3m

# 200 users (stress test)
locust -f locustfile.py --host <URL> --users 200 --spawn-rate 10 --run-time 5m
```

### Resilience Test (Part 3 only)
1. Run load test
2. Stop one ECS task in console
3. Observe auto-recovery

## Monitoring
- CloudWatch: CPU, Memory metrics
- ECS Console: Task status
- Target Group: Health checks
- Auto Scaling: Scaling activities

## Cleanup
```bash
cd terraform/part2 && terraform destroy
cd terraform/part3 && terraform destroy
```

## Key Differences

| Aspect | Part 2 | Part 3 |
|--------|--------|--------|
| Instances | 1 | 2-4 |
| Load Balancer | None | ALB |
| Auto Scaling | No | Yes |
| Access | Public IP | ALB DNS |
| HA | No | Yes |

## Common Issues

**No public IP on Part 2**: Check `assign_public_ip = true`

**Unhealthy targets in Part 3**: Verify security groups allow ALB → ECS traffic

**Auto Scaling not triggering**: CPU must exceed 70% for 2 evaluation periods

**Both parts simultaneously**: Yes, they use different VPCs
