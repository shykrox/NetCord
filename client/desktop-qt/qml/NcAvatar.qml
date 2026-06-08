import QtQuick
import QtQuick.Controls

Rectangle {
    id: root
    property string label: "?"
    property string status: "offline"
    property int avatarSize: 34

    width: avatarSize
    height: avatarSize
    radius: Theme.radiusMd
    color: Theme.bg3
    border.color: Theme.border

    Label {
        anchors.centerIn: parent
        text: root.label.length > 0 ? root.label.substring(0, 2).toUpperCase() : "?"
        color: Theme.text
        font.pixelSize: 11
        font.bold: true
    }

    Rectangle {
        anchors.right: parent.right
        anchors.bottom: parent.bottom
        width: 9
        height: 9
        radius: 5
        color: root.status === "online" ? Theme.success : Theme.textSubtle
        border.color: Theme.bg1
        border.width: 1
    }
}
