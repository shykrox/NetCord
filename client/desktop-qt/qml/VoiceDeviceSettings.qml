import QtQuick
import QtQuick.Controls
import QtQuick.Layouts

Rectangle {
    color: "#202632"
    radius: 6
    border.color: "#394354"

    ColumnLayout {
        anchors.fill: parent
        anchors.margins: 10
        spacing: 8

        Label {
            text: "Voice devices"
            color: "#f3f6fb"
            font.bold: true
        }

        ComboBox {
            Layout.fillWidth: true
            model: ["Default microphone"]
        }

        ComboBox {
            Layout.fillWidth: true
            model: ["Default speakers"]
        }
    }
}
