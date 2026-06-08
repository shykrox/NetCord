# LiveKit Voice

NetCord uses LiveKit as the planned voice/WebRTC backend.

## Backend

Configure:

```env
NETCORD_LIVEKIT_URL=wss://livekit.example.com
NETCORD_LIVEKIT_API_KEY=...
NETCORD_LIVEKIT_API_SECRET=...
```

Endpoints:

- `POST /voice/join` with `channel_id`
- `POST /voice/leave` with `channel_id`

Join validates:

- authenticated user
- membership in the channel server
- channel type is `voice`
- `CONNECT_VOICE` permission

If LiveKit env is empty, `/voice/join` returns:

```json
{"error":{"code":"voice_not_configured","message":"LiveKit voice is not configured"}}
```

## Client

The Qt client includes voice channel rows, join/leave controls, participant list UI, and device settings placeholders.

Audio transport is scaffold-only. The current client obtains backend state and shows clear configured/not-configured status, but it does not yet connect a native LiveKit Qt/WebRTC audio stack.

## Next Step

Choose a Qt-compatible LiveKit/WebRTC integration path:

- LiveKit native SDK through a C++ wrapper, or
- Qt WebEngine bridge only if acceptable for voice internals, or
- custom WebRTC native module.

Do not mark audio as working until microphone capture, playback, mute, deafen, reconnect, and device selection are tested.
