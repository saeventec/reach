# NATS and JetStream

## What it is

NATS is a message server: a "post office" between programs. A program **publishes** a message to a **subject** (an address, like `chat.message.sent`), and every program that **subscribed** to that subject gets a copy. The sender doesn't know who listens, and the listeners don't know who sent it.

- **Subject:** a name made of words with dots: `chat.message.sent`. Subscribers can use wildcards: `chat.*.sent` matches one word in that place, and `chat.>` matches everything that starts with `chat.`.
- **Core NATS** delivers a message only to programs listening *right now*. If nobody listens, the message is gone. Fast and simple, like a radio broadcast.
- **JetStream** is built into NATS and adds memory. A **stream** saves every message on certain subjects to disk. A **consumer** reads a stream and remembers how far it got, so a program that was down can catch up afterwards. Like a mailbox instead of a radio.

## Why we use it

`chat` must tell `realtime` "a message was sent", without calling it directly. Then `chat` doesn't need to know how many copies of `realtime` run, or whether any run at all. JetStream keeps the events, so `realtime` can be restarted without missing them. See [ADR 0007](../decisions/0007-use-nats-for-service-messaging.md), and the [milestone 3 design](../design/m03-real-time-chat.md#5-events-nats-jetstream-and-the-transactional-outbox).

## Our setup

In [compose.yaml](../../compose.yaml), NATS runs as one container, started with three options:

| Option | What it does |
| --- | --- |
| `-js` | turns JetStream on |
| `-sd /data` | keeps JetStream's data in `/data`, which is the named volume `nats-data`, so it survives restarts |
| `-m 8222` | serves the monitoring pages, and `/healthz` for Docker's health check |

- Services connect on port **4222**, inside Compose only: `nats://nats:4222`.
- The monitoring pages are published on your computer: open http://localhost:8222. Try `/varz` (the server's settings and counters), `/jsz` (JetStream: streams and how much they store) and `/connz` (who is connected).
- We use the Alpine variant of the image, because it has `wget` for the health check. The default image contains only the server program.
- The container runs as root, as the image does by default. Our own images run as `nobody`; for a ready-made image we keep its defaults (Golden rule 12). On a real server (milestone 11) we look at this again.

## Tiny example

The idea in Go, with the official `nats.go` library, core NATS only (`chat` uses JetStream, see below):

```go
connection, err := nats.Connect("nats://nats:4222")

// realtime: listen to every chat event.
subscription, err := connection.Subscribe("chat.>", func(message *nats.Msg) {
	fmt.Println("got an event on", message.Subject)
})

// chat: announce a new message.
err = connection.Publish("chat.message.sent", eventBytes)
```

## JetStream in chat

`chat` publishes its events with JetStream, in [services/chat/natspublisher/publisher.go](../../services/chat/natspublisher/publisher.go). The library is `nats.go` (v1.55.0), with its `jetstream` package. It replaces an older JetStream API (`nats.JetStreamContext`), which the library calls "legacy".

- **The stream is made by code.** When `chat` starts, `CreateOrUpdateStream` creates the stream `CHAT`, or updates its settings if it exists. Nobody sets it up by hand (Golden rule 8). Look at it at http://localhost:8222/jsz?streams=true.
- **Its settings:** it keeps every subject that starts with `chat.` (`chat.>`), on disk (the default), for **24 hours** (`MaxAge`). Events are only for live updates: the full history is in Postgres. Without a limit, the stream would keep every event forever.
- **The ack.** A JetStream publish waits for an answer: "stored". Only then does the relay mark the event published. With core NATS, a publish is "fire and forget": nobody confirms anything, so the relay couldn't know. If no answer comes within 5 seconds (the library's default), the publish fails and the relay tries again.
- **Dropping copies.** Each event carries a message id, the `Nats-Msg-Id` header: its outbox row id. JetStream remembers the ids of the last **2 minutes** (the default "duplicate window") and drops a second message with a known id. The answer still says "stored" (with `Duplicate: true`), so the relay marks it published as normal.
- **While NATS is down,** the library reconnects on its own and holds new publishes until then. The events wait in the outbox meanwhile. See [the transactional outbox](transactional-outbox.md).

```go
// Publish and wait for the ack. WithMsgID lets JetStream drop a copy.
_, err := jetStream.Publish(ctx, "chat.message.sent", payload, jetstream.WithMsgID("42"))
```

Core NATS subscribers also get every message published to the stream's subjects. So `realtime` can simply subscribe to `chat.>` (step 9).

## Delivery guarantees

- **At most once** (core NATS): a message arrives once or not at all.
- **At least once** (JetStream): a message is sent again until the consumer confirms it ("acknowledges" it). So it can arrive twice, and the receiver must ignore duplicates. Our events carry sequence numbers for that.
- **Exactly once** is very hard across a network. Systems usually fake it with "at least once" plus ignoring duplicates (idempotency).

## Alternatives

- **Kafka:** the industry standard for event streams, kept forever and replayed. Heavier to run. A later learning module (milestone 12).
- **RabbitMQ:** a mature message queue, with more concepts up front (exchanges, bindings).
- **Redis or Valkey Pub/Sub:** simple, but like core NATS, messages are lost if nobody listens.

## Learn more

- https://docs.nats.io/nats-concepts/overview
- https://docs.nats.io/nats-concepts/subjects
- https://docs.nats.io/nats-concepts/jetstream
- https://docs.nats.io/running-a-nats-service/nats_admin/monitoring
