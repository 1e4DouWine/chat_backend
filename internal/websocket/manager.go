package websocket

import (
	"chat_backend/pkg/logger"
	"sync"
	"time"
)

// ConnectionManager WebSocket连接管理器
// 负责管理所有用户的WebSocket连接，提供连接的增删查改功能
type ConnectionManager struct {
	// connections 存储所有用户连接的映射表，key为用户ID，value为连接对象
	connections map[string]*UserConnection
	// mu 读写锁，用于保护connections的并发访问
	mu sync.RWMutex
}

// 单例模式相关变量
var (
	// instance 连接管理器单例实例
	instance *ConnectionManager
	// once 用于确保单例只初始化一次
	once sync.Once
)

// GetConnectionManager 获取连接管理器单例实例
// 使用sync.Once确保线程安全的单例初始化
// 返回:
//   - *ConnectionManager: 连接管理器实例
func GetConnectionManager() *ConnectionManager {
	once.Do(func() {
		instance = &ConnectionManager{
			connections: make(map[string]*UserConnection),
		}
	})
	return instance
}

// AddConnection 添加用户连接到管理器
// 如果用户已有连接，会先关闭旧连接
// 参数:
//   - userID: 用户唯一标识符
//   - conn: 用户连接对象
func (cm *ConnectionManager) AddConnection(userID string, conn *UserConnection) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// 检查是否已存在该用户的连接
	if existingConn, exists := cm.connections[userID]; exists {
		logger.GetLogger().Infow("Closing existing connection", "user_id", userID)
		existingConn.Close()
	}

	// 存储新连接
	cm.connections[userID] = conn
	logger.GetLogger().Infow("Connection added", "user_id", userID, "total_connections", len(cm.connections))
}

// RemoveConnection 从管理器移除用户连接
// 参数:
//   - userID: 用户唯一标识符
func (cm *ConnectionManager) RemoveConnection(userID string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if conn, exists := cm.connections[userID]; exists {
		// 关闭连接
		conn.Close()
		// 从映射表中删除
		delete(cm.connections, userID)
		logger.GetLogger().Infow("Connection removed", "user_id", userID, "total_connections", len(cm.connections))
	}
}

// GetConnection 获取指定用户的连接
// 参数:
//   - userID: 用户唯一标识符
// 返回:
//   - *UserConnection: 用户连接对象
//   - bool: 是否存在该连接
func (cm *ConnectionManager) GetConnection(userID string) (*UserConnection, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	conn, exists := cm.connections[userID]
	return conn, exists
}

// IsOnline 检查用户是否在线
// 参数:
//   - userID: 用户唯一标识符
// 返回:
//   - bool: 用户是否在线
func (cm *ConnectionManager) IsOnline(userID string) bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	conn, exists := cm.connections[userID]
	if !exists {
		return false
	}
	// 还需要检查连接是否已关闭
	return !conn.IsClosed()
}

// SendToUser 向指定用户发送消息
// 参数:
//   - userID: 目标用户ID
//   - msg: 要发送的消息
// 返回:
//   - bool: 是否成功发送
func (cm *ConnectionManager) SendToUser(userID string, msg WSMessage) bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	conn, exists := cm.connections[userID]
	if !exists {
		return false
	}

	return conn.Send(msg)
}

// Broadcast 广播消息给所有在线用户
// 参数:
//   - msg: 要广播的消息
//   - excludeUserIDs: 需要排除的用户ID列表
func (cm *ConnectionManager) Broadcast(msg WSMessage, excludeUserIDs ...string) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	// 构建排除用户集合
	exclude := make(map[string]bool)
	for _, uid := range excludeUserIDs {
		exclude[uid] = true
	}

	// 统计成功发送的消息数量
	sentCount := 0
	for userID, conn := range cm.connections {
		// 跳过被排除的用户
		if exclude[userID] {
			continue
		}
		// 尝试发送消息
		if conn.Send(msg) {
			sentCount++
		}
	}

	logger.GetLogger().Debugw("Broadcast completed", "sent_count", sentCount, "total_connections", len(cm.connections))
}

// GetOnlineUserCount 获取在线用户数量
// 返回:
//   - int: 在线用户数量
func (cm *ConnectionManager) GetOnlineUserCount() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	count := 0
	for _, conn := range cm.connections {
		if !conn.IsClosed() {
			count++
		}
	}
	return count
}

// GetOnlineUserIDs 获取所有在线用户的ID列表
// 返回:
//   - []string: 在线用户ID列表
func (cm *ConnectionManager) GetOnlineUserIDs() []string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	userIDs := make([]string, 0, len(cm.connections))
	for userID, conn := range cm.connections {
		if !conn.IsClosed() {
			userIDs = append(userIDs, userID)
		}
	}
	return userIDs
}

// CloseAll 关闭所有连接
// 用于系统关闭时的清理操作
func (cm *ConnectionManager) CloseAll() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	logger.GetLogger().Infow("Closing all connections", "count", len(cm.connections))

	for userID, conn := range cm.connections {
		conn.Close()
		delete(cm.connections, userID)
		logger.GetLogger().Infow("Connection closed", "user_id", userID)
	}

	logger.GetLogger().Infow("All connections closed")
}

// CleanupStaleConnections 清理过期的连接
// 参数:
//   - timeout: 连接超时时间
func (cm *ConnectionManager) CleanupStaleConnections(timeout time.Duration) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	now := time.Now()
	cleanedCount := 0

	for userID, conn := range cm.connections {
		// 清理已关闭的连接
		if conn.IsClosed() {
			delete(cm.connections, userID)
			cleanedCount++
			continue
		}

		// 清理超时的连接
		if now.Sub(conn.ConnectedAt) > timeout {
			conn.Close()
			delete(cm.connections, userID)
			cleanedCount++
			logger.GetLogger().Infow("Stale connection removed", "user_id", userID, "connected_at", conn.ConnectedAt)
		}
	}

	if cleanedCount > 0 {
		logger.GetLogger().Infow("Cleanup completed", "cleaned_count", cleanedCount, "remaining_connections", len(cm.connections))
	}
}
