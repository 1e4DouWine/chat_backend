package websocket

import (
	"chat_backend/internal/database"
	"chat_backend/internal/model"
	"chat_backend/internal/service"
	"chat_backend/pkg/logger"
	"context"
	"encoding/json"
	"net/http"
	"time"
	"unicode/utf8"

	"github.com/coder/websocket"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// upgrader WebSocket升级配置选项
// 定义了WebSocket连接升级时的配置参数
var upgrader = &websocket.AcceptOptions{
	// OriginPatterns 允许所有来源的连接
	// 在生产环境中应该限制为特定的域名
	OriginPatterns: []string{"*"},
}

// HandleWebSocket WebSocket连接处理主函数
// 处理WebSocket连接的升级、消息读写和连接管理
// 参数:
//   - c: Echo框架的上下文对象
//
// 返回:
//   - error: 错误信息
func HandleWebSocket(c echo.Context) error {
	ctx := c.Request().Context()
	userID := c.Get("user_id")
	if userID == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "未授权",
		})
	}
	// 类型断言，将interface{}转换为string
	userIDStr, ok := userID.(string)
	if !ok {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "用户ID格式错误",
		})
	}

	// 将HTTP连接升级为WebSocket连接
	conn, err := websocket.Accept(c.Response(), c.Request(), upgrader)
	if err != nil {
		logger.GetLogger().Errorw("WebSocket upgrade failed", "user_id", userIDStr, "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "WebSocket升级失败",
		})
	}

	// 创建用户连接对象
	userConn := NewUserConnection(userIDStr, conn)

	// 获取连接管理器并添加连接
	cm := GetConnectionManager()
	cm.AddConnection(userIDStr, userConn)

	// 延迟执行：连接关闭时从管理器移除
	defer cm.RemoveConnection(userIDStr)

	// 发送连接成功消息
	connectedMsg := WSMessage{
		Type:      MessageTypeConnected,
		Timestamp: time.Now().UnixMilli(),
	}
	data, err := json.Marshal(connectedMsg)
	if err == nil {
		err = conn.Write(ctx, websocket.MessageText, data)
		if err != nil {
			return err
		}
	}

	// 处理未送达消息
	messageService := service.NewMessageService(database.GetDB())
	undeliveredMessages, err := messageService.GetUndeliveredMessages(ctx, userIDStr)
	if err != nil {
		logger.GetLogger().Errorw("Failed to get undelivered messages", "user_id", userIDStr, "error", err)
	} else if len(undeliveredMessages) > 0 {
		logger.GetLogger().Infow("Sending undelivered messages", "user_id", userIDStr, "count", len(undeliveredMessages))

		messageIDs := make([]string, 0, len(undeliveredMessages))
		for _, msg := range undeliveredMessages {
			var chatType ChatType
			if msg.Type == string(model.MessageTypePrivate) {
				chatType = ChatTypePrivate
			} else {
				chatType = ChatTypeGroup
			}

			wsMsg := WSMessage{
				Type:      MessageTypeText,
				ChatType:  chatType,
				From:      msg.FromUserID,
				To:        msg.TargetID,
				Content:   msg.Content,
				MessageID: msg.MessageID,
				Timestamp: msg.CreatedAt.UnixMilli(),
			}

			userConn.Send(wsMsg)
			messageIDs = append(messageIDs, msg.MessageID)
		}

		// 标记消息为已送达
		if err = messageService.MarkMessagesAsDelivered(ctx, userIDStr, messageIDs); err != nil {
			logger.GetLogger().Errorw("Failed to mark messages as delivered", "user_id", userIDStr, "error", err)
		}
	}

	// 启动写入泵（goroutine）
	go userConn.WritePump(ctx)

	// 启动读取泵（阻塞调用）
	userConn.ReadPump(ctx, func(msg WSMessage) {
		handleMessage(userConn, msg)
	})

	logger.GetLogger().Infow("WebSocket connection closed", "user_id", userIDStr)

	return nil
}

const (
	maxMessageLength = 10000
)

// validateMessageContent 验证消息内容
func validateMessageContent(msgType MessageType, content string) bool {
	if content == "" {
		return false
	}

	length := utf8.RuneCountInString(content)
	if length > maxMessageLength {
		return false
	}

	return true
}

// handleMessage 处理接收到的WebSocket消息
// 根据消息类型和聊天类型进行分发处理
// 参数:
//   - conn: 用户连接对象
//   - msg: 接收到的消息
func handleMessage(conn *UserConnection, msg WSMessage) {
	logger.GetLogger().Infow("Received message", "type", msg.Type, "chat_type", msg.ChatType, "from", msg.From, "to", msg.To)

	// 根据消息类型进行处理
	switch msg.Type {
	case MessageTypeHeartbeat:
		logger.GetLogger().Debugw("Heartbeat received", "user_id", conn.UserID)
		heartbeatMsg := WSMessage{
			Type:      MessageTypeHeartbeat,
			From:      "system",
			To:        conn.UserID,
			Timestamp: time.Now().UnixMilli(),
		}
		conn.Send(heartbeatMsg)

	case MessageTypeText, MessageTypeImage, MessageTypeFile:
		// 验证消息内容
		if !validateMessageContent(msg.Type, msg.Content) {
			logger.GetLogger().Warnw("Invalid message content", "type", msg.Type, "from", msg.From, "content_length", utf8.RuneCountInString(msg.Content))
			SendSystemMessage(conn.UserID, "消息内容无效")
			return
		}

		// 统一使用后端生成的消息ID
		messageID := uuid.New().String()
		// 创建新的消息对象用于广播，避免修改原始消息对象
		broadcastMsg := WSMessage{
			Type:      msg.Type,
			ChatType:  msg.ChatType,
			From:      msg.From,
			To:        msg.To,
			Content:   msg.Content,
			MessageID: messageID,
			Timestamp: time.Now().UnixMilli(),
		}

		messageService := service.NewMessageService(database.GetDB())
		cm := GetConnectionManager()

		// 根据聊天类型分发消息
		switch msg.ChatType {
		case ChatTypePrivate:
			// 检查是否有接收者
			if msg.To == "" {
				logger.GetLogger().Warnw("Private message missing target", "from", msg.From)
				SendSystemMessage(conn.UserID, "缺少目标用户")
				return
			}

			// 验证好友关系
			userService := service.NewUserService(database.GetDB())
			isFriend, err := userService.IsFriend(context.Background(), msg.From, msg.To)
			if err != nil {
				logger.GetLogger().Errorw("Failed to check friend relationship", "from", msg.From, "to", msg.To, "error", err)
				SendSystemMessage(conn.UserID, "验证好友关系失败")
				return
			}
			if !isFriend {
				logger.GetLogger().Warnw("Not friends", "from", msg.From, "to", msg.To)
				SendSystemMessage(conn.UserID, "只能向好友发送消息")
				return
			}

			// 检查接收者是否在线
			isTargetOnline := cm.IsOnline(msg.To)

			// 存储消息到数据库并处理离线消息
			_, err = messageService.SendPrivateMessage(context.Background(), msg.From, msg.To, msg.Content, messageID, isTargetOnline)
			if err != nil {
				logger.GetLogger().Errorw("Failed to store message", "message_id", messageID, "error", err)
				SendSystemMessage(conn.UserID, "消息发送失败")
				return
			}

			// 在线则直接发送
			if isTargetOnline {
				cm.SendToUser(broadcastMsg.To, broadcastMsg)
				logger.GetLogger().Infow("Message sent to online user", "to", broadcastMsg.To, "message_id", messageID)
			} else {
				logger.GetLogger().Infow("User offline, message stored", "to", broadcastMsg.To, "message_id", messageID)
			}

			// 发送确认消息给发送者
			ackMsg := WSMessage{
				Type:      MessageTypeAck,
				MessageID: messageID,
				From:      "system",
				To:        conn.UserID,
				Content:   "message_sent",
				Timestamp: time.Now().UnixMilli(),
			}
			conn.Send(ackMsg)

		case ChatTypeGroup:
			// 检查是否有目标群组
			if msg.To == "" {
				logger.GetLogger().Warnw("Group message missing target", "from", msg.From)
				SendSystemMessage(conn.UserID, "缺少目标群组")
				return
			}

			// 验证发送者是否是群组成员
			groupService := service.NewGroupService(database.GetDB())
			isMember, err := groupService.IsGroupMember(context.Background(), msg.To, msg.From)
			if err != nil {
				logger.GetLogger().Errorw("Failed to check group membership", "from", msg.From, "group_id", msg.To, "error", err)
				SendSystemMessage(conn.UserID, "验证群组成员失败")
				return
			}
			if !isMember {
				logger.GetLogger().Warnw("Not group member", "from", msg.From, "group_id", msg.To)
				SendSystemMessage(conn.UserID, "只有群组成员才能发送消息")
				return
			}

			// 获取群组成员列表
			memberIDs, err := groupService.GetGroupMemberIDs(context.Background(), msg.To)
			if err != nil {
				logger.GetLogger().Errorw("Failed to get group members", "group_id", msg.To, "error", err)
				SendSystemMessage(conn.UserID, "获取群组成员失败")
				return
			}

			// 构建接收者列表（排除发送者）和在线用户映射
			recipientIDs := make([]string, 0, len(memberIDs))
			onlineUserIDs := make(map[string]bool)
			for _, memberID := range memberIDs {
				if memberID != conn.UserID {
					recipientIDs = append(recipientIDs, memberID)
				}
				onlineUserIDs[memberID] = cm.IsOnline(memberID)
			}

			// 存储群组消息到数据库并处理离线消息
			_, err = messageService.SendGroupMessage(context.Background(), msg.From, msg.To, msg.Content, messageID, recipientIDs, onlineUserIDs)
			if err != nil {
				logger.GetLogger().Errorw("Failed to store group message", "message_id", messageID, "error", err)
				SendSystemMessage(conn.UserID, "消息发送失败")
				return
			}

			// 广播消息给群组内所有在线用户（排除发送者自己）
			cm.BroadcastToGroup(broadcastMsg, recipientIDs)
			logger.GetLogger().Infow("Group message broadcasted", "group_id", broadcastMsg.To, "message_id", messageID, "recipient_count", len(recipientIDs))

			// 发送确认消息给发送者
			ackMsg := WSMessage{
				Type:      MessageTypeAck,
				MessageID: messageID,
				From:      "system",
				To:        conn.UserID,
				Content:   "message_sent",
				Timestamp: time.Now().UnixMilli(),
			}
			conn.Send(ackMsg)

		default:
			logger.GetLogger().Warnw("Unknown chat type", "chat_type", msg.ChatType, "from", msg.From)
			SendSystemMessage(conn.UserID, "未知的聊天类型")
		}

	// 未知消息类型
	default:
		logger.GetLogger().Warnw("Unknown message type", "type", msg.Type, "from", msg.From)
		SendSystemMessage(conn.UserID, "未知的消息类型")
	}
}

// GetOnlineUsers 获取所有在线用户信息
// HTTP API接口，用于查询当前在线用户
// 参数:
//   - c: Echo框架的上下文对象
//
// 返回:
//   - error: 错误信息
func GetOnlineUsers(c echo.Context) error {
	cm := GetConnectionManager()
	onlineCount := cm.GetOnlineUserCount()
	onlineUserIDs := cm.GetOnlineUserIDs()

	return c.JSON(http.StatusOK, map[string]interface{}{
		"online_count":    onlineCount,
		"online_user_ids": onlineUserIDs,
	})
}

// IsUserOnline 检查指定用户是否在线
// HTTP API接口，用于查询单个用户的在线状态
// 参数:
//   - c: Echo框架的上下文对象
//
// 返回:
//   - error: 错误信息
func IsUserOnline(c echo.Context) error {
	userID := c.Param("id")
	if userID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "用户ID不能为空",
		})
	}

	cm := GetConnectionManager()
	isOnline := cm.IsOnline(userID)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"user_id":   userID,
		"is_online": isOnline,
	})
}

// SendSystemMessage 向指定用户发送系统消息
// 参数:
//   - userID: 目标用户ID
//   - content: 消息内容
//
// 返回:
//   - bool: 是否成功发送
func SendSystemMessage(userID string, content string) bool {
	cm := GetConnectionManager()
	// 检查用户是否在线
	if !cm.IsOnline(userID) {
		return false
	}

	// 构建系统消息
	msg := WSMessage{
		Type:      MessageTypeSystem,
		From:      "system",
		To:        userID,
		Content:   content,
		Timestamp: time.Now().UnixMilli(),
	}

	// 发送消息
	return cm.SendToUser(userID, msg)
}

// BroadcastSystemMessage 广播系统消息给所有在线用户
// 参数:
//   - content: 消息内容
//   - excludeUserIDs: 需要排除的用户ID列表
func BroadcastSystemMessage(content string, excludeUserIDs ...string) {
	// 构建系统消息
	msg := WSMessage{
		Type:      MessageTypeSystem,
		From:      "system",
		Content:   content,
		Timestamp: time.Now().UnixMilli(),
	}

	// 获取连接管理器并广播消息
	cm := GetConnectionManager()
	cm.Broadcast(msg, excludeUserIDs...)
}
