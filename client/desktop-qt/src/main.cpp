#include <QGuiApplication>
#include <QQmlApplicationEngine>
#include <QQmlContext>

#include "netcordclient.h"

int main(int argc, char *argv[])
{
    QGuiApplication app(argc, argv);
    QCoreApplication::setOrganizationName("NetCord");
    QCoreApplication::setApplicationName("NetCordDesktop");

    NetCordClient client;

    QQmlApplicationEngine engine;
    engine.rootContext()->setContextProperty("netcord", &client);
    engine.loadFromModule("NetCord", "Main");
    if (engine.rootObjects().isEmpty()) {
        return -1;
    }

    return app.exec();
}
