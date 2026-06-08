import QtQuick
import QtQuick.Controls
import QtQuick.Dialogs
import QtQuick.Layouts

Rectangle {
    id: appRoot
    color: "#111318"
    property string editMessageId: ""

    FileDialog {
        id: uploadDialog
        title: "Attach file"
        onAccepted: netcord.uploadFile(selectedFile)
    }

    Dialog {
        id: createServerDialog
        title: "Create server"
        modal: true
        standardButtons: Dialog.Ok | Dialog.Cancel
        anchors.centerIn: parent
        onAccepted: {
            netcord.createServer(serverName.text, serverDescription.text)
            serverName.text = ""
            serverDescription.text = ""
        }

        ColumnLayout {
            spacing: 10
            TextField { id: serverName; placeholderText: "Server name"; Layout.preferredWidth: 320 }
            TextField { id: serverDescription; placeholderText: "Description" }
        }
    }

    Dialog {
        id: createChannelDialog
        title: "Create channel"
        modal: true
        standardButtons: Dialog.Ok | Dialog.Cancel
        anchors.centerIn: parent
        onAccepted: {
            netcord.createChannel(channelName.text, channelType.currentText)
            channelName.text = ""
        }

        ColumnLayout {
            spacing: 10
            TextField { id: channelName; placeholderText: "Channel name"; Layout.preferredWidth: 320 }
            ComboBox { id: channelType; model: ["text", "voice"] }
        }
    }

    Dialog {
        id: editMessageDialog
        title: "Edit message"
        modal: true
        standardButtons: Dialog.Ok | Dialog.Cancel
        anchors.centerIn: parent
        onAccepted: netcord.editMessage(appRoot.editMessageId, editMessageText.text)

        TextArea {
            id: editMessageText
            Layout.preferredWidth: 420
            Layout.preferredHeight: 120
            wrapMode: TextArea.Wrap
        }
    }

    Dialog {
        id: profileDialog
        title: "User settings"
        modal: true
        standardButtons: Dialog.Ok | Dialog.Cancel
        anchors.centerIn: parent
        onAccepted: netcord.updateProfile(profileName.text, profileStatus.currentText)

        ColumnLayout {
            spacing: 10
            TextField {
                id: profileName
                text: netcord.currentUser.display_name || netcord.currentUser.username || ""
                placeholderText: "Display name"
                Layout.preferredWidth: 320
            }
            ComboBox {
                id: profileStatus
                model: ["online", "idle", "dnd", "offline"]
            }
            Button {
                text: "Clear local cache"
                onClicked: netcord.clearCache()
            }
        }
    }

    Dialog {
        id: roleDialog
        title: "Create role"
        modal: true
        standardButtons: Dialog.Ok | Dialog.Cancel
        anchors.centerIn: parent
        onAccepted: netcord.createRole(roleName.text, roleAdmin.checked ? 524288 : 0)

        ColumnLayout {
            spacing: 10
            TextField { id: roleName; placeholderText: "Role name"; Layout.preferredWidth: 320 }
            CheckBox { id: roleAdmin; text: "Administrator" }
        }
    }

    Dialog {
        id: joinInviteDialog
        title: "Join invite"
        modal: true
        standardButtons: Dialog.Ok | Dialog.Cancel
        anchors.centerIn: parent
        onAccepted: netcord.joinInvite(inviteCode.text)

        TextField {
            id: inviteCode
            placeholderText: "Invite code"
            Layout.preferredWidth: 320
        }
    }

    DropArea {
        anchors.fill: parent
        onDropped: function(drop) {
            if (drop.urls.length > 0) {
                netcord.uploadFile(drop.urls[0])
            }
        }
    }

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

                Button {
                    text: "+"
                    Layout.fillWidth: true
                    onClicked: createServerDialog.open()
                }

                Button {
                    text: "Refresh"
                    Layout.fillWidth: true
                    enabled: !netcord.busy
                    onClicked: netcord.refreshServers()
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
                    text: "Logout"
                    Layout.fillWidth: true
                    onClicked: netcord.logout()
                }

                Button {
                    text: "Settings"
                    Layout.fillWidth: true
                    onClicked: profileDialog.open()
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

                RowLayout {
                    Layout.fillWidth: true

                    Label {
                        text: netcord.selectedServer.name || "NetCord"
                        color: "#f2f5fb"
                        font.pixelSize: 18
                        font.bold: true
                        elide: Text.ElideRight
                        Layout.fillWidth: true
                    }

                    Button {
                        text: "+"
                        enabled: !!netcord.selectedServer.id
                        onClicked: createChannelDialog.open()
                    }

                    Button {
                        text: "Refresh"
                        enabled: !!netcord.selectedServer.id && !netcord.busy
                        onClicked: netcord.refreshChannels()
                    }
                }

                Rectangle {
                    Layout.fillWidth: true
                    Layout.preferredHeight: 1
                    color: "#2a303c"
                }

                Label {
                    text: "Channels"
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

                    delegate: Loader {
                        width: ListView.view.width
                        height: 38
                        sourceComponent: modelData.type === "voice" ? voiceChannelDelegate : textChannelDelegate
                        property var channelData: modelData
                    }

                    Component {
                        id: textChannelDelegate
                        Rectangle {
                            width: parent.width
                            height: 38
                            radius: 6
                            color: channelData.id === netcord.selectedChannel.id ? "#2b3240" : "transparent"

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
                                    text: channelData.name
                                    color: channelData.id === netcord.selectedChannel.id ? "#f3f6fb" : "#aab3c2"
                                    elide: Text.ElideRight
                                    Layout.fillWidth: true
                                }
                            }

                            MouseArea {
                                anchors.fill: parent
                                onClicked: netcord.selectChannel(channelData.id)
                            }
                        }
                    }

                    Component {
                        id: voiceChannelDelegate
                        VoiceChannelItem {
                            width: parent.width
                            channel: channelData
                            onJoinRequested: function(channelId) { netcord.joinVoice(channelId) }
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

                        Label {
                            text: netcord.gatewayConnected ? "Live" : "Offline"
                            color: netcord.gatewayConnected ? "#9af0d4" : "#a8b0bf"
                            font.pixelSize: 12
                        }

                        Button {
                            text: "Reconnect"
                            enabled: netcord.authenticated
                            onClicked: netcord.reconnectGateway()
                        }

                        Button {
                            text: "Refresh"
                            enabled: !!netcord.selectedChannel.id && !netcord.busy
                            onClicked: netcord.refreshMessages()
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
                    onCountChanged: Qt.callLater(function() { messageList.positionViewAtEnd() })
                    onHeightChanged: Qt.callLater(function() { messageList.positionViewAtEnd() })

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
                                        text: modelData.content || ""
                                        visible: text.length > 0
                                        color: "#eef2f7"
                                        wrapMode: Text.Wrap
                                        Layout.fillWidth: true
                                    }

                                    RowLayout {
                                        spacing: 6

                                        Button {
                                            text: "Edit"
                                            visible: modelData.author_id === netcord.currentUser.id
                                            Layout.preferredHeight: 26
                                            onClicked: {
                                                appRoot.editMessageId = modelData.id
                                                editMessageText.text = modelData.content || ""
                                                editMessageDialog.open()
                                            }
                                        }

                                        Button {
                                            text: "Delete"
                                            visible: modelData.author_id === netcord.currentUser.id
                                            Layout.preferredHeight: 26
                                            onClicked: netcord.deleteMessage(modelData.id)
                                        }

                                        Label {
                                            text: modelData.edited_at ? "edited" : ""
                                            color: "#788498"
                                            font.pixelSize: 11
                                        }
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

                                            MouseArea {
                                                anchors.fill: parent
                                                onClicked: netcord.openAttachment(modelData.download_url)
                                            }
                                        }
                                    }
                                }
                            }
                        }
                    }
                }

                Rectangle {
                    visible: netcord.searchResults.length > 0
                    Layout.fillWidth: true
                    Layout.preferredHeight: visible ? 96 : 0
                    color: "#181c25"
                    border.color: "#2a303c"

                    ListView {
                        anchors.fill: parent
                        anchors.margins: 8
                        clip: true
                        model: netcord.searchResults
                        delegate: Label {
                            width: ListView.view.width
                            text: (modelData.created_at || "").substring(0, 19) + "  " + (modelData.content || "")
                            color: "#d8e1ee"
                            elide: Text.ElideRight
                        }
                    }
                }

                Rectangle {
                    Layout.fillWidth: true
                    Layout.preferredHeight: netcord.pendingAttachments.length > 0 ? 168 : 114
                    color: "#151820"

                    ColumnLayout {
                        anchors.fill: parent
                        anchors.margins: 14
                        spacing: 8

                        ListView {
                            visible: netcord.pendingAttachments.length > 0
                            Layout.fillWidth: true
                            Layout.preferredHeight: visible ? 36 : 0
                            orientation: ListView.Horizontal
                            spacing: 8
                            clip: true
                            model: netcord.pendingAttachments

                            delegate: Rectangle {
                                width: Math.min(240, attachmentLabel.implicitWidth + 54)
                                height: 32
                                radius: 6
                                color: "#232936"
                                border.color: "#394354"

                                RowLayout {
                                    anchors.fill: parent
                                    anchors.leftMargin: 10
                                    anchors.rightMargin: 8
                                    spacing: 6

                                    Label {
                                        id: attachmentLabel
                                        text: modelData.original_filename || "file"
                                        color: "#dce4ef"
                                        elide: Text.ElideRight
                                        Layout.fillWidth: true
                                    }

                                    Button {
                                        text: "X"
                                        Layout.preferredWidth: 30
                                        Layout.preferredHeight: 24
                                        onClicked: netcord.removePendingAttachment(modelData.id)
                                    }
                                }
                            }
                        }

                        RowLayout {
                            Layout.fillWidth: true
                            spacing: 8

                            Button {
                                text: "Older"
                                enabled: netcord.messages.length > 0 && !netcord.busy
                                onClicked: netcord.loadOlderMessages()
                            }

                            TextField {
                                id: searchBox
                                placeholderText: "Search current channel"
                                Layout.fillWidth: true
                                onAccepted: netcord.searchMessages(text)
                            }

                            Button {
                                text: "Search"
                                enabled: searchBox.text.trim().length > 0
                                onClicked: netcord.searchMessages(searchBox.text)
                            }

                            Button {
                                text: "Clear"
                                onClicked: {
                                    searchBox.text = ""
                                    netcord.clearSearchResults()
                                }
                            }
                        }

                        Label {
                            text: netcord.typingText
                            visible: text.length > 0
                            color: "#91a0b5"
                            Layout.fillWidth: true
                        }

                        RowLayout {
                            Layout.fillWidth: true
                            Layout.fillHeight: true
                            spacing: 10

                            Button {
                                text: "Attach"
                                enabled: !!netcord.selectedChannel.id && !netcord.busy
                                Layout.preferredWidth: 84
                                Layout.fillHeight: true
                                onClicked: uploadDialog.open()
                            }

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
                                onTextChanged: {
                                    if (text.trim().length > 0) {
                                        netcord.sendTypingStart()
                                    }
                                }
                            }

                            Button {
                                text: "Send"
                                enabled: (composer.text.trim().length > 0 || netcord.pendingAttachments.length > 0) && !!netcord.selectedChannel.id && !netcord.busy
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
            Layout.preferredWidth: 280
            Layout.fillHeight: true
            color: "#171a22"
            border.color: "#262b36"

            ColumnLayout {
                anchors.fill: parent
                anchors.margins: 14
                spacing: 10

                RowLayout {
                    Layout.fillWidth: true
                    Label {
                        text: "Info"
                        color: "#f3f6fb"
                        font.bold: true
                        Layout.fillWidth: true
                    }
                    Button {
                        text: "Refresh"
                        onClicked: {
                            netcord.loadFriends()
                            netcord.loadFriendRequests()
                            netcord.loadDMs()
                        }
                    }
                }

                VoiceControls {
                    Layout.fillWidth: true
                    connected: netcord.voiceConnected
                    status: netcord.voiceStatus
                    onLeaveRequested: netcord.leaveVoice()
                }

                RowLayout {
                    Layout.fillWidth: true
                    Button {
                        text: "Invite"
                        enabled: !!netcord.selectedServer.id
                        onClicked: netcord.createInvite()
                    }
                    Button {
                        text: "Join"
                        onClicked: joinInviteDialog.open()
                    }
                    Button {
                        text: "Role"
                        enabled: !!netcord.selectedServer.id
                        onClicked: roleDialog.open()
                    }
                }

                Label { text: "Members"; color: "#95a0b2"; font.bold: true }
                ListView {
                    Layout.fillWidth: true
                    Layout.preferredHeight: 120
                    clip: true
                    model: netcord.serverMembers
                    delegate: RowLayout {
                        width: ListView.view.width
                        Rectangle {
                            Layout.preferredWidth: 8
                            Layout.preferredHeight: 8
                            radius: 4
                            color: modelData.status === "online" ? "#37d7a7" : "#6f7788"
                        }
                        Label {
                            text: modelData.username || modelData.user_id
                            color: "#d8e1ee"
                            elide: Text.ElideRight
                            Layout.fillWidth: true
                        }
                    }
                }

                Label { text: "Roles"; color: "#95a0b2"; font.bold: true }
                ListView {
                    Layout.fillWidth: true
                    Layout.preferredHeight: 80
                    clip: true
                    model: netcord.roles
                    delegate: Label {
                        width: ListView.view.width
                        text: modelData.name || "role"
                        color: "#d8e1ee"
                        elide: Text.ElideRight
                    }
                }

                TextField {
                    id: friendIdField
                    Layout.fillWidth: true
                    placeholderText: "User UUID"
                    onAccepted: netcord.sendFriendRequest(text)
                }

                RowLayout {
                    Layout.fillWidth: true
                    Button {
                        text: "Add"
                        enabled: friendIdField.text.trim().length > 0
                        onClicked: netcord.sendFriendRequest(friendIdField.text)
                    }
                    Button {
                        text: "DM"
                        enabled: friendIdField.text.trim().length > 0
                        onClicked: netcord.createDM(friendIdField.text)
                    }
                }

                Label { text: "Requests"; color: "#95a0b2"; font.bold: true }
                ListView {
                    Layout.fillWidth: true
                    Layout.preferredHeight: 96
                    clip: true
                    model: netcord.friendRequests
                    delegate: RowLayout {
                        width: ListView.view.width
                        Label {
                            text: (modelData.requester_id || modelData.recipient_id || "").substring(0, 8)
                            color: "#d8e1ee"
                            Layout.fillWidth: true
                        }
                        Button {
                            text: "OK"
                            visible: modelData.recipient_id === netcord.currentUser.id
                            onClicked: netcord.acceptFriendRequest(modelData.id)
                        }
                        Button {
                            text: "No"
                            visible: modelData.recipient_id === netcord.currentUser.id
                            onClicked: netcord.declineFriendRequest(modelData.id)
                        }
                    }
                }

                Label { text: "Friends"; color: "#95a0b2"; font.bold: true }
                ListView {
                    Layout.fillWidth: true
                    Layout.preferredHeight: 110
                    clip: true
                    model: netcord.friends
                    delegate: Label {
                        width: ListView.view.width
                        text: modelData.username || modelData.user_id
                        color: "#d8e1ee"
                        elide: Text.ElideRight
                    }
                }

                Label { text: "DMs"; color: "#95a0b2"; font.bold: true }
                ListView {
                    Layout.fillWidth: true
                    Layout.preferredHeight: 90
                    clip: true
                    model: netcord.dmConversations
                    delegate: Label {
                        width: ListView.view.width
                        text: (modelData.name || modelData.type || "dm") + "  " + (modelData.id || "").substring(0, 8)
                        color: "#d8e1ee"
                        elide: Text.ElideRight
                    }
                }

                Label { text: "AI jobs"; color: "#95a0b2"; font.bold: true }
                ListView {
                    Layout.fillWidth: true
                    Layout.fillHeight: true
                    clip: true
                    model: netcord.aiJobs
                    delegate: Label {
                        width: ListView.view.width
                        text: (modelData.command || "job") + " - " + (modelData.status || "")
                        color: "#d8e1ee"
                        elide: Text.ElideRight
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
