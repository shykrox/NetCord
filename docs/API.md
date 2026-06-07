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
  "messages": [
    {
      "id": "uuid",
      "server_id": "uuid",
      "channel_id": "uuid",
      "author_id": "uuid",
      "content": "hello",
      "attachments": [],
      "created_at": "timestamp"
    }
  ]
}
```

Only members of the channel's server can read messages.

## POST /channels/{channel_id}/messages

Requires `Authorization: Bearer <token>`.

Request:

```json
{
  "content": "hello",
  "attachments": ["attachment-uuid"]
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
  "attachments": [
    {
      "id": "uuid",
      "original_filename": "hello.txt",
      "content_type": "text/plain; charset=utf-8",
      "size_bytes": 12,
      "download_url": "/files/uuid",
      "created_at": "timestamp"
    }
  ],
  "created_at": "timestamp"
}
```

`attachments` is optional. Each attachment ID must come from `POST /files/upload`, must belong to the authenticated user, and must not already be attached to another message.

## POST /files/upload

Requires `Authorization: Bearer <token>`.

Uploads one private attachment object to MinIO using `multipart/form-data`.

Request:

```text
file=<binary file>
```

Response `201 Created`:

```json
{
  "id": "uuid",
  "original_filename": "hello.txt",
  "content_type": "text/plain; charset=utf-8",
  "size_bytes": 12,
  "download_url": "/files/uuid",
  "created_at": "timestamp"
}
```

The API detects MIME type server-side and stores the object using a non-predictable MinIO object key. Bucket and object key are never returned to clients.

## GET /files/{id}

Requires `Authorization: Bearer <token>`.

Downloads an attachment through the API. The authenticated user can access the file if they uploaded it or if it is attached to a message in a server where they are a member.
