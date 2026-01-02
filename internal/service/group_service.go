package service

import (
	"chat_backend/internal/dao"
	"chat_backend/internal/dto"
	"chat_backend/internal/model"
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

const (
	RoleOwner  = "owner"
	RoleMember = "member"
)

const (
	StatusPending  = "pending"
	StatusRejected = "rejected"
	StatusApproved = "approved"
)

type GroupService struct {
	db *gorm.DB
}

func NewGroupService(db *gorm.DB) *GroupService {
	return &GroupService{
		db: db,
	}
}

// CreateGroup 创建群组
func (s *GroupService) CreateGroup(ctx context.Context, userID string, name string) (*dto.GroupResponse, error) {
	var group model.Group

	err := s.db.Transaction(func(tx *gorm.DB) error {
		q := dao.Use(tx).Group
		do := q.WithContext(ctx)

		group = model.Group{
			Name:        name,
			OwnerID:     userID,
			MemberCount: 1,
		}

		if err := do.Create(&group); err != nil {
			return err
		}

		mq := dao.Use(tx).GroupMember
		mdo := mq.WithContext(ctx)

		groupMember := model.GroupMember{
			GroupID: group.ID,
			UserID:  userID,
			Role:    RoleOwner,
		}

		if err := mdo.Create(&groupMember); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &dto.GroupResponse{
		GroupID:     group.ID,
		Name:        group.Name,
		OwnerID:     group.OwnerID,
		MemberCount: group.MemberCount,
		CreatedAt:   group.CreatedAt.Format(time.RFC3339),
	}, nil
}

// GetGroupList 获取群组列表
func (s *GroupService) GetGroupList(ctx context.Context, userID string, role string) ([]*dto.GroupListResponse, error) {
	mq := dao.Use(s.db).GroupMember
	mdo := mq.WithContext(ctx)

	var groupMembers []model.GroupMember

	// 根据角色筛选
	query := mdo.Where(mq.UserID.Eq(userID))
	switch role {
	case RoleOwner:
		query = query.Where(mq.Role.Eq(RoleOwner))
	case RoleMember:
		query = query.Where(mq.Role.Eq(RoleMember))
	}

	err := query.Scan(&groupMembers)
	if err != nil {
		return nil, err
	}

	if len(groupMembers) == 0 {
		return []*dto.GroupListResponse{}, nil
	}

	gq := dao.Use(s.db).Group
	gdo := gq.WithContext(ctx)

	// 获取群组详细信息
	var responses []*dto.GroupListResponse
	for _, gm := range groupMembers {
		group, err := gdo.Where(gq.ID.Eq(gm.GroupID)).First()
		if err != nil {
			continue
		}

		responses = append(responses, &dto.GroupListResponse{
			GroupID:     group.ID,
			Name:        group.Name,
			OwnerID:     group.OwnerID,
			MemberCount: group.MemberCount,
			Role:        gm.Role,
			CreatedAt:   group.CreatedAt.Format(time.RFC3339),
		})
	}

	return responses, nil
}

// GetGroupDetail 获取群组详情
func (s *GroupService) GetGroupDetail(ctx context.Context, userID string, groupID string) (*dto.GroupDetailResponse, error) {
	// 检查用户是否在群组中
	mq := dao.Use(s.db).GroupMember
	mdo := mq.WithContext(ctx)

	_, err := mdo.Where(mq.GroupID.Eq(groupID), mq.UserID.Eq(userID)).First()
	if err != nil {
		return nil, fmt.Errorf("you are not in this group")
	}

	// 获取群组信息
	gq := dao.Use(s.db).Group
	gdo := gq.WithContext(ctx)

	group, err := gdo.Where(gq.ID.Eq(groupID)).First()
	if err != nil {
		return nil, err
	}

	// 获取群主用户名
	uq := dao.Use(s.db).User
	udo := uq.WithContext(ctx)

	owner, err := udo.Where(uq.ID.Eq(group.OwnerID)).First()
	if err != nil {
		return nil, err
	}

	// 获取所有成员
	var members []model.GroupMember
	err = mdo.Where(mq.GroupID.Eq(groupID)).Scan(&members)
	if err != nil {
		return nil, err
	}

	// 构建成员信息列表
	var memberInfos []dto.GroupMemberInfo
	for _, m := range members {
		user, err := udo.Where(uq.ID.Eq(m.UserID)).First()
		if err != nil {
			continue
		}
		memberInfos = append(memberInfos, dto.GroupMemberInfo{
			UserID:   m.UserID,
			Username: user.Username,
			Role:     m.Role,
			JoinedAt: m.CreatedAt.Format(time.RFC3339),
		})
	}

	return &dto.GroupDetailResponse{
		GroupID:     group.ID,
		Name:        group.Name,
		OwnerID:     group.OwnerID,
		OwnerName:   owner.Username,
		MemberCount: group.MemberCount,
		CreatedAt:   group.CreatedAt.Format(time.RFC3339),
		Members:     memberInfos,
	}, nil
}

// JoinGroup 加入群组
func (s *GroupService) JoinGroup(ctx context.Context, userID string, groupID string, inviteCode string) (*dto.JoinGroupResponse, error) {
	var group *model.Group

	err := s.db.Transaction(func(tx *gorm.DB) error {
		gq := dao.Use(tx).Group
		gdo := gq.WithContext(ctx)

		var err error
		group, err = gdo.Where(gq.ID.Eq(groupID)).First()
		if err != nil {
			return fmt.Errorf("group not found")
		}

		mq := dao.Use(tx).GroupMember
		mdo := mq.WithContext(ctx)

		_, err = mdo.Where(mq.GroupID.Eq(groupID), mq.UserID.Eq(userID)).First()
		if err == nil {
			return fmt.Errorf("already in group")
		}

		groupMember := model.GroupMember{
			GroupID: groupID,
			UserID:  userID,
			Role:    RoleMember,
		}

		if err = mdo.Create(&groupMember); err != nil {
			return err
		}

		_, err = gdo.Where(gq.ID.Eq(groupID)).UpdateSimple(gq.MemberCount.Add(1))
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &dto.JoinGroupResponse{
		GroupID: group.ID,
		Name:    group.Name,
		Status:  "joined",
	}, nil
}

// RequestJoinGroup 申请加入群组
func (s *GroupService) RequestJoinGroup(ctx context.Context, userID string, groupID string, message string) (*dto.RequestJoinGroupResponse, error) {
	gq := dao.Use(s.db).Group
	gdo := gq.WithContext(ctx)

	group, err := gdo.Where(gq.ID.Eq(groupID)).First()
	if err != nil {
		return nil, fmt.Errorf("group not found")
	}

	mq := dao.Use(s.db).GroupMember
	mdo := mq.WithContext(ctx)

	_, err = mdo.Where(mq.GroupID.Eq(groupID), mq.UserID.Eq(userID)).First()
	if err == nil {
		return nil, fmt.Errorf("already in group")
	}

	rq := dao.Use(s.db).GroupJoinRequest
	rdo := rq.WithContext(ctx)

	sevenDaysAgo := time.Now().AddDate(0, 0, -7)

	_, err = rdo.Where(
		rq.SenderID.Eq(userID),
		rq.TargetGroupID.Eq(groupID),
		rq.Status.Eq(StatusPending),
		rq.CreatedAt.Gte(sevenDaysAgo),
	).First()
	if err == nil {
		return nil, fmt.Errorf("pending request already exists")
	}

	_, err = rdo.Where(
		rq.SenderID.Eq(userID),
		rq.TargetGroupID.Eq(groupID),
		rq.Status.Eq(StatusRejected),
		rq.CreatedAt.Gte(sevenDaysAgo),
	).First()
	if err == nil {
		return nil, fmt.Errorf("cannot request within cooldown period")
	}

	joinRequest := model.GroupJoinRequest{
		SenderID:      userID,
		TargetGroupID: groupID,
		Status:        StatusPending,
		Message:       message,
	}

	if err := rdo.Create(&joinRequest); err != nil {
		return nil, err
	}

	return &dto.RequestJoinGroupResponse{
		GroupID: group.ID,
		Name:    group.Name,
		Status:  "pending",
	}, nil
}

// LeaveGroup 退出群组
func (s *GroupService) LeaveGroup(ctx context.Context, userID string, groupID string) error {
	err := s.db.Transaction(func(tx *gorm.DB) error {
		gq := dao.Use(tx).Group
		gdo := gq.WithContext(ctx)

		_, err := gdo.Where(gq.ID.Eq(groupID)).First()
		if err != nil {
			return fmt.Errorf("group not found")
		}

		mq := dao.Use(tx).GroupMember
		mdo := mq.WithContext(ctx)

		member, err := mdo.Where(mq.GroupID.Eq(groupID), mq.UserID.Eq(userID)).First()
		if err != nil {
			return fmt.Errorf("not in group")
		}

		if member.Role == RoleOwner {
			return fmt.Errorf("cannot leave as owner")
		}

		_, err = mdo.Where(mq.GroupID.Eq(groupID), mq.UserID.Eq(userID)).Delete()
		if err != nil {
			return err
		}

		_, err = gdo.Where(gq.ID.Eq(groupID)).UpdateSimple(gq.MemberCount.Sub(1))
		if err != nil {
			return err
		}

		return nil
	})

	return err
}

// DisbandGroup 解散群组
func (s *GroupService) DisbandGroup(ctx context.Context, userID string, groupID string) error {
	err := s.db.Transaction(func(tx *gorm.DB) error {
		gq := dao.Use(tx).Group
		gdo := gq.WithContext(ctx)

		group, err := gdo.Where(gq.ID.Eq(groupID)).First()
		if err != nil {
			return fmt.Errorf("group not found")
		}

		if group.OwnerID != userID {
			return fmt.Errorf("permission denied")
		}

		mq := dao.Use(tx).GroupMember
		mdo := mq.WithContext(ctx)

		_, err = mdo.Where(mq.GroupID.Eq(groupID)).Delete()
		if err != nil {
			return err
		}

		_, err = gdo.Where(gq.ID.Eq(groupID)).Delete()
		if err != nil {
			return err
		}

		return nil
	})

	return err
}

// TransferGroup 转让群组
func (s *GroupService) TransferGroup(ctx context.Context, userID string, groupID string, newOwnerID string) error {
	err := s.db.Transaction(func(tx *gorm.DB) error {
		gq := dao.Use(tx).Group
		gdo := gq.WithContext(ctx)

		group, err := gdo.Where(gq.ID.Eq(groupID)).First()
		if err != nil {
			return fmt.Errorf("group not found")
		}

		if group.OwnerID != userID {
			return fmt.Errorf("permission denied")
		}

		mq := dao.Use(tx).GroupMember
		mdo := mq.WithContext(ctx)

		_, err = mdo.Where(mq.GroupID.Eq(groupID), mq.UserID.Eq(newOwnerID)).First()
		if err != nil {
			return fmt.Errorf("target user not in group")
		}

		_, err = mdo.Where(mq.GroupID.Eq(groupID), mq.UserID.Eq(newOwnerID)).Update(mq.Role, RoleOwner)
		if err != nil {
			return err
		}

		_, err = mdo.Where(mq.GroupID.Eq(groupID), mq.UserID.Eq(userID)).Update(mq.Role, RoleMember)
		if err != nil {
			return err
		}

		_, err = gdo.Where(gq.ID.Eq(groupID)).Update(gq.OwnerID, newOwnerID)
		if err != nil {
			return err
		}

		return nil
	})

	return err
}

// RemoveMember 移除群组成员
func (s *GroupService) RemoveMember(ctx context.Context, userID string, groupID string, targetUserID string) error {
	err := s.db.Transaction(func(tx *gorm.DB) error {
		gq := dao.Use(tx).Group
		gdo := gq.WithContext(ctx)

		group, err := gdo.Where(gq.ID.Eq(groupID)).First()
		if err != nil {
			return fmt.Errorf("group not found")
		}

		if group.OwnerID != userID {
			return fmt.Errorf("permission denied")
		}

		if targetUserID == userID {
			return fmt.Errorf("cannot remove yourself")
		}

		mq := dao.Use(tx).GroupMember
		mdo := mq.WithContext(ctx)

		targetMember, err := mdo.Where(mq.GroupID.Eq(groupID), mq.UserID.Eq(targetUserID)).First()
		if err != nil {
			return fmt.Errorf("target user not in group")
		}

		if targetMember.Role == RoleOwner {
			return fmt.Errorf("cannot remove owner")
		}

		_, err = mdo.Where(mq.GroupID.Eq(groupID), mq.UserID.Eq(targetUserID)).Delete()
		if err != nil {
			return err
		}

		_, err = gdo.Where(gq.ID.Eq(groupID)).UpdateSimple(gq.MemberCount.Sub(1))
		if err != nil {
			return err
		}

		return nil
	})

	return err
}

// JoinGroupByCode 通过邀请码加入群组
func (s *GroupService) JoinGroupByCode(ctx context.Context, userID string, inviteCode string) (*dto.JoinGroupByCodeResponse, error) {
	var group *model.Group
	var groupID string

	err := s.db.Transaction(func(tx *gorm.DB) error {
		icQ := dao.Use(tx).InvitationCode
		icDo := icQ.WithContext(ctx)

		invitationCode, err := icDo.Where(icQ.Code.Eq(inviteCode)).First()
		if err != nil {
			return fmt.Errorf("invalid invite code")
		}

		groupID = invitationCode.UUID

		gq := dao.Use(tx).Group
		gdo := gq.WithContext(ctx)

		group, err = gdo.Where(gq.ID.Eq(groupID)).First()
		if err != nil {
			return fmt.Errorf("group not found")
		}

		mq := dao.Use(tx).GroupMember
		mdo := mq.WithContext(ctx)

		_, err = mdo.Where(mq.GroupID.Eq(groupID), mq.UserID.Eq(userID)).First()
		if err == nil {
			return fmt.Errorf("already in group")
		}

		groupMember := model.GroupMember{
			GroupID: groupID,
			UserID:  userID,
			Role:    RoleMember,
		}

		if err = mdo.Create(&groupMember); err != nil {
			return err
		}

		_, err = gdo.Where(gq.ID.Eq(groupID)).UpdateSimple(gq.MemberCount.Add(1))
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &dto.JoinGroupByCodeResponse{
		GroupID: group.ID,
		Name:    group.Name,
		Status:  "joined",
	}, nil
}

// SearchGroup 搜索群组
func (s *GroupService) SearchGroup(ctx context.Context, groupName string) ([]*dto.SearchGroupResponse, error) {
	gq := dao.Use(s.db).Group
	gdo := gq.WithContext(ctx)

	var groups []model.Group

	err := gdo.Where(gq.Name.Like("%" + groupName + "%")).Scan(&groups)
	if err != nil {
		return nil, err
	}

	if len(groups) == 0 {
		return []*dto.SearchGroupResponse{}, nil
	}

	responses := make([]*dto.SearchGroupResponse, 0, len(groups))
	for _, group := range groups {
		responses = append(responses, &dto.SearchGroupResponse{
			GroupID:     group.ID,
			Name:        group.Name,
			OwnerID:     group.OwnerID,
			MemberCount: group.MemberCount,
			CreatedAt:   group.CreatedAt.Format(time.RFC3339),
		})
	}

	return responses, nil
}

// GetPendingJoinRequests 获取待审核的入群请求
func (s *GroupService) GetPendingJoinRequests(ctx context.Context, userID string) ([]*dto.PendingJoinRequest, error) {
	gq := dao.Use(s.db).Group
	gdo := gq.WithContext(ctx)

	groups, err := gdo.Where(gq.OwnerID.Eq(userID)).Find()
	if err != nil {
		return nil, err
	}

	if len(groups) == 0 {
		return []*dto.PendingJoinRequest{}, nil
	}

	groupIDs := make([]string, 0, len(groups))
	for _, group := range groups {
		groupIDs = append(groupIDs, group.ID)
	}

	rq := dao.Use(s.db).GroupJoinRequest
	rdo := rq.WithContext(ctx)

	sevenDaysAgo := time.Now().AddDate(0, 0, -7)

	var requests []model.GroupJoinRequest
	err = rdo.Where(
		rq.TargetGroupID.In(groupIDs...),
		rq.Status.Eq(StatusPending),
		rq.CreatedAt.Gte(sevenDaysAgo),
	).Scan(&requests)
	if err != nil {
		return nil, err
	}

	if len(requests) == 0 {
		return []*dto.PendingJoinRequest{}, nil
	}

	uq := dao.Use(s.db).User
	udo := uq.WithContext(ctx)

	responses := make([]*dto.PendingJoinRequest, 0, len(requests))
	for _, req := range requests {
		user, err := udo.Where(uq.ID.Eq(req.SenderID)).First()
		if err != nil {
			continue
		}

		groupName := ""
		for _, group := range groups {
			if group.ID == req.TargetGroupID {
				groupName = group.Name
				break
			}
		}

		responses = append(responses, &dto.PendingJoinRequest{
			RequestID: req.ID,
			GroupID:   req.TargetGroupID,
			GroupName: groupName,
			UserID:    req.SenderID,
			Username:  user.Username,
			Message:   req.Message,
			CreatedAt: req.CreatedAt.Format(time.RFC3339),
		})
	}

	return responses, nil
}

// ApproveJoinRequest 审批入群请求
func (s *GroupService) ApproveJoinRequest(ctx context.Context, userID string, groupID string, senderID string, action string) error {
	gq := dao.Use(s.db).Group
	gdo := gq.WithContext(ctx)

	group, err := gdo.Where(gq.ID.Eq(groupID)).First()
	if err != nil {
		return fmt.Errorf("group not found")
	}

	if group.OwnerID != userID {
		return fmt.Errorf("permission denied")
	}

	rq := dao.Use(s.db).GroupJoinRequest
	rdo := rq.WithContext(ctx)

	sevenDaysAgo := time.Now().AddDate(0, 0, -7)

	joinRequest, err := rdo.Where(
		rq.TargetGroupID.Eq(groupID),
		rq.SenderID.Eq(senderID),
		rq.Status.Eq(StatusPending),
		rq.CreatedAt.Gte(sevenDaysAgo),
	).First()
	if err != nil {
		return fmt.Errorf("join request not found")
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		switch action {
		case "approve":
			mq := dao.Use(tx).GroupMember
			mdo := mq.WithContext(ctx)

			_, err = mdo.Where(mq.GroupID.Eq(groupID), mq.UserID.Eq(senderID)).First()
			if err == nil {
				return fmt.Errorf("already in group")
			}

			groupMember := model.GroupMember{
				GroupID: groupID,
				UserID:  senderID,
				Role:    RoleMember,
			}

			if err = mdo.Create(&groupMember); err != nil {
				return err
			}

			_, err = gdo.Where(gq.ID.Eq(groupID)).UpdateSimple(gq.MemberCount.Add(1))
			if err != nil {
				return err
			}

			_, err = rdo.Where(
				rq.ID.Eq(joinRequest.ID),
			).Delete()
			if err != nil {
				return err
			}
		case "reject":
			_, err = rdo.Where(
				rq.ID.Eq(joinRequest.ID),
			).Update(rq.Status, StatusRejected)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("invalid action")
		}

		return nil
	})

	return err
}
