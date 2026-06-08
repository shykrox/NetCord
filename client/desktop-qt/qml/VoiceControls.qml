import QtQuick
import QtQuick.Controls
import QtQuick.Layouts

Rectangle {
    id: root
    property bool connected: false
    property string status: ""
    signal leaveRequested()

    height: 72
    radius: 6
    color: "#202632"
    border.color: connected ? "#37d7a7" : "#394354"

    RowLayout {
        anchors.fill: parent
        anchors.margins: 10
        spacing: 8

        Rectangle {
            Layout.preferredWidth: 10
            Layout.preferredHeight: 10
            radius: 5
            color: connected ? "#37d7a7" : "#778092"
        }

        Label {
            text: status.length > 0 ? status : (connected ? "Voice connected" : "Voice idle")
            color: "#d8e1ee"
            elide: Text.ElideRight
            Layout.fillWidth: true
        }

        Button {
            text: "Mute"
            enabled: connected
        }

        Button {
            text: "Leave"
            enabled: connected
            onClicked: root.leaveRequested()
        }
    }
}
