package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"
)

var kafkaWriter *kafka.Writer

var topic string = func() string {
	t := os.Getenv("KAFKA_TOPIC")
	if t == "" {
		return "orders"
	}
	return t
}()

var brokerAddress string = func() string {
	addr := os.Getenv("KAFKA_BROKER_ADDRESS")
	if addr == "" {
		return "127.0.0.1"
	}
	return addr
}()

var brokerPort int = func() int {
	port, err := strconv.Atoi(os.Getenv("KAFKA_BROKER_PORT"))
	if err != nil {
		return 9092
	}
	return port
}()

func ensureTopics() {
	if topic == "" {
		topic = "orders"
	}
	if brokerAddress == "" {
		brokerAddress = "127.0.0.1"
	}

	var conn *kafka.Conn
	var err error

	// Retry connecting to Kafka for up to 20 seconds
	for i := 0; i < 10; i++ {
		conn, err = kafka.Dial("tcp", brokerAddress+":"+strconv.Itoa(brokerPort))
		if err == nil {
			break
		}
		log.Printf("Failed to dial leader (attempt %d/10): %v. Retrying...", i+1, err)
		time.Sleep(2 * time.Second)
	}
	// log.Printf("Successfully connected to kafka on connection: %s", conn)

	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		log.Printf("Failed to get controler: %v", err)
		return
	}

	// log.Printf("Successfully got the controller: %c", controller)

	controllerConn, err := kafka.Dial("tcp", net.JoinHostPort(brokerAddress, strconv.Itoa(controller.Port)))
	if err != nil {
		log.Printf("Failed to dail controller: %v", err)
		return
	}
	// log.Printf("Successfully got the controller connection: %c", controllerConn)
	defer controllerConn.Close()

	topicConfig := kafka.TopicConfig{
		Topic:             topic,
		NumPartitions:     1,
		ReplicationFactor: 1,
	}
	err = controllerConn.CreateTopics(topicConfig)
	if err != nil {
		log.Printf("Failed to create topic or topic already exists: %v", err)
	} else {
		log.Printf("Topic %s ensured", topic)
	}
}

func orderHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var order Order
	decodeErr := json.NewDecoder(r.Body).Decode(&order)
	if decodeErr != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	product, productErr := store.GetProduct(order.ProductID)
	if productErr != nil {
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}
	log.Printf("Received order for product: %s", product.Name)

	if order.ID == "" {
		order.ID = fmt.Sprintf("order-%d", time.Now().UnixNano())
	}
	order.Status = "PENDING"

	payload, jsonErr := json.Marshal(order)
	if jsonErr != nil {
		log.Printf("Failed to marshal order: %v", jsonErr)
		http.Error(w, "Failed to process order", http.StatusInternalServerError)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	message := kafka.Message{
		Key:   []byte(order.ID),
		Value: payload,
	}

	err := kafkaWriter.WriteMessages(ctx, message)
	if err != nil {
		log.Printf("Failed to write message to Kafka: %v", err)
		http.Error(w, "Failed to process order", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":  "Order placed successfully",
		"order_id": order.ID,
		"product":  product.Name,
	})
}

func startConsumer(ctx context.Context) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{brokerAddress + ":" + strconv.Itoa(brokerPort)},
		Topic:     topic,
		GroupID:   "order-processors",
		Partition: 0,
		MinBytes:  10e3, // 10KB
		MaxBytes:  10e6, // 10MB
	})
	defer reader.Close()

	fmt.Println("Starting Kafka consumer...")

	for {
		m, err := reader.ReadMessage(ctx)
		if err != nil {
			log.Printf("Error reading message: %v", err)
			continue
		}
		log.Printf("Received message: %s", string(m.Value))

		var order Order
		if err = json.Unmarshal(m.Value, &order); err != nil {
			log.Printf("Failed to unmarshal order: %v", err)
			continue
		}

		log.Printf("Processing order ID: %s for product ID: %s Qty=%d\n", order.ID, order.ProductID, order.Quantity)

		result := store.ProcessOrder(&order)

		if result.Success {
			log.Printf("✅ Order %s processed successfully: %s\n", order.ID, result.Message)
		} else {
			log.Printf("❌ Order %s failed: %s\n", order.ID, result.Message)
		}

	}
}

func main() {
	log.Print("Starting Cloud Native Example Application...")
	ensureTopics()
	kafkaWriter = &kafka.Writer{
		Addr:     kafka.TCP(brokerAddress + ":" + strconv.Itoa(brokerPort)),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}

	defer kafkaWriter.Close()
	ctx := context.Background()
	go startConsumer(ctx)

	http.HandleFunc("/order", orderHandler)

	fmt.Println("Starting server on :7880...")
	if err := http.ListenAndServe(":7880", nil); err != nil {
		fmt.Println("Error starting server:", err)
		panic(err)
	}
}
