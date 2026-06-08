import QtQuick
import QtQuick.Controls
import QtQuick.Layouts

Rectangle {
    color: "#111318"

    RowLayout {
        anchors.fill: parent
        spacing: 0

        Rectangle {
            Layout.preferredWidth: 72
            Layout.fillHeight: true
            color: "#141720"

            ColumnLayout {
                anchors.fill: parent
                anchors.margins: 8
                spacing: 10

                Rectangle {
                    Layout.preferredWidth: 48
                    Layout.preferredHeight: 48
                    Layout.alignment: Qt.AlignHCenter
                    radius: 14
                    color: "#37d7a7"

                    Label {
                        anchors.centerIn: parent
                        text: "N"
                        color: "#081311"
                        font.bold: true
                        font.pixelSize: 22
                    }
                }

                ListView {
                    id: serverList
                    Layout.fillWidth: true
                    Layout.fillHeight: true
                    spacing: 10
                    clip: true
                    model: netcord.servers

                    delegate: Rectangle {
                        width: 48
                        height: 48
                        radius: 14
                        color: modelData.id === netcord.selectedServer.id ? "#37d7a7" : "#252a35"
                        border.color: "#333947"
                        border.width: 1

                        Label {
                            anchors.centerIn: parent
                            text: (modelData.name || "?").substring(0, 1).toUpperCase()
                            color: modelData.id === netcord.selectedServer.id ? "#081311" : "#e9edf5"
                            font.bold: true
                            font.pixelSize: 18
                        }

                        MouseArea {
                            anchors.fill: parent
                            onClicked: netcord.selectServer(modelData.id)
                        }
                    }
                }

                Button {
                    text: "Out"
                    Layout.fillWidth: true
                    onClicked: netcord.logout()
                }
            }
        }

        Rectangle {
            Layout.preferredWidth: 260
            Layout.fillHeight: true
            color: "#1b1f29"

            ColumnLayout {
                anchors.fill: parent
                anchors.margins: 16
                spacing: 12

                Label {
                    text: netcord.selectedServer.name || "NetCord"
                    color: "#f2f5fb"
                    font.pixelSize: 18
                    font.bold: true
                    elide: Text.ElideRight
                    Layout.fillWidth: true
                }

                Rectangle {
                    Layout.fillWidth: true
                    Layout.preferredHeight: 1
                    color: "#2a303c"
                }

                Label {
                    text: "Text channels"
                    color: "#95a0b2"
                    font.pixelSize: 12
                    font.bold: true
                    Layout.fillWidth: true
                }

                ListView {
                    Layout.fillWidth: true
                    Layout.fillHeight: true
                    clip: true
                    spacing: 4
                    model: netcord.channels

                    delegate: Rectangle {
                        width: ListView.view.width
                        height: 38
                        radius: 6
                        color: modelData.id === netcord.selectedChannel.id ? "#2b3240" : "transparent"

                        RowLayout {
                            anchors.fill: parent
                            anchors.leftMargin: 10
                            anchors.rightMargin: 10
                            spacing: 8

                            Label {
                                text: "#"
                                color: "#6f7b8f"
                                font.pixelSize: 18
                            }

                            Label {
                                text: modelData.name
                                color: modelData.id === netcord.selectedChannel.id ? "#f3f6fb" : "#aab3c2"
                                elide: Text.ElideRight
                                Layout.fillWidth: true
                            }
                        }

                        MouseArea {
                            anchors.fill: parent
                            onClicked: netcord.selectChannel(modelData.id)
                        }
                    }
                }

                Rectangle {
                    Layout.fillWidth: true
                    Layout.preferredHeight: 42
                    radius: 6
                    color: "#232936"

                    RowLayout {
                        anchors.fill: parent
                        anchors.margins: 8
                        spacing: 8

                        Rectangle {
                            Layout.preferredWidth: 28
                            Layout.preferredHeight: 28
                            radius: 9
                            color: "#37d7a7"

                            Label {
                                anchors.centerIn: parent
                                text: (netcord.currentUser.username || "?").substring(0, 1).toUpperCase()
                                color: "#081311"
                                font.bold: true
                            }
                        }

                        Label {
                            text: netcord.currentUser.username || ""
                            color: "#e9edf5"
                            elide: Text.ElideRight
                            Layout.fillWidth: true
                        }
                    }
                }
            }
        }

        Rectangle {
            Layout.fillWidth: true
            Layout.fillHeight: true
            color: "#111318"

            ColumnLayout {
                anchors.fill: parent
                spacing: 0

                Rectangle {
                    Layout.fillWidth: true
                    Layout.preferredHeight: 58
                    color: "#171a22"
                    border.color: "#262b36"
                    border.width: 1

                    RowLayout {
                        anchors.fill: parent
                        anchors.leftMargin: 18
                        anchors.rightMargin: 18
                        spacing: 12

                        Label {
                            text: netcord.selectedChannel.name ? "#" : ""
                            color: "#7c879a"
                            font.pixelSize: 22
                        }

                        Label {
                            text: netcord.selectedChannel.name || "No channel selected"
                            color: "#f3f6fb"
                            font.pixelSize: 18
                            font.bold: true
                            Layout.fillWidth: true
                        }

                        Rectangle {
                            Layout.preferredWidth: 10
                            Layout.preferredHeight: 10
                            radius: 5
                            color: netcord.gatewayConnected ? "#37d7a7" : "#6f7788"
                        }
                    }
                }

                ListView {
                    id: messageList
                    Layout.fillWidth: true
                    Layout.fillHeight: true
                    clip: true
                    spacing: 6
                    model: netcord.messages
                    onCountChanged: Qt.callLater(positionViewAtEnd)

                    delegate: Rectangle {
                        width: messageList.width
                        implicitHeight: messageColumn.implicitHeight + 16
                        color: "transparent"

                        ColumnLayout {
                            id: messageColumn
                            anchors.left: parent.left
                            anchors.right: parent.right
                            anchors.verticalCenter: parent.verticalCenter
                            anchors.leftMargin: 20
                            anchors.rightMargin: 20
                            spacing: 4

                            RowLayout {
                                Layout.fillWidth: true
                                spacing: 8

                                Rectangle {
                                    Layout.preferredWidth: 34
                                    Layout.preferredHeight: 34
                                    radius: 10
                                    color: "#2b3240"

                                    Label {
                                        anchors.centerIn: parent
                                        text: (modelData.author_id || "?").substring(0, 2).toUpperCase()
                                        color: "#c7d0df"
                                        font.pixelSize: 11
                                        font.bold: true
                                    }
                                }

                                ColumnLayout {
                                    Layout.fillWidth: true
                                    spacing: 3

                                    Label {
                                        text: (modelData.author_id || "").substring(0, 8)
                                        color: "#dce2ec"
                                        font.bold: true
                                        elide: Text.ElideRight
                                        Layout.fillWidth: true
                                    }

                                    Label {
                                        text: modelData.content
                                        color: "#eef2f7"
                                        wrapMode: Text.Wrap
                                        Layout.fillWidth: true
                                    }

                                    Repeater {
                                        model: modelData.attachments || []

                                        delegate: Rectangle {
                                            Layout.fillWidth: true
                                            implicitHeight: 34
                                            radius: 6
                                            color: "#202632"
                                            border.color: "#303746"

                                            RowLayout {
                                                anchors.fill: parent
                                                anchors.leftMargin: 10
                                                anchors.rightMargin: 10
                                                spacing: 8

                                                Label {
                                                    text: modelData.original_filename
                                                    color: "#d8e0ea"
                                                    elide: Text.ElideRight
                                                    Layout.fillWidth: true
                                                }

                                                Label {
                                                    text: Math.ceil((modelData.size_bytes || 0) / 1024) + " KB"
                                                    color: "#91a0b5"
                                                }
                                            }
                                        }
                                    }
                                }
                            }
                        }
                    }
                }

                Rectangle {
                    Layout.fillWidth: true
                    Layout.preferredHeight: 84
                    color: "#151820"

                    RowLayout {
                        anchors.fill: parent
                        anchors.margins: 14
                        spacing: 10

                        TextArea {
                            id: composer
                            enabled: !!netcord.selectedChannel.id
                            placeholderText: netcord.selectedChannel.id ? "Message #" + netcord.selectedChannel.name : ""
                            color: "#f2f5fb"
                            placeholderTextColor: "#707b8f"
                            wrapMode: TextArea.Wrap
                            Layout.fillWidth: true
                            Layout.fillHeight: true
                            background: Rectangle {
                                radius: 8
                                color: "#232936"
                                border.color: "#303747"
                            }
                        }

                        Button {
                            text: "Send"
                            enabled: composer.text.trim().length > 0 && !!netcord.selectedChannel.id && !netcord.busy
                            Layout.preferredWidth: 96
                            Layout.fillHeight: true
                            highlighted: true
                            onClicked: {
                                netcord.sendMessage(composer.text)
                                composer.text = ""
                            }
                        }
                    }
                }
            }
        }
    }

    Rectangle {
        anchors.horizontalCenter: parent.horizontalCenter
        anchors.bottom: parent.bottom
        anchors.bottomMargin: 96
        width: Math.min(parent.width - 48, statusText.implicitWidth + 28)
        height: statusText.implicitHeight + 18
        radius: 8
        color: "#30232a"
        border.color: "#6f4450"
        visible: netcord.statusMessage.length > 0

        Label {
            id: statusText
            anchors.centerIn: parent
            text: netcord.statusMessage
            color: "#ffd4d8"
            wrapMode: Text.Wrap
            width: Math.min(520, parent.parent.width - 76)
        }

        MouseArea {
            anchors.fill: parent
            onClicked: netcord.clearStatus()
        }
    }
}
