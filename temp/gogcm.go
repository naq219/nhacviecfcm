package main

import (
	"context"
	"log"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

func main() {
	ctx := context.Background()

	opt := option.WithCredentialsFile("serviceAccountKey.json")

	// 🔑 Truyền ProjectID tường minh
	config := &firebase.Config{
		ProjectID: "quan-than", // ← bắt buộc với FCM client
	}

	app, err := firebase.NewApp(ctx, config, opt)
	if err != nil {
		log.Fatalf("error initializing app: %v", err)
	}

	client, err := app.Messaging(ctx)
	if err != nil {
		log.Fatalf("error getting Messaging client: %v", err)
	}

	deviceToken := "fC2NUSQbSdSauB-ADEfAkW:APA91bGkbVOq5BvTywBKodHsYgOHIaZTsX6s7Fc3Rj0wR59OcF7A5RPd-7a4YwUT7cQcdvl4-XIhi6rEPaz67Hk1kpOKf8drsIw-y8VA17SgRS2yQOlCRqw"

	message := &messaging.Message{
		Notification: &messaging.Notification{
			Title: "✅ Test from Go",
			Body:  "ProjectID fixed — now it works!",
		},
		Token: deviceToken,
	}

	response, err := client.Send(ctx, message)
	if err != nil {
		log.Fatalf("error sending message: %v", err)
	}

	log.Printf("✅ Success! Message ID: %s", response)
}
