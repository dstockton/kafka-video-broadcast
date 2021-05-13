package video

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
	"nhooyr.io/websocket"
)

// Lovingly borrowed from https://github.com/confluentinc/examples/blob/6.1.1-post/clients/cloud/go/producer.go#L40
func CreateTopic(p *kafka.Producer, topic string) {
	a, err := kafka.NewAdminClientFromProducer(p)
	if err != nil {
		log.Fatalf("Failed to create new admin client from producer: %v", err)
	}
	// Contexts are used to abort or limit the amount of time
	// the Admin call blocks waiting for a result.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// Create topics on cluster.
	// Set Admin options to wait up to 60s for the operation to finish on the remote cluster
	maxDur, err := time.ParseDuration("60s")
	if err != nil {
		log.Fatalf("ParseDuration(60s): %v", err)
	}
	results, err := a.CreateTopics(
		ctx,
		// Multiple topics can be created simultaneously
		// by providing more TopicSpecification structs here.
		[]kafka.TopicSpecification{{
			Topic:             topic,
			NumPartitions:     1,
			ReplicationFactor: 3}},
		// Admin options
		kafka.SetAdminOperationTimeout(maxDur))
	if err != nil {
		log.Fatalf("Admin Client request error: %v", err)
	}
	for _, result := range results {
		if result.Error.Code() != kafka.ErrNoError && result.Error.Code() != kafka.ErrTopicAlreadyExists {
			log.Fatalf("Failed to create topic: %v", result.Error)
		}
		log.Printf("CreateTopic Result: %v", result)
	}
	a.Close()
}

func VideoConnections(w http.ResponseWriter, r *http.Request, bootstrap string, mechanisms string, protocol string, username string, password string) {
	ws, err := websocket.Accept(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer closeWS(ws)
	sessionID := r.URL.Query().Get("sessionID")

	topicName := fmt.Sprintf("video-%s", sessionID)

	kafkaConfig := &kafka.ConfigMap{
		"bootstrap.servers": bootstrap,
		"sasl.mechanisms":   mechanisms,
		"security.protocol": protocol,
		"sasl.username":     username,
		"sasl.password":     password,
		"group.id":          topicName + "_group",
		"auto.offset.reset": "earliest",
	}

	// Create Consumer instance
	c, err := kafka.NewConsumer(kafkaConfig)
	if err != nil {
		log.Fatalf("Failed to create consumer: %v", err)
	}
	err = c.SubscribeTopics([]string{topicName}, nil)
	if err != nil {
		log.Fatalf("Failed to subscribe to topics: %v", err)
	}

	// Create Producer instance
	p, err := kafka.NewProducer(kafkaConfig)
	if err != nil {
		log.Fatalf("Failed to create producer: %v", err)
	}

	// Ensure the topic exists
	CreateTopic(p, topicName)

	ctx := context.Background()
	cctx, cancelFunc := context.WithCancel(ctx)
	go wsLoop(ctx, cancelFunc, ws, p, topicName, sessionID)
	consumerLoop(cctx, ctx, ws, c, topicName, sessionID)
}

func wsLoop(ctx context.Context, cancelFunc context.CancelFunc, ws *websocket.Conn, producer *kafka.Producer, topicName string, sessionID string) {
	log.Printf("Starting wsLoop for %s...", sessionID)

	for {
		if _, message, err := ws.Read(ctx); err != nil {
			// could check for 'close' here and tell peer we have closed
			log.Printf("Error reading message: %v", err)
			break
		} else {
			// Take the websocket message (video bytes) and publish it to Kafka
			producer.Produce(&kafka.Message{
				TopicPartition: kafka.TopicPartition{Topic: &topicName, Partition: kafka.PartitionAny},
				Key:            []byte(sessionID),
				Value:          message,
			}, nil)
		}
	}
	cancelFunc()
	log.Printf("Shutting down wsLoop for %s...", sessionID)
}

func consumerLoop(cctx, ctx context.Context, ws *websocket.Conn, consumer *kafka.Consumer, topicName string, sessionID string) {
	log.Printf("Starting consumerLoop for %s...", sessionID)

	for cctx.Err() != context.Canceled {
		msg, err := consumer.ReadMessage(-1)
		if err == nil {
			if err := ws.Write(ctx, websocket.MessageBinary, msg.Value); err != nil {
				log.Printf("Error writing message to %s: %v", sessionID, err)
			}
		}
	}
}

func closeWS(ws *websocket.Conn) {
	if err := ws.Close(websocket.StatusNormalClosure, ""); err != nil {
		log.Printf("Error closing: %s", err)
	}
}
