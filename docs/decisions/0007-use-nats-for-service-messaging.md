# 0007. Use NATS for service messaging

- **Status:** Accepted
- **Date:** 2026-10-30

## Context

Services need to tell each other what happened (for example, "a message was created") without calling each other directly.

## Decision

Use NATS, with JetStream for messages that must be kept, as the event "post office". Event names and payloads are defined in Protobuf. The publishing code sits behind a small interface, so Kafka can replace or join NATS later.

## Why

- Tiny, simple and fast. Runs as one small container.
- Publish/subscribe is easy to understand and to see working.
- JetStream adds storage and replay when we need them.

## Alternatives considered

- **Kafka:** the industry standard for event streams, but heavier. Planned as a later learning module.
- **RabbitMQ:** mature message queue, more concepts to learn up front.
- **Redis Pub/Sub:** simple, but messages are lost if nobody is listening.

## Consequences

One more container to run. Events must be designed carefully, because other services depend on them.
