#include "netcordclient.h"

#include <QAbstractSocket>
#include <QDesktopServices>
#include <QDir>
#include <QFile>
#include <QFileInfo>
#include <QHttpMultiPart>
#include <QJsonArray>
#include <QJsonDocument>
#include <QNetworkReply>
#include <QNetworkRequest>
#include <QSqlError>
#include <QSqlQuery>
#include <QStandardPaths>
#include <QUrlQuery>
#include <QtGlobal>

namespace {
constexpr auto settingsApiBaseUrlKey = "apiBaseUrl";
constexpr auto settingsTokenKey = "jwt";
constexpr auto cacheConnectionName = "netcord_cache";

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
        m_gatewayReconnectAttempts = 0;
        m_gatewayReconnectTimer.stop();
        setGatewayConnected(true);
        setStatusMessage(QStringLiteral("Gateway connected"));
    });
    connect(&m_gateway, &QWebSocket::disconnected, this, [this]() {
        setGatewayConnected(false);
        m_heartbeatTimer.stop();
        if (m_authenticated && !m_manualGatewayClose) {
            scheduleGatewayReconnect();
        }
    });
    connect(&m_gateway, &QWebSocket::textMessageReceived, this, &NetCordClient::handleGatewayTextMessage);
    connect(&m_gateway, &QWebSocket::errorOccurred, this, [this](QAbstractSocket::SocketError) {
        if (m_authenticated && !m_manualGatewayClose) {
            setStatusMessage(QStringLiteral("Gateway disconnected. Reconnecting..."));
        } else {
            setStatusMessage(QStringLiteral("Gateway connection failed"));
        }
    });
    connect(&m_heartbeatTimer, &QTimer::timeout, this, &NetCordClient::sendHeartbeat);
    m_gatewayReconnectTimer.setSingleShot(true);
    connect(&m_gatewayReconnectTimer, &QTimer::timeout, this, &NetCordClient::connectGateway);
    m_typingStopTimer.setSingleShot(true);
    connect(&m_typingStopTimer, &QTimer::timeout, this, &NetCordClient::sendTypingStop);
    m_typingIndicatorTimer.setSingleShot(true);
    connect(&m_typingIndicatorTimer, &QTimer::timeout, this, &NetCordClient::clearTypingLater);
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

QVariantList NetCordClient::searchResults() const
{
    return m_searchResults;
}

QVariantList NetCordClient::friendRequests() const
{
    return m_friendRequests;
}

QVariantList NetCordClient::friends() const
{
    return m_friends;
}

QVariantList NetCordClient::dmConversations() const
{
    return m_dmConversations;
}

QVariantList NetCordClient::aiJobs() const
{
    return m_aiJobs;
}

QVariantList NetCordClient::serverMembers() const
{
    return m_serverMembers;
}

QVariantList NetCordClient::roles() const
{
    return m_roles;
}

QVariantList NetCordClient::pendingAttachments() const
{
    return m_pendingAttachments;
}

QVariantMap NetCordClient::selectedServer() const
{
    return m_selectedServer;
}

QVariantMap NetCordClient::selectedChannel() const
{
    return m_selectedChannel;
}

QString NetCordClient::typingText() const
{
    return m_typingText;
}

bool NetCordClient::voiceConnected() const
{
    return m_voiceConnected;
}

QString NetCordClient::voiceStatus() const
{
    return m_voiceStatus;
}

void NetCordClient::initialize()
{
    initializeCache();
    if (m_token.isEmpty()) {
        return;
    }

    getJson(QStringLiteral("/users/me"), true, [this](const QJsonObject &payload) {
        setAuthenticated(true);
        setCurrentUser(payload.toVariantMap());
        loadServers();
        loadFriends();
        loadFriendRequests();
        loadDMs();
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

    const QVariantList cached = cachedServers();
    if (!cached.isEmpty() && m_servers.isEmpty()) {
        setServers(cached);
    }

    getJson(QStringLiteral("/servers"), true, [this](const QJsonObject &payload) {
        const QVariantList servers = jsonArrayToVariantList(payload.value(QStringLiteral("servers")).toArray());
        setServers(servers);
        cacheServers(servers);

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

void NetCordClient::refreshServers()
{
    loadServers();
}

void NetCordClient::refreshChannels()
{
    const QString serverId = m_selectedServer.value(QStringLiteral("id")).toString();
    if (!serverId.isEmpty()) {
        selectServer(serverId);
    }
}

void NetCordClient::refreshMessages()
{
    const QString channelId = m_selectedChannel.value(QStringLiteral("id")).toString();
    if (!channelId.isEmpty()) {
        selectChannel(channelId);
    }
}

void NetCordClient::reconnectGateway()
{
    disconnectGateway();
    connectGateway();
}

void NetCordClient::selectServer(const QString &serverId)
{
    const QVariantMap server = objectById(m_servers, serverId);
    if (server.isEmpty()) {
        return;
    }

    const QString previousChannelId = m_selectedChannel.value(QStringLiteral("id")).toString();
        setSelectedServer(server);
        setChannels(cachedChannels(serverId));
        setSelectedChannel({});
        setMessages({});
        loadServerMembers();
        loadRoles();

    getJson(QStringLiteral("/servers/%1/channels").arg(serverId), true, [this, previousChannelId](const QJsonObject &payload) {
        const QVariantList channels = jsonArrayToVariantList(payload.value(QStringLiteral("channels")).toArray());
        setChannels(channels);
        cacheChannels(m_selectedServer.value(QStringLiteral("id")).toString(), channels);
        if (channels.isEmpty()) {
            return;
        }

        QString channelId = previousChannelId;
        if (objectById(channels, channelId).isEmpty()) {
            channelId = channels.first().toMap().value(QStringLiteral("id")).toString();
        }
        selectChannel(channelId);
    });
}

void NetCordClient::selectChannel(const QString &channelId)
{
    const QVariantMap channel = objectById(m_channels, channelId);
    if (channel.isEmpty()) {
        return;
    }

    setSelectedChannel(channel);
    setMessages(cachedMessages(channelId));
    setPendingAttachments({});

    getJson(QStringLiteral("/channels/%1/messages").arg(channelId), true, [this](const QJsonObject &payload) {
        QVariantList messages;
        const QJsonArray array = payload.value(QStringLiteral("messages")).toArray();
        messages.reserve(array.size());
        for (const QJsonValue &value : array) {
            messages.append(normalizeMessage(value.toObject()));
        }
        setMessages(messages);
        cacheMessages(m_selectedChannel.value(QStringLiteral("id")).toString(), messages);
    });
}

void NetCordClient::sendMessage(const QString &content)
{
    const QString trimmed = content.trimmed();
    const QString channelId = m_selectedChannel.value(QStringLiteral("id")).toString();
    if ((trimmed.isEmpty() && m_pendingAttachments.isEmpty()) || channelId.isEmpty()) {
        return;
    }

    QJsonArray attachments;
    for (const QVariant &attachment : m_pendingAttachments) {
        const QString id = attachment.toMap().value(QStringLiteral("id")).toString();
        if (!id.isEmpty()) {
            attachments.append(id);
        }
    }

    QJsonObject payload{{QStringLiteral("content"), trimmed}};
    if (!attachments.isEmpty()) {
        payload.insert(QStringLiteral("attachments"), attachments);
    }

    postJson(QStringLiteral("/channels/%1/messages").arg(channelId),
             payload,
             true,
             [this](const QJsonObject &payload) {
                 setPendingAttachments({});
                 sendTypingStop();
                 addOrUpdateMessage(normalizeMessage(payload));
             });
}

void NetCordClient::editMessage(const QString &messageId, const QString &content)
{
    const QString trimmed = content.trimmed();
    if (messageId.isEmpty() || trimmed.isEmpty()) {
        return;
    }

    patchJson(QStringLiteral("/messages/%1").arg(messageId),
              QJsonObject{{QStringLiteral("content"), trimmed}},
              true,
              [this](const QJsonObject &payload) { addOrUpdateMessage(normalizeMessage(payload)); });
}

void NetCordClient::deleteMessage(const QString &messageId)
{
    if (messageId.isEmpty()) {
        return;
    }

    deleteRequest(QStringLiteral("/messages/%1").arg(messageId), true, [this, messageId](const QJsonObject &) {
        removeMessage(messageId);
    });
}

void NetCordClient::loadOlderMessages()
{
    const QString channelId = m_selectedChannel.value(QStringLiteral("id")).toString();
    if (channelId.isEmpty() || m_messages.isEmpty()) {
        return;
    }
    const QString before = m_messages.first().toMap().value(QStringLiteral("id")).toString();
    getJson(QStringLiteral("/channels/%1/messages?before=%2&limit=50").arg(channelId, before), true, [this](const QJsonObject &payload) {
        QVariantList older;
        const QJsonArray array = payload.value(QStringLiteral("messages")).toArray();
        older.reserve(array.size() + m_messages.size());
        for (const QJsonValue &value : array) {
            older.append(normalizeMessage(value.toObject()));
        }
        older.append(m_messages);
        setMessages(older);
        cacheMessages(m_selectedChannel.value(QStringLiteral("id")).toString(), older);
    });
}

void NetCordClient::searchMessages(const QString &query)
{
    const QString channelId = m_selectedChannel.value(QStringLiteral("id")).toString();
    const QString trimmed = query.trimmed();
    if (channelId.isEmpty() || trimmed.isEmpty()) {
        setSearchResults({});
        return;
    }
    QUrl url = apiUrl(QStringLiteral("/channels/%1/messages/search").arg(channelId));
    QUrlQuery urlQuery;
    urlQuery.addQueryItem(QStringLiteral("q"), trimmed);
    urlQuery.addQueryItem(QStringLiteral("limit"), QStringLiteral("50"));
    url.setQuery(urlQuery);

    QNetworkRequest request(url);
    request.setHeader(QNetworkRequest::ContentTypeHeader, QStringLiteral("application/json"));
    request.setRawHeader("Authorization", QByteArrayLiteral("Bearer ") + m_token.toUtf8());
    beginRequest();
    QNetworkReply *reply = m_network.get(request);
    connect(reply, &QNetworkReply::finished, this, [this, reply]() {
        handleReply(reply, true, [this](const QJsonObject &payload) {
            QVariantList results;
            const QJsonArray array = payload.value(QStringLiteral("messages")).toArray();
            results.reserve(array.size());
            for (const QJsonValue &value : array) {
                results.append(normalizeMessage(value.toObject()));
            }
            setSearchResults(results);
        });
    });
}

void NetCordClient::clearSearchResults()
{
    setSearchResults({});
}

void NetCordClient::createServer(const QString &name, const QString &description)
{
    postJson(QStringLiteral("/servers"),
             QJsonObject{
                 {QStringLiteral("name"), name.trimmed()},
                 {QStringLiteral("description"), description.trimmed()},
             },
             true,
             [this](const QJsonObject &) { loadServers(); });
}

void NetCordClient::createChannel(const QString &name, const QString &type)
{
    const QString serverId = m_selectedServer.value(QStringLiteral("id")).toString();
    if (serverId.isEmpty()) {
        return;
    }
    postJson(QStringLiteral("/servers/%1/channels").arg(serverId),
             QJsonObject{
                 {QStringLiteral("name"), name.trimmed()},
                 {QStringLiteral("type"), type.trimmed().isEmpty() ? QStringLiteral("text") : type.trimmed()},
             },
             true,
             [this](const QJsonObject &) { refreshChannels(); });
}

void NetCordClient::updateProfile(const QString &displayName, const QString &status)
{
    patchJson(QStringLiteral("/users/me"),
              QJsonObject{
                  {QStringLiteral("display_name"), displayName.trimmed()},
                  {QStringLiteral("status"), status.trimmed().isEmpty() ? QStringLiteral("offline") : status.trimmed()},
              },
              true,
              [this](const QJsonObject &payload) {
                  setCurrentUser(payload.toVariantMap());
                  setStatusMessage(QStringLiteral("Profile updated"));
              });
}

void NetCordClient::loadServerMembers()
{
    const QString serverId = m_selectedServer.value(QStringLiteral("id")).toString();
    if (serverId.isEmpty()) {
        setServerMembers({});
        return;
    }
    getJson(QStringLiteral("/servers/%1/members").arg(serverId), true, [this](const QJsonObject &payload) {
        setServerMembers(jsonArrayToVariantList(payload.value(QStringLiteral("members")).toArray()));
    });
}

void NetCordClient::loadRoles()
{
    const QString serverId = m_selectedServer.value(QStringLiteral("id")).toString();
    if (serverId.isEmpty()) {
        setRoles({});
        return;
    }
    getJson(QStringLiteral("/servers/%1/roles").arg(serverId), true, [this](const QJsonObject &payload) {
        setRoles(jsonArrayToVariantList(payload.value(QStringLiteral("roles")).toArray()));
    });
}

void NetCordClient::createRole(const QString &name, qint64 permissions)
{
    const QString serverId = m_selectedServer.value(QStringLiteral("id")).toString();
    if (serverId.isEmpty()) {
        return;
    }
    postJson(QStringLiteral("/servers/%1/roles").arg(serverId),
             QJsonObject{
                 {QStringLiteral("name"), name.trimmed()},
                 {QStringLiteral("permissions"), permissions},
             },
             true,
             [this](const QJsonObject &) { loadRoles(); });
}

void NetCordClient::createInvite()
{
    const QString serverId = m_selectedServer.value(QStringLiteral("id")).toString();
    if (serverId.isEmpty()) {
        return;
    }
    postJson(QStringLiteral("/servers/%1/invites").arg(serverId), {}, true, [this](const QJsonObject &payload) {
        setStatusMessage(QStringLiteral("Invite code: %1").arg(payload.value(QStringLiteral("code")).toString()));
    });
}

void NetCordClient::joinInvite(const QString &code)
{
    const QString trimmed = code.trimmed();
    if (trimmed.isEmpty()) {
        return;
    }
    postJson(QStringLiteral("/invites/%1/join").arg(trimmed), {}, true, [this](const QJsonObject &) { loadServers(); });
}

void NetCordClient::joinVoice(const QString &channelId)
{
    if (channelId.isEmpty()) {
        return;
    }
    postJson(QStringLiteral("/voice/join"),
             QJsonObject{{QStringLiteral("channel_id"), channelId}},
             true,
             [this, channelId](const QJsonObject &payload) {
                 m_voiceChannelId = channelId;
                 setVoiceConnected(true);
                 setVoiceStatus(QStringLiteral("Voice room: %1").arg(payload.value(QStringLiteral("room")).toString()));
             });
}

void NetCordClient::leaveVoice()
{
    if (m_voiceChannelId.isEmpty()) {
        setVoiceConnected(false);
        setVoiceStatus({});
        return;
    }
    const QString channelId = m_voiceChannelId;
    postJson(QStringLiteral("/voice/leave"),
             QJsonObject{{QStringLiteral("channel_id"), channelId}},
             true,
             [this](const QJsonObject &) {
                 m_voiceChannelId.clear();
                 setVoiceConnected(false);
                 setVoiceStatus(QStringLiteral("Voice disconnected"));
             });
}

void NetCordClient::clearCache()
{
    if (!m_cache.isOpen()) {
        return;
    }
    QSqlQuery query(m_cache);
    query.exec(QStringLiteral("DELETE FROM messages"));
    query.exec(QStringLiteral("DELETE FROM channels"));
    query.exec(QStringLiteral("DELETE FROM servers"));
    setStatusMessage(QStringLiteral("Local cache cleared"));
}

void NetCordClient::sendTypingStart()
{
    const QString channelId = m_selectedChannel.value(QStringLiteral("id")).toString();
    if (!m_gatewayConnected || channelId.isEmpty()) {
        return;
    }
    if (m_lastTypingChannelId != channelId) {
        m_lastTypingChannelId = channelId;
        sendGatewayEvent(QJsonObject{
            {QStringLiteral("type"), QStringLiteral("typing.start")},
            {QStringLiteral("data"), QJsonObject{{QStringLiteral("channel_id"), channelId}}},
        });
    }
    m_typingStopTimer.start(2200);
}

void NetCordClient::sendTypingStop()
{
    const QString channelId = m_lastTypingChannelId;
    if (!m_gatewayConnected || channelId.isEmpty()) {
        m_lastTypingChannelId.clear();
        return;
    }
    sendGatewayEvent(QJsonObject{
        {QStringLiteral("type"), QStringLiteral("typing.stop")},
        {QStringLiteral("data"), QJsonObject{{QStringLiteral("channel_id"), channelId}}},
    });
    m_lastTypingChannelId.clear();
    m_typingStopTimer.stop();
}

void NetCordClient::openAttachment(const QString &downloadUrl)
{
    if (downloadUrl.isEmpty()) {
        return;
    }
    QDesktopServices::openUrl(apiUrl(downloadUrl));
}

void NetCordClient::loadFriends()
{
    getJson(QStringLiteral("/friends"), true, [this](const QJsonObject &payload) {
        setFriends(jsonArrayToVariantList(payload.value(QStringLiteral("friends")).toArray()));
    });
}

void NetCordClient::loadFriendRequests()
{
    getJson(QStringLiteral("/friends/requests"), true, [this](const QJsonObject &payload) {
        setFriendRequests(jsonArrayToVariantList(payload.value(QStringLiteral("requests")).toArray()));
    });
}

void NetCordClient::sendFriendRequest(const QString &recipientId)
{
    if (recipientId.trimmed().isEmpty()) {
        return;
    }
    postJson(QStringLiteral("/friends/requests"),
             QJsonObject{{QStringLiteral("recipient_id"), recipientId.trimmed()}},
             true,
             [this](const QJsonObject &) {
                 loadFriendRequests();
                 setStatusMessage(QStringLiteral("Friend request sent"));
             });
}

void NetCordClient::acceptFriendRequest(const QString &requestId)
{
    if (requestId.isEmpty()) {
        return;
    }
    postJson(QStringLiteral("/friends/requests/%1/accept").arg(requestId), {}, true, [this](const QJsonObject &) {
        loadFriendRequests();
        loadFriends();
    });
}

void NetCordClient::declineFriendRequest(const QString &requestId)
{
    if (requestId.isEmpty()) {
        return;
    }
    postJson(QStringLiteral("/friends/requests/%1/decline").arg(requestId), {}, true, [this](const QJsonObject &) {
        loadFriendRequests();
    });
}

void NetCordClient::loadDMs()
{
    getJson(QStringLiteral("/dm"), true, [this](const QJsonObject &payload) {
        setDMConversations(jsonArrayToVariantList(payload.value(QStringLiteral("conversations")).toArray()));
    });
}

void NetCordClient::createDM(const QString &userId)
{
    if (userId.trimmed().isEmpty()) {
        return;
    }
    postJson(QStringLiteral("/dm"),
             QJsonObject{{QStringLiteral("member_ids"), QJsonArray{userId.trimmed()}}},
             true,
             [this](const QJsonObject &) { loadDMs(); });
}

void NetCordClient::uploadFile(const QUrl &fileUrl)
{
    if (!m_authenticated) {
        setStatusMessage(QStringLiteral("Sign in before uploading files"));
        return;
    }

    const QString localPath = fileUrl.toLocalFile();
    if (localPath.isEmpty()) {
        setStatusMessage(QStringLiteral("Only local files can be uploaded"));
        return;
    }

    auto *file = new QFile(localPath);
    if (!file->open(QIODevice::ReadOnly)) {
        setStatusMessage(QStringLiteral("Could not open file: %1").arg(QFileInfo(localPath).fileName()));
        file->deleteLater();
        return;
    }

    auto *multiPart = new QHttpMultiPart(QHttpMultiPart::FormDataType);
    QHttpPart filePart;
    QString filename = QFileInfo(localPath).fileName();
    filename.replace(QLatin1Char('"'), QLatin1Char('_'));
    filePart.setRawHeader("Content-Disposition",
                          QStringLiteral("form-data; name=\"file\"; filename=\"%1\"").arg(filename).toUtf8());
    filePart.setBodyDevice(file);
    file->setParent(multiPart);
    multiPart->append(filePart);

    beginRequest();
    QNetworkReply *reply = m_network.post(makeRequest(QStringLiteral("/files/upload"), true, false), multiPart);
    multiPart->setParent(reply);
    connect(reply, &QNetworkReply::finished, this, [this, reply]() {
        handleReply(reply, true, [this](const QJsonObject &payload) {
            addPendingAttachment(payload.toVariantMap());
            setStatusMessage(QStringLiteral("File ready to attach"));
        });
    });
}

void NetCordClient::removePendingAttachment(const QString &attachmentId)
{
    QVariantList updated;
    for (const QVariant &attachment : m_pendingAttachments) {
        if (attachment.toMap().value(QStringLiteral("id")).toString() != attachmentId) {
            updated.append(attachment);
        }
    }
    setPendingAttachments(updated);
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
    query.addQueryItem(QStringLiteral("token"), m_token);
    url.setQuery(query);
    return url;
}

QNetworkRequest NetCordClient::makeRequest(const QString &path, bool withAuth, bool jsonContent) const
{
    QNetworkRequest request(apiUrl(path));
    if (jsonContent) {
        request.setHeader(QNetworkRequest::ContentTypeHeader, QStringLiteral("application/json"));
    }
    if (withAuth && !m_token.isEmpty()) {
        request.setRawHeader("Authorization", QByteArrayLiteral("Bearer ") + m_token.toUtf8());
    }
    return request;
}

void NetCordClient::getJson(const QString &path, bool withAuth, JsonCallback onSuccess)
{
    beginRequest();
    QNetworkReply *reply = m_network.get(makeRequest(path, withAuth));
    connect(reply, &QNetworkReply::finished, this, [this, reply, withAuth, onSuccess]() {
        handleReply(reply, withAuth, onSuccess);
    });
}

void NetCordClient::postJson(const QString &path, const QJsonObject &payload, bool withAuth, JsonCallback onSuccess)
{
    beginRequest();
    QNetworkReply *reply = m_network.post(makeRequest(path, withAuth), QJsonDocument(payload).toJson(QJsonDocument::Compact));
    connect(reply, &QNetworkReply::finished, this, [this, reply, withAuth, onSuccess]() {
        handleReply(reply, withAuth, onSuccess);
    });
}

void NetCordClient::patchJson(const QString &path, const QJsonObject &payload, bool withAuth, JsonCallback onSuccess)
{
    beginRequest();
    QNetworkReply *reply = m_network.sendCustomRequest(makeRequest(path, withAuth), QByteArrayLiteral("PATCH"), QJsonDocument(payload).toJson(QJsonDocument::Compact));
    connect(reply, &QNetworkReply::finished, this, [this, reply, withAuth, onSuccess]() {
        handleReply(reply, withAuth, onSuccess);
    });
}

void NetCordClient::deleteRequest(const QString &path, bool withAuth, JsonCallback onSuccess)
{
    beginRequest();
    QNetworkReply *reply = m_network.deleteResource(makeRequest(path, withAuth, false));
    connect(reply, &QNetworkReply::finished, this, [this, reply, withAuth, onSuccess]() {
        handleReply(reply, withAuth, onSuccess);
    });
}

void NetCordClient::handleReply(QNetworkReply *reply, bool withAuth, JsonCallback onSuccess)
{
    const QByteArray body = reply->readAll();
    const int statusCode = reply->attribute(QNetworkRequest::HttpStatusCodeAttribute).toInt();
    const QString networkError = reply->error() == QNetworkReply::NoError ? QString() : reply->errorString();
    const QJsonDocument document = QJsonDocument::fromJson(body);
    const QJsonObject payload = document.object();
    reply->deleteLater();
    endRequest();

    if (statusCode >= 200 && statusCode < 300) {
        onSuccess(payload);
        return;
    }

    if (withAuth && statusCode == 401) {
        clearSession(true);
    }
    setStatusMessage(errorMessageFromPayload(payload, statusCode, networkError));
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
    loadFriends();
    loadFriendRequests();
    loadDMs();
    connectGateway();
}

void NetCordClient::connectGateway()
{
    if (!m_authenticated || m_token.isEmpty()) {
        return;
    }

    if (m_gateway.state() != QAbstractSocket::UnconnectedState) {
        m_manualGatewayClose = true;
        m_gateway.abort();
    }
    m_manualGatewayClose = false;
    m_gateway.open(gatewayUrl());
}

void NetCordClient::disconnectGateway()
{
    m_manualGatewayClose = true;
    m_gatewayReconnectTimer.stop();
    m_heartbeatTimer.stop();
    if (m_gateway.state() != QAbstractSocket::UnconnectedState) {
        m_gateway.close();
    }
    setGatewayConnected(false);
}

void NetCordClient::scheduleGatewayReconnect()
{
    if (m_gatewayReconnectTimer.isActive()) {
        return;
    }

    const int delayMs = qMin(30000, 1000 * (1 << qMin(m_gatewayReconnectAttempts, 5)));
    ++m_gatewayReconnectAttempts;
    m_gatewayReconnectTimer.start(delayMs);
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

    if (type == QStringLiteral("message.updated")) {
        const QJsonObject data = envelope.value(QStringLiteral("data")).toObject();
        if (data.value(QStringLiteral("channel_id")).toString() == m_selectedChannel.value(QStringLiteral("id")).toString()) {
            addOrUpdateMessage(normalizeMessage(data));
        }
        return;
    }

    if (type == QStringLiteral("message.deleted")) {
        const QJsonObject data = envelope.value(QStringLiteral("data")).toObject();
        if (data.value(QStringLiteral("channel_id")).toString() == m_selectedChannel.value(QStringLiteral("id")).toString()) {
            removeMessage(data.value(QStringLiteral("message_id")).toString());
        }
        return;
    }

    if (type == QStringLiteral("typing.start")) {
        const QJsonObject data = envelope.value(QStringLiteral("data")).toObject();
        if (data.value(QStringLiteral("channel_id")).toString() == m_selectedChannel.value(QStringLiteral("id")).toString()
            && data.value(QStringLiteral("user_id")).toString() != m_currentUser.value(QStringLiteral("id")).toString()) {
            const QString username = data.value(QStringLiteral("username")).toString(QStringLiteral("Someone"));
            setTypingText(QStringLiteral("%1 is typing...").arg(username));
            m_typingIndicatorTimer.start(3500);
        }
        return;
    }

    if (type == QStringLiteral("typing.stop")) {
        const QJsonObject data = envelope.value(QStringLiteral("data")).toObject();
        if (data.value(QStringLiteral("channel_id")).toString() == m_selectedChannel.value(QStringLiteral("id")).toString()) {
            setTypingText({});
            m_typingIndicatorTimer.stop();
        }
        return;
    }

    if (type == QStringLiteral("friend.requested") || type == QStringLiteral("friend.accepted")) {
        loadFriendRequests();
        loadFriends();
        return;
    }

    if (type == QStringLiteral("dm.message.created")) {
        loadDMs();
        return;
    }

    if (type == QStringLiteral("voice.joined")) {
        const QJsonObject data = envelope.value(QStringLiteral("data")).toObject();
        setVoiceStatus(QStringLiteral("Voice active: %1").arg(data.value(QStringLiteral("room")).toString()));
        return;
    }

    if (type == QStringLiteral("voice.left")) {
        const QJsonObject data = envelope.value(QStringLiteral("data")).toObject();
        if (data.value(QStringLiteral("user_id")).toString() == m_currentUser.value(QStringLiteral("id")).toString()) {
            setVoiceConnected(false);
        }
        setVoiceStatus(QStringLiteral("Voice user left"));
        return;
    }

    if (type == QStringLiteral("job.progress") || type == QStringLiteral("job.completed") || type == QStringLiteral("ai.job.updated")) {
        addOrUpdateAIJob(envelope.value(QStringLiteral("data")).toObject().toVariantMap());
        return;
    }

    if (type == QStringLiteral("error")) {
        const QJsonObject data = envelope.value(QStringLiteral("data")).toObject();
        setStatusMessage(data.value(QStringLiteral("message")).toString(QStringLiteral("Gateway error")));
    }
}

void NetCordClient::sendGatewayEvent(const QJsonObject &event)
{
    if (m_gateway.state() != QAbstractSocket::ConnectedState) {
        return;
    }
    m_gateway.sendTextMessage(QString::fromUtf8(QJsonDocument(event).toJson(QJsonDocument::Compact)));
}

void NetCordClient::sendHeartbeat()
{
    if (m_gateway.state() != QAbstractSocket::ConnectedState) {
        return;
    }

    const QJsonObject event{{QStringLiteral("type"), QStringLiteral("heartbeat")}};
    sendGatewayEvent(event);
}

void NetCordClient::clearTypingLater()
{
    setTypingText({});
}

void NetCordClient::initializeCache()
{
    if (QSqlDatabase::contains(cacheConnectionName)) {
        m_cache = QSqlDatabase::database(cacheConnectionName);
    } else {
        m_cache = QSqlDatabase::addDatabase(QStringLiteral("QSQLITE"), cacheConnectionName);
        const QString dataDir = QStandardPaths::writableLocation(QStandardPaths::AppDataLocation);
        QDir().mkpath(dataDir);
        m_cache.setDatabaseName(dataDir + QStringLiteral("/netcord-cache.sqlite"));
    }
    if (!m_cache.open()) {
        setStatusMessage(QStringLiteral("Local cache unavailable: %1").arg(m_cache.lastError().text()));
        return;
    }

    QSqlQuery query(m_cache);
    query.exec(QStringLiteral("CREATE TABLE IF NOT EXISTS servers (id TEXT PRIMARY KEY, payload TEXT NOT NULL, updated_at INTEGER NOT NULL)"));
    query.exec(QStringLiteral("CREATE TABLE IF NOT EXISTS channels (id TEXT PRIMARY KEY, server_id TEXT NOT NULL, payload TEXT NOT NULL, updated_at INTEGER NOT NULL)"));
    query.exec(QStringLiteral("CREATE TABLE IF NOT EXISTS messages (id TEXT PRIMARY KEY, channel_id TEXT NOT NULL, payload TEXT NOT NULL, created_at TEXT, updated_at INTEGER NOT NULL)"));
    query.exec(QStringLiteral("CREATE INDEX IF NOT EXISTS channels_server_idx ON channels(server_id)"));
    query.exec(QStringLiteral("CREATE INDEX IF NOT EXISTS messages_channel_idx ON messages(channel_id, created_at)"));
}

void NetCordClient::cacheServers(const QVariantList &servers)
{
    if (!m_cache.isOpen()) {
        return;
    }
    QSqlQuery query(m_cache);
    query.prepare(QStringLiteral("INSERT OR REPLACE INTO servers (id, payload, updated_at) VALUES (?, ?, strftime('%s','now'))"));
    for (const QVariant &item : servers) {
        const QVariantMap server = item.toMap();
        query.addBindValue(server.value(QStringLiteral("id")).toString());
        query.addBindValue(QString::fromUtf8(QJsonDocument::fromVariant(server).toJson(QJsonDocument::Compact)));
        query.exec();
    }
}

void NetCordClient::cacheChannels(const QString &serverId, const QVariantList &channels)
{
    if (!m_cache.isOpen() || serverId.isEmpty()) {
        return;
    }
    QSqlQuery query(m_cache);
    query.prepare(QStringLiteral("INSERT OR REPLACE INTO channels (id, server_id, payload, updated_at) VALUES (?, ?, ?, strftime('%s','now'))"));
    for (const QVariant &item : channels) {
        const QVariantMap channel = item.toMap();
        query.addBindValue(channel.value(QStringLiteral("id")).toString());
        query.addBindValue(serverId);
        query.addBindValue(QString::fromUtf8(QJsonDocument::fromVariant(channel).toJson(QJsonDocument::Compact)));
        query.exec();
    }
}

void NetCordClient::cacheMessages(const QString &channelId, const QVariantList &messages)
{
    if (!m_cache.isOpen() || channelId.isEmpty()) {
        return;
    }
    QSqlQuery query(m_cache);
    query.prepare(QStringLiteral("INSERT OR REPLACE INTO messages (id, channel_id, payload, created_at, updated_at) VALUES (?, ?, ?, ?, strftime('%s','now'))"));
    for (const QVariant &item : messages) {
        const QVariantMap message = item.toMap();
        query.addBindValue(message.value(QStringLiteral("id")).toString());
        query.addBindValue(channelId);
        query.addBindValue(QString::fromUtf8(QJsonDocument::fromVariant(message).toJson(QJsonDocument::Compact)));
        query.addBindValue(message.value(QStringLiteral("created_at")).toString());
        query.exec();
    }
}

QVariantList NetCordClient::cachedServers() const
{
    QVariantList servers;
    if (!m_cache.isOpen()) {
        return servers;
    }
    QSqlQuery query(m_cache);
    query.exec(QStringLiteral("SELECT payload FROM servers ORDER BY updated_at ASC"));
    while (query.next()) {
        servers.append(QJsonDocument::fromJson(query.value(0).toString().toUtf8()).object().toVariantMap());
    }
    return servers;
}

QVariantList NetCordClient::cachedChannels(const QString &serverId) const
{
    QVariantList channels;
    if (!m_cache.isOpen() || serverId.isEmpty()) {
        return channels;
    }
    QSqlQuery query(m_cache);
    query.prepare(QStringLiteral("SELECT payload FROM channels WHERE server_id = ? ORDER BY updated_at ASC"));
    query.addBindValue(serverId);
    query.exec();
    while (query.next()) {
        channels.append(QJsonDocument::fromJson(query.value(0).toString().toUtf8()).object().toVariantMap());
    }
    return channels;
}

QVariantList NetCordClient::cachedMessages(const QString &channelId) const
{
    QVariantList messages;
    if (!m_cache.isOpen() || channelId.isEmpty()) {
        return messages;
    }
    QSqlQuery query(m_cache);
    query.prepare(QStringLiteral("SELECT payload FROM messages WHERE channel_id = ? ORDER BY created_at ASC LIMIT 100"));
    query.addBindValue(channelId);
    query.exec();
    while (query.next()) {
        messages.append(QJsonDocument::fromJson(query.value(0).toString().toUtf8()).object().toVariantMap());
    }
    return messages;
}

void NetCordClient::beginRequest()
{
    ++m_activeRequests;
    setBusy(true);
}

void NetCordClient::endRequest()
{
    if (m_activeRequests > 0) {
        --m_activeRequests;
    }
    setBusy(m_activeRequests > 0);
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

void NetCordClient::setSearchResults(const QVariantList &messages)
{
    m_searchResults = messages;
    emit searchResultsChanged();
}

void NetCordClient::setFriendRequests(const QVariantList &requests)
{
    m_friendRequests = requests;
    emit friendRequestsChanged();
}

void NetCordClient::setFriends(const QVariantList &friends)
{
    m_friends = friends;
    emit friendsChanged();
}

void NetCordClient::setDMConversations(const QVariantList &conversations)
{
    m_dmConversations = conversations;
    emit dmConversationsChanged();
}

void NetCordClient::setAIJobs(const QVariantList &jobs)
{
    m_aiJobs = jobs;
    emit aiJobsChanged();
}

void NetCordClient::setServerMembers(const QVariantList &members)
{
    m_serverMembers = members;
    emit serverMembersChanged();
}

void NetCordClient::setRoles(const QVariantList &roles)
{
    m_roles = roles;
    emit rolesChanged();
}

void NetCordClient::setPendingAttachments(const QVariantList &attachments)
{
    m_pendingAttachments = attachments;
    emit pendingAttachmentsChanged();
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

void NetCordClient::setTypingText(const QString &typingText)
{
    if (m_typingText == typingText) {
        return;
    }
    m_typingText = typingText;
    emit typingTextChanged();
}

void NetCordClient::setVoiceConnected(bool connected)
{
    if (m_voiceConnected == connected) {
        return;
    }
    m_voiceConnected = connected;
    emit voiceConnectedChanged();
}

void NetCordClient::setVoiceStatus(const QString &status)
{
    if (m_voiceStatus == status) {
        return;
    }
    m_voiceStatus = status;
    emit voiceStatusChanged();
}

void NetCordClient::addOrUpdateMessage(const QVariantMap &message)
{
    const QString id = message.value(QStringLiteral("id")).toString();
    QVariantList updated = m_messages;
    for (int i = 0; i < updated.size(); ++i) {
        if (updated.at(i).toMap().value(QStringLiteral("id")).toString() == id) {
            updated[i] = message;
            setMessages(updated);
            cacheMessages(m_selectedChannel.value(QStringLiteral("id")).toString(), updated);
            return;
        }
    }

    updated.append(message);
    setMessages(updated);
    cacheMessages(m_selectedChannel.value(QStringLiteral("id")).toString(), updated);
}

void NetCordClient::removeMessage(const QString &messageId)
{
    QVariantList updated;
    for (const QVariant &message : m_messages) {
        if (message.toMap().value(QStringLiteral("id")).toString() != messageId) {
            updated.append(message);
        }
    }
    setMessages(updated);
    cacheMessages(m_selectedChannel.value(QStringLiteral("id")).toString(), updated);
}

void NetCordClient::addOrUpdateAIJob(const QVariantMap &job)
{
    const QString id = job.value(QStringLiteral("id")).toString();
    QVariantList updated = m_aiJobs;
    for (int i = 0; i < updated.size(); ++i) {
        if (updated.at(i).toMap().value(QStringLiteral("id")).toString() == id) {
            updated[i] = job;
            setAIJobs(updated);
            return;
        }
    }
    updated.prepend(job);
    setAIJobs(updated);
}

void NetCordClient::addPendingAttachment(const QVariantMap &attachment)
{
    QVariantList updated = m_pendingAttachments;
    updated.append(attachment);
    setPendingAttachments(updated);
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
    setSearchResults({});
    setFriendRequests({});
    setFriends({});
    setDMConversations({});
    setAIJobs({});
    setServerMembers({});
    setRoles({});
    setPendingAttachments({});
    setSelectedServer({});
    setSelectedChannel({});
    setTypingText({});
    setVoiceConnected(false);
    setVoiceStatus({});
    m_voiceChannelId.clear();
}

QString NetCordClient::errorMessageFromPayload(const QJsonObject &payload, int statusCode, const QString &networkError) const
{
    const QJsonObject error = payload.value(QStringLiteral("error")).toObject();
    const QString message = error.value(QStringLiteral("message")).toString();
    if (!message.isEmpty()) {
        return message;
    }
    if (!networkError.isEmpty()) {
        return QStringLiteral("Network error: %1").arg(networkError);
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
