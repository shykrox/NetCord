import QtQuick
import QtQuick.Controls

ApplicationWindow {
    id: root
    width: 1180
    height: 720
    minimumWidth: 940
    minimumHeight: 560
    visible: true
    title: "NetCord"
    color: "#111318"

    Component.onCompleted: netcord.initialize()

    Loader {
        anchors.fill: parent
        sourceComponent: netcord.authenticated ? appView : loginView
    }

    Component {
        id: loginView
        LoginView {}
    }

    Component {
        id: appView
        AppView {}
    }
}
