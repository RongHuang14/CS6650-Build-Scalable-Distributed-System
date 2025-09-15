# Homework 1: Let's Get GOING!

**Name:** Rong Huang  
**Course:** CS6650  

---

## Part I: Local Development with Go

In this part, I built a simple RESTful API using Go and the Gin framework. I ran the server both locally and on Google Cloud Platform. The service provided basic album endpoints and was tested with `curl` commands.

Key points learned:  
- Using `go mod init` to initialize a module and manage dependencies.  
- Using `go get .` to fetch required libraries.  
- Structuring server code as a RESTful API with endpoints.  
- Understanding `localhost` as the loopback address and the difference between local execution and cloud deployment.  
- Practicing lightweight API testing with multiple terminals and `curl`.

---

## Part II: Deploying on AWS EC2

In this part, I set up my AWS Academy Learner Lab account and launched an EC2 instance:  
- **Instance type:** Amazon Linux 2023, `t2.micro`.  
- **Key pair:** Generated RSA `.pem` file for SSH access.  
- **Security group:** Allowed SSH on port 22 and HTTP requests on port 8080.  
- **IAM role:** Configured with `LabInstanceProfile`.  

I cross-compiled my Go server for Linux (`GOOS=linux GOARCH=amd64`) and uploaded the binary to EC2 using `scp`. After SSHing into the instance, I started the server successfully:

```
[GIN-debug] Listening and serving HTTP on 0.0.0.0:8080
```

I then tested the server externally with `curl http://<EC2_PUBLIC_IP>:8080/albums`, confirming that the API was reachable from my local machine.

**Screenshots included:**  
- EC2 instance running in AWS Console.  
- SSH session with server running.  
- Local `curl` command returning JSON from EC2.  

---

## Part III: Performance Testing and Response Times

I wrote a Python script (`load_test.py`) to measure the performance of my EC2-hosted Go server. The script:  
- Sent continuous GET requests for 30 seconds.  
- Recorded response times.  
- Plotted the results in a histogram and scatter plot.  

**Screenshots included:**  
- Load test terminal output (first ~30 requests).  
- Generated histogram and scatter plot of response times.  

### Notes / Observations
- **Distribution Shape:** The histogram showed a narrow cluster of response times, without a heavy long tail.  
- **Consistency:** Response times were very stable (mostly between 35–45ms).  
- **Median vs 95th percentile:** The gap was small, indicating low variability.  
- **Impact of t2.micro:** Since this is a small instance type, the server handled sequential requests fine, but performance may degrade under high concurrency.  
- **Scaling Implications:** With 100 concurrent users, response times would likely increase and tail latency would become more visible.  
- **Network vs Processing:** The delays observed are mostly from network latency, as the server’s processing load is minimal.  

---

## Part IV: Reading Reflection

I read the Introduction and Chapter 1 of *Distributed Systems for Fun and Profit*. One insight I found interesting was how distributed systems are fundamentally about handling **partial failure**. Unlike a single machine, where failure is all-or-nothing, distributed systems must continue operating despite individual component failures. This idea highlights why concepts like replication, fault tolerance, and consensus are so important.

---

## Included Files

- `main.go`, `go.mod`, `go.sum`  
- `hw1server` (compiled binary for EC2)  
- `load_test.py` (Python performance testing script)  
- `README.md` (this file)  
- `screenshots/` folder with Part I & II & III images  

---
