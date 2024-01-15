package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"

	"cloud.google.com/go/pubsub"
)

const (
	topicID  = "demo"
	template = `<html><body>
        <form action="/" method="post">
            <input type="text" name="message"/>
            <input type="submit" value="Send"/>
        </form>
    </body></html>`
)

func main() {
	ctx := context.Background()

	logger := NewLogger()

	projectID := os.Getenv("TF_VAR_project_id")
	if projectID == "" {
		logger.Error("TF_VAR_project_id environment variable is not set")
		return
	}

	// Create Pub/Sub client
	pubsubClient, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		logger.Error("Failed to create Pub/Sub client", "error", err)
	}
	defer pubsubClient.Close()

	// Ensure the topic exists
	topic := ensureTopicExists(ctx, pubsubClient, topicID, logger)

	// Start the subscriber in a goroutine
	go subscribeAndLog(ctx, pubsubClient, topicID, logger)

	// Set up web server
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			r.ParseForm()
			message := r.FormValue("message")
			publishMessage(ctx, topic, message, logger)
			fmt.Fprintln(w, template)
		}
		if r.Method == http.MethodGet {
			fmt.Fprintln(w, template)
		}
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func NewLogger() *slog.Logger {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	return logger
}

func ensureTopicExists(ctx context.Context, client *pubsub.Client, topicID string, log *slog.Logger) *pubsub.Topic {
	topic := client.Topic(topicID)
	exists, err := topic.Exists(ctx)
	if err != nil {
		log.Error("Error checking if topic exists", "error", err)
	}
	if !exists {
		_, err := client.CreateTopic(ctx, topicID)
		if err != nil {
			log.Error("Failed to create the topic", "error", err)
		}
	}
	return topic
}

func publishMessage(ctx context.Context, topic *pubsub.Topic, message string, log *slog.Logger) {
	result := topic.Publish(ctx, &pubsub.Message{Data: []byte(message)})
	_, err := result.Get(ctx)
	if err != nil {
		log.Error("Could not publish message", "error", err)
		return
	}
}

func subscribeAndLog(ctx context.Context, client *pubsub.Client, topicID string, log *slog.Logger) {
	subID := topicID + "-sub"
	sub := client.Subscription(subID)

	exists, err := sub.Exists(ctx)
	if err != nil {
		log.Error("Error checking if subscription exists", "error", err)
		return
	}

	if !exists {
		_, err := client.CreateSubscription(ctx, subID, pubsub.SubscriptionConfig{Topic: client.Topic(topicID)})
		if err != nil {
			log.Error("Failed to create the subscription", "error", err)
			return
		}
	}

	cctx, cancel := context.WithCancel(ctx)
	defer cancel()

	err = sub.Receive(cctx, func(ctx context.Context, m *pubsub.Message) {
		log.Info("Received message", "data", string(m.Data))
		m.Ack() // Acknowledge that we've consumed the message.
	})
	if err != nil {
		log.Error("Could not subscribe to the topic", "error", err)
		return
	}
}
