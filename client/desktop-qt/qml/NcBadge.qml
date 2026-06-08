import QtQuick
import QtQuick.Controls

Rectangle {
    id: root
    property string text: ""
    property color badgeColor: Theme.accent2

    height: 20
    width: Math.max(24, label.implicitWidth + 12)
    radius: 10
    color: badgeColor

    Label {
        id: label
        anchors.centerIn: parent
        text: root.text
        color: Theme.bg0
        font.pixelSize: 11
        font.bold: true
    }
}
