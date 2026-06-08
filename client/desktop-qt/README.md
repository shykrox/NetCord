# NetCord Desktop Qt

Initial Qt 6/QML desktop client for NetCord.

## Features

- Login and register against the Go backend.
- Configurable API URL on the login screen.
- Stores the JWT locally with `QSettings`.
- Calls `/users/me`, `/servers`, `/servers/{id}/channels`, `/channels/{id}/messages`.
- Sends messages with `POST /channels/{id}/messages`.
- Uploads files with `POST /files/upload` and attaches uploaded file IDs to messages.
- Displays message attachments returned by the backend.
- Connects to `/gateway/ws` and receives `message.created` in real time.
- Handles `message.updated`, `message.deleted`, typing events, friend/DM events, and AI job progress events.
- Automatic WebSocket reconnect with a visible connected/offline indicator.
- Manual refresh buttons for servers, channels, and messages.
- Create server/channel UI, message search, load older messages, edit/delete own messages.
- Local SQLite cache for servers, channels, and message metadata.
- Basic friends, friend requests, DMs, and AI job status panel.
- Dark NetCord UI with server rail, channel list, and central chat.

## Requirements

- Qt 6.5 or newer recommended.
- Qt modules: `Quick`, `QuickControls2`, `QuickDialogs2`, `Network`, `WebSockets`, `Sql`.
- CMake 3.21 or newer.
- A C++17 compiler.
- Running NetCord backend, usually at `http://127.0.0.1:8080`.

## Qt Creator Build

1. Open Qt Creator.
2. Choose `File > Open File or Project`.
3. Open `client/desktop-qt/CMakeLists.txt`.
4. Select a Qt 6 desktop kit.
   - Windows: Desktop Qt 6.11.1 MinGW 64-bit is validated.
   - Linux: GCC or Clang desktop kit.
5. Configure the project.
6. Build and run `NetCordDesktop`.

If CMake cannot find Qt, set `CMAKE_PREFIX_PATH` to your Qt installation, for example:

```text
C:\Qt\6.7.3\msvc2019_64
```

or on Linux:

```text
/home/you/Qt/6.7.3/gcc_64
```

## Windows CLI Build

From a Qt Developer PowerShell or a shell where Qt, CMake, Ninja, and a compiler are on `PATH`:

```powershell
cd G:\NetCord\client\desktop-qt
set PATH=C:\Qt\Tools\mingw1310_64\bin;C:\Qt\6.11.1\mingw_64\bin;%PATH%
cmake -S . -B build -G "Ninja" -DCMAKE_BUILD_TYPE=Release -DCMAKE_PREFIX_PATH=C:\Qt\6.11.1\mingw_64
cmake --build build --config Release
.\build\NetCordDesktop.exe
```

With an explicit Qt path:

```powershell
cmake -S . -B build -G "Ninja" -DCMAKE_BUILD_TYPE=Release -DCMAKE_PREFIX_PATH=C:\Qt\6.11.1\mingw_64
cmake --build build --config Release
```

## Linux CLI Build

```bash
cd /path/to/NetCord/client/desktop-qt
cmake -S . -B build -G Ninja -DCMAKE_BUILD_TYPE=Release
cmake --build build
./build/NetCordDesktop
```

With an explicit Qt path:

```bash
cmake -S . -B build -G Ninja -DCMAKE_BUILD_TYPE=Release -DCMAKE_PREFIX_PATH="$HOME/Qt/6.7.3/gcc_64"
cmake --build build
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

For HTTPS backends, the client automatically uses `wss://` for the gateway.

## Troubleshooting

- `No CMAKE_CXX_COMPILER could be found`: install/select a C++ compiler kit in Qt Creator, or run from a developer shell.
- `Could not find Qt6`: install Qt 6 with the required modules or pass `-DCMAKE_PREFIX_PATH=...`.
- `Could not find Qt6Sql`: install the Qt SQL module for the selected kit.
- Login/network errors are shown in the app status banner. Confirm the backend is running and the API URL is correct.
- Gateway offline indicator: the client will reconnect automatically, but REST calls still require the backend to be reachable.
