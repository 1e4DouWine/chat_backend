package service

import (
	"chat_backend/internal/dao"
	"chat_backend/internal/dto"
	"chat_backend/internal/model"
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"time"

	"gorm.io/gorm"
)

const (
	FriendStatusAccepted = "accepted"
	FriendStatusPending  = "pending"
	FriendStatusRejected = "rejected"

	ActionParamAccept = "accept"
	ActionParamReject = "reject"
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
func (s *UserService) GetMe(ctx context.Context, userID string, username string) (*dto.UserInfoResponse, error) {
	avatarUrl, err := s.GetUserAvatarUrl(userID, username)
	if err != nil {
		return nil, err
	}
	return &dto.UserInfoResponse{
		UserID:   userID,
		Username: username,
		Avatar:   avatarUrl,
	}, nil
}

//// GetUserInfo 获取用户资料
//func (s *UserService) GetUserInfo(c context.Context, userID string) (*dto.UserInfoResponse, error) {
//	avatarUrl, err := s.GetUserAvatarUrl(c, userID, username)
//	if err != nil {
//		return nil, err
//	}
//	return &dto.UserInfoResponse{
//		UserID:   userID,
//		Username: username,
//		Avatar:   avatarUrl,
//	}, nil
//}

// GetUsernameByUserID 通过 userID 查询 username
func (s *UserService) GetUsernameByUserID(ctx context.Context, userID string) (string, error) {
	q := dao.Use(s.db).User
	do := q.WithContext(ctx)

	user, err := do.Where(q.ID.Eq(userID)).First()
	if err != nil {
		return "", err
	}

	return user.Username, nil
}

// GetUserIDByUsername 通过 username 查询 userID
func (s *UserService) GetUserIDByUsername(ctx context.Context, username string) (string, error) {
	q := dao.Use(s.db).User
	do := q.WithContext(ctx)

	user, err := do.Where(q.Username.Eq(username)).First()
	if err != nil {
		return "", err
	}

	return user.ID, nil
}

// GetUserAvatarUrl 获取用户头像
func (s *UserService) GetUserAvatarUrl(userID string, username string) (string, error) {
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

// AddFriend 好友申请
func (s *UserService) AddFriend(ctx context.Context, userID string, friendID string) (*dto.AddFriendResponse, error) {
	q := dao.Use(s.db).Friend
	do := q.WithContext(ctx)

	// 检查是否已存在好友关系或申请记录
	existingRecord, err := do.Where(
		q.User1ID.Eq(userID),
		q.User2ID.Eq(friendID),
	).Or(
		q.User1ID.Eq(friendID),
		q.User2ID.Eq(userID),
	).First()

	if err == nil {
		// 记录已存在，检查状态
		switch existingRecord.Status {
		case FriendStatusPending:
			if existingRecord.User1ID == userID {
				// 我已经发送过申请，还在等待中
				return nil, fmt.Errorf("您已经发送过好友申请，正在等待对方处理")
			} else {
				// 对方已经给我发送了申请，我应该去处理，而不是重复发送
				return nil, fmt.Errorf("对方已经向您发送了好友申请，请在待处理申请中查看")
			}
		case FriendStatusAccepted:
			return nil, fmt.Errorf("你们已经是好友了")
		case FriendStatusRejected:
			// 如果之前被拒绝，允许重新发送，但需要更新记录而不是创建新的
			existingRecord.Status = FriendStatusPending
			existingRecord.CreatedAt = time.Now() // 重置时间
			if err := do.Save(existingRecord); err != nil {
				return nil, err
			}
			return &dto.AddFriendResponse{
				FriendID: friendID,
				Status:   FriendStatusPending,
			}, nil
		}
	}

	// 不存在记录，创建新的申请
	record := model.Friend{
		User1ID: userID,
		User2ID: friendID,
		Status:  FriendStatusPending,
	}

	if err := do.Create(&record); err != nil {
		return nil, err
	}

	return &dto.AddFriendResponse{
		FriendID: friendID,
		Status:   FriendStatusPending,
	}, nil
}

// GetFriendList 获取好友列表
func (s *UserService) GetFriendList(ctx context.Context, userID string, status string) ([]*dto.UserInfoResponse, error) {
	q := dao.Use(s.db).Friend
	do := q.WithContext(ctx)

	var friends []model.Friend

	// 根据状态处理不同的查询逻辑
	if status == FriendStatusPending {
		// 查询待处理的好友申请：别人发给我的申请（User2ID = userID）
		err := do.Where(
			q.User2ID.Eq(userID),
			q.Status.Eq(status),
		).Scan(&friends)
		if err != nil {
			return nil, err
		}
	} else if status == FriendStatusAccepted {
		// 查询已接受的好友：需要查询两个方向
		// 1. 我发出的被接受的：User1ID = userID, status = accepted
		// 2. 别人发给我被我接受的：User2ID = userID, status = accepted
		err := do.Where(
			q.User1ID.Eq(userID),
			q.Status.Eq(status),
		).Or(
			q.User2ID.Eq(userID),
			q.Status.Eq(status),
		).Scan(&friends)
		if err != nil {
			return nil, err
		}
	} else {
		// 其他状态的查询，理论上不会到这
		err := do.Where(
			q.User1ID.Eq(userID),
			q.Status.Eq(status),
		).Or(
			q.User2ID.Eq(userID),
			q.Status.Eq(status),
		).Scan(&friends)
		if err != nil {
			return nil, err
		}
	}

	if len(friends) == 0 {
		return []*dto.UserInfoResponse{}, nil
	}

	responses := make([]*dto.UserInfoResponse, 0, len(friends))
	for _, f := range friends {
		// 确定对方的 UserID
		var friendUserID string
		if f.User1ID == userID {
			friendUserID = f.User2ID
		} else {
			friendUserID = f.User1ID
		}

		username, err := s.GetUsernameByUserID(ctx, friendUserID)
		if err != nil {
			username = "unknown"
		}
		avatarUrl, err := s.GetUserAvatarUrl(friendUserID, username)
		if err != nil {
			avatarUrl = "unknown"
		}

		responses = append(responses, &dto.UserInfoResponse{
			UserID:   friendUserID,
			Username: username,
			Avatar:   avatarUrl,
		})
	}

	return responses, nil
}

// ProcessFriendRequest 处理好友申请
func (s *UserService) ProcessFriendRequest(ctx context.Context, userID string, friendID string, action string) (*dto.AddFriendResponse, error) {
	// UserA 向 UserB 发送好友申请，此时 friend 表中的 status 应该是 pending
	// UserB 对 UserA 发出的好友申请进行操作，传入的 action 为 accept 表示通过好友申请，reject 表示拒绝
	// 注意：好友申请记录是 (UserA, UserB)，所以当 UserB 处理时，需要查询 User2ID = userID 的记录
	q := dao.Use(s.db).Friend
	do := q.WithContext(ctx)

	// 查找好友申请记录：UserA 向 UserB 发送的申请
	// 所以 User1ID = friendID, User2ID = userID
	_, err := do.Where(
		q.User1ID.Eq(friendID),
		q.User2ID.Eq(userID),
		q.Status.Eq(FriendStatusPending),
	).First()
	if err != nil {
		return nil, fmt.Errorf("未找到待处理的好友申请")
	}

	var newStatus string
	if action == ActionParamAccept {
		newStatus = FriendStatusAccepted
	} else if action == ActionParamReject {
		newStatus = FriendStatusRejected
	}

	// 更新状态
	_, err = do.Where(
		q.User1ID.Eq(friendID),
		q.User2ID.Eq(userID),
	).Update(q.Status, newStatus)
	if err != nil {
		return nil, err
	}

	return &dto.AddFriendResponse{
		FriendID: friendID,
		Status:   newStatus,
	}, nil
}

// DeleteFriend 删除好友
func (s *UserService) DeleteFriend(ctx context.Context, userID string, friendID string) error {
	q := dao.Use(s.db).Friend
	do := q.WithContext(ctx)

	// 查找好友关系记录（双向查找）
	// 情况1: userID 是 User1ID, friendID 是 User2ID
	// 情况2: userID 是 User2ID, friendID 是 User1ID
	record, err := do.Where(
		q.User1ID.Eq(userID),
		q.User2ID.Eq(friendID),
	).Or(
		q.User1ID.Eq(friendID),
		q.User2ID.Eq(userID),
	).First()

	if err != nil {
		return fmt.Errorf("好友关系不存在")
	}

	// 验证权限：只能删除状态为 accepted 的好友关系，或者删除自己发出的 pending 申请
	if record.Status == FriendStatusPending && record.User1ID != userID {
		return fmt.Errorf("不能删除别人发给您的申请，请使用拒绝功能")
	}

	// 执行删除（软删除，GORM 会自动处理 DeletedAt）
	_, err = do.Where(
		q.User1ID.Eq(record.User1ID),
		q.User2ID.Eq(record.User2ID),
	).Delete()

	if err != nil {
		return fmt.Errorf("删除好友失败: %v", err)
	}

	return nil
}
