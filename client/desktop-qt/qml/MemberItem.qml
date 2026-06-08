import QtQuick
import QtQuick.Controls
import QtQuick.Layouts

RowLayout {
    id: root
    property var member

    width: parent ? parent.width : 220
    height: 32
    spacing: 8

    NcAvatar {
        avatarSize: 26
        label: member.username || member.user_id || "?"
        status: member.status || "offline"
    }

    Label {
        text: member.username || member.user_id || "user"
        color: Theme.text
        elide: Text.ElideRight
        Layout.fillWidth: true
    }

    Label {
        text: member.role || ""
        color: Theme.textSubtle
        font.pixelSize: 11
    }
}
