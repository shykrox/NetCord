import QtQuick
import QtQuick.Controls

Rectangle {
    id: root
    property string message: ""

    visible: message.length > 0
    radius: Theme.radiusMd
    color: "#33252C"
    border.color: "#794858"
    width: Math.min(parent ? parent.width - 48 : 520, textItem.implicitWidth + 28)
    height: textItem.implicitHeight + 18

    Label {
        id: textItem
        anchors.centerIn: parent
        width: Math.min(520, root.width - 28)
        text: root.message
        color: "#FFD8DE"
        wrapMode: Text.Wrap
    }
}
