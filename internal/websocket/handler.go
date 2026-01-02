package websocket

import (
	"chat_backend/internal/database"
	"chat_backend/internal/service"
	"chat_backend/pkg/logger"
	"context"
	"encoding/json"
	"net/http"
	"time"

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
	// 获取请求上下文
	ctx := c.Request().Context()

	// 从上下文中获取用户ID（由认证中间件设置）
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

	// 启动写入泵（goroutine）
	go userConn.WritePump(ctx)

	// 启动读取泵（阻塞调用）
	userConn.ReadPump(ctx, func(msg WSMessage) {
		handleMessage(userConn, msg)
	})

	logger.GetLogger().Infow("WebSocket connection closed", "user_id", userIDStr)

	return nil
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
	// 心跳消息处理
	case MessageTypeHeartbeat:
		logger.GetLogger().Debugw("Heartbeat received", "user_id", conn.UserID)

	// 业务消息处理（文本、图片、文件）
	case MessageTypeText, MessageTypeImage, MessageTypeFile:
		// 生成消息ID（如果未提供）
		if msg.MessageID == "" {
			msg.MessageID = uuid.New().String()
		}

		cm := GetConnectionManager()

		// 根据聊天类型分发消息
		switch msg.ChatType {
		// 私聊消息处理
		case ChatTypePrivate:
			// 检查是否有接收者
			if msg.To == "" {
				logger.GetLogger().Warnw("Private message missing target", "from", msg.From)
				return
			}

			// 检查接收者是否在线
			if cm.IsOnline(msg.To) {
				// 在线则直接发送
				cm.SendToUser(msg.To, msg)
				logger.GetLogger().Infow("Message sent to online user", "to", msg.To, "message_id", msg.MessageID)
			} else {
				// 离线则记录日志（实际应该存储到数据库）
				logger.GetLogger().Infow("User offline, message should be stored", "to", msg.To, "message_id", msg.MessageID)
			}

			// 发送确认消息给发送者
			ackMsg := WSMessage{
				Type:      MessageTypeAck,
				MessageID: msg.MessageID,
				From:      "system",
				To:        conn.UserID,
				Content:   "message_sent",
				Timestamp: time.Now().UnixMilli(),
			}
			conn.Send(ackMsg)

		// 群聊消息处理
		case ChatTypeGroup:
			// 检查是否有目标群组
			if msg.To == "" {
				logger.GetLogger().Warnw("Group message missing target", "from", msg.From)
				return
			}

			// 获取群组成员列表
			groupService := service.NewGroupService(database.GetDB())
			memberIDs, err := groupService.GetGroupMemberIDs(context.Background(), msg.To)
			if err != nil {
				logger.GetLogger().Errorw("Failed to get group members", "group_id", msg.To, "error", err)
				return
			}

			// 广播消息给群组内所有在线用户（排除发送者自己）
			cm.BroadcastToGroup(msg, memberIDs, conn.UserID)
			logger.GetLogger().Infow("Group message broadcasted", "group_id", msg.To, "message_id", msg.MessageID, "member_count", len(memberIDs))

			// 发送确认消息给发送者
			ackMsg := WSMessage{
				Type:      MessageTypeAck,
				MessageID: msg.MessageID,
				From:      "system",
				To:        conn.UserID,
				Content:   "message_sent",
				Timestamp: time.Now().UnixMilli(),
			}
			conn.Send(ackMsg)

		// 未知聊天类型
		default:
			logger.GetLogger().Warnw("Unknown chat type", "chat_type", msg.ChatType, "from", msg.From)
		}

	// 未知消息类型
	default:
		logger.GetLogger().Warnw("Unknown message type", "type", msg.Type, "from", msg.From)
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
