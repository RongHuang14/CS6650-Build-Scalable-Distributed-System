# Homework 7: When Your Startup's Flash Sale Almost Failed

## Project Structure
```
hw7/
├── src/                    
│   ├── api/                    # Order API Service
│   │   ├── main.go
│   │   └── Dockerfile
│   └── processor/              # Order Processor Service
│       ├── main.go
│       └── Dockerfile
│
├── lambda/                     # Lambda Function
│   ├── handler.go             # Lambda handler code
│   ├── Makefile               # Lambda build script
│   └── function.zip           # Generated deployment package (gitignored)
│
├── terraform/                  # Infrastructure
│   ├── main.tf                # VPC, networking, security groups
│   ├── ecs.tf                 # ECS cluster, tasks, services, ECR repositories
│   ├── messaging.tf           # SNS, SQS configuration
│   ├── lambda.tf              # Lambda function and permissions
│   ├── variables.tf           # Input variables
│   └── outputs.tf             # Output values
│
├── tests/                      # Tests
│   ├── locustfile.py          # Load testing script
│   └── requirements.txt       # Python dependencies
│
├── go.mod                      # Go dependency management
├── go.sum
├── .gitignore
├── README.md                   # Project documentation
└── REPORT.md                   # Assignment report
```

## Quick Start Guide

### 1. Deploy Infrastructure
```bash
cd terraform
terraform init
terraform plan
terraform apply
```

### 2. Build and Deploy Lambda
```bash
cd lambda
make build
make deploy
```

### 3. Run Load Tests
```bash
cd tests
pip install -r requirements.txt
locust -f locustfile.py --host=http://YOUR_ALB_URL
```

## Test Scenarios

### Part II: Synchronous vs Asynchronous Systems

#### Phase 1-2: Synchronous System Bottleneck Test
- **Normal Operation**: 5 concurrent users, 30 seconds
- **Flash Sale Test**: 20 concurrent users, 60 seconds
- **Expected**: Observe synchronous system bottlenecks

#### Phase 3-5: Asynchronous System Test
- **Order Acceptance**: 20 concurrent users, 60 seconds
- **Worker Scaling**: 1, 5, 20, 100 goroutines
- **Expected**: 100% order acceptance rate, observe queue behavior

### Part III: Lambda vs ECS

#### Lambda Cold Start Test
- Send 5-10 test orders
- Observe cold start time in CloudWatch logs
- Analyze cost effectiveness

## Monitoring Metrics
- CloudWatch SQS queue depth (`ApproximateNumberOfMessagesVisible`)
- ECS task resource utilization
- Lambda cold start time (`Init Duration`)
- API response time

## Architecture Comparison

### Synchronous Architecture
```
Customer → API → Payment (3s) → Response
```

### Asynchronous Architecture (ECS)
```
Customer → API → SNS → SQS → ECS Workers → Payment (3s)
```

### Asynchronous Architecture (Lambda)
```
Customer → API → SNS → Lambda → Payment (3s)
```

## Cost Analysis
- **ECS**: 2 tasks × $8.50/month = $17/month
- **Lambda**: Free tier = $0/month (up to 267K orders/month)