# NetCord WebSocket Gateway

The local gateway endpoint is:

```text
GET /gateway/ws
```

Example URL:

```text
ws://127.0.0.1:8080/gateway/ws
```

## Authentication

The gateway requires a valid JWT for every connection.

Preferred:

```text
Authorization: Bearer <token>
```

Fallback for clients that cannot set WebSocket headers:

```text
ws://127.0.0.1:8080/gateway/ws?token=<jwt>
```

Invalid or missing tokens receive `401 Unauthorized` before the WebSocket upgrade.

## Subscriptions

On connect, the backend loads the authenticated user's current server memberships and subscribes the connection to those servers in the local in-memory gateway hub.

Only messages from servers where the user is a member are broadcast to that connection.

Redis pub/sub is not used yet. This means broadcasts are local to the current API process.

## Event Envelope

All gateway messages are JSON objects with a `type` field and optional `data`.

```json
{
  "type": "event.name",
  "data": {}
}
```

## Server Events

### hello

Sent immediately after a successful connection.

```json
{
  "type": "hello",
  "data": {
    "user_id": "uuid",
    "server_ids": ["uuid"],
    "heartbeat_interval_ms": 30000
  }
}
```

### heartbeat_ack

Sent in response to a client `heartbeat`.

```json
{
  "type": "heartbeat_ack"
}
```

### message.created

Broadcast after `POST /channels/{channel_id}/messages` succeeds.

```json
{
  "type": "message.created",
  "data": {
    "id": "uuid",
    "server_id": "uuid",
    "channel_id": "uuid",
    "author_id": "uuid",
    "content": "hello",
    "attachments": [],
    "created_at": "timestamp"
  }
}
```

### error

Sent for unsupported gateway events or protocol-level issues that can be reported before disconnecting.

```json
{
  "type": "error",
  "data": {
    "code": "unknown_event",
    "message": "unsupported gateway event"
  }
}
```

## Client Events

### heartbeat

Clients should send this at the interval announced in `hello.data.heartbeat_interval_ms`.

```json
{
  "type": "heartbeat"
}
```

The server also uses WebSocket ping/pong frames to detect broken connections.
