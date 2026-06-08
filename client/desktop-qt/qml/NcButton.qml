import QtQuick
import QtQuick.Controls

Button {
    id: root
    property bool danger: false

    font.pixelSize: 13
    padding: 10

    contentItem: Label {
        text: root.text
        color: root.enabled ? (root.highlighted ? Theme.bg0 : Theme.text) : Theme.textSubtle
        horizontalAlignment: Text.AlignHCenter
        verticalAlignment: Text.AlignVCenter
        elide: Text.ElideRight
    }

    background: Rectangle {
        radius: Theme.radiusSm
        color: !root.enabled ? Theme.bg2
             : root.danger ? Theme.danger
             : root.highlighted ? Theme.accent
             : root.hovered ? Theme.bg3
             : Theme.bg2
        border.color: root.activeFocus ? Theme.accent2 : Theme.border
    }
}
