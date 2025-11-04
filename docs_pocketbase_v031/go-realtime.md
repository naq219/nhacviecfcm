# PocketBase Go Realtime Messaging

This document explains how to send custom realtime messages to connected clients in PocketBase using Go.

## Overview

By default, PocketBase sends realtime events for record create, update, and delete operations. However, you can also send custom messages to clients using the `app.SubscriptionsBroker()`.

### Key Components

- **`app.SubscriptionsBroker()`:** Provides access to the realtime message broker.
- **`Clients()`:** Returns all connected clients, indexed by their unique connection ID.
- **`ChunkedClients(size)`:** Returns clients in chunks, allowing for concurrent processing.
- **`client.Get(apis.RealtimeClientAuthKey)`:** Retrieves the authenticated user associated with a client.

---

## Sending Custom Messages

You can send a JSON payload to all clients subscribed to a specific topic. The following example demonstrates how to send a message to the "example" topic.

### Go Example

```go
import (
    "encoding/json"
    "golang.org/x/sync/errgroup"
    "github.com/pocketbase/pocketbase/core"
    "github.com/pocketbase/pocketbase/tools/subscriptions"
)

func notify(app core.App, subscription string, data any) error {
    rawData, err := json.Marshal(data)
    if err != nil {
        return err
    }

    message := subscriptions.Message{
        Name: subscription,
        Data: rawData,
    }

    group := new(errgroup.Group)
    chunks := app.SubscriptionsBroker().ChunkedClients(300)

    for _, chunk := range chunks {
        group.Go(func() error {
            for _, client := range chunk {
                if client.HasSubscription(subscription) {
                    client.Send(message)
                }
            }
            return nil
        })
    }

    return group.Wait()
}

// Usage
err := notify(app, "example", map[string]any{"test": 123})
if err != nil {
    return err
}
```

---

## Client-Side Subscription

Clients can subscribe to custom topics to receive messages.

### JavaScript Example

```javascript
import PocketBase from 'pocketbase';

const pb = new PocketBase('http://127.0.0.1:8090');

await pb.realtime.subscribe('example', (e) => {
    console.log(e);
});
```

### Dart Example

```dart
import 'package:pocketbase/pocketbase.dart';

final pb = PocketBase('http://127.0.0.1:8090');

await pb.realtime.subscribe('example', (e) {
    print(e);
});
```