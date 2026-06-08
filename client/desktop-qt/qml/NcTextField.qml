import QtQuick
import QtQuick.Controls

TextField {
    id: root
    color: Theme.text
    placeholderTextColor: Theme.textSubtle
    selectionColor: Theme.accent2
    selectedTextColor: Theme.text
    font.pixelSize: 13

    background: Rectangle {
        radius: Theme.radiusSm
        color: Theme.bg2
        border.color: root.activeFocus ? Theme.accent2 : Theme.border
    }
}
