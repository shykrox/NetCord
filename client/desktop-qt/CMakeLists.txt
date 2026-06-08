cmake_minimum_required(VERSION 3.21)

project(NetCordDesktop VERSION 0.1.0 LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_AUTOMOC ON)

find_package(Qt6 REQUIRED COMPONENTS Quick Network WebSockets Sql)

if(Qt6_VERSION VERSION_GREATER_EQUAL 6.5)
    qt_policy(SET QTP0001 NEW)
endif()
if(Qt6_VERSION VERSION_GREATER_EQUAL 6.8)
    qt_policy(SET QTP0004 NEW)
endif()

set_source_files_properties(qml/Theme.qml PROPERTIES QT_QML_SINGLETON_TYPE TRUE)

qt_add_executable(NetCordDesktop
    src/main.cpp
    src/netcordclient.cpp
    src/netcordclient.h
)

qt_add_qml_module(NetCordDesktop
    URI NetCord
    VERSION 1.0
    QML_FILES
        qml/Theme.qml
        qml/Main.qml
        qml/LoginView.qml
        qml/AppView.qml
        qml/NcButton.qml
        qml/NcTextField.qml
        qml/NcAvatar.qml
        qml/NcBadge.qml
        qml/NcToast.qml
        qml/ServerIcon.qml
        qml/ChannelItem.qml
        qml/AttachmentCard.qml
        qml/MessageItem.qml
        qml/MemberItem.qml
        qml/VoicePanel.qml
        qml/VoiceChannelItem.qml
        qml/VoiceControls.qml
        qml/VoiceParticipantList.qml
        qml/VoiceDeviceSettings.qml
        qml/CameraPreview.qml
        qml/ScreenSharePicker.qml
)

target_link_libraries(NetCordDesktop
    PRIVATE
        Qt6::Quick
        Qt6::Network
        Qt6::WebSockets
        Qt6::Sql
)

set_target_properties(NetCordDesktop PROPERTIES
    WIN32_EXECUTABLE TRUE
    MACOSX_BUNDLE TRUE
)
