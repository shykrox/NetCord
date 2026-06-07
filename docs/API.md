# NetCord API

Base URL for local development:

```text
http://127.0.0.1:8080
```

All error responses use this shape:

```json
{
  "error": {
    "code": "validation_error",
    "message": "request validation failed",
    "fields": {
      "email": "email must be valid"
    }
  }
}
```

## GET /health

Returns API health.

Response:

```json
{
  "service": "netcord-api",
  "status": "ok"
}
```

## POST /auth/register

Request:

```json
{
  "username": "shykrox",
  "email": "user@example.com",
  "password": "strong-password"
}
```

Response `201 Created`:

```json
{
  "token": "jwt",
  "user": {
    "id": "uuid",
    "username": "shykrox",
    "email": "user@example.com",
    "display_name": "shykrox",
    "avatar_url": null,
    "status": "offline",
    "created_at": "timestamp"
  }
}
```

## POST /auth/login

Request:

```json
{
  "email": "user@example.com",
  "password": "strong-password"
}
```

Response `200 OK`: same shape as register.

Invalid email or password returns `401 Unauthorized` without revealing whether the email exists.

## GET /users/me

Headers:

```text
Authorization: Bearer <token>
```

Response `200 OK`:

```json
{
  "id": "uuid",
  "username": "shykrox",
  "email": "user@example.com",
  "display_name": "shykrox",
  "avatar_url": null,
  "status": "offline",
  "created_at": "timestamp"
}
```

## POST /servers

Requires `Authorization: Bearer <token>`.

Request:

```json
{
  "name": "NetCord",
  "description": "private server"
}
```

Response `201 Created`:

```json
{
  "id": "uuid",
  "owner_id": "uuid",
  "name": "NetCord",
  "description": "private server",
  "icon_url": null,
  "created_at": "timestamp"
}
```

The authenticated creator is inserted into `server_members` with role `owner`.

## GET /servers

Requires `Authorization: Bearer <token>`.

Response `200 OK`:

```json
{
  "servers": []
}
```

Only servers where the authenticated user is a member are returned.

## GET /servers/{server_id}

Requires `Authorization: Bearer <token>`.

Returns the server only if the authenticated user is a member. Non-members receive `404 Not Found`.

## POST /servers/{server_id}/channels

Requires `Authorization: Bearer <token>`.

Request:

```json
{
  "name": "general"
}
```

Response `201 Created`:

```json
{
  "id": "uuid",
  "server_id": "uuid",
  "name": "general",
  "type": "text",
  "position": 0,
  "created_at": "timestamp"
}
```

Only text channels are supported for now.

## GET /servers/{server_id}/channels

Requires `Authorization: Bearer <token>`.

Response `200 OK`:

```json
{
  "channels": []
}
```

## GET /channels/{channel_id}/messages

Requires `Authorization: Bearer <token>`.

Response `200 OK`:

```json
{
  "messages": []
}
```

Only members of the channel's server can read messages.

## POST /channels/{channel_id}/messages

Requires `Authorization: Bearer <token>`.

Request:

```json
{
  "content": "hello"
}
```

Response `201 Created`:

```json
{
  "id": "uuid",
  "server_id": "uuid",
  "channel_id": "uuid",
  "author_id": "uuid",
  "content": "hello",
  "created_at": "timestamp"
}
```
