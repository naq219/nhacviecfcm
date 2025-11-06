package services

import (
	"context"
	"errors"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

// FCMServiceInterface defines the interface for FCM service
type FCMServiceInterface interface {
	SendNotification(token, title, body string) error
	SendNotificationWithData(token, title, body string, data map[string]string) error
	SendMulticast(tokens []string, title, body string) (*messaging.BatchResponse, error)
}

// FCMService handles Firebase Cloud Messaging
type FCMService struct {
	client *messaging.Client
}

// NewFCMService creates a new FCM service
func NewFCMService(credentialsPath string) (*FCMService, error) {
	ctx := context.Background()

	// Initialize Firebase app
	opt := option.WithCredentialsFile(credentialsPath)
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return nil, err
	}

	// Get messaging client
	client, err := app.Messaging(ctx)
	if err != nil {
		return nil, err
	}

	return &FCMService{client: client}, nil
}

// SendNotification sends a notification to a device
func (s *FCMService) SendNotification(token, title, body string) error {
	if token == "" {
		return errors.New("token is empty")
	}

	message := &messaging.Message{
		Token: token,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Android: &messaging.AndroidConfig{
			Priority: "high",
			Notification: &messaging.AndroidNotification{
				Sound: "default",
			},
		},
		APNS: &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Sound: "default",
				},
			},
		},
	}

	// Send message
	_, err := s.client.Send(context.Background(), message)
	return err
}

// SendNotificationWithData sends a notification with custom data
func (s *FCMService) SendNotificationWithData(token, title, body string, data map[string]string) error {
	if token == "" {
		return errors.New("token is empty")
	}

	message := &messaging.Message{
		Token: token,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: data,
		Android: &messaging.AndroidConfig{
			Priority: "high",
			Notification: &messaging.AndroidNotification{
				Sound: "default",
			},
		},
		APNS: &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Sound: "default",
				},
			},
		},
	}

	_, err := s.client.Send(context.Background(), message)
	return err
}

// SendMulticast sends the same notification to multiple devices
func (s *FCMService) SendMulticast(tokens []string, title, body string) (*messaging.BatchResponse, error) {
	if len(tokens) == 0 {
		return nil, errors.New("no tokens provided")
	}

	message := &messaging.MulticastMessage{
		Tokens: tokens,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Android: &messaging.AndroidConfig{
			Priority: "high",
			Notification: &messaging.AndroidNotification{
				Sound: "default",
			},
		},
		APNS: &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Sound: "default",
				},
			},
		},
	}

	return s.client.SendEachForMulticast(context.Background(), message)
}
