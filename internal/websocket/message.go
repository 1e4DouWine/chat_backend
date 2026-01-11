package websocket

import "time"

// WebSocket 配置常量
const (
	// SendChanBufferSize 发送通道缓冲区大小
	// 用于缓存待发送的消息，防止消息丢失
	SendChanBufferSize = 256
	// HeartbeatInterval 心跳间隔时间
	// 用于保持 WebSocket 连接活跃
	HeartbeatInterval = 30 * time.Second
	// WriteTimeout 写入超时时间
	// 用于防止写入操作长时间阻塞
	WriteTimeout = 10 * time.Second
	// PingTimeout Ping 操作超时时间
	// 用于检测连接是否仍然活跃
	PingTimeout = 10 * time.Second
)

// MessageType 定义了WebSocket消息的类型
// 用于区分不同类型的消息内容
type MessageType string

const (
	// MessageTypeText 文本消息
	MessageTypeText MessageType = "text"
	// MessageTypeImage 图片消息
	MessageTypeImage MessageType = "image"
	// MessageTypeFile 文件消息
	MessageTypeFile MessageType = "file"
	// MessageTypeSystem 系统消息
	MessageTypeSystem MessageType = "system"
	// MessageTypeAck 确认消息，用于确认消息接收状态
	MessageTypeAck MessageType = "ack"
	// MessageTypeConnected 连接成功消息，用于通知客户端连接已建立
	MessageTypeConnected MessageType = "connected"
)

// ChatType 定义了聊天的类型
// 用于区分私聊和群聊
type ChatType string

const (
	// ChatTypePrivate 私聊类型
	ChatTypePrivate ChatType = "private"
	// ChatTypeGroup 群聊类型
	ChatTypeGroup ChatType = "group"
)

// WSMessage WebSocket消息结构体
// 定义了WebSocket通信中传输的消息格式
type WSMessage struct {
	// Type 消息类型，如文本、图片、文件等
	Type MessageType `json:"type"`
	// ChatType 聊天类型，私聊或群聊
	ChatType ChatType `json:"chatType"`
	// From 发送者用户ID
	From string `json:"from"`
	// FromUsername 发送者用户名
	FromUsername string `json:"fromUsername,omitempty"`
	// FromAvatar 发送者头像URL
	FromAvatar string `json:"fromAvatar,omitempty"`
	// To 接收者用户ID（私聊）或群组ID（群聊）
	To string `json:"to"`
	// Content 消息内容
	Content string `json:"content"`
	// MessageID 消息唯一标识符，用于消息去重和确认
	MessageID string `json:"messageId"`
	// Timestamp 消息发送时间戳（毫秒）
	Timestamp int64 `json:"timestamp"`
}

// AckMessage 确认消息结构体
// 用于向发送方确认消息接收状态
type AckMessage struct {
	// MessageID 被确认的消息ID
	MessageID string `json:"messageId"`
	// Success 是否成功接收
	Success bool `json:"success"`
	// Error 错误信息（可选）
	Error string `json:"error,omitempty"`
}

// SystemMessage 系统消息结构体
// 用于发送系统级别的通知和提示
type SystemMessage struct {
	// Type 系统消息类型
	Type string `json:"type"`
	// Content 系统消息内容，可以是任意类型
	Content interface{} `json:"content"`
}
