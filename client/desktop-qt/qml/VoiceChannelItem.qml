import QtQuick
import QtQuick.Controls
import QtQuick.Layouts

Rectangle {
    id: root
    property var channel
    signal joinRequested(string channelId)

    width: parent ? parent.width : 220
    height: 38
    radius: 6
    color: "transparent"

    RowLayout {
        anchors.fill: parent
        anchors.leftMargin: 10
        anchors.rightMargin: 8
        spacing: 8

        Label {
            text: ">"
            color: "#8ea4c2"
            font.bold: true
        }

        Label {
            text: channel.name || "voice"
            color: "#aab3c2"
            elide: Text.ElideRight
            Layout.fillWidth: true
        }

        Button {
            text: "Join"
            Layout.preferredHeight: 28
            onClicked: root.joinRequested(channel.id)
        }
    }
}
