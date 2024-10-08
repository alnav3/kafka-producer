package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/a-h/templ"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

var producer *kafka.Producer
var err error

func main() {
	// Kafka producer configuration
	producer, err = kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": "localhost:9092"})
	if err != nil {
		log.Fatalf("Failed to create Kafka producer: %s\n", err)
	}
	defer producer.Close()

	http.Handle("/", templ.Handler(indexPage()))
	http.HandleFunc("/produce", handleProduce)

	fmt.Println("Starting server on port 42069")
	if err := http.ListenAndServe(":42069", nil); err != nil {
		fmt.Println("Failed to start server:", err)
	}
}

func handleProduce(w http.ResponseWriter, r *http.Request) {
	topic := r.FormValue("topic")
    message := r.FormValue("message")
	if topic == "" || message == "" {
		http.Error(w, "Missing data in request", http.StatusBadRequest)
		return
	}

	produceMessage(topic, message, w)
}
func removeNewlinesAndExtraSpaces(input []byte) []byte {
	// Remove newlines and carriage returns
	input = bytes.ReplaceAll(input, []byte("\n"), []byte(""))
	input = bytes.ReplaceAll(input, []byte("\r"), []byte(""))

	// Replace multiple spaces with a single space
	input = bytes.Join(bytes.Fields(input), []byte(" "))

	return input
}

func produceMessage(topic string, body string, w http.ResponseWriter) {
	// Parse JSON body into a map
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(body), &data); err != nil {
		http.Error(w, "Invalid JSON in request body", http.StatusBadRequest)
		return
	}

	// Convert the map to a JSON string
	message, err := json.Marshal(data)
	if err != nil {
		http.Error(w, "Failed to serialize data", http.StatusInternalServerError)
		return
	}

	// Produce Kafka message
	kafkaMessage := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          message,
	}
	err = producer.Produce(kafkaMessage, nil)
	if err != nil {
		http.Error(w, "Failed to produce Kafka message", http.StatusInternalServerError)
		return
	}

	// Wait for message to be delivered
	e := <-producer.Events()
	m := e.(*kafka.Message)
	if m.TopicPartition.Error != nil {
		http.Error(w, "Delivery failed: "+m.TopicPartition.Error.Error(), http.StatusInternalServerError)
	} else {
		fmt.Fprintf(w, "Message delivered to topic %s [%d] at offset %v\n",
			*m.TopicPartition.Topic, m.TopicPartition.Partition, m.TopicPartition.Offset)
	}
}
