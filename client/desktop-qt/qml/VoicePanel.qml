import QtQuick
import QtQuick.Controls
import QtQuick.Layouts

Rectangle {
    id: root
    property bool connected: false
    property string status: ""
    property var participants: []
    signal leaveRequested()

    property bool muted: false
    property bool deafened: false
    property bool cameraOn: false

    radius: Theme.radiusMd
    color: Theme.bg2
    border.color: connected ? Theme.accent : Theme.border
    implicitHeight: content.implicitHeight + 20

    ColumnLayout {
        id: content
        anchors.fill: parent
        anchors.margins: 10
        spacing: 10

        RowLayout {
            Layout.fillWidth: true
            Rectangle {
                Layout.preferredWidth: 10
                Layout.preferredHeight: 10
                radius: 5
                color: connected ? Theme.accent : Theme.textSubtle
            }
            Label {
                text: status.length > 0 ? status : (connected ? "Voice connected" : "Not connected")
                color: Theme.text
                elide: Text.ElideRight
                Layout.fillWidth: true
            }
        }

        RowLayout {
            Layout.fillWidth: true
            NcButton {
                text: root.muted ? "Unmute" : "Mute"
                enabled: connected
                onClicked: root.muted = !root.muted
            }
            NcButton {
                text: root.deafened ? "Undeafen" : "Deafen"
                enabled: connected
                onClicked: root.deafened = !root.deafened
            }
            NcButton {
                text: root.cameraOn ? "Cam off" : "Cam"
                enabled: connected
                onClicked: root.cameraOn = !root.cameraOn
            }
            NcButton {
                text: "Leave"
                danger: true
                enabled: connected
                onClicked: root.leaveRequested()
            }
        }

        CameraPreview {
            Layout.fillWidth: true
            active: root.cameraOn
        }

        ScreenSharePicker {
            Layout.fillWidth: true
            enabled: connected
        }

        VoiceDeviceSettings {
            Layout.fillWidth: true
            Layout.preferredHeight: 128
        }

        Label {
            text: "Voice participants"
            color: Theme.textMuted
            font.bold: true
        }

        VoiceParticipantList {
            Layout.fillWidth: true
            Layout.preferredHeight: 90
            model: root.participants
        }
    }
}
