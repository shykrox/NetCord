#include "netcordclient.h"

#include <QAbstractSocket>
#include <QJsonDocument>
#include <QNetworkReply>
#include <QNetworkRequest>
#include <QUrlQuery>

namespace {
constexpr auto settingsApiBaseUrlKey = "apiBaseUrl";
constexpr auto settingsTokenKey = "jwt";

QString normalizedBaseUrl(const QString &value)
{
    QString trimmed = value.trimmed();
    if (trimmed.isEmpty()) {
        return QStringLiteral("http://127.0.0.1:8080");
    }
    while (trimmed.endsWith('/')) {
        trimmed.chop(1);
    }
    return trimmed;
}
}

NetCordClient::NetCordClient(QObject *parent)
    : QObject(parent)
    , m_settings(QStringLiteral("NetCord"), QStringLiteral("NetCordDesktop"))
{
    m_apiBaseUrl = normalizedBaseUrl(m_settings.value(settingsApiBaseUrlKey, QStringLiteral("http://127.0.0.1:8080")).toString());
    m_token = m_settings.value(settingsTokenKey).toString();

    connect(&m_gateway, &QWebSocket::connected, this, [this]() {
        setGatewayConnected(true);
        setStatusMessage(QStringLiteral("Gateway connected"));
    });
    connect(&m_gateway, &QWebSocket::disconnected, this, [this]() {
        setGatewayConnected(false);
        m_heartbeatTimer.stop();
    });
    connect(&m_gateway, &QWebSocket::textMessageReceived, this, &NetCordClient::handleGatewayTextMessage);
    connect(&m_gateway, &QWebSocket::errorOccurred, this, [this](QAbstractSocket::SocketError) {
        setStatusMessage(QStringLiteral("Gateway connection failed"));
    });
    connect(&m_heartbeatTimer, &QTimer::timeout, this, &NetCordClient::sendHeartbeat);
}

QString NetCordClient::apiBaseUrl() const
{
    return m_apiBaseUrl;
}

void NetCordClient::setApiBaseUrl(const QString &apiBaseUrl)
{
    const QString normalized = normalizedBaseUrl(apiBaseUrl);
    if (m_apiBaseUrl == normalized) {
        return;
    }

    m_apiBaseUrl = normalized;
    m_settings.setValue(settingsApiBaseUrlKey, m_apiBaseUrl);
    emit apiBaseUrlChanged();

    if (m_authenticated) {
        disconnectGateway();
        connectGateway();
    }
}

bool NetCordClient::authenticated() const
{
    return m_authenticated;
}

bool NetCordClient::busy() const
{
    return m_busy;
}

bool NetCordClient::gatewayConnected() const
{
    return m_gatewayConnected;
}

QString NetCordClient::statusMessage() const
{
    return m_statusMessage;
}

QVariantMap NetCordClient::currentUser() const
{
    return m_currentUser;
}

QVariantList NetCordClient::servers() const
{
    return m_servers;
}

QVariantList NetCordClient::channels() const
{
    return m_channels;
}

QVariantList NetCordClient::messages() const
{
    return m_messages;
}

QVariantMap NetCordClient::selectedServer() const
{
    return m_selectedServer;
}

QVariantMap NetCordClient::selectedChannel() const
{
    return m_selectedChannel;
}

void NetCordClient::initialize()
{
    if (m_token.isEmpty()) {
        return;
    }

    getJson(QStringLiteral("/users/me"), true, [this](const QJsonObject &payload) {
        setAuthenticated(true);
        setCurrentUser(payload.toVariantMap());
        loadServers();
        connectGateway();
    });
}

void NetCordClient::login(const QString &email, const QString &password)
{
    clearStatus();
    postJson(QStringLiteral("/auth/login"),
             QJsonObject{
                 {QStringLiteral("email"), email.trimmed()},
                 {QStringLiteral("password"), password},
             },
             false,
             [this](const QJsonObject &payload) { handleAuthSuccess(payload); });
}

void NetCordClient::registerAccount(const QString &username, const QString &email, const QString &password)
{
    clearStatus();
    postJson(QStringLiteral("/auth/register"),
             QJsonObject{
                 {QStringLiteral("username"), username.trimmed()},
                 {QStringLiteral("email"), email.trimmed()},
                 {QStringLiteral("password"), password},
             },
             false,
             [this](const QJsonObject &payload) { handleAuthSuccess(payload); });
}

void NetCordClient::logout()
{
    clearSession(true);
    setStatusMessage(QStringLiteral("Signed out"));
}

void NetCordClient::loadServers()
{
    if (!m_authenticated) {
        return;
    }

    getJson(QStringLiteral("/servers"), true, [this](const QJsonObject &payload) {
        const QVariantList servers = jsonArrayToVariantList(payload.value(QStringLiteral("servers")).toArray());
        setServers(servers);

        if (servers.isEmpty()) {
            setSelectedServer({});
            setChannels({});
            setSelectedChannel({});
            setMessages({});
            return;
        }

        QString serverId = m_selectedServer.value(QStringLiteral("id")).toString();
        if (objectById(servers, serverId).isEmpty()) {
            serverId = servers.first().toMap().value(QStringLiteral("id")).toString();
        }
        selectServer(serverId);
    });
}

void NetCordClient::selectServer(const QString &serverId)
{
    const QVariantMap server = objectById(m_servers, serverId);
    if (server.isEmpty()) {
        return;
    }

    setSelectedServer(server);
    setChannels({});
    setSelectedChannel({});
    setMessages({});

    getJson(QStringLiteral("/servers/%1/channels").arg(serverId), true, [this](const QJsonObject &payload) {
        const QVariantList channels = jsonArrayToVariantList(payload.value(QStringLiteral("channels")).toArray());
        setChannels(channels);
        if (!channels.isEmpty()) {
            selectChannel(channels.first().toMap().value(QStringLiteral("id")).toString());
        }
    });
}

void NetCordClient::selectChannel(const QString &channelId)
{
    const QVariantMap channel = objectById(m_channels, channelId);
    if (channel.isEmpty()) {
        return;
    }

    setSelectedChannel(channel);
    setMessages({});

    getJson(QStringLiteral("/channels/%1/messages").arg(channelId), true, [this](const QJsonObject &payload) {
        QVariantList messages;
        const QJsonArray array = payload.value(QStringLiteral("messages")).toArray();
        messages.reserve(array.size());
        for (const QJsonValue &value : array) {
            messages.append(normalizeMessage(value.toObject()));
        }
        setMessages(messages);
    });
}

void NetCordClient::sendMessage(const QString &content)
{
    const QString trimmed = content.trimmed();
    const QString channelId = m_selectedChannel.value(QStringLiteral("id")).toString();
    if (trimmed.isEmpty() || channelId.isEmpty()) {
        return;
    }

    postJson(QStringLiteral("/channels/%1/messages").arg(channelId),
             QJsonObject{{QStringLiteral("content"), trimmed}},
             true,
             [this](const QJsonObject &payload) { addOrUpdateMessage(normalizeMessage(payload)); });
}

void NetCordClient::clearStatus()
{
    setStatusMessage({});
}

QUrl NetCordClient::apiUrl(const QString &path) const
{
    QUrl url(m_apiBaseUrl);
    QString basePath = url.path();
    if (basePath.endsWith('/')) {
        basePath.chop(1);
    }
    url.setPath(basePath + path);
    return url;
}

QUrl NetCordClient::gatewayUrl() const
{
    QUrl url(m_apiBaseUrl);
    if (url.scheme() == QStringLiteral("https")) {
        url.setScheme(QStringLiteral("wss"));
    } else {
        url.setScheme(QStringLiteral("ws"));
    }

    QString basePath = url.path();
    if (basePath.endsWith('/')) {
        basePath.chop(1);
    }
    url.setPath(basePath + QStringLiteral("/gateway/ws"));
    QUrlQuery query;
    query.setQueryItem(QStringLiteral("token"), m_token);
    url.setQuery(query);
    return url;
}

QNetworkRequest NetCordClient::makeRequest(const QString &path, bool withAuth) const
{
    QNetworkRequest request(apiUrl(path));
    request.setHeader(QNetworkRequest::ContentTypeHeader, QStringLiteral("application/json"));
    if (withAuth && !m_token.isEmpty()) {
        request.setRawHeader("Authorization", QByteArrayLiteral("Bearer ") + m_token.toUtf8());
    }
    return request;
}

void NetCordClient::getJson(const QString &path, bool withAuth, JsonCallback onSuccess)
{
    setBusy(true);
    QNetworkReply *reply = m_network.get(makeRequest(path, withAuth));
    connect(reply, &QNetworkReply::finished, this, [this, reply, withAuth, onSuccess]() {
        handleReply(reply, withAuth, onSuccess);
    });
}

void NetCordClient::postJson(const QString &path, const QJsonObject &payload, bool withAuth, JsonCallback onSuccess)
{
    setBusy(true);
    QNetworkReply *reply = m_network.post(makeRequest(path, withAuth), QJsonDocument(payload).toJson(QJsonDocument::Compact));
    connect(reply, &QNetworkReply::finished, this, [this, reply, withAuth, onSuccess]() {
        handleReply(reply, withAuth, onSuccess);
    });
}

void NetCordClient::handleReply(QNetworkReply *reply, bool withAuth, JsonCallback onSuccess)
{
    const QByteArray body = reply->readAll();
    const int statusCode = reply->attribute(QNetworkRequest::HttpStatusCodeAttribute).toInt();
    const QJsonDocument document = QJsonDocument::fromJson(body);
    const QJsonObject payload = document.object();
    reply->deleteLater();
    setBusy(false);

    if (statusCode >= 200 && statusCode < 300) {
        onSuccess(payload);
        return;
    }

    if (withAuth && statusCode == 401) {
        clearSession(true);
    }
    setStatusMessage(errorMessageFromPayload(payload, statusCode));
}

void NetCordClient::handleAuthSuccess(const QJsonObject &payload)
{
    m_token = payload.value(QStringLiteral("token")).toString();
    if (m_token.isEmpty()) {
        setStatusMessage(QStringLiteral("Authentication response did not include a token"));
        return;
    }

    m_settings.setValue(settingsTokenKey, m_token);
    setAuthenticated(true);
    setCurrentUser(payload.value(QStringLiteral("user")).toObject().toVariantMap());
    loadServers();
    connectGateway();
}

void NetCordClient::connectGateway()
{
    if (!m_authenticated || m_token.isEmpty()) {
        return;
    }

    if (m_gateway.state() != QAbstractSocket::UnconnectedState) {
        m_gateway.abort();
    }
    m_gateway.open(gatewayUrl());
}

void NetCordClient::disconnectGateway()
{
    m_heartbeatTimer.stop();
    if (m_gateway.state() != QAbstractSocket::UnconnectedState) {
        m_gateway.close();
    }
    setGatewayConnected(false);
}

void NetCordClient::handleGatewayTextMessage(const QString &message)
{
    const QJsonDocument document = QJsonDocument::fromJson(message.toUtf8());
    const QJsonObject envelope = document.object();
    const QString type = envelope.value(QStringLiteral("type")).toString();

    if (type == QStringLiteral("hello")) {
        const QJsonObject data = envelope.value(QStringLiteral("data")).toObject();
        int interval = data.value(QStringLiteral("heartbeat_interval_ms")).toInt(30000);
        if (interval < 5000) {
            interval = 5000;
        }
        m_heartbeatTimer.start(interval);
        return;
    }

    if (type == QStringLiteral("heartbeat_ack")) {
        return;
    }

    if (type == QStringLiteral("message.created")) {
        const QJsonObject data = envelope.value(QStringLiteral("data")).toObject();
        if (data.value(QStringLiteral("channel_id")).toString() == m_selectedChannel.value(QStringLiteral("id")).toString()) {
            addOrUpdateMessage(normalizeMessage(data));
        }
        return;
    }

    if (type == QStringLiteral("error")) {
        const QJsonObject data = envelope.value(QStringLiteral("data")).toObject();
        setStatusMessage(data.value(QStringLiteral("message")).toString(QStringLiteral("Gateway error")));
    }
}

void NetCordClient::sendHeartbeat()
{
    if (m_gateway.state() != QAbstractSocket::ConnectedState) {
        return;
    }

    const QJsonObject event{{QStringLiteral("type"), QStringLiteral("heartbeat")}};
    m_gateway.sendTextMessage(QString::fromUtf8(QJsonDocument(event).toJson(QJsonDocument::Compact)));
}

void NetCordClient::setAuthenticated(bool authenticated)
{
    if (m_authenticated == authenticated) {
        return;
    }
    m_authenticated = authenticated;
    emit authenticatedChanged();
}

void NetCordClient::setBusy(bool busy)
{
    if (m_busy == busy) {
        return;
    }
    m_busy = busy;
    emit busyChanged();
}

void NetCordClient::setGatewayConnected(bool connected)
{
    if (m_gatewayConnected == connected) {
        return;
    }
    m_gatewayConnected = connected;
    emit gatewayConnectedChanged();
}

void NetCordClient::setStatusMessage(const QString &message)
{
    if (m_statusMessage == message) {
        return;
    }
    m_statusMessage = message;
    emit statusMessageChanged();
}

void NetCordClient::setCurrentUser(const QVariantMap &user)
{
    m_currentUser = user;
    emit currentUserChanged();
}

void NetCordClient::setServers(const QVariantList &servers)
{
    m_servers = servers;
    emit serversChanged();
}

void NetCordClient::setChannels(const QVariantList &channels)
{
    m_channels = channels;
    emit channelsChanged();
}

void NetCordClient::setMessages(const QVariantList &messages)
{
    m_messages = messages;
    emit messagesChanged();
}

void NetCordClient::setSelectedServer(const QVariantMap &server)
{
    m_selectedServer = server;
    emit selectedServerChanged();
}

void NetCordClient::setSelectedChannel(const QVariantMap &channel)
{
    m_selectedChannel = channel;
    emit selectedChannelChanged();
}

void NetCordClient::addOrUpdateMessage(const QVariantMap &message)
{
    const QString id = message.value(QStringLiteral("id")).toString();
    QVariantList updated = m_messages;
    for (int i = 0; i < updated.size(); ++i) {
        if (updated.at(i).toMap().value(QStringLiteral("id")).toString() == id) {
            updated[i] = message;
            setMessages(updated);
            return;
        }
    }

    updated.append(message);
    setMessages(updated);
}

void NetCordClient::clearSession(bool clearStoredToken)
{
    disconnectGateway();
    m_token.clear();
    if (clearStoredToken) {
        m_settings.remove(settingsTokenKey);
    }
    setAuthenticated(false);
    setCurrentUser({});
    setServers({});
    setChannels({});
    setMessages({});
    setSelectedServer({});
    setSelectedChannel({});
}

QString NetCordClient::errorMessageFromPayload(const QJsonObject &payload, int statusCode) const
{
    const QJsonObject error = payload.value(QStringLiteral("error")).toObject();
    const QString message = error.value(QStringLiteral("message")).toString();
    if (!message.isEmpty()) {
        return message;
    }
    if (statusCode == 0) {
        return QStringLiteral("Backend is unreachable");
    }
    return QStringLiteral("Request failed with HTTP %1").arg(statusCode);
}

QVariantList NetCordClient::jsonArrayToVariantList(const QJsonArray &array)
{
    QVariantList list;
    list.reserve(array.size());
    for (const QJsonValue &value : array) {
        list.append(value.toObject().toVariantMap());
    }
    return list;
}

QVariantMap NetCordClient::objectById(const QVariantList &items, const QString &id)
{
    for (const QVariant &item : items) {
        const QVariantMap object = item.toMap();
        if (object.value(QStringLiteral("id")).toString() == id) {
            return object;
        }
    }
    return {};
}

QVariantMap NetCordClient::normalizeMessage(const QJsonObject &message)
{
    QVariantMap normalized = message.toVariantMap();
    if (!normalized.contains(QStringLiteral("attachments"))) {
        normalized.insert(QStringLiteral("attachments"), QVariantList{});
    }
    return normalized;
}
