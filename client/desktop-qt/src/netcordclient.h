#pragma once

#include <QJsonArray>
#include <QJsonObject>
#include <QNetworkAccessManager>
#include <QObject>
#include <QSettings>
#include <QSqlDatabase>
#include <QTimer>
#include <QUrl>
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
    Q_PROPERTY(QVariantList searchResults READ searchResults NOTIFY searchResultsChanged)
    Q_PROPERTY(QVariantList friendRequests READ friendRequests NOTIFY friendRequestsChanged)
    Q_PROPERTY(QVariantList friends READ friends NOTIFY friendsChanged)
    Q_PROPERTY(QVariantList dmConversations READ dmConversations NOTIFY dmConversationsChanged)
    Q_PROPERTY(QVariantList aiJobs READ aiJobs NOTIFY aiJobsChanged)
    Q_PROPERTY(QVariantList serverMembers READ serverMembers NOTIFY serverMembersChanged)
    Q_PROPERTY(QVariantList roles READ roles NOTIFY rolesChanged)
    Q_PROPERTY(QVariantList pendingAttachments READ pendingAttachments NOTIFY pendingAttachmentsChanged)
    Q_PROPERTY(QVariantMap selectedServer READ selectedServer NOTIFY selectedServerChanged)
    Q_PROPERTY(QVariantMap selectedChannel READ selectedChannel NOTIFY selectedChannelChanged)
    Q_PROPERTY(QString typingText READ typingText NOTIFY typingTextChanged)
    Q_PROPERTY(bool voiceConnected READ voiceConnected NOTIFY voiceConnectedChanged)
    Q_PROPERTY(QString voiceStatus READ voiceStatus NOTIFY voiceStatusChanged)

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
    QVariantList searchResults() const;
    QVariantList friendRequests() const;
    QVariantList friends() const;
    QVariantList dmConversations() const;
    QVariantList aiJobs() const;
    QVariantList serverMembers() const;
    QVariantList roles() const;
    QVariantList pendingAttachments() const;
    QVariantMap selectedServer() const;
    QVariantMap selectedChannel() const;
    QString typingText() const;
    bool voiceConnected() const;
    QString voiceStatus() const;

    Q_INVOKABLE void initialize();
    Q_INVOKABLE void login(const QString &email, const QString &password);
    Q_INVOKABLE void registerAccount(const QString &username, const QString &email, const QString &password);
    Q_INVOKABLE void logout();
    Q_INVOKABLE void loadServers();
    Q_INVOKABLE void refreshServers();
    Q_INVOKABLE void refreshChannels();
    Q_INVOKABLE void refreshMessages();
    Q_INVOKABLE void reconnectGateway();
    Q_INVOKABLE void selectServer(const QString &serverId);
    Q_INVOKABLE void selectChannel(const QString &channelId);
    Q_INVOKABLE void sendMessage(const QString &content);
    Q_INVOKABLE void editMessage(const QString &messageId, const QString &content);
    Q_INVOKABLE void deleteMessage(const QString &messageId);
    Q_INVOKABLE void loadOlderMessages();
    Q_INVOKABLE void searchMessages(const QString &query);
    Q_INVOKABLE void clearSearchResults();
    Q_INVOKABLE void createServer(const QString &name, const QString &description);
    Q_INVOKABLE void createChannel(const QString &name, const QString &type);
    Q_INVOKABLE void updateProfile(const QString &displayName, const QString &status);
    Q_INVOKABLE void loadServerMembers();
    Q_INVOKABLE void loadRoles();
    Q_INVOKABLE void createRole(const QString &name, qint64 permissions);
    Q_INVOKABLE void createInvite();
    Q_INVOKABLE void joinInvite(const QString &code);
    Q_INVOKABLE void joinVoice(const QString &channelId);
    Q_INVOKABLE void leaveVoice();
    Q_INVOKABLE void clearCache();
    Q_INVOKABLE void sendTypingStart();
    Q_INVOKABLE void sendTypingStop();
    Q_INVOKABLE void openAttachment(const QString &downloadUrl);
    Q_INVOKABLE void loadFriends();
    Q_INVOKABLE void loadFriendRequests();
    Q_INVOKABLE void sendFriendRequest(const QString &recipientId);
    Q_INVOKABLE void acceptFriendRequest(const QString &requestId);
    Q_INVOKABLE void declineFriendRequest(const QString &requestId);
    Q_INVOKABLE void loadDMs();
    Q_INVOKABLE void createDM(const QString &userId);
    Q_INVOKABLE void uploadFile(const QUrl &fileUrl);
    Q_INVOKABLE void removePendingAttachment(const QString &attachmentId);
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
    void searchResultsChanged();
    void friendRequestsChanged();
    void friendsChanged();
    void dmConversationsChanged();
    void aiJobsChanged();
    void serverMembersChanged();
    void rolesChanged();
    void pendingAttachmentsChanged();
    void selectedServerChanged();
    void selectedChannelChanged();
    void typingTextChanged();
    void voiceConnectedChanged();
    void voiceStatusChanged();

private:
    using JsonCallback = std::function<void(const QJsonObject &)>;

    QUrl apiUrl(const QString &path) const;
    QUrl gatewayUrl() const;
    QNetworkRequest makeRequest(const QString &path, bool withAuth, bool jsonContent = true) const;
    void getJson(const QString &path, bool withAuth, JsonCallback onSuccess);
    void postJson(const QString &path, const QJsonObject &payload, bool withAuth, JsonCallback onSuccess);
    void patchJson(const QString &path, const QJsonObject &payload, bool withAuth, JsonCallback onSuccess);
    void deleteRequest(const QString &path, bool withAuth, JsonCallback onSuccess);
    void handleReply(QNetworkReply *reply, bool withAuth, JsonCallback onSuccess);
    void handleAuthSuccess(const QJsonObject &payload);
    void connectGateway();
    void disconnectGateway();
    void scheduleGatewayReconnect();
    void handleGatewayTextMessage(const QString &message);
    void sendGatewayEvent(const QJsonObject &event);
    void sendHeartbeat();
    void clearTypingLater();
    void initializeCache();
    void cacheServers(const QVariantList &servers);
    void cacheChannels(const QString &serverId, const QVariantList &channels);
    void cacheMessages(const QString &channelId, const QVariantList &messages);
    QVariantList cachedServers() const;
    QVariantList cachedChannels(const QString &serverId) const;
    QVariantList cachedMessages(const QString &channelId) const;
    void beginRequest();
    void endRequest();
    void setAuthenticated(bool authenticated);
    void setBusy(bool busy);
    void setGatewayConnected(bool connected);
    void setStatusMessage(const QString &message);
    void setCurrentUser(const QVariantMap &user);
    void setServers(const QVariantList &servers);
    void setChannels(const QVariantList &channels);
    void setMessages(const QVariantList &messages);
    void setSearchResults(const QVariantList &messages);
    void setFriendRequests(const QVariantList &requests);
    void setFriends(const QVariantList &friends);
    void setDMConversations(const QVariantList &conversations);
    void setAIJobs(const QVariantList &jobs);
    void setServerMembers(const QVariantList &members);
    void setRoles(const QVariantList &roles);
    void setPendingAttachments(const QVariantList &attachments);
    void setSelectedServer(const QVariantMap &server);
    void setSelectedChannel(const QVariantMap &channel);
    void setTypingText(const QString &typingText);
    void setVoiceConnected(bool connected);
    void setVoiceStatus(const QString &status);
    void addOrUpdateMessage(const QVariantMap &message);
    void removeMessage(const QString &messageId);
    void addOrUpdateAIJob(const QVariantMap &job);
    void addPendingAttachment(const QVariantMap &attachment);
    void clearSession(bool clearStoredToken);
    QString errorMessageFromPayload(const QJsonObject &payload, int statusCode, const QString &networkError) const;
    static QVariantList jsonArrayToVariantList(const QJsonArray &array);
    static QVariantMap objectById(const QVariantList &items, const QString &id);
    static QVariantMap normalizeMessage(const QJsonObject &message);

    QNetworkAccessManager m_network;
    QWebSocket m_gateway;
    QTimer m_heartbeatTimer;
    QTimer m_gatewayReconnectTimer;
    QTimer m_typingStopTimer;
    QTimer m_typingIndicatorTimer;
    QSettings m_settings;
    QSqlDatabase m_cache;
    QString m_apiBaseUrl;
    QString m_token;
    bool m_authenticated = false;
    bool m_busy = false;
    bool m_gatewayConnected = false;
    bool m_manualGatewayClose = false;
    int m_activeRequests = 0;
    int m_gatewayReconnectAttempts = 0;
    QString m_statusMessage;
    QVariantMap m_currentUser;
    QVariantList m_servers;
    QVariantList m_channels;
    QVariantList m_messages;
    QVariantList m_searchResults;
    QVariantList m_friendRequests;
    QVariantList m_friends;
    QVariantList m_dmConversations;
    QVariantList m_aiJobs;
    QVariantList m_serverMembers;
    QVariantList m_roles;
    QVariantList m_pendingAttachments;
    QVariantMap m_selectedServer;
    QVariantMap m_selectedChannel;
    QString m_typingText;
    QString m_lastTypingChannelId;
    QString m_voiceChannelId;
    bool m_voiceConnected = false;
    QString m_voiceStatus;
};
