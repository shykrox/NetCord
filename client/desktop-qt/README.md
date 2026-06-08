# NetCord Desktop Qt

Initial Qt 6/QML desktop client for NetCord.

## Features

- Login and register against the Go backend.
- Stores the JWT locally with `QSettings`.
- Calls `/users/me`, `/servers`, `/servers/{id}/channels`, `/channels/{id}/messages`.
- Sends messages with `POST /channels/{id}/messages`.
- Connects to `/gateway/ws` and receives `message.created` in real time.
- Dark NetCord UI with server rail, channel list, and central chat.

## Requirements

- Qt 6 with `Quick`, `Network`, and `WebSockets` modules.
- CMake 3.21 or newer.
- A C++17 compiler.
- Running NetCord backend, usually at `http://127.0.0.1:8080`.

## Windows Build

From a Qt Developer PowerShell or a shell where Qt and CMake are on `PATH`:

```powershell
cd G:\NetCord\client\desktop-qt
cmake -S . -B build -G "Ninja" -DCMAKE_BUILD_TYPE=Release
cmake --build build --config Release
.\build\NetCordDesktop.exe
```

If you use Visual Studio generators, replace the generator line with the one matching your installed Visual Studio toolchain.

## Linux Build

```bash
cd /path/to/NetCord/client/desktop-qt
cmake -S . -B build -G Ninja -DCMAKE_BUILD_TYPE=Release
cmake --build build
./build/NetCordDesktop
```

## Backend URL

The login screen includes the backend URL field. It defaults to:

```text
http://127.0.0.1:8080
```

The WebSocket gateway URL is derived automatically from that value:

```text
ws://127.0.0.1:8080/gateway/ws
```

## Notes

This first client does not include file upload UI yet. It can display attachment metadata returned by messages, and the backend API already supports file upload/download.
