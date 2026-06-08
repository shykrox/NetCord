# NetCord Client UI

The desktop client uses Qt 6/QML, not Electron.

## Layout

```text
Servers | Channels/DMs | Chat | Members/Info
```

Current UI includes:

- login/register with API URL
- persistent JWT with `QSettings`
- server rail
- text and voice channel list
- central chat with attachments
- message edit/delete/search/load older
- typing indicator
- WebSocket status
- file drag-and-drop upload
- right-side members, roles, social, DMs, AI jobs, voice controls
- local SQLite cache

## Voice Components

- `VoiceChannelItem.qml`
- `VoiceControls.qml`
- `VoiceParticipantList.qml`
- `VoiceDeviceSettings.qml`

These are UI/state components. Real LiveKit audio is not implemented yet.

## Roles and Invites

The client can create a basic role and create/join invites. Advanced role assignment UI is still scaffold-level.

## AI

Slash commands `/ask` and `/draw` create backend jobs. The AI jobs panel displays queued/progress events.

## Notifications

The current notification layer is in-app status/toast messages. Native Windows notifications are not wired yet.
