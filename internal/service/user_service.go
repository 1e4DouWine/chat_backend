package service

import (
	"chat_backend/internal/dao"
	"chat_backend/internal/dto"
	"chat_backend/internal/model"
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"

	"gorm.io/gorm"
)

const (
	friendStatusAccepted = "accepted"
	friendStatusRejected = "rejected"
	friendStatusPending  = "pending"
)

type UserService struct {
	db *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{
		db: db,
	}
}

// GetMe 获取自己的资料
func (s *UserService) GetMe(c context.Context, userID string, username string) (*dto.UserInfoResponse, error) {
	avatarUrl, err := s.GetUserAvatarUrl(c, userID, username)
	if err != nil {
		return nil, err
	}
	return &dto.UserInfoResponse{
		Username: username,
		Avatar:   avatarUrl,
	}, nil
}

// AddFriend 好友申请
func (s *UserService) AddFriend(c context.Context, userID string, friendID string) (*dto.AddFriendResponse, error) {
	q := dao.Use(s.db).Friend
	do := q.WithContext(c)

	record := model.Friend{
		User1ID: userID,
		User2ID: friendID,
		Status:  friendStatusPending,
	}

	if err := do.Create(&record); err != nil {
		return nil, err
	}

	return &dto.AddFriendResponse{
		FriendID: friendID,
		Status:   friendStatusPending,
	}, nil
}

// GetUserIDByUsername 通过 username 查询 userID
func (s *UserService) GetUserIDByUsername(c context.Context, username string) (string, error) {
	q := dao.Use(s.db).User
	do := q.WithContext(c)

	user, err := do.Where(q.Username.Eq(username)).First()
	if err != nil {
		return "", err
	}

	return user.ID, nil
}

// GetUserAvatarUrl 获取用户头像
func (s *UserService) GetUserAvatarUrl(c context.Context, userID string, username string) (string, error) {
	color := colorFromUUID(userID)
	avatar := fmt.Sprintf(
		"https://ui-avatars.com/api/?name=%s&background=%s&rounded=true&size=128",
		username,
		color,
	)
	return avatar, nil
}
func colorFromUUID(uuid string) string {
	hash := md5.Sum([]byte(uuid))
	// 取前 3 个字节作为 RGB
	r := hash[0]
	g := hash[1]
	b := hash[2]

	r = clamp(r, 50, 200)
	g = clamp(g, 50, 200)
	b = clamp(b, 50, 200)

	return hex.EncodeToString([]byte{r, g, b})
}

// 限制颜色亮度
func clamp(v, min, max byte) byte {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
