import QtQuick
import QtQuick.Controls
import QtQuick.Layouts

Rectangle {
    color: "#111318"

    property bool registerMode: false

    Rectangle {
        width: 420
        height: registerMode ? 500 : 440
        anchors.centerIn: parent
        color: "#1d2028"
        radius: 8
        border.color: "#2b303c"
        border.width: 1

        ColumnLayout {
            anchors.fill: parent
            anchors.margins: 28
            spacing: 14

            Label {
                text: "NetCord"
                color: "#f3f6fb"
                font.pixelSize: 32
                font.bold: true
                Layout.alignment: Qt.AlignHCenter
            }

            RowLayout {
                Layout.fillWidth: true
                spacing: 8

                Button {
                    text: "Login"
                    checkable: true
                    checked: !registerMode
                    Layout.fillWidth: true
                    onClicked: registerMode = false
                }

                Button {
                    text: "Register"
                    checkable: true
                    checked: registerMode
                    Layout.fillWidth: true
                    onClicked: registerMode = true
                }
            }

            TextField {
                id: apiBaseUrl
                text: netcord.apiBaseUrl
                placeholderText: "Backend URL"
                Layout.fillWidth: true
                onEditingFinished: netcord.apiBaseUrl = text
            }

            TextField {
                id: username
                visible: registerMode
                Layout.fillWidth: true
                placeholderText: "Username"
            }

            TextField {
                id: email
                Layout.fillWidth: true
                placeholderText: "Email"
                inputMethodHints: Qt.ImhEmailCharactersOnly
            }

            TextField {
                id: password
                Layout.fillWidth: true
                placeholderText: "Password"
                echoMode: TextInput.Password
                onAccepted: submitButton.clicked()
            }

            Button {
                id: submitButton
                text: registerMode ? "Create account" : "Sign in"
                enabled: !netcord.busy
                Layout.fillWidth: true
                highlighted: true
                onClicked: {
                    netcord.apiBaseUrl = apiBaseUrl.text
                    if (registerMode) {
                        netcord.registerAccount(username.text, email.text, password.text)
                    } else {
                        netcord.login(email.text, password.text)
                    }
                }
            }

            BusyIndicator {
                running: netcord.busy
                visible: running
                Layout.alignment: Qt.AlignHCenter
            }

            Label {
                text: netcord.statusMessage
                visible: text.length > 0
                color: "#ffb4b4"
                wrapMode: Text.Wrap
                Layout.fillWidth: true
            }
        }
    }
}
