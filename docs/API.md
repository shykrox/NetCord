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
