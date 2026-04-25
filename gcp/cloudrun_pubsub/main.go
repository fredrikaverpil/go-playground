package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/pubsub/v2"
	"cloud.google.com/go/pubsub/v2/apiv1/pubsubpb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

	pubsubClient, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		logger.Error("Failed to create Pub/Sub client", "error", err)
		return
	}
	defer func() {
		if err := pubsubClient.Close(); err != nil {
			logger.Error("Failed to close Pub/Sub client", "error", err)
		}
	}()

	publisher := ensureTopicExists(ctx, pubsubClient, projectID, topicID, logger)
	defer publisher.Stop()

	go subscribeAndLog(ctx, pubsubClient, projectID, topicID, logger)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB
			if err := r.ParseForm(); err != nil {
				http.Error(w, "failed to parse form", http.StatusBadRequest)
				return
			}
			message := r.FormValue("message")
			publishMessage(ctx, publisher, message, logger)
			if _, err := fmt.Fprintln(w, template); err != nil {
				logger.Error("Failed to write response", "error", err)
			}
		}
		if r.Method == http.MethodGet {
			if _, err := fmt.Fprintln(w, template); err != nil {
				logger.Error("Failed to write response", "error", err)
			}
		}
	})
	server := &http.Server{
		Addr:              ":8080",
		ReadHeaderTimeout: 10 * time.Second,
	}
	if err := server.ListenAndServe(); err != nil {
		logger.Error("Server failed", "error", err)
	}
}

func NewLogger() *slog.Logger {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	return logger
}

func ensureTopicExists(
	ctx context.Context,
	client *pubsub.Client,
	projectID string,
	topicID string,
	log *slog.Logger,
) *pubsub.Publisher {
	topicName := fmt.Sprintf("projects/%s/topics/%s", projectID, topicID)
	_, err := client.TopicAdminClient.GetTopic(ctx, &pubsubpb.GetTopicRequest{Topic: topicName})
	if status.Code(err) == codes.NotFound {
		_, err = client.TopicAdminClient.CreateTopic(ctx, &pubsubpb.Topic{Name: topicName})
	}
	if err != nil {
		log.Error("Failed to ensure topic exists", "error", err)
	}
	return client.Publisher(topicName)
}

func publishMessage(ctx context.Context, publisher *pubsub.Publisher, message string, log *slog.Logger) {
	result := publisher.Publish(ctx, &pubsub.Message{Data: []byte(message)})
	_, err := result.Get(ctx)
	if err != nil {
		log.Error("Could not publish message", "error", err)
		return
	}
}

func subscribeAndLog(ctx context.Context, client *pubsub.Client, projectID, topicID string, log *slog.Logger) {
	subID := topicID + "-sub"
	subName := fmt.Sprintf("projects/%s/subscriptions/%s", projectID, subID)
	topicName := fmt.Sprintf("projects/%s/topics/%s", projectID, topicID)

	_, err := client.SubscriptionAdminClient.GetSubscription(
		ctx,
		&pubsubpb.GetSubscriptionRequest{Subscription: subName},
	)
	if status.Code(err) == codes.NotFound {
		_, err = client.SubscriptionAdminClient.CreateSubscription(ctx, &pubsubpb.Subscription{
			Name:  subName,
			Topic: topicName,
		})
	}
	if err != nil {
		log.Error("Failed to ensure subscription exists", "error", err)
		return
	}

	cctx, cancel := context.WithCancel(ctx)
	defer cancel()

	subscriber := client.Subscriber(subName)
	err = subscriber.Receive(cctx, func(_ context.Context, m *pubsub.Message) {
		log.Info("Received message", "data", string(m.Data))
		m.Ack() // Acknowledge that we've consumed the message.
	})
	if err != nil {
		log.Error("Could not subscribe to the topic", "error", err)
		return
	}
}
