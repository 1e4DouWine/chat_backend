package service

import (
	"chat_backend/internal/dao"
	"chat_backend/internal/dto"
	"chat_backend/internal/model"
	"context"
	"time"

	"gorm.io/gorm"
)

type MessageService struct {
	db *gorm.DB
}

func NewMessageService(db *gorm.DB) *MessageService {
	return &MessageService{
		db: db,
	}
}

func (s *MessageService) GetPrivateMessages(ctx context.Context, userID string, targetUserID string, limit int, cursor string) (*dto.GetMessagesResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	q := dao.Use(s.db).Message
	do := q.WithContext(ctx)

	query := do.Where(
		q.Type.Eq(string(model.MessageTypePrivate)),
		q.FromUserID.Eq(userID),
		q.TargetID.Eq(targetUserID),
	)
	query = query.Or(
		q.Type.Eq(string(model.MessageTypePrivate)),
		q.FromUserID.Eq(targetUserID),
		q.TargetID.Eq(userID),
	)

	if cursor != "" {
		cursorTime, err := time.Parse(time.RFC3339Nano, cursor)
		if err == nil {
			query = query.Where(q.CreatedAt.Lt(cursorTime))
		}
	}

	messages, err := query.Order(q.CreatedAt.Desc()).Limit(limit).Find()
	if err != nil {
		return nil, err
	}

	var messageResponses []dto.MessageResponse
	var nextCursor string
	hasMore := false

	for _, msg := range messages {
		fromUser, _ := s.getUserInfo(ctx, msg.FromUserID)
		targetUser, _ := s.getUserInfo(ctx, msg.TargetID)

		messageResponses = append(messageResponses, dto.MessageResponse{
			MessageID:  msg.ID,
			FromUserID: msg.FromUserID,
			TargetID:   msg.TargetID,
			Type:       string(msg.Type),
			Content:    msg.Content,
			CreatedAt:  msg.CreatedAt,
			FromUser:   fromUser,
			TargetUser: targetUser,
		})
	}

	if len(messages) == limit {
		nextCursor = messages[len(messages)-1].CreatedAt.Format(time.RFC3339Nano)
		hasMore = true
	}

	return &dto.GetMessagesResponse{
		Messages:   messageResponses,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

func (s *MessageService) GetGroupMessages(ctx context.Context, groupID string, limit int, cursor string) (*dto.GetMessagesResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	q := dao.Use(s.db).Message
	do := q.WithContext(ctx)

	query := do.Where(
		q.Type.Eq(string(model.MessageTypeGroup)),
		q.TargetID.Eq(groupID),
	)

	if cursor != "" {
		cursorTime, err := time.Parse(time.RFC3339Nano, cursor)
		if err == nil {
			query = query.Where(q.CreatedAt.Lt(cursorTime))
		}
	}

	messages, err := query.Order(q.CreatedAt.Desc()).Limit(limit).Find()
	if err != nil {
		return nil, err
	}

	var messageResponses []dto.MessageResponse
	var nextCursor string
	hasMore := false

	for _, msg := range messages {
		fromUser, _ := s.getUserInfo(ctx, msg.FromUserID)
		groupInfo, _ := s.getGroupInfo(ctx, msg.TargetID)

		messageResponses = append(messageResponses, dto.MessageResponse{
			MessageID:   msg.ID,
			FromUserID:  msg.FromUserID,
			TargetID:    msg.TargetID,
			Type:        string(msg.Type),
			Content:     msg.Content,
			CreatedAt:   msg.CreatedAt,
			FromUser:    fromUser,
			TargetGroup: groupInfo,
		})
	}

	if len(messages) == limit {
		nextCursor = messages[len(messages)-1].CreatedAt.Format(time.RFC3339Nano)
		hasMore = true
	}

	return &dto.GetMessagesResponse{
		Messages:   messageResponses,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

func (s *MessageService) getUserInfo(ctx context.Context, userID string) (*dto.UserInfo, error) {
	q := dao.Use(s.db).User
	do := q.WithContext(ctx)

	user, err := do.Where(q.ID.Eq(userID)).First()
	if err != nil {
		return nil, err
	}

	avatarUrl := s.generateAvatarUrl(user.ID, user.Username)

	return &dto.UserInfo{
		UserID:   user.ID,
		Username: user.Username,
		Avatar:   avatarUrl,
	}, nil
}

func (s *MessageService) getGroupInfo(ctx context.Context, groupID string) (*dto.GroupInfo, error) {
	q := dao.Use(s.db).Group
	do := q.WithContext(ctx)

	group, err := do.Where(q.ID.Eq(groupID)).First()
	if err != nil {
		return nil, err
	}

	return &dto.GroupInfo{
		GroupID: group.ID,
		Name:    group.Name,
	}, nil
}

func (s *MessageService) generateAvatarUrl(userID string, username string) string {
	color := colorFromUUID(userID)
	return "https://ui-avatars.com/api/?name=" + username + "&background=" + color + "&rounded=true&size=128"
}
