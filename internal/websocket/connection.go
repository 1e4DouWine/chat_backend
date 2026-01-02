package websocket

import (
	"chat_backend/pkg/logger"
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/coder/websocket"
)

// UserConnection 表示单个用户的WebSocket连接
// 封装了用户连接的所有相关信息和操作
type UserConnection struct {
	// UserID 用户唯一标识符
	UserID string
	// Conn WebSocket连接对象
	Conn *websocket.Conn
	// ConnectedAt 连接建立时间
	ConnectedAt time.Time
	// SendChan 发送消息通道，用于向客户端发送数据
	SendChan chan WSMessage
	// CloseChan 关闭信号通道，用于通知连接关闭
	CloseChan chan struct{}
	// mu 互斥锁，用于保护并发访问
	mu sync.RWMutex
	// closed 连接关闭状态标记
	closed bool
}

// NewUserConnection 创建新的用户连接实例
// 参数:
//   - userID: 用户唯一标识符
//   - conn: WebSocket连接对象
//
// 返回:
//   - *UserConnection: 初始化的用户连接对象
func NewUserConnection(userID string, conn *websocket.Conn) *UserConnection {
	return &UserConnection{
		UserID:      userID,
		Conn:        conn,
		ConnectedAt: time.Now(),
		SendChan:    make(chan WSMessage, 256), // 带缓冲的发送通道，容量256
		CloseChan:   make(chan struct{}),       // 无缓冲关闭通道
		closed:      false,                     // 初始状态为未关闭
	}
}

// ReadPump 读取消息泵
// 持续从WebSocket连接读取消息，并调用消息处理函数
// 参数:
//   - ctx: 上下文，用于控制goroutine生命周期
//   - messageHandler: 消息处理回调函数（可选）
func (uc *UserConnection) ReadPump(ctx context.Context, messageHandler func(WSMessage)) {
	// 延迟执行恢复机制，捕获可能的panic
	defer func() {
		if r := recover(); r != nil {
			logger.GetLogger().Errorw("ReadPanic recover", "user_id", uc.UserID, "error", r)
		}
		// 确保连接关闭
		uc.Close()
	}()

	for {
		select {
		// 上下文取消，退出读取循环
		case <-ctx.Done():
			logger.GetLogger().Infow("ReadPump context done", "user_id", uc.UserID)
			return
		// 关闭信号，退出读取循环
		case <-uc.CloseChan:
			logger.GetLogger().Infow("ReadPump close signal", "user_id", uc.UserID)
			return
		// 默认分支：尝试读取消息
		default:
			// 从WebSocket连接读取消息
			_, data, err := uc.Conn.Read(ctx)
			if err != nil {
				status := websocket.CloseStatus(err)
				if status == websocket.StatusNormalClosure || status == websocket.StatusNoStatusRcvd || status == websocket.StatusGoingAway {
					logger.GetLogger().Infow("Connection closed normally", "user_id", uc.UserID, "status", status)
				} else {
					logger.GetLogger().Errorw("Read error", "user_id", uc.UserID, "error", err)
				}
				return
			}

			// 反序列化JSON消息
			var msg WSMessage
			if err := json.Unmarshal(data, &msg); err != nil {
				logger.GetLogger().Errorw("Unmarshal message error", "user_id", uc.UserID, "error", err)
				continue // 继续读取下一条消息
			}

			// 设置消息发送者和时间戳
			msg.From = uc.UserID
			msg.Timestamp = time.Now().UnixMilli()

			// 调用消息处理回调
			if messageHandler != nil {
				messageHandler(msg)
			}
		}
	}
}

// WritePump 写入消息泵
// 持续向WebSocket连接写入消息，包括业务消息和心跳
// 参数:
//   - ctx: 上下文，用于控制goroutine生命周期
func (uc *UserConnection) WritePump(ctx context.Context) {
	// 延迟执行恢复机制，捕获可能的panic
	defer func() {
		if r := recover(); r != nil {
			logger.GetLogger().Errorw("WritePanic recover", "user_id", uc.UserID, "error", r)
		}
		// 确保连接关闭
		uc.Close()
	}()

	// 创建心跳定时器，30秒间隔
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		// 上下文取消，退出写入循环
		case <-ctx.Done():
			logger.GetLogger().Infow("WritePump context done", "user_id", uc.UserID)
			return
		// 关闭信号，退出写入循环
		case <-uc.CloseChan:
			logger.GetLogger().Infow("WritePump close signal", "user_id", uc.UserID)
			return
		// 业务消息发送
		case msg, ok := <-uc.SendChan:
			if !ok {
				logger.GetLogger().Infow("SendChan closed", "user_id", uc.UserID)
				return
			}

			// 序列化消息为JSON
			data, err := json.Marshal(msg)
			if err != nil {
				logger.GetLogger().Errorw("Marshal message error", "user_id", uc.UserID, "error", err)
				continue
			}

			// 创建带超时的写入上下文
			writeCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			// 向WebSocket连接写入消息
			if err = uc.Conn.Write(writeCtx, websocket.MessageText, data); err != nil {
				cancel()
				logger.GetLogger().Errorw("Write message error", "user_id", uc.UserID, "error", err)
				return
			}
			cancel()

		// 心跳消息发送
		case <-ticker.C:
			// 构建心跳消息
			heartbeatMsg := WSMessage{
				Type:      MessageTypeHeartbeat,
				Timestamp: time.Now().UnixMilli(),
			}
			data, err := json.Marshal(heartbeatMsg)
			if err != nil {
				logger.GetLogger().Errorw("Marshal heartbeat error", "user_id", uc.UserID, "error", err)
				continue
			}

			// 创建带超时的写入上下文
			writeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			// 发送心跳消息
			if err := uc.Conn.Write(writeCtx, websocket.MessageText, data); err != nil {
				cancel()
				logger.GetLogger().Errorw("Write heartbeat error", "user_id", uc.UserID, "error", err)
				return
			}
			cancel()
		}
	}
}

// Send 发送消息到用户的发送通道
// 参数:
//   - msg: 要发送的消息
//
// 返回:
//   - bool: 是否成功发送到通道
func (uc *UserConnection) Send(msg WSMessage) bool {
	uc.mu.RLock()
	defer uc.mu.RUnlock()

	// 检查连接是否已关闭
	if uc.closed {
		return false
	}

	// 非阻塞发送到通道
	select {
	case uc.SendChan <- msg:
		return true
	default:
		// 通道已满，消息被丢弃
		logger.GetLogger().Warnw("SendChan full, message dropped", "user_id", uc.UserID)
		return false
	}
}

// Close 关闭用户连接
// 清理资源，关闭通道和WebSocket连接
func (uc *UserConnection) Close() {
	uc.mu.Lock()
	defer uc.mu.Unlock()

	// 防止重复关闭
	if uc.closed {
		return
	}

	// 标记为已关闭
	uc.closed = true
	// 关闭信号通道
	close(uc.CloseChan)
	// 关闭发送通道
	close(uc.SendChan)

	// 关闭WebSocket连接
	if uc.Conn != nil {
		err := uc.Conn.Close(websocket.StatusNormalClosure, "")
		if err != nil {
			return
		}
	}

	logger.GetLogger().Infow("UserConnection closed", "user_id", uc.UserID)
}

// IsClosed 检查连接是否已关闭
// 返回:
//   - bool: 连接关闭状态
func (uc *UserConnection) IsClosed() bool {
	uc.mu.RLock()
	defer uc.mu.RUnlock()
	return uc.closed
}
