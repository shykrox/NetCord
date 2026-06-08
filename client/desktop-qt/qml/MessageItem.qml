import QtQuick
import QtQuick.Controls
import QtQuick.Layouts

Rectangle {
    id: root
    property var message
    property string currentUserId: ""
    signal editRequested(string id, string content)
    signal deleteRequested(string id)
    signal attachmentOpenRequested(string id, string filename)

    width: parent ? parent.width : 640
    implicitHeight: messageColumn.implicitHeight + 14
    color: mouse.containsMouse ? "#151B25" : "transparent"

    RowLayout {
        anchors.fill: parent
        anchors.leftMargin: 18
        anchors.rightMargin: 18
        anchors.topMargin: 7
        anchors.bottomMargin: 7
        spacing: 10

        NcAvatar {
            label: message.author_id || "?"
            status: "offline"
        }

        ColumnLayout {
            id: messageColumn
            Layout.fillWidth: true
            spacing: 4

            RowLayout {
                Layout.fillWidth: true
                Label {
                    text: (message.author_id || "").substring(0, 8)
                    color: Theme.text
                    font.bold: true
                }
                Label {
                    text: (message.created_at || "").substring(11, 19)
                    color: Theme.textSubtle
                    font.pixelSize: 11
                }
                Label {
                    text: message.edited_at ? "edited" : ""
                    color: Theme.textSubtle
                    font.pixelSize: 11
                    Layout.fillWidth: true
                }
                NcButton {
                    text: "Edit"
                    visible: message.author_id === root.currentUserId
                    Layout.preferredWidth: 58
                    onClicked: root.editRequested(message.id, message.content || "")
                }
                NcButton {
                    text: "Delete"
                    danger: true
                    visible: message.author_id === root.currentUserId
                    Layout.preferredWidth: 72
                    onClicked: root.deleteRequested(message.id)
                }
            }

            Label {
                text: message.content || ""
                visible: text.length > 0
                color: Theme.text
                wrapMode: Text.Wrap
                Layout.fillWidth: true
            }

            Repeater {
                model: message.attachments || []
                delegate: AttachmentCard {
                    Layout.fillWidth: true
                    attachment: modelData
                    onOpenRequested: function(id, filename) { root.attachmentOpenRequested(id, filename) }
                }
            }
        }
    }

    MouseArea {
        id: mouse
        anchors.fill: parent
        hoverEnabled: true
        acceptedButtons: Qt.NoButton
    }
}
