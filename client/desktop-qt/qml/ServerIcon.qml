import QtQuick
import QtQuick.Controls

Rectangle {
    id: root
    property string name: "?"
    property bool selected: false

    width: 48
    height: 48
    radius: selected ? Theme.radiusMd : Theme.radiusLg
    color: selected ? Theme.accent : Theme.bg3
    border.color: selected ? Theme.accent : Theme.border

    Rectangle {
        visible: selected
        anchors.left: parent.left
        anchors.verticalCenter: parent.verticalCenter
        anchors.leftMargin: -8
        width: 4
        height: 28
        radius: 2
        color: Theme.accent
    }

    Label {
        anchors.centerIn: parent
        text: root.name.length > 0 ? root.name.substring(0, 1).toUpperCase() : "?"
        color: selected ? Theme.bg0 : Theme.text
        font.bold: true
        font.pixelSize: 18
    }
}
