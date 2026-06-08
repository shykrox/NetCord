#pragma once

#include <QJsonArray>
#include <QJsonObject>
#include <QNetworkAccessManager>
#include <QObject>
#include <QSettings>
#include <QTimer>
#include <QVariantList>
#include <QVariantMap>
#include <QWebSocket>

#include <functional>

class QNetworkReply;

class NetCordClient : public QObject
{
    Q_OBJECT
    Q_PROPERTY(QString apiBaseUrl READ apiBaseUrl WRITE setApiBaseUrl NOTIFY apiBaseUrlChanged)
    Q_PROPERTY(bool authenticated READ authenticated NOTIFY authenticatedChanged)
    Q_PROPERTY(bool busy READ busy NOTIFY busyChanged)
    Q_PROPERTY(bool gatewayConnected READ gatewayConnected NOTIFY gatewayConnectedChanged)
    Q_PROPERTY(QString statusMessage READ statusMessage NOTIFY statusMessageChanged)
    Q_PROPERTY(QVariantMap currentUser READ currentUser NOTIFY currentUserChanged)
    Q_PROPERTY(QVariantList servers READ servers NOTIFY serversChanged)
    Q_PROPERTY(QVariantList channels READ channels NOTIFY channelsChanged)
    Q_PROPERTY(QVariantList messages READ messages NOTIFY messagesChanged)
    Q_PROPERTY(QVariantMap selectedServer READ selectedServer NOTIFY selectedServerChanged)
    Q_PROPERTY(QVariantMap selectedChannel READ selectedChannel NOTIFY selectedChannelChanged)

public:
    explicit NetCordClient(QObject *parent = nullptr);

    QString apiBaseUrl() const;
    void setApiBaseUrl(const QString &apiBaseUrl);

    bool authenticated() const;
    bool busy() const;
    bool gatewayConnected() const;
    QString statusMessage() const;
    QVariantMap currentUser() const;
    QVariantList servers() const;
    QVariantList channels() const;
    QVariantList messages() const;
    QVariantMap selectedServer() const;
    QVariantMap selectedChannel() const;

    Q_INVOKABLE void initialize();
    Q_INVOKABLE void login(const QString &email, const QString &password);
    Q_INVOKABLE void registerAccount(const QString &username, const QString &email, const QString &password);
    Q_INVOKABLE void logout();
    Q_INVOKABLE void loadServers();
    Q_INVOKABLE void selectServer(const QString &serverId);
    Q_INVOKABLE void selectChannel(const QString &channelId);
    Q_INVOKABLE void sendMessage(const QString &content);
    Q_INVOKABLE void clearStatus();

signals:
    void apiBaseUrlChanged();
    void authenticatedChanged();
    void busyChanged();
    void gatewayConnectedChanged();
    void statusMessageChanged();
    void currentUserChanged();
    void serversChanged();
    void channelsChanged();
    void messagesChanged();
    void selectedServerChanged();
    void selectedChannelChanged();

private:
    using JsonCallback = std::function<void(const QJsonObject &)>;

    QUrl apiUrl(const QString &path) const;
    QUrl gatewayUrl() const;
    QNetworkRequest makeRequest(const QString &path, bool withAuth) const;
    void getJson(const QString &path, bool withAuth, JsonCallback onSuccess);
    void postJson(const QString &path, const QJsonObject &payload, bool withAuth, JsonCallback onSuccess);
    void handleReply(QNetworkReply *reply, bool withAuth, JsonCallback onSuccess);
    void handleAuthSuccess(const QJsonObject &payload);
    void connectGateway();
    void disconnectGateway();
    void handleGatewayTextMessage(const QString &message);
    void sendHeartbeat();
    void setAuthenticated(bool authenticated);
    void setBusy(bool busy);
    void setGatewayConnected(bool connected);
    void setStatusMessage(const QString &message);
    void setCurrentUser(const QVariantMap &user);
    void setServers(const QVariantList &servers);
    void setChannels(const QVariantList &channels);
    void setMessages(const QVariantList &messages);
    void setSelectedServer(const QVariantMap &server);
    void setSelectedChannel(const QVariantMap &channel);
    void addOrUpdateMessage(const QVariantMap &message);
    void clearSession(bool clearStoredToken);
    QString errorMessageFromPayload(const QJsonObject &payload, int statusCode) const;
    static QVariantList jsonArrayToVariantList(const QJsonArray &array);
    static QVariantMap objectById(const QVariantList &items, const QString &id);
    static QVariantMap normalizeMessage(const QJsonObject &message);

    QNetworkAccessManager m_network;
    QWebSocket m_gateway;
    QTimer m_heartbeatTimer;
    QSettings m_settings;
    QString m_apiBaseUrl;
    QString m_token;
    bool m_authenticated = false;
    bool m_busy = false;
    bool m_gatewayConnected = false;
    QString m_statusMessage;
    QVariantMap m_currentUser;
    QVariantList m_servers;
    QVariantList m_channels;
    QVariantList m_messages;
    QVariantMap m_selectedServer;
    QVariantMap m_selectedChannel;
};
