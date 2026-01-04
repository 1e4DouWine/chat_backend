package database

import (
	"chat_backend/internal/model"
	"fmt"

	"gorm.io/gorm"
)

// Migrate 执行数据库迁移
func Migrate() error {
	db := GetDB()

	// 自动迁移所有模型
	err := db.AutoMigrate(
		&model.User{},
		&model.Message{},
		&model.Friend{},
		&model.FriendRequest{},
		&model.Group{},
		&model.GroupMember{},
		&model.InvitationCode{},
		&model.GroupJoinRequest{},
		&model.MessageReceipt{},
	)

	if err != nil {
		return fmt.Errorf("数据库迁移失败: %w", err)
	}

	// 创建索引
	if err := createIndexes(db); err != nil {
		return fmt.Errorf("创建索引失败: %w", err)
	}

	return nil
}

// createIndexes 创建额外的数据库索引
func createIndexes(db *gorm.DB) error {
	// 为Friend表创建复合索引（基于UserA和UserB字段）
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_friends_user_a_user_b 
		ON friends (user_a, user_b) 
		WHERE deleted_at IS NULL
	`).Error; err != nil {
		return fmt.Errorf("创建friends复合索引失败: %w", err)
	}

	// 为Friend表创建反向复合索引
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_friends_user_b_user_a 
		ON friends (user_b, user_a) 
		WHERE deleted_at IS NULL
	`).Error; err != nil {
		return fmt.Errorf("创建friends反向复合索引失败: %w", err)
	}

	// 为GroupMember表创建复合索引
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_group_members_group_user 
		ON group_members (group_id, user_id) 
		WHERE deleted_at IS NULL
	`).Error; err != nil {
		return fmt.Errorf("创建group_members复合索引失败: %w", err)
	}

	// 为GroupMember表创建用户索引（方便查询用户所属的所有群组）
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_group_members_user_group 
		ON group_members (user_id, group_id) 
		WHERE deleted_at IS NULL
	`).Error; err != nil {
		return fmt.Errorf("创建group_members用户索引失败: %w", err)
	}

	// 为Message表创建复合索引（按类型和创建时间降序）
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_messages_type_created 
		ON messages (type, created_at DESC)
	`).Error; err != nil {
		return fmt.Errorf("创建messages类型时间索引失败: %w", err)
	}

	// 为Message表创建复合索引（用于私聊消息查询）
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_messages_from_target_created 
		ON messages (from_user_id, target_id, created_at DESC)
	`).Error; err != nil {
		return fmt.Errorf("创建messages发送者接收者索引失败: %w", err)
	}

	// 为Message表创建复合索引（用于群组消息查询）
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_messages_target_type_created 
		ON messages (target_id, type, created_at DESC)
	`).Error; err != nil {
		return fmt.Errorf("创建messages接收者类型索引失败: %w", err)
	}

	return nil
}

// DropTables 删除所有表（仅用于开发环境）
func DropTables() error {
	db := GetDB()

	// 删除表（需要按依赖顺序删除）
	tables := []interface{}{
		&model.GroupMember{},
		&model.Message{},
		&model.Friend{},
		&model.Group{},
		&model.User{},
		&model.InvitationCode{},
		&model.FriendRequest{},
		&model.GroupJoinRequest{},
		&model.MessageReceipt{},
	}

	for _, table := range tables {
		if err := db.Migrator().DropTable(table); err != nil {
			return fmt.Errorf("删除表失败: %w", err)
		}
	}

	return nil
}

// ResetDatabase 重置数据库（删除并重新创建所有表）
func ResetDatabase() error {
	if err := DropTables(); err != nil {
		return fmt.Errorf("删除表失败: %w", err)
	}

	if err := Migrate(); err != nil {
		return fmt.Errorf("迁移失败: %w", err)
	}

	return nil
}
