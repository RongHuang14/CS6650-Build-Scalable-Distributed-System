package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
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

func HandleRequest(ctx context.Context, snsEvent events.SNSEvent) error {
	log.Printf("Processing %d SNS records", len(snsEvent.Records))

	for _, record := range snsEvent.Records {
		var order Order
		if err := json.Unmarshal([]byte(record.SNS.Message), &order); err != nil {
			log.Printf("Error unmarshaling order: %v", err)
			return err
		}

		log.Printf("Processing order %s for customer %d", order.OrderID, order.CustomerID)

		// Simulate 3-second payment processing
		time.Sleep(3 * time.Second)

		log.Printf("Order %s completed", order.OrderID)
	}

	return nil
}

func main() {
	lambda.Start(HandleRequest)
}
