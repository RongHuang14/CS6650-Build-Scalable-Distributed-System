package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
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

type OrderAPI struct {
	snsClient    *sns.SNS
	snsTopicArn  string
	paymentLimit chan struct{} // Buffered channel to limit concurrent payment processing
}

func NewOrderAPI() *OrderAPI {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("AWS_REGION")),
	}))

	return &OrderAPI{
		snsClient:    sns.New(sess),
		snsTopicArn:  os.Getenv("SNS_TOPIC_ARN"),
		paymentLimit: make(chan struct{}, 1), // Limit to 1 concurrent payment (bottleneck!)
	}
}

// Synchronous order processing with payment bottleneck
func (api *OrderAPI) HandleSyncOrder(w http.ResponseWriter, r *http.Request) {
	var order Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Generate order ID
	order.OrderID = uuid.New().String()
	order.Status = "processing"
	order.CreatedAt = time.Now()

	// Simulate payment processing bottleneck using buffered channel
	// This ensures only 1 payment can be processed at a time

	// Use goroutine with timeout to allow queueing while still enforcing timeout
	done := make(chan struct{})

	go func() {
		// Block waiting for token - this allows requests to queue
		api.paymentLimit <- struct{}{}
		time.Sleep(3 * time.Second) // Payment verification takes 3 seconds
		<-api.paymentLimit          // Release token
		close(done)
	}()

	// Wait for completion or timeout
	select {
	case <-done:
		order.Status = "completed"
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Order processed successfully",
			"order":   order,
		})
	case <-time.After(30 * time.Second):
		// Timeout waiting for payment processor
		// 30s allows ~10 orders to queue (normal load should succeed)
		http.Error(w, "Payment processor timeout", http.StatusServiceUnavailable)
	}
}

// Asynchronous order processing
func (api *OrderAPI) HandleAsyncOrder(w http.ResponseWriter, r *http.Request) {
	var order Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Generate order ID
	order.OrderID = uuid.New().String()
	order.Status = "pending"
	order.CreatedAt = time.Now()

	// Publish to SNS
	orderJSON, err := json.Marshal(order)
	if err != nil {
		http.Error(w, "Failed to process order", http.StatusInternalServerError)
		return
	}

	input := &sns.PublishInput{
		TopicArn: aws.String(api.snsTopicArn),
		Message:  aws.String(string(orderJSON)),
		MessageAttributes: map[string]*sns.MessageAttributeValue{
			"order_id": {
				DataType:    aws.String("String"),
				StringValue: aws.String(order.OrderID),
			},
		},
	}

	if _, err := api.snsClient.Publish(input); err != nil {
		log.Printf("Failed to publish to SNS: %v", err)
		http.Error(w, "Failed to queue order", http.StatusInternalServerError)
		return
	}

	// Return immediate response
	w.WriteHeader(http.StatusAccepted)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Order accepted for processing",
		"order":   order,
	})
}

// Health check endpoint
func (api *OrderAPI) HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

// Metrics endpoint for monitoring
func (api *OrderAPI) HandleMetrics(w http.ResponseWriter, r *http.Request) {
	// Simple metrics for monitoring
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"service":   "order-api",
		"timestamp": time.Now().Unix(),
		"status":    "running",
	})
}

func main() {
	api := NewOrderAPI()

	router := mux.NewRouter()
	router.HandleFunc("/orders/sync", api.HandleSyncOrder).Methods("POST")
	router.HandleFunc("/orders/async", api.HandleAsyncOrder).Methods("POST")
	router.HandleFunc("/health", api.HandleHealth).Methods("GET")
	router.HandleFunc("/metrics", api.HandleMetrics).Methods("GET")

	// Logging middleware
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			log.Printf("[%s] %s %s", r.Method, r.URL.Path, r.RemoteAddr)
			next.ServeHTTP(w, r)
			log.Printf("Request processed in %v", time.Since(start))
		})
	})

	port := "8080"
	log.Printf("Order API starting on port %s", port)
	log.Printf("SNS Topic ARN: %s", api.snsTopicArn)

	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatal(err)
	}
}
