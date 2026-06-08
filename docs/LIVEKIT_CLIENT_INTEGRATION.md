# LiveKit Native Qt Client Integration

Current state:

- Backend validates membership and permissions.
- Backend returns LiveKit URL, room, and token when env is configured.
- Qt UI has voice channel join, mute/deafen/camera/screen share controls, participants, and device settings.
- Native audio/video transport is not wired yet.

Required implementation steps:

1. Add a C++ `LiveKitClient` wrapper around a Qt-compatible LiveKit/WebRTC native SDK.
2. Expose it to QML as `NetCordVoice`.
3. Implement `connectToRoom(url, token)` and `disconnect()`.
4. Publish microphone audio with mute support.
5. Subscribe to remote audio tracks and route them to the selected output device.
6. Enumerate input/output devices and persist selected IDs in `QSettings`.
7. Add local camera preview via Qt Multimedia, then publish camera track through LiveKit.
8. Add screen/window picker and publish screen share through LiveKit.
9. Add reconnect handling, participant state, and error reporting.
10. Test on Windows MinGW and Linux before marking voice as production-ready.

Do not stream audio, camera, or screen frames through NetCord REST or WebSocket.
