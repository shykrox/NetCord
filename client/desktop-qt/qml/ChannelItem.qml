import QtQuick
import QtQuick.Controls
import QtQuick.Layouts

Rectangle {
    id: root
    property string name: ""
    property string type: "text"
    property bool selected: false
    signal clicked()

    height: 38
    radius: Theme.radiusSm
    color: selected ? Theme.bg3 : (mouse.containsMouse ? Theme.bg2 : "transparent")

    RowLayout {
        anchors.fill: parent
        anchors.leftMargin: 10
        anchors.rightMargin: 10
        spacing: 8

        Label {
            text: root.type === "voice" ? ">" : "#"
            color: root.type === "voice" ? Theme.accent2 : Theme.textSubtle
            font.pixelSize: 18
        }

        Label {
            text: root.name
            color: selected ? Theme.text : Theme.textMuted
            elide: Text.ElideRight
            Layout.fillWidth: true
        }
    }

    MouseArea {
        id: mouse
        anchors.fill: parent
        hoverEnabled: true
        onClicked: root.clicked()
    }
}
