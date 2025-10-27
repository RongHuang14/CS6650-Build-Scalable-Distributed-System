# CS6650: Building Scalable Distributed Systems

This repository contains my coursework for CS6650 - Building Scalable Distributed Systems at Northeastern University. The course covers practical aspects of designing, implementing, and deploying distributed systems that can handle real-world scale.

I'm keeping this public because distributed systems knowledge has been incredibly valuable for interviews and technical discussions. Each assignment builds on previous concepts, progressing from basic client-server architecture to complex distributed architectures with message queues, auto-scaling, and serverless computing.

## Repository Structure

### [hw1](./hw1) - Getting Started with Go
Built my first REST API using Go and Gin framework. Deployed to AWS EC2 and ran performance tests to understand response time distributions and network latency. This was my introduction to cross-compiling, cloud deployment, and systematic performance measurement.

**Key concepts**: RESTful APIs, EC2 deployment, performance testing, response time analysis

### [hw2](./hw2) - Infrastructure as Code
Learned to automate infrastructure deployment using Terraform. Created reproducible AWS environments and containerized applications with Docker. Discovered why stateless services need external storage by running multiple EC2 instances and observing they don't share state.

**Key concepts**: Terraform, Docker, infrastructure automation, stateless services

### [hw3](./hw3) - Concurrency and Load Testing
Deep dive into Go's concurrency primitives. Compared atomic operations vs regular operations, different synchronization mechanisms (mutex, RWMutex, sync.Map), and buffered vs unbuffered I/O. Used Locust to stress test the API and understand concurrent request handling.

**Key concepts**: Go concurrency, race conditions, synchronization, load testing with Locust

### [hw4](./hw4) - Message Passing and MapReduce
Implemented MapReduce for distributed word counting using RabbitMQ for message passing. Built separate mapper, reducer, and splitter services that communicate asynchronously. Learned how message queues enable decoupled, scalable processing.

**Key concepts**: Message queues, RabbitMQ, MapReduce, asynchronous processing

### [hw5](./hw5) - Cloud-Native Deployment
Built a production-ready product catalog API using FastAPI. Deployed to AWS ECS Fargate with proper VPC networking, security groups, and CloudWatch monitoring. Compared HttpUser vs FastHttpUser in Locust to understand client-side performance bottlenecks.

**Key concepts**: ECS Fargate, VPC networking, container orchestration, API design

### [hw6](./hw6) - Auto Scaling and Performance Optimization
Explored horizontal scaling with Application Load Balancer and ECS auto-scaling policies. Deliberately created bottlenecks with a single instance, then solved them by scaling horizontally. Measured the impact of adding compute capacity on throughput and latency.

**Key concepts**: Auto-scaling, load balancing, performance bottlenecks, CloudWatch metrics

### [hw7](./hw7) - Asynchronous Architecture and Serverless
Built a flash sale system comparing synchronous vs asynchronous architectures. Used SNS/SQS for event-driven order processing with both ECS workers and Lambda functions. Analyzed trade-offs between different processing models and their cost implications.

**Key concepts**: Event-driven architecture, SNS/SQS, Lambda, cold starts, cost optimization

### [midterm](./midterm) - System Reliability Engineering
Diagnosed and fixed reliability issues in a distributed shopping cart service. Implemented circuit breakers, proper error handling, and timeout management. Learned how cascading failures happen and how to prevent them.

**Key concepts**: Circuit breakers, fault tolerance, reliability patterns, cascading failures

## Technologies Used

- **Languages**: Go, Python
- **Cloud**: AWS (EC2, ECS, Lambda, SNS, SQS, CloudWatch)
- **Infrastructure**: Terraform, Docker
- **Message Queues**: RabbitMQ, AWS SQS
- **Testing**: Locust, curl, Postman
- **Frameworks**: Gin (Go), FastAPI (Python)

## Why This Matters

Distributed systems concepts show up everywhere in technical interviews and production systems. Understanding how to:
- Design for failure (circuit breakers, retries, timeouts)
- Scale horizontally vs vertically
- Choose between synchronous and asynchronous processing
- Optimize for cost vs performance
- Measure and improve system reliability

...these skills translate directly to real-world engineering challenges.

## Running the Code

Each homework folder has its own README with setup instructions. Most assignments follow this pattern:

```bash
# Navigate to assignment
cd hw7

# Deploy infrastructure (if applicable)
cd terraform
terraform init
terraform apply

# Run tests
cd tests
pip install -r requirements.txt
locust -f locustfile.py

# Clean up
terraform destroy
```

## Course Information

**Institution**: Northeastern University  
**Course**: CS6650 - Building Scalable Distributed Systems  
**Semester**: Fall 2025

---

Feel free to explore the code and reach out if you have questions about any of the implementations!

