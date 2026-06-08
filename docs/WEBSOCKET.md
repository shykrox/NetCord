# NetCord WebSocket Gateway

Endpoint:

```text
GET /gateway/ws
```

Auth:

```text
Authorization: Bearer <token>
```

or:

```text
ws://127.0.0.1:8080/gateway/ws?token=<jwt>
```

All messages are JSON:

```json
{"type":"event.name","data":{}}
```

## Client Events

- `heartbeat`
- `typing.start` with `data.channel_id`
- `typing.stop` with `data.channel_id`

The server validates the channel and membership before broadcasting typing events.

## Server Events

- `hello`: includes `user_id`, `server_ids`, `heartbeat_interval_ms`
- `heartbeat_ack`
- `presence.update`
- `typing.start`
- `typing.stop`
- `message.created`
- `message.updated`
- `message.deleted`
- `friend.requested`
- `friend.accepted`
- `dm.message.created`
- `voice.joined`
- `voice.left`
- `voice.state`
- `job.progress`
- `job.completed`
- `ai.job.updated` (reserved alias for job status updates)
- `error`

Subscriptions are in memory:

- Server events are sent to connected users who are members of that server.
- User-private events are sent to all open connections for that user.
- Redis pub/sub is not implemented yet, so broadcasts are local to one API process.
