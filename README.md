# learn-pub-sub-starter (Peril)

This is the starter code used in Boot.dev's [Learn Pub/Sub](https://learn.boot.dev/learn-pub-sub) course.

# Peril – Pub/Sub Game Simulator

Peril is a multiplayer, event-driven strategy game built to demonstrate **publish/subscribe systems**, **RabbitMQ routing**, **backpressure**, and **horizontal scaling**.  
It was developed as part of the Boot.dev Pub/Sub course.

The project uses **RabbitMQ**, **AMQP**, and **Go generics** to coordinate game state across multiple clients and servers.

---

## Features

- Event-driven architecture using RabbitMQ
- Multiple exchanges and routing strategies:
  - Direct exchange for pause/resume
  - Topic exchange for moves, wars, and logs
- Real-time multiplayer interaction
- Distributed war resolution
- Centralized game logging
- Backpressure demonstration with slow consumers
- Horizontal scaling with multiple server instances
- Dead-letter queues for failed messages
- Gob and JSON serialization

---

## Architecture Overview

### Exchanges

| Exchange        | Type    | Purpose |
|-----------------|---------|--------|
| `peril_direct`  | direct  | Pause / resume messages |
| `peril_topic`   | topic   | Moves, wars, and logs |
| `peril_dlx`     | fanout  | Dead-letter exchange |

### Queues

| Queue        | Type     | Purpose |
|--------------|----------|--------|
| `pause.*`    | transient | Pause updates per client |
| `army_moves.*` | transient | Player move broadcasts |
| `war`        | durable  | Shared war resolution |
| `game_logs`  | durable  | Centralized game logs |
| `peril_dlq`  | durable  | Dead-letter queue |

---

## Requirements

- Go 1.21+
- Docker
- RabbitMQ (provided via Docker)

---

## Getting Started

### 1. Start RabbitMQ

```bash
./rabbit.sh start
````

RabbitMQ UI will be available at:
[http://localhost:15672](http://localhost:15672)
(username: `guest`, password: `guest`)

---

### 2. Run the Server

```bash
go run ./cmd/server
```

The server:

* Subscribes to game logs
* Writes logs to `game.log`
* Provides a REPL for pause/resume commands

---

### 3. Run Clients

Open one or more terminals:

```bash
go run ./cmd/client
```

Choose a username when prompted.

---

## Client Commands

| Command                       | Description                |
| ----------------------------- | -------------------------- |
| `spawn <location> <unit>`     | Spawn a unit               |
| `move <location> <unitID...>` | Move units                 |
| `status`                      | Show current state         |
| `pause`                       | Pause game (server only)   |
| `resume`                      | Resume game (server only)  |
| `spam <n>`                    | Publish `n` malicious logs |
| `help`                        | Show commands              |
| `quit`                        | Exit                       |

---

## Locations

* americas
* europe
* africa
* asia
* antarctica
* australia

## Units

* infantry
* cavalry
* artillery

---

## War System

* Wars are triggered when units overlap
* A shared durable `war` queue ensures **only one client resolves each war**
* Non-involved clients requeue the event
* Results are logged to the `game_logs` queue
* Logs are written to disk by the server

---

## Game Logs

* Logs are published using **Gob serialization**
* Stored in a durable queue (`game_logs`)
* Written to `game.log` by the server
* Backpressure is demonstrated with a 1s artificial delay per log

> `game.log` is ignored by git (`*.log`)

---

## Backpressure Demonstration

### Spam logs

From a client:

```text
spam 10000
```

Watch the `game_logs` queue grow in the RabbitMQ UI.

---

### Scale servers

```bash
./multiserver.sh 100
```

* Starts 100 server instances
* Demonstrates consumer scaling
* Prefetch is limited to 10 messages per consumer
* Throughput approaches ~100 logs/sec

Stop with `Ctrl+C` once the queue drains.

---

## Dead Letter Queue

* Failed messages are routed to `peril_dlx`
* Stored in `peril_dlq`
* Useful for debugging and inspection

---

## Project Structure

```text
cmd/
  client/   # Game client
  server/   # Game server
internal/
  gamelogic/ # Core game rules
  pubsub/    # RabbitMQ helpers
  routing/   # Exchange and routing constants
```

---

## Notes

* Durable queues must be deleted manually if redeclared with new arguments
* RabbitMQ routing keys are central to system behavior
* This project intentionally demonstrates failure modes and recovery

---