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
	ActionParamAccept = "accept"
	ActionParamReject = "reject"

	FriendRequestExpireDays = 7

	FriendRequestStatusPending  = "pending"
	FriendRequestStatusRejected = "rejected"

	FriendStatusNormal    = "normal"
	FriendStatusBlacklist = "blacklist"
	FriendStatusRemoved   = "removed"
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

// GetUserInfo 获取用户资料
func (s *UserService) GetUserInfo(ctx context.Context, userID string, username string) (*dto.UserInfoResponse, error) {
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

// SendAddFriendRequest 发送添加好友申请
func (s *UserService) SendAddFriendRequest(ctx context.Context, userID string, friendID string) (*dto.AddFriendResponse, error) {
	// 1. 检查 Friend 表是否已存在好友关系
	friendQ := dao.Use(s.db).Friend
	friendDo := friendQ.WithContext(ctx)
	existingFriend, err := friendDo.Where(
		friendQ.UserA.Eq(userID),
		friendQ.UserB.Eq(friendID),
	).Or(
		friendQ.UserA.Eq(friendID),
		friendQ.UserB.Eq(userID),
	).First()
	if err == nil && existingFriend.Status == FriendStatusNormal {
		return nil, fmt.Errorf("你们已经是好友了")
	}

	// 2. 检查 FriendRequest 表是否有未过期的 pending 申请
	requestQ := dao.Use(s.db).FriendRequest
	requestDo := requestQ.WithContext(ctx)
	expireTime := time.Now().AddDate(0, 0, -FriendRequestExpireDays)
	// 检查我是否发送过未过期的申请
	r, err := requestDo.Where(
		requestQ.SenderID.Eq(userID),
		requestQ.ReceiverID.Eq(friendID),
		requestQ.CreatedAt.Gt(expireTime),
	).First()
	if err != nil {
		return nil, err
	}
	switch r.Status {
	case FriendRequestStatusPending:
		return nil, fmt.Errorf("您已经发送过好友申请，正在等待对方处理")
	case FriendRequestStatusRejected:
		return nil, fmt.Errorf("对方已拒绝您的好友申请，请七天后再试")
	}
	// 检查对方是否发送过未过期的申请
	_, err = requestDo.Where(
		requestQ.SenderID.Eq(friendID),
		requestQ.ReceiverID.Eq(userID),
		requestQ.Status.Eq(FriendRequestStatusPending),
		requestQ.CreatedAt.Gt(expireTime),
	).First()
	if err == nil {
		return nil, fmt.Errorf("对方已经向您发送了好友申请，请在待处理申请中查看")
	}

	// 3. 向 FriendRequest 表插入 pending 记录
	record := model.FriendRequest{
		SenderID:   userID,
		ReceiverID: friendID,
		Status:     FriendRequestStatusPending,
	}
	if err := requestDo.Create(&record); err != nil {
		return nil, err
	}

	return &dto.AddFriendResponse{
		FriendID: friendID,
		Status:   FriendRequestStatusPending,
	}, nil
}

// GetFriendList 获取好友列表
func (s *UserService) GetFriendList(ctx context.Context, userID string, status string) ([]*dto.FriendInfoResponse, error) {
	switch status {
	case FriendRequestStatusPending:
		// 查询待处理的好友申请：别人发给我的申请（ReceiverID = userID）
		return s.getPendingFriendRequests(ctx, userID)
	case FriendStatusNormal:
		// 查询已接受的好友：Friend表中Status为normal的记录
		return s.getAcceptedFriends(ctx, userID)
	}

	return nil, fmt.Errorf("无效的status参数")
}

func (s *UserService) getPendingFriendRequests(ctx context.Context, userID string) ([]*dto.FriendInfoResponse, error) {
	requestQ := dao.Use(s.db).FriendRequest
	requestDo := requestQ.WithContext(ctx)

	expireTime := time.Now().AddDate(0, 0, -FriendRequestExpireDays)

	var requests []model.FriendRequest
	err := requestDo.Where(
		requestQ.ReceiverID.Eq(userID),
		requestQ.Status.Eq(FriendRequestStatusPending),
		requestQ.CreatedAt.Gt(expireTime),
	).Scan(&requests)

	if err != nil {
		return nil, err
	}

	if len(requests) == 0 {
		return []*dto.FriendInfoResponse{}, nil
	}

	responses := make([]*dto.FriendInfoResponse, 0, len(requests))
	for _, r := range requests {
		username, err := s.GetUsernameByUserID(ctx, r.SenderID)
		if err != nil {
			username = "unknown"
		}
		avatarUrl, err := s.GetUserAvatarUrl(r.SenderID, username)
		if err != nil {
			avatarUrl = "unknown"
		}

		responses = append(responses, &dto.FriendInfoResponse{
			UserID:   r.SenderID,
			Username: username,
			Avatar:   avatarUrl,
			Status:   r.Status,
			CreateAt: r.CreatedAt.Unix(),
		})
	}

	return responses, nil
}

func (s *UserService) getAcceptedFriends(ctx context.Context, userID string) ([]*dto.FriendInfoResponse, error) {
	friendQ := dao.Use(s.db).Friend
	friendDo := friendQ.WithContext(ctx)

	var friends []model.Friend
	err := friendDo.Where(
		friendQ.UserA.Eq(userID),
		friendQ.Status.Eq(FriendStatusNormal),
	).Or(
		friendQ.UserB.Eq(userID),
		friendQ.Status.Eq(FriendStatusNormal),
	).Scan(&friends)

	if err != nil {
		return nil, err
	}

	if len(friends) == 0 {
		return []*dto.FriendInfoResponse{}, nil
	}

	responses := make([]*dto.FriendInfoResponse, 0, len(friends))
	for _, f := range friends {
		var friendUserID string
		if f.UserA == userID {
			friendUserID = f.UserB
		} else {
			friendUserID = f.UserA
		}

		username, err := s.GetUsernameByUserID(ctx, friendUserID)
		if err != nil {
			username = "unknown"
		}
		avatarUrl, err := s.GetUserAvatarUrl(friendUserID, username)
		if err != nil {
			avatarUrl = "unknown"
		}

		responses = append(responses, &dto.FriendInfoResponse{
			UserID:   friendUserID,
			Username: username,
			Avatar:   avatarUrl,
			Status:   f.Status,
			CreateAt: f.CreatedAt.Unix(),
		})
	}

	return responses, nil
}

// ProcessFriendRequest 处理好友申请
func (s *UserService) ProcessFriendRequest(ctx context.Context, userID string, friendID string, action string) (*dto.AddFriendResponse, error) {
	// 查询 FriendRequest 表中待处理的申请
	requestQ := dao.Use(s.db).FriendRequest
	requestDo := requestQ.WithContext(ctx)

	expireTime := time.Now().AddDate(0, 0, -FriendRequestExpireDays)

	// 查找好友申请记录：SenderID = friendID, ReceiverID = userID, Status = pending, 未过期
	_, err := requestDo.Where(
		requestQ.SenderID.Eq(friendID),
		requestQ.ReceiverID.Eq(userID),
		requestQ.Status.Eq(FriendRequestStatusPending),
		requestQ.CreatedAt.Gt(expireTime),
	).First()

	if err != nil {
		return nil, fmt.Errorf("未找到待处理的好友申请")
	}

	switch action {
	case ActionParamAccept:
		// 接受好友申请：在Friend表创建normal记录
		friendQ := dao.Use(s.db).Friend
		friendDo := friendQ.WithContext(ctx)

		friendRecord := model.Friend{
			UserA:  friendID,
			UserB:  userID,
			Status: FriendStatusNormal,
		}

		if err := friendDo.Create(&friendRecord); err != nil {
			return nil, fmt.Errorf("创建好友关系失败: %v", err)
		}

		// 删除FriendRequest表中的对应记录（软删除）
		if _, err := requestDo.Where(
			requestQ.SenderID.Eq(friendID),
			requestQ.ReceiverID.Eq(userID),
		).Delete(); err != nil {
			return nil, fmt.Errorf("删除好友申请记录失败: %v", err)
		}

		return &dto.AddFriendResponse{
			FriendID: friendID,
			Status:   FriendStatusNormal,
		}, nil
	case ActionParamReject:
		// 拒绝好友申请：更新FriendRequest状态为rejected
		if _, err := requestDo.Where(
			requestQ.SenderID.Eq(friendID),
			requestQ.ReceiverID.Eq(userID),
		).Update(requestQ.Status, FriendRequestStatusRejected); err != nil {
			return nil, fmt.Errorf("拒绝好友申请失败: %v", err)
		}

		return &dto.AddFriendResponse{
			FriendID: friendID,
			Status:   FriendRequestStatusRejected,
		}, nil
	}
	// 理论上不会跑到这里
	return nil, fmt.Errorf("无效的参数")
}

// DeleteFriend 删除好友
func (s *UserService) DeleteFriend(ctx context.Context, userID string, friendID string) error {
	friendQ := dao.Use(s.db).Friend
	friendDo := friendQ.WithContext(ctx)

	// 查找好友关系记录（双向查找）
	// 情况1: userID 是 UserA, friendID 是 UserB
	// 情况2: userID 是 UserB, friendID 是 UserA
	record, err := friendDo.Where(
		friendQ.UserA.Eq(userID),
		friendQ.UserB.Eq(friendID),
	).Or(
		friendQ.UserA.Eq(friendID),
		friendQ.UserB.Eq(userID),
	).First()

	if err != nil {
		return fmt.Errorf("好友关系不存在")
	}

	// 验证权限：只能删除状态为 normal 的好友关系
	if record.Status != FriendStatusNormal {
		return fmt.Errorf("只能删除正常的好友关系")
	}

	// 更新状态为 removed
	_, err = friendDo.Where(
		friendQ.UserA.Eq(record.UserA),
		friendQ.UserB.Eq(record.UserB),
	).Update(friendQ.Status, FriendStatusRemoved)

	if err != nil {
		return fmt.Errorf("删除好友失败: %v", err)
	}

	return nil
}

// IsFriend 检查两个用户是否是好友关系
func (s *UserService) IsFriend(ctx context.Context, userID string, targetUserID string) (bool, error) {
	friendQ := dao.Use(s.db).Friend
	friendDo := friendQ.WithContext(ctx)

	count, err := friendDo.Where(
		friendQ.Status.Eq(FriendStatusNormal),
		friendQ.UserA.Eq(userID),
		friendQ.UserB.Eq(targetUserID),
	).Or(
		friendQ.Status.Eq(FriendStatusNormal),
		friendQ.UserA.Eq(targetUserID),
		friendQ.UserB.Eq(userID),
	).Count()

	if err != nil {
		return false, err
	}

	return count > 0, nil
}
