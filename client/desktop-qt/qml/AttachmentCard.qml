import QtQuick
import QtQuick.Controls
import QtQuick.Layouts

Rectangle {
    id: root
    property var attachment
    signal openRequested(string id, string filename)

    implicitHeight: 42
    radius: Theme.radiusSm
    color: Theme.bg2
    border.color: Theme.border

    RowLayout {
        anchors.fill: parent
        anchors.leftMargin: 10
        anchors.rightMargin: 10
        spacing: 8

        Label {
            text: "file"
            color: Theme.accent2
            font.pixelSize: 11
            font.bold: true
        }

        ColumnLayout {
            Layout.fillWidth: true
            spacing: 1

            Label {
                text: attachment.original_filename || "attachment"
                color: Theme.text
                elide: Text.ElideRight
                Layout.fillWidth: true
            }

            Label {
                text: Math.ceil((attachment.size_bytes || 0) / 1024) + " KB"
                color: Theme.textSubtle
                font.pixelSize: 11
            }
        }

        NcButton {
            text: "Open"
            Layout.preferredWidth: 64
            onClicked: root.openRequested(attachment.id, attachment.original_filename || "attachment")
        }
    }
}
