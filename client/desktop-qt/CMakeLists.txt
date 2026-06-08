cmake_minimum_required(VERSION 3.21)

project(NetCordDesktop VERSION 0.1.0 LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_AUTOMOC ON)

find_package(Qt6 REQUIRED COMPONENTS Quick Network WebSockets Sql)

qt_add_executable(NetCordDesktop
    src/main.cpp
    src/netcordclient.cpp
    src/netcordclient.h
)

qt_add_qml_module(NetCordDesktop
    URI NetCord
    VERSION 1.0
    QML_FILES
        qml/Main.qml
        qml/LoginView.qml
        qml/AppView.qml
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
