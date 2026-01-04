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
	).Where(
		do.Where(
			q.FromUserID.Eq(userID),
			q.TargetID.Eq(targetUserID),
		).Or(
			q.FromUserID.Eq(targetUserID),
			q.TargetID.Eq(userID),
		),
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

func (s *MessageService) SendPrivateMessage(ctx context.Context, fromUserID string, targetUserID string, content string, messageID string, isTargetOnline bool) (*model.Message, error) {
	var message *model.Message

	err := s.db.Transaction(func(tx *gorm.DB) error {
		q := dao.Use(tx).Message
		do := q.WithContext(ctx)

		message = &model.Message{
			FromUserID: fromUserID,
			TargetID:   targetUserID,
			Type:       model.MessageTypePrivate,
			Content:    content,
		}

		if messageID != "" {
			message.ID = messageID
		}

		if err := do.Create(message); err != nil {
			return err
		}

		if !isTargetOnline {
			rq := dao.Use(tx).MessageReceipt
			rdo := rq.WithContext(ctx)

			receipt := &model.MessageReceipt{
				MessageID:   message.ID,
				UserID:      targetUserID,
				IsDelivered: false,
			}

			if err := rdo.Create(receipt); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return message, nil
}

func (s *MessageService) SendGroupMessage(ctx context.Context, fromUserID string, groupID string, content string, messageID string, recipientIDs []string, onlineUserIDs map[string]bool) (*model.Message, error) {
	var message *model.Message

	err := s.db.Transaction(func(tx *gorm.DB) error {
		q := dao.Use(tx).Message
		do := q.WithContext(ctx)

		message = &model.Message{
			FromUserID: fromUserID,
			TargetID:   groupID,
			Type:       model.MessageTypeGroup,
			Content:    content,
		}

		if messageID != "" {
			message.ID = messageID
		}

		if err := do.Create(message); err != nil {
			return err
		}

		rq := dao.Use(tx).MessageReceipt
		rdo := rq.WithContext(ctx)

		var receipts []*model.MessageReceipt
		for _, recipientID := range recipientIDs {
			if !onlineUserIDs[recipientID] {
				receipt := &model.MessageReceipt{
					MessageID:   message.ID,
					UserID:      recipientID,
					IsDelivered: false,
				}
				receipts = append(receipts, receipt)
			}
		}

		if len(receipts) > 0 {
			if err := rdo.CreateInBatches(receipts, 100); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return message, nil
}

func (s *MessageService) GetUndeliveredMessages(ctx context.Context, userID string) ([]dto.MessageResponse, error) {
	rq := dao.Use(s.db).MessageReceipt
	rdo := rq.WithContext(ctx)

	receipts, err := rdo.Where(
		rq.UserID.Eq(userID),
		rq.IsDelivered.Is(false),
	).Order(rq.CreatedAt.Asc()).Find()
	if err != nil {
		return nil, err
	}

	if len(receipts) == 0 {
		return []dto.MessageResponse{}, nil
	}

	messageIDs := make([]string, 0, len(receipts))
	for _, receipt := range receipts {
		messageIDs = append(messageIDs, receipt.MessageID)
	}

	q := dao.Use(s.db).Message
	do := q.WithContext(ctx)

	messages, err := do.Where(q.ID.In(messageIDs...)).Order(q.CreatedAt.Asc()).Find()
	if err != nil {
		return nil, err
	}

	messageResponses := make([]dto.MessageResponse, 0, len(messages))
	for _, msg := range messages {
		fromUser, _ := s.getUserInfo(ctx, msg.FromUserID)

		if msg.Type == model.MessageTypePrivate {
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
		} else {
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
	}

	return messageResponses, nil
}

func (s *MessageService) MarkMessagesAsDelivered(ctx context.Context, userID string, messageIDs []string) error {
	rq := dao.Use(s.db).MessageReceipt
	rdo := rq.WithContext(ctx)

	_, err := rdo.Where(
		rq.UserID.Eq(userID),
		rq.MessageID.In(messageIDs...),
	).Update(rq.IsDelivered, true)

	return err
}

func (s *MessageService) GetConversationList(ctx context.Context, userID string) (*dto.GetConversationListResponse, error) {
	privateConversations, err := s.getPrivateConversations(ctx, userID)
	if err != nil {
		return nil, err
	}

	groupConversations, err := s.getGroupConversations(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &dto.GetConversationListResponse{
		PrivateConversations: privateConversations,
		GroupConversations:   groupConversations,
	}, nil
}

func (s *MessageService) getPrivateConversations(ctx context.Context, userID string) ([]dto.PrivateConversation, error) {
	type PrivateChat struct {
		PartnerID   string
		LastContent string
		LastTime    time.Time
	}

	var privateChats []PrivateChat

	err := s.db.WithContext(ctx).Raw(`
		SELECT partner_id, content as last_content, created_at as last_time
		FROM (
			SELECT 
				CASE 
					WHEN from_user_id = ? THEN target_id 
					ELSE from_user_id 
				END as partner_id,
				content,
				created_at,
				ROW_NUMBER() OVER (PARTITION BY CASE 
					WHEN from_user_id = ? THEN target_id 
					ELSE from_user_id 
				END ORDER BY created_at DESC) as rn
			FROM messages
			WHERE type = ? AND (from_user_id = ? OR target_id = ?)
		) t
		WHERE rn = 1
		ORDER BY last_time DESC
	`, userID, userID, string(model.MessageTypePrivate), userID, userID).Scan(&privateChats).Error

	if err != nil {
		return nil, err
	}

	if len(privateChats) == 0 {
		return []dto.PrivateConversation{}, nil
	}

	partnerIDs := make([]string, 0, len(privateChats))
	for _, chat := range privateChats {
		partnerIDs = append(partnerIDs, chat.PartnerID)
	}

	userQ := dao.Use(s.db).User
	userDo := userQ.WithContext(ctx)

	users, err := userDo.Where(userQ.ID.In(partnerIDs...)).Find()
	if err != nil {
		return nil, err
	}

	userMap := make(map[string]*model.User)
	for _, user := range users {
		userMap[user.ID] = user
	}

	chatMap := make(map[string]PrivateChat)
	for _, chat := range privateChats {
		chatMap[chat.PartnerID] = chat
	}

	conversations := make([]dto.PrivateConversation, 0, len(partnerIDs))
	for _, partnerID := range partnerIDs {
		chat := chatMap[partnerID]
		user := userMap[partnerID]

		conversations = append(conversations, dto.PrivateConversation{
			UserID:      user.ID,
			Username:    user.Username,
			Avatar:      s.generateAvatarUrl(user.ID, user.Username),
			LastContent: chat.LastContent,
			LastTime:    chat.LastTime,
		})
	}

	return conversations, nil
}

func (s *MessageService) getGroupConversations(ctx context.Context, userID string) ([]dto.GroupConversation, error) {
	gmQ := dao.Use(s.db).GroupMember
	gmDo := gmQ.WithContext(ctx)

	groupMembers, err := gmDo.Where(gmQ.UserID.Eq(userID)).Find()
	if err != nil {
		return nil, err
	}

	if len(groupMembers) == 0 {
		return []dto.GroupConversation{}, nil
	}

	groupIDs := make([]string, 0, len(groupMembers))
	for _, gm := range groupMembers {
		groupIDs = append(groupIDs, gm.GroupID)
	}

	type LastGroupMessage struct {
		GroupID      string
		LastContent  string
		LastTime     time.Time
		LastSenderID string
	}

	var lastGroupMessages []LastGroupMessage

	err = s.db.WithContext(ctx).Raw(`
		SELECT target_id as group_id, content as last_content, created_at as last_time, from_user_id as last_sender_id
		FROM (
			SELECT 
				target_id,
				content,
				created_at,
				from_user_id,
				ROW_NUMBER() OVER (PARTITION BY target_id ORDER BY created_at DESC) as rn
			FROM messages
			WHERE type = ? AND target_id IN (?)
		) t
		WHERE rn = 1
		ORDER BY last_time DESC
	`, string(model.MessageTypeGroup), groupIDs).Scan(&lastGroupMessages).Error

	if err != nil {
		return nil, err
	}

	groupQ := dao.Use(s.db).Group
	groupDo := groupQ.WithContext(ctx)

	groups, err := groupDo.Where(groupQ.ID.In(groupIDs...)).Find()
	if err != nil {
		return nil, err
	}

	groupMap := make(map[string]*model.Group)
	for _, group := range groups {
		groupMap[group.ID] = group
	}

	messageMap := make(map[string]LastGroupMessage)
	for _, msg := range lastGroupMessages {
		messageMap[msg.GroupID] = msg
	}

	senderIDs := make([]string, 0, len(lastGroupMessages))
	for _, msg := range lastGroupMessages {
		senderIDs = append(senderIDs, msg.LastSenderID)
	}

	var senders []*model.User
	var senderMap map[string]*model.User

	if len(senderIDs) > 0 {
		userQ := dao.Use(s.db).User
		userDo := userQ.WithContext(ctx)
		senders, err = userDo.Where(userQ.ID.In(senderIDs...)).Find()
		if err != nil {
			return nil, err
		}
		senderMap = make(map[string]*model.User)
		for _, sender := range senders {
			senderMap[sender.ID] = sender
		}
	} else {
		senderMap = make(map[string]*model.User)
	}

	conversations := make([]dto.GroupConversation, 0, len(groupIDs))
	for _, groupID := range groupIDs {
		group := groupMap[groupID]
		if msg, ok := messageMap[groupID]; ok {
			if sender, senderOk := senderMap[msg.LastSenderID]; senderOk {
				conversations = append(conversations, dto.GroupConversation{
					GroupID:        group.ID,
					GroupName:      group.Name,
					LastContent:    msg.LastContent,
					LastTime:       msg.LastTime,
					LastSenderID:   msg.LastSenderID,
					LastSenderName: sender.Username,
				})
			} else {
				conversations = append(conversations, dto.GroupConversation{
					GroupID:        group.ID,
					GroupName:      group.Name,
					LastContent:    msg.LastContent,
					LastTime:       msg.LastTime,
					LastSenderID:   msg.LastSenderID,
					LastSenderName: "",
				})
			}
		} else {
			conversations = append(conversations, dto.GroupConversation{
				GroupID:        group.ID,
				GroupName:      group.Name,
				LastContent:    "",
				LastTime:       time.Time{},
				LastSenderID:   "",
				LastSenderName: "",
			})
		}
	}

	return conversations, nil
}
