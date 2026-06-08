import QtQuick
import QtQuick.Controls

ListView {
    id: root
    clip: true
    spacing: 4

    delegate: Label {
        width: ListView.view.width
        text: (modelData.username || modelData.user_id || "user") + " - " + (modelData.status || "offline")
        color: "#d8e1ee"
        elide: Text.ElideRight
    }
}
