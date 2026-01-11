package service

import (
	"chat_backend/internal/cache"
	"chat_backend/internal/dao"
	"chat_backend/internal/dto"
	"chat_backend/internal/model"
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
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

const (
	errAlreadyFriends              = "你们已经是好友了"
	errFriendRequestPendingExists  = "您已经发送过好友申请，正在等待对方处理"
	errRequestRejected             = "对方已拒绝您的好友申请，请七天后再试"
	errIncomingRequestExists       = "对方已经向您发送了好友申请，请在待处理申请中查看"
	errUnknownUser                 = "unknown"
	errInvalidStatus               = "无效的status参数"
	errNoPendingRequest            = "未找到待处理的好友申请"
	errCreateFriendRelationFailed  = "创建好友关系失败"
	errDeleteFriendRequestFailed   = "删除好友申请记录失败"
	errRejectFriendRequestFailed   = "拒绝好友申请失败"
	errInvalidFriendRequestAction  = "无效的参数"
	errFriendRelationNotFound      = "好友关系不存在"
	errCannotDeleteNonNormalFriend = "只能删除正常的好友关系"
	errDeleteFriendFailed          = "删除好友失败"
)

type UserService struct {
	db          *gorm.DB
	userCache   *cache.UserCacheManager
	friendCache *cache.FriendCacheManager
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{
		db:          db,
		userCache:   cache.NewUserCacheManager(),
		friendCache: cache.NewFriendCacheManager(),
	}
}

// GetMe 获取自己的资料
func (s *UserService) GetMe(ctx context.Context, userID string, username string) (*dto.UserInfoResponse, error) {
	// 先尝试从缓存获取
	userInfo, err := s.userCache.GetOrLoadUserInfo(ctx, userID, func(id string) (*cache.UserInfo, error) {
		q := dao.Use(s.db).User
		do := q.WithContext(ctx)
		user, err := do.Where(q.ID.Eq(id)).First()
		if err != nil {
			return nil, err
		}
		avatarUrl, err := s.GetUserAvatarUrl(user.ID, user.Username)
		if err != nil {
			return nil, err
		}
		return &cache.UserInfo{
			UserID:   user.ID,
			Username: user.Username,
			Avatar:   avatarUrl,
		}, nil
	})
	if err != nil {
		return nil, err
	}

	return &dto.UserInfoResponse{
		UserID:   userInfo.UserID,
		Username: userInfo.Username,
		Avatar:   userInfo.Avatar,
	}, nil
}

// GetUserInfo 获取用户资料
func (s *UserService) GetUserInfo(ctx context.Context, userID string, username string) (*dto.UserInfoResponse, error) {
	// 先尝试从缓存获取
	userInfo, err := s.userCache.GetOrLoadUserInfo(ctx, userID, func(id string) (*cache.UserInfo, error) {
		q := dao.Use(s.db).User
		do := q.WithContext(ctx)
		user, err := do.Where(q.ID.Eq(id)).First()
		if err != nil {
			return nil, err
		}
		avatarUrl, err := s.GetUserAvatarUrl(user.ID, user.Username)
		if err != nil {
			return nil, err
		}
		return &cache.UserInfo{
			UserID:   user.ID,
			Username: user.Username,
			Avatar:   avatarUrl,
		}, nil
	})
	if err != nil {
		return nil, err
	}

	return &dto.UserInfoResponse{
		UserID:   userInfo.UserID,
		Username: userInfo.Username,
		Avatar:   userInfo.Avatar,
	}, nil
}

// GetUsernameByUserID 通过 userID 查询 username
func (s *UserService) GetUsernameByUserID(ctx context.Context, userID string) (string, error) {
	// 先尝试从缓存获取
	userInfo, err := s.userCache.GetOrLoadUserInfo(ctx, userID, func(id string) (*cache.UserInfo, error) {
		q := dao.Use(s.db).User
		do := q.WithContext(ctx)
		user, err := do.Where(q.ID.Eq(id)).First()
		if err != nil {
			return nil, err
		}
		avatarUrl, err := s.GetUserAvatarUrl(user.ID, user.Username)
		if err != nil {
			return nil, err
		}
		return &cache.UserInfo{
			UserID:   user.ID,
			Username: user.Username,
			Avatar:   avatarUrl,
		}, nil
	})
	if err != nil {
		return "", err
	}
	return userInfo.Username, nil
}

// GetUserIDByUsername 通过 username 查询 userID
func (s *UserService) GetUserIDByUsername(ctx context.Context, username string) (string, error) {
	// 先尝试从缓存获取
	userID, err := s.userCache.GetOrLoadUserIDByUsername(ctx, username, func(name string) (string, error) {
		q := dao.Use(s.db).User
		do := q.WithContext(ctx)
		user, err := do.Where(q.Username.Eq(name)).First()
		if err != nil {
			return "", err
		}
		// 缓存用户信息
		avatarUrl, err := s.GetUserAvatarUrl(user.ID, user.Username)
		if err == nil {
			_ = s.userCache.SetUserInfo(ctx, user.ID, user.Username, avatarUrl)
		}
		return user.ID, nil
	})
	if err != nil {
		return "", err
	}
	return userID, nil
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
	// 1. 先从缓存检查是否已经是好友
	isFriend, err := s.friendCache.IsFriend(ctx, userID, friendID)
	if err == nil && isFriend {
		return nil, fmt.Errorf(errAlreadyFriends)
	}

	// 2. 检查 Friend 表是否已存在好友关系（缓存未命中时）
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
		// 更新缓存
		_ = s.friendCache.AddFriend(ctx, userID, friendID)
		_ = s.friendCache.AddFriend(ctx, friendID, userID)
		return nil, fmt.Errorf(errAlreadyFriends)
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
	if err == nil {
		switch r.Status {
		case FriendRequestStatusPending:
			return nil, fmt.Errorf(errFriendRequestPendingExists)
		case FriendRequestStatusRejected:
			return nil, fmt.Errorf(errRequestRejected)
		}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	_, err = requestDo.Where(
		requestQ.SenderID.Eq(friendID),
		requestQ.ReceiverID.Eq(userID),
		requestQ.Status.Eq(FriendRequestStatusPending),
		requestQ.CreatedAt.Gt(expireTime),
	).First()
	if err == nil {
		return nil, fmt.Errorf(errIncomingRequestExists)
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

	// 4. 获取发送者信息并缓存好友申请
	senderInfo, err := s.GetUserInfo(ctx, userID, "")
	if err != nil {
		return nil, err
	}
	friendRequest := &cache.FriendInfo{
		UserID:   userID,
		Username: senderInfo.Username,
		Avatar:   senderInfo.Avatar,
		Status:   FriendRequestStatusPending,
		CreateAt: record.CreatedAt.Unix(),
	}
	_ = s.friendCache.AddFriendRequest(ctx, friendID, friendRequest)

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

	return nil, fmt.Errorf(errInvalidStatus)
}

func (s *UserService) getPendingFriendRequests(ctx context.Context, userID string) ([]*dto.FriendInfoResponse, error) {
	// 先尝试从缓存获取
	cachedRequests, err := s.friendCache.GetFriendRequests(ctx, userID)
	if err == nil && len(cachedRequests) > 0 {
		responses := make([]*dto.FriendInfoResponse, 0, len(cachedRequests))
		for _, req := range cachedRequests {
			responses = append(responses, &dto.FriendInfoResponse{
				UserID:   req.UserID,
				Username: req.Username,
				Avatar:   req.Avatar,
				Status:   FriendRequestStatusPending,
				CreateAt: req.CreateAt,
			})
		}
		return responses, nil
	}

	// 缓存未命中，从数据库查询
	requestQ := dao.Use(s.db).FriendRequest
	requestDo := requestQ.WithContext(ctx)

	expireTime := time.Now().AddDate(0, 0, -FriendRequestExpireDays)

	var requests []model.FriendRequest
	err = requestDo.Where(
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
	friendInfos := make([]*cache.FriendInfo, 0, len(requests))

	for _, r := range requests {
		username, err := s.GetUsernameByUserID(ctx, r.SenderID)
		if err != nil {
			username = errUnknownUser
		}
		avatarUrl, err := s.GetUserAvatarUrl(r.SenderID, username)
		if err != nil {
			avatarUrl = errUnknownUser
		}

		responses = append(responses, &dto.FriendInfoResponse{
			UserID:   r.SenderID,
			Username: username,
			Avatar:   avatarUrl,
			Status:   r.Status,
			CreateAt: r.CreatedAt.Unix(),
		})

		// 添加到缓存
		friendInfos = append(friendInfos, &cache.FriendInfo{
			UserID:   r.SenderID,
			Username: username,
			Avatar:   avatarUrl,
			Status:   r.Status,
			CreateAt: r.CreatedAt.Unix(),
		})
	}

	// 批量缓存好友申请
	_ = s.friendCache.BatchAddFriendRequests(ctx, userID, friendInfos)

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
			username = errUnknownUser
		}
		avatarUrl, err := s.GetUserAvatarUrl(friendUserID, username)
		if err != nil {
			avatarUrl = errUnknownUser
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
		return nil, fmt.Errorf(errNoPendingRequest)
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
			return nil, fmt.Errorf("%s: %v", errCreateFriendRelationFailed, err)
		}

		// 删除FriendRequest表中的对应记录（软删除）
		if _, err := requestDo.Where(
			requestQ.SenderID.Eq(friendID),
			requestQ.ReceiverID.Eq(userID),
		).Delete(); err != nil {
			return nil, fmt.Errorf("%s: %v", errDeleteFriendRequestFailed, err)
		}

		// 更新缓存：添加好友关系
		_ = s.friendCache.AddFriend(ctx, userID, friendID)
		_ = s.friendCache.AddFriend(ctx, friendID, userID)
		// 删除好友申请缓存
		_ = s.friendCache.RemoveFriendRequest(ctx, userID, friendID)

		return &dto.AddFriendResponse{
			FriendID: friendID,
			Status:   FriendStatusNormal,
		}, nil
	case ActionParamReject:
		if _, err := requestDo.Where(
			requestQ.SenderID.Eq(friendID),
			requestQ.ReceiverID.Eq(userID),
		).Update(requestQ.Status, FriendRequestStatusRejected); err != nil {
			return nil, fmt.Errorf("%s: %v", errRejectFriendRequestFailed, err)
		}

		// 删除好友申请缓存
		_ = s.friendCache.RemoveFriendRequest(ctx, userID, friendID)

		return &dto.AddFriendResponse{
			FriendID: friendID,
			Status:   FriendRequestStatusRejected,
		}, nil
	}
	return nil, fmt.Errorf(errInvalidFriendRequestAction)
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
		return fmt.Errorf(errFriendRelationNotFound)
	}

	// 验证权限：只能删除状态为 normal 的好友关系
	if record.Status != FriendStatusNormal {
		return fmt.Errorf(errCannotDeleteNonNormalFriend)
	}

	// 更新状态为 removed
	_, err = friendDo.Where(
		friendQ.UserA.Eq(record.UserA),
		friendQ.UserB.Eq(record.UserB),
	).Update(friendQ.Status, FriendStatusRemoved)

	if err != nil {
		return fmt.Errorf("%s: %v", errDeleteFriendFailed, err)
	}

	// 更新缓存：移除好友关系
	_ = s.friendCache.RemoveFriend(ctx, userID, friendID)
	_ = s.friendCache.RemoveFriend(ctx, friendID, userID)

	return nil
}

// IsFriend 检查两个用户是否是好友关系
func (s *UserService) IsFriend(ctx context.Context, userID string, targetUserID string) (bool, error) {
	// 先从缓存检查
	isFriend, err := s.friendCache.IsFriend(ctx, userID, targetUserID)
	if err == nil && isFriend {
		return true, nil
	}

	// 缓存未命中，从数据库查询
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

	isFriend = count > 0
	// 更新缓存
	if isFriend {
		_ = s.friendCache.AddFriend(ctx, userID, targetUserID)
		_ = s.friendCache.AddFriend(ctx, targetUserID, userID)
	}

	return isFriend, nil
}
