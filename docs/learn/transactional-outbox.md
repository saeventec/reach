# The transactional outbox

## What it is

A way to save a change in the database **and** announce it on a message server, without ever doing only one of the two.

Analogy: an office's "outbox" tray. You don't run to the post office after writing each letter. You put the letter in the tray, in the same move as filing your copy. A clerk empties the tray regularly and takes the letters to the post office. If the clerk is ill for a day, the letters wait in the tray: none is lost.

In `chat`:

- **The tray** is the `outbox` table in the chat database.
- **The letter** is an event, like "message 42 was sent".
- **The clerk** is the relay, a loop in `chat` ([services/chat/outbox/relay.go](../../services/chat/outbox/relay.go)).
- **The post office** is NATS JetStream.

## Why we use it

Postgres and NATS are two separate systems. No transaction covers both. Look at the two simple orders:

| Order | What goes wrong on a crash in between |
| --- | --- |
| Save the message, then publish | The message exists, but nobody is told. Bob's screen stays empty until he reloads. |
| Publish, then save | Bob is told about a message that doesn't exist, because the save failed. |

With an outbox, the message and its event are written in **one Postgres transaction**. Either both are saved, or neither. Publishing comes later, and is retried until it works. See [ADR 0031](../decisions/0031-transactional-outbox-for-chat-events.md).

## How it works in chat

```
SendMessage ─▶ one transaction: ┌ conversations: last_sequence + 1
                                ├ messages: the new message
                                └ outbox: the MessageSent event   ──commit──▶ wake the relay

relay loop:  read the waiting events (oldest first)
             ─▶ publish one to JetStream, wait for its "ack"
             ─▶ mark it published (published_at = now())
             ─▶ the next one ...
```

1. [services/chat/postgres/message_store.go](../../services/chat/postgres/message_store.go): `SaveMessage` adds the event inside the message's transaction (`saveMessageSentEvent` in [outbox_store.go](../../services/chat/postgres/outbox_store.go)). The event lists the conversation's members, so `realtime` (step 9) knows who to tell.
2. After the commit, the store **wakes the relay** through a Go channel. So the event goes out within milliseconds, not at the next check.
3. The relay also checks **every second**, on its own. That catches events whose wake-up was lost, for example when `chat` crashed right after a commit, or when NATS was down.
4. The relay publishes **in order and stops at the first failure**. So event 43 never overtakes event 42 because 42 failed once.

### At least once, and copies

The relay can publish an event and then crash before marking it published. After a restart, it publishes the event again. So events are delivered **at least once**: maybe twice, never zero times. Two things handle copies:

- **JetStream drops most copies itself.** Each event is published with its outbox id as the message id (`Nats-Msg-Id`). JetStream refuses a second message with the same id within 2 minutes. See [nats.md](nats.md#jetstream-in-chat).
- **Receivers ignore copies anyway.** A copy can come later than 2 minutes. The browser ignores a message id or sequence number it already has.

### The wake-up channel

The channel is `make(chan struct{}, 1)`: a channel with room for one empty value. `struct{}` is a type that holds nothing; only the signal itself matters.

- `chan<- struct{}` (the store's side) can only send. `<-chan struct{}` (the relay's side) can only receive. Go checks this, so neither side uses the channel the wrong way.
- The store sends with `select { case channel <- struct{}{}: default: }`. If a wake-up is already waiting, `default` runs and the send is skipped. That's fine: the relay will read *all* waiting events when it wakes, this one too. So a busy chat never blocks on the channel.

### Which events chat publishes

| Subject | When | What it carries |
| --- | --- | --- |
| `chat.message.sent` | a message was stored | the whole message, and the members |
| `chat.receipt.updated` | a member's delivered or read watermark moved up | both watermarks, and the members |
| `chat.conversation.created` | a group, or a new 1:1 chat, was created (not an existing 1:1 found again) | only the conversation's id, and the members |

The first two carry everything a receiver needs ("event-carried state"). `chat.conversation.created` only says that something happened ("event notification"): the app then asks `ListConversations` for the details, like the usernames, which `chat` doesn't store. Both styles are common. Carrying the data saves a call; a notification keeps the event small and never out of date.

## What we tested

- **Unit tests** ([relay_test.go](../../services/chat/outbox/relay_test.go)) with a fake store and a fake publisher: publishing in order, stopping at a failure, and waking up.
- **Postgres** ([outbox_store_integration_test.go](../../services/chat/postgres/outbox_store_integration_test.go)): a sent message saves one event with the right members; a retried message saves none.
- **NATS** ([publisher_integration_test.go](../../services/chat/natspublisher/publisher_integration_test.go)): an event published twice is stored once. Without the message id, this test fails.
- **All together** ([relay_integration_test.go](../../services/chat/postgres/relay_integration_test.go)): a real Postgres, the relay and a real NATS, wired as in `main.go`. A new group and a message in it arrive on NATS as two events, in that order, and the outbox is empty afterwards.
- **By hand, with the whole system:** messages sent while NATS was stopped for 15 seconds were saved at once. The relay logged an error every 5 seconds and kept trying. When NATS was started again, the waiting events were published, in order, without copies.

## Tiny example

```go
err := store.inTransaction(ctx, func(transactionQueries *queries.Queries) error {
	row, err := transactionQueries.CreateMessage(ctx, parameters)  // the change
	if err != nil {
		return err
	}
	return transactionQueries.CreateOutboxEvent(ctx, event)       // its event, same transaction
})
// Committed: both are saved. Now the relay can publish the event.
```

## Alternatives

- **Publish right after the commit:** the simplest. A crash in between loses the event. (The browser would still catch up on its next fetch, so the message is late, not lost.)
- **Read Postgres's change log (CDC, "change data capture", for example Debezium):** a tool reads every committed change from Postgres's own log and publishes it. No outbox code, but a big extra system to run.
- **Claiming rows with `SELECT ... FOR UPDATE SKIP LOCKED`:** with several copies of `chat`, each relay would take different rows. We don't need it yet: with two copies, both relays may publish the same event, and JetStream drops the copy. A good exercise when we run two copies.
- **Postgres `LISTEN`/`NOTIFY` to wake the relay:** works across several copies of `chat`. The Go channel only wakes the relay in the same program, which is all we need with one copy.

## Later

- Published rows pile up. For now we keep them. Deleting rows older than a few days is a small later job.

## Learn more

- https://microservices.io/patterns/data/transactional-outbox.html
- https://docs.nats.io/nats-concepts/jetstream/streams (the "Duplicate Window" setting)
- https://go.dev/tour/concurrency/5 (select)
