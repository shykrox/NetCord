import QtQuick
import QtQuick.Controls
import QtQuick.Layouts

Rectangle {
    id: root
    property bool active: false

    height: 86
    radius: Theme.radiusMd
    color: Theme.bg2
    border.color: active ? Theme.warning : Theme.border

    RowLayout {
        anchors.fill: parent
        anchors.margins: 10
        spacing: 8

        ColumnLayout {
            Layout.fillWidth: true
            Label {
                text: active ? "Screen share ready" : "Screen share"
                color: Theme.text
                font.bold: true
            }
            Label {
                text: "Picker/transport pending native LiveKit integration"
                color: Theme.textSubtle
                font.pixelSize: 11
                wrapMode: Text.Wrap
                Layout.fillWidth: true
            }
        }

        NcButton {
            text: active ? "Stop" : "Start"
            onClicked: root.active = !root.active
        }
    }
}
