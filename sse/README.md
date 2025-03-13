# Server Sent Events

## SSE Fields

- Message fields are separated by a newline.
- Messages are separated by two newlines.

See the
[docs](https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events/Using_server-sent_events#fields)
for more details.

## Run v1

This is the simplest implementation.

```sh
go run ./cmd/serverv1
```

Now open `indexv1.html`. You should see the server sending the events to the
browser.

## Run v2

This has some nicer events data, leveraging JSON.

```sh
go run ./cmd/serverv2
```

Now open `indexv2.html`. You should see the server sending the events to the
browser.

## Inspecting events and the event source in the web browser

- Go to the network tab.
- Look for the `events` in the left hand side list of request items.
- Inspect the associated properties.
