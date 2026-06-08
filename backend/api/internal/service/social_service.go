package service

import (
	"context"
	"errors"
	"strings"

	"netcord/backend/api/internal/models"
	"netcord/backend/api/internal/repository"

	"github.com/google/uuid"
)

type CreateFriendRequestInput struct {
	RecipientID uuid.UUID `json:"recipient_id"`
}

type CreateDMInput struct {
	MemberIDs []uuid.UUID `json:"member_ids"`
	Name      string      `json:"name"`
}

type CreateDMMessageInput struct {
	Content string `json:"content"`
}

type FriendRequestsResponse struct {
	Requests []models.PublicFriendRequest `json:"requests"`
}

type FriendsResponse struct {
	Friends []models.PublicFriend `json:"friends"`
}

type DMConversationsResponse struct {
	Conversations []models.PublicDMConversation `json:"conversations"`
}

type DMMessagesResponse struct {
	Messages []models.PublicDMMessage `json:"messages"`
}

type CreateDMMessageResult struct {
	Message   models.PublicDMMessage
	MemberIDs []uuid.UUID
}

type SocialService struct {
	social repository.SocialRepository
	users  repository.UserRepository
}

func NewSocialService(social repository.SocialRepository, users repository.UserRepository) *SocialService {
	return &SocialService{social: social, users: users}
}

func (s *SocialService) CreateFriendRequest(ctx context.Context, requesterID uuid.UUID, input CreateFriendRequestInput) (models.PublicFriendRequest, error) {
	if fields := validateCreateFriendRequestInput(requesterID, input); len(fields) > 0 {
		return models.PublicFriendRequest{}, &ValidationError{Fields: fields}
	}
	if _, err := s.users.GetByID(ctx, input.RecipientID); err != nil {
		return models.PublicFriendRequest{}, mapSocialRepositoryError(err)
	}

	request := models.FriendRequest{
		ID:          uuid.New(),
		RequesterID: requesterID,
		RecipientID: input.RecipientID,
		Status:      models.FriendRequestPending,
	}
	created, err := s.social.CreateFriendRequest(ctx, request)
	if err != nil {
		return models.PublicFriendRequest{}, mapSocialRepositoryError(err)
	}
	return created.Public(), nil
}

func (s *SocialService) ListFriendRequests(ctx context.Context, userID uuid.UUID) (FriendRequestsResponse, error) {
	requests, err := s.social.ListFriendRequestsForUser(ctx, userID)
	if err != nil {
		return FriendRequestsResponse{}, mapSocialRepositoryError(err)
	}
	return FriendRequestsResponse{Requests: publicFriendRequests(requests)}, nil
}

func (s *SocialService) AcceptFriendRequest(ctx context.Context, userID, requestID uuid.UUID) (models.PublicFriendRequest, error) {
	request, err := s.social.AcceptFriendRequest(ctx, requestID, userID)
	if err != nil {
		return models.PublicFriendRequest{}, mapSocialRepositoryError(err)
	}
	return request.Public(), nil
}

func (s *SocialService) DeclineFriendRequest(ctx context.Context, userID, requestID uuid.UUID) (models.PublicFriendRequest, error) {
	request, err := s.social.DeclineFriendRequest(ctx, requestID, userID)
	if err != nil {
		return models.PublicFriendRequest{}, mapSocialRepositoryError(err)
	}
	return request.Public(), nil
}

func (s *SocialService) ListFriends(ctx context.Context, userID uuid.UUID) (FriendsResponse, error) {
	friends, err := s.social.ListFriends(ctx, userID)
	if err != nil {
		return FriendsResponse{}, mapSocialRepositoryError(err)
	}
	return FriendsResponse{Friends: publicFriends(friends)}, nil
}

func (s *SocialService) RemoveFriend(ctx context.Context, userID, friendID uuid.UUID) error {
	return mapSocialRepositoryError(s.social.RemoveFriend(ctx, userID, friendID))
}

func (s *SocialService) CreateDM(ctx context.Context, userID uuid.UUID, input CreateDMInput) (models.PublicDMConversation, error) {
	input.Name = strings.TrimSpace(input.Name)
	memberIDs := normalizeMemberIDs(userID, input.MemberIDs)
	if fields := validateCreateDMInput(memberIDs, input); len(fields) > 0 {
		return models.PublicDMConversation{}, &ValidationError{Fields: fields}
	}

	for _, memberID := range memberIDs {
		if _, err := s.users.GetByID(ctx, memberID); err != nil {
			return models.PublicDMConversation{}, mapSocialRepositoryError(err)
		}
		if memberID == userID {
			continue
		}
		blocked, err := s.social.UsersBlocked(ctx, userID, memberID)
		if err != nil {
			return models.PublicDMConversation{}, mapSocialRepositoryError(err)
		}
		if blocked {
			return models.PublicDMConversation{}, ErrForbidden
		}
	}

	conversationType := models.DMTypeGroup
	var name *string
	if len(memberIDs) == 2 {
		conversationType = models.DMTypeDirect
	} else if input.Name != "" {
		name = &input.Name
	}

	conversation := models.DMConversation{
		ID:        uuid.New(),
		Type:      conversationType,
		Name:      name,
		CreatedBy: userID,
	}
	created, err := s.social.CreateDMConversation(ctx, conversation, memberIDs)
	if err != nil {
		return models.PublicDMConversation{}, mapSocialRepositoryError(err)
	}
	return created.Public(), nil
}

func (s *SocialService) ListDMs(ctx context.Context, userID uuid.UUID) (DMConversationsResponse, error) {
	conversations, err := s.social.ListDMConversationsForUser(ctx, userID)
	if err != nil {
		return DMConversationsResponse{}, mapSocialRepositoryError(err)
	}
	return DMConversationsResponse{Conversations: publicDMConversations(conversations)}, nil
}

func (s *SocialService) ListDMMessages(ctx context.Context, userID, conversationID uuid.UUID, limit int) (DMMessagesResponse, error) {
	if limit <= 0 {
		limit = defaultMessageLimit
	}
	if limit > maxMessageLimit {
		limit = maxMessageLimit
	}
	messages, err := s.social.ListDMMessagesForUser(ctx, conversationID, userID, limit)
	if err != nil {
		return DMMessagesResponse{}, mapSocialRepositoryError(err)
	}
	return DMMessagesResponse{Messages: publicDMMessages(messages)}, nil
}

func (s *SocialService) CreateDMMessage(ctx context.Context, userID, conversationID uuid.UUID, input CreateDMMessageInput) (CreateDMMessageResult, error) {
	input.Content = strings.TrimSpace(input.Content)
	if fields := validateCreateDMMessageInput(input); len(fields) > 0 {
		return CreateDMMessageResult{}, &ValidationError{Fields: fields}
	}

	message := models.DMMessage{
		ID:             uuid.New(),
		ConversationID: conversationID,
		AuthorID:       userID,
		Content:        input.Content,
	}
	created, err := s.social.CreateDMMessage(ctx, message)
	if err != nil {
		return CreateDMMessageResult{}, mapSocialRepositoryError(err)
	}
	memberIDs, err := s.social.ListDMMemberIDs(ctx, conversationID)
	if err != nil {
		return CreateDMMessageResult{}, mapSocialRepositoryError(err)
	}
	return CreateDMMessageResult{Message: created.Public(), MemberIDs: memberIDs}, nil
}

func mapSocialRepositoryError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, repository.ErrFriendRequestNotFound),
		errors.Is(err, repository.ErrFriendNotFound),
		errors.Is(err, repository.ErrDMConversationNotFound),
		errors.Is(err, repository.ErrUserNotFound):
		return ErrNotFound
	case errors.Is(err, repository.ErrFriendRequestConflict):
		return &ValidationError{Fields: map[string]string{"recipient_id": "friend request already exists"}}
	case errors.Is(err, repository.ErrForbidden):
		return ErrForbidden
	default:
		return err
	}
}

func normalizeMemberIDs(userID uuid.UUID, requested []uuid.UUID) []uuid.UUID {
	seen := map[uuid.UUID]struct{}{userID: {}}
	members := []uuid.UUID{userID}
	for _, id := range requested {
		if id == uuid.Nil {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		members = append(members, id)
	}
	return members
}

func publicFriendRequests(requests []models.FriendRequest) []models.PublicFriendRequest {
	public := make([]models.PublicFriendRequest, 0, len(requests))
	for _, request := range requests {
		public = append(public, request.Public())
	}
	return public
}

func publicFriends(friends []models.Friend) []models.PublicFriend {
	public := make([]models.PublicFriend, 0, len(friends))
	for _, friend := range friends {
		public = append(public, friend.Public())
	}
	return public
}

func publicDMConversations(conversations []models.DMConversation) []models.PublicDMConversation {
	public := make([]models.PublicDMConversation, 0, len(conversations))
	for _, conversation := range conversations {
		public = append(public, conversation.Public())
	}
	return public
}

func publicDMMessages(messages []models.DMMessage) []models.PublicDMMessage {
	public := make([]models.PublicDMMessage, 0, len(messages))
	for _, message := range messages {
		public = append(public, message.Public())
	}
	return public
}
