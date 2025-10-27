package main

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

type Order struct {
	OrderID    string    `json:"order_id"`
	CustomerID int       `json:"customer_id"`
	Status     string    `json:"status"`
	Items      []Item    `json:"items"`
	CreatedAt  time.Time `json:"created_at"`
}

type Item struct {
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

type OrderProcessor struct {
	sqsClient   *sqs.SQS
	queueURL    string
	workerCount int
	semaphore   chan struct{} // Buffered channel to limit concurrent processing
	wg          sync.WaitGroup
	stats       *ProcessingStats
}

type ProcessingStats struct {
	mu              sync.Mutex
	processedCount  int64
	failedCount     int64
	startTime       time.Time
	lastProcessTime time.Time
}

func NewOrderProcessor() *OrderProcessor {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("AWS_REGION")),
	}))

	// Read worker count from environment variable for Phase 5 testing
	// This controls the number of concurrent goroutines within this single task
	workerCount := 1
	if count := os.Getenv("WORKER_COUNT"); count != "" {
		if parsed, err := strconv.Atoi(count); err == nil {
			workerCount = parsed
			log.Printf("Using WORKER_COUNT from environment: %d", workerCount)
		}
	}

	return &OrderProcessor{
		sqsClient:   sqs.New(sess),
		queueURL:    os.Getenv("SQS_QUEUE_URL"),
		workerCount: workerCount,
		semaphore:   make(chan struct{}, workerCount), // Buffered channel for concurrency control
		stats: &ProcessingStats{
			startTime: time.Now(),
		},
	}
}

func (p *OrderProcessor) Start() {
	log.Printf("Starting Order Processor with %d workers", p.workerCount)
	log.Printf("SQS Queue URL: %s", p.queueURL)

	// Start stats reporter
	go p.reportStats()

	// Start worker goroutines
	for i := 0; i < p.workerCount; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}

	// Wait for all workers to complete (they won't in normal operation)
	p.wg.Wait()
}

func (p *OrderProcessor) worker(id int) {
	defer p.wg.Done()
	log.Printf("Worker %d started", id)

	for {
		// Receive messages from SQS (long polling)
		result, err := p.sqsClient.ReceiveMessage(&sqs.ReceiveMessageInput{
			QueueUrl:            aws.String(p.queueURL),
			MaxNumberOfMessages: aws.Int64(10),
			WaitTimeSeconds:     aws.Int64(20), // Long polling
			VisibilityTimeout:   aws.Int64(30),
		})

		if err != nil {
			log.Printf("Worker %d: Error receiving messages: %v", id, err)
			time.Sleep(5 * time.Second)
			continue
		}

		if len(result.Messages) == 0 {
			log.Printf("Worker %d: No messages received", id)
			continue
		}

		log.Printf("Worker %d: Received %d messages", id, len(result.Messages))

		// Process each message with semaphore-based concurrency control
		for _, msg := range result.Messages {
			// Acquire semaphore slot (blocks if all workers are busy)
			p.semaphore <- struct{}{}

			go func(message *sqs.Message) {
				defer func() {
					<-p.semaphore // Release semaphore slot
				}()
				p.processMessage(id, message)
			}(msg)
		}
	}
}

func (p *OrderProcessor) processMessage(workerID int, message *sqs.Message) {
	startTime := time.Now()

	// Extract SNS message
	var snsMessage struct {
		Message string `json:"Message"`
	}

	if err := json.Unmarshal([]byte(*message.Body), &snsMessage); err != nil {
		log.Printf("Worker %d: Failed to parse SNS message: %v", workerID, err)
		p.deleteMessage(message)
		p.updateStats(false)
		return
	}

	// Parse order from SNS message
	var order Order
	if err := json.Unmarshal([]byte(snsMessage.Message), &order); err != nil {
		log.Printf("Worker %d: Failed to parse order: %v", workerID, err)
		p.deleteMessage(message)
		p.updateStats(false)
		return
	}

	log.Printf("Worker %d: Processing order %s for customer %d", workerID, order.OrderID, order.CustomerID)

	// Simulate payment processing (3 seconds)
	time.Sleep(3 * time.Second)

	// Mark order as completed
	order.Status = "completed"

	log.Printf("Worker %d: Completed order %s in %v", workerID, order.OrderID, time.Since(startTime))

	// Delete message from queue
	p.deleteMessage(message)
	p.updateStats(true)
}

func (p *OrderProcessor) deleteMessage(message *sqs.Message) {
	if _, err := p.sqsClient.DeleteMessage(&sqs.DeleteMessageInput{
		QueueUrl:      aws.String(p.queueURL),
		ReceiptHandle: message.ReceiptHandle,
	}); err != nil {
		log.Printf("Failed to delete message: %v", err)
	}
}

func (p *OrderProcessor) updateStats(success bool) {
	p.stats.mu.Lock()
	defer p.stats.mu.Unlock()

	if success {
		p.stats.processedCount++
	} else {
		p.stats.failedCount++
	}
	p.stats.lastProcessTime = time.Now()
}

func (p *OrderProcessor) reportStats() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		p.stats.mu.Lock()
		uptime := time.Since(p.stats.startTime)
		rate := float64(p.stats.processedCount) / uptime.Seconds()
		log.Printf("=== STATS === Processed: %d | Failed: %d | Rate: %.2f/sec | Workers: %d | Uptime: %v",
			p.stats.processedCount,
			p.stats.failedCount,
			rate,
			p.workerCount,
			uptime.Round(time.Second))
		p.stats.mu.Unlock()
	}
}

func main() {
	processor := NewOrderProcessor()
	processor.Start()
}
