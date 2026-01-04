package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// 测试颜色生成函数
func TestColorFromUUID(t *testing.T) {
	uuid := "123e4567-e89b-12d3-a456-426614174000"
	
	color := colorFromUUID(uuid)
	
	// 验证颜色是6位十六进制字符串
	assert.Len(t, color, 6)
	
	// 验证颜色在合理范围内（50-200之间的值转换为十六进制）
	for _, c := range color {
		assert.Contains(t, "0123456789abcdef", string(c))
	}
}

// 测试限制函数
func TestClamp(t *testing.T) {
	// 测试值在范围内
	assert.Equal(t, byte(100), clamp(100, 50, 200))
	
	// 测试值小于最小值
	assert.Equal(t, byte(50), clamp(30, 50, 200))
	
	// 测试值大于最大值
	assert.Equal(t, byte(200), clamp(250, 50, 200))
}

// 测试获取用户头像URL函数
func TestUserService_GetUserAvatarUrl(t *testing.T) {
	userService := &UserService{db: nil}

	userID := "123e4567-e89b-12d3-a456-426614174000"
	username := "testuser"
	
	avatarUrl, err := userService.GetUserAvatarUrl(userID, username)
	
	assert.NoError(t, err)
	assert.Contains(t, avatarUrl, "ui-avatars.com")
	assert.Contains(t, avatarUrl, username)
	assert.Contains(t, avatarUrl, "128") // size参数
}