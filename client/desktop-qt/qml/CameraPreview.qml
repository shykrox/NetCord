import QtQuick
import QtQuick.Controls
import QtQuick.Layouts

Rectangle {
    id: root
    property bool active: false

    height: 120
    radius: Theme.radiusMd
    color: Theme.bg0
    border.color: active ? Theme.accent2 : Theme.border

    ColumnLayout {
        anchors.centerIn: parent
        spacing: 6

        Label {
            text: active ? "Camera preview placeholder" : "Camera off"
            color: Theme.textMuted
        }

        Label {
            text: "Native camera preview is pending Qt Multimedia wiring"
            color: Theme.textSubtle
            font.pixelSize: 11
        }
    }
}
