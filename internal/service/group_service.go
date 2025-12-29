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
	// 创建群组
	group := model.Group{
		Name:        name,
		OwnerID:     userID,
		MemberCount: 1,
	}

	q := dao.Use(s.db).Group
	do := q.WithContext(ctx)

	if err := do.Create(&group); err != nil {
		return nil, err
	}

	// 创建群组成员关系（群主）
	groupMember := model.GroupMember{
		GroupID: group.ID,
		UserID:  userID,
		Role:    RoleOwner,
	}

	mq := dao.Use(s.db).GroupMember
	mdo := mq.WithContext(ctx)

	if err := mdo.Create(&groupMember); err != nil {
		// 如果创建成员关系失败，删除群组
		_, _ = do.Where(q.ID.Eq(group.ID)).Delete()
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
	if role == RoleOwner {
		query = query.Where(mq.Role.Eq(RoleOwner))
	} else if role == RoleMember {
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
	gq := dao.Use(s.db).Group
	gdo := gq.WithContext(ctx)

	// 检查群组是否存在
	group, err := gdo.Where(gq.ID.Eq(groupID)).First()
	if err != nil {
		return nil, fmt.Errorf("group not found")
	}

	// 检查是否已在群组中
	mq := dao.Use(s.db).GroupMember
	mdo := mq.WithContext(ctx)

	_, err = mdo.Where(mq.GroupID.Eq(groupID), mq.UserID.Eq(userID)).First()
	if err == nil {
		return nil, fmt.Errorf("already in group")
	}

	// TODO: 如果群组需要邀请码，检查邀请码（这里简化处理）
	// 实际项目中需要实现邀请码验证逻辑

	// 创建成员关系
	groupMember := model.GroupMember{
		GroupID: groupID,
		UserID:  userID,
		Role:    RoleMember,
	}

	if err := mdo.Create(&groupMember); err != nil {
		return nil, err
	}

	// 更新群组成员数
	_, err = gdo.Where(gq.ID.Eq(groupID)).UpdateSimple(gq.MemberCount.Add(1))
	if err != nil {
		return nil, err
	}

	return &dto.JoinGroupResponse{
		GroupID: group.ID,
		Name:    group.Name,
		Status:  "joined",
	}, nil
}

// LeaveGroup 退出群组
func (s *GroupService) LeaveGroup(ctx context.Context, userID string, groupID string) error {
	gq := dao.Use(s.db).Group
	gdo := gq.WithContext(ctx)

	// 检查群组是否存在
	_, err := gdo.Where(gq.ID.Eq(groupID)).First()
	if err != nil {
		return fmt.Errorf("group not found")
	}

	// 检查是否是群主
	mq := dao.Use(s.db).GroupMember
	mdo := mq.WithContext(ctx)

	member, err := mdo.Where(mq.GroupID.Eq(groupID), mq.UserID.Eq(userID)).First()
	if err != nil {
		return fmt.Errorf("not in group")
	}

	if member.Role == RoleOwner {
		return fmt.Errorf("cannot leave as owner")
	}

	// 删除成员关系
	_, err = mdo.Where(mq.GroupID.Eq(groupID), mq.UserID.Eq(userID)).Delete()
	if err != nil {
		return err
	}

	// 更新群组成员数
	_, err = gdo.Where(gq.ID.Eq(groupID)).UpdateSimple(gq.MemberCount.Sub(1))
	if err != nil {
		return err
	}

	return nil
}

// DisbandGroup 解散群组
func (s *GroupService) DisbandGroup(ctx context.Context, userID string, groupID string) error {
	gq := dao.Use(s.db).Group
	gdo := gq.WithContext(ctx)

	// 检查群组是否存在
	group, err := gdo.Where(gq.ID.Eq(groupID)).First()
	if err != nil {
		return fmt.Errorf("group not found")
	}

	// 检查权限
	if group.OwnerID != userID {
		return fmt.Errorf("permission denied")
	}

	// 删除所有成员
	mq := dao.Use(s.db).GroupMember
	mdo := mq.WithContext(ctx)

	_, err = mdo.Where(mq.GroupID.Eq(groupID)).Delete()
	if err != nil {
		return err
	}

	// 删除群组
	_, err = gdo.Where(gq.ID.Eq(groupID)).Delete()
	if err != nil {
		return err
	}

	return nil
}

// TransferGroup 转让群组
func (s *GroupService) TransferGroup(ctx context.Context, userID string, groupID string, newOwnerID string) error {
	gq := dao.Use(s.db).Group
	gdo := gq.WithContext(ctx)

	// 检查群组是否存在
	group, err := gdo.Where(gq.ID.Eq(groupID)).First()
	if err != nil {
		return fmt.Errorf("group not found")
	}

	// 检查权限
	if group.OwnerID != userID {
		return fmt.Errorf("permission denied")
	}

	// 检查新群主是否在群组中
	mq := dao.Use(s.db).GroupMember
	mdo := mq.WithContext(ctx)

	_, err = mdo.Where(mq.GroupID.Eq(groupID), mq.UserID.Eq(newOwnerID)).First()
	if err != nil {
		return fmt.Errorf("target user not in group")
	}

	// 更新新群主角色
	_, err = mdo.Where(mq.GroupID.Eq(groupID), mq.UserID.Eq(newOwnerID)).Update(mq.Role, RoleOwner)
	if err != nil {
		return err
	}

	// 更新原群主角色
	_, err = mdo.Where(mq.GroupID.Eq(groupID), mq.UserID.Eq(userID)).Update(mq.Role, RoleMember)
	if err != nil {
		return err
	}

	// 更新群组的owner_id
	_, err = gdo.Where(gq.ID.Eq(groupID)).Update(gq.OwnerID, newOwnerID)
	if err != nil {
		return err
	}

	return nil
}

// RemoveMember 移除群组成员
func (s *GroupService) RemoveMember(ctx context.Context, userID string, groupID string, targetUserID string) error {
	gq := dao.Use(s.db).Group
	gdo := gq.WithContext(ctx)

	// 检查群组是否存在
	group, err := gdo.Where(gq.ID.Eq(groupID)).First()
	if err != nil {
		return fmt.Errorf("group not found")
	}

	// 检查权限
	if group.OwnerID != userID {
		return fmt.Errorf("permission denied")
	}

	// 不能移除自己
	if targetUserID == userID {
		return fmt.Errorf("cannot remove yourself")
	}

	// 检查目标用户是否在群组中
	mq := dao.Use(s.db).GroupMember
	mdo := mq.WithContext(ctx)

	targetMember, err := mdo.Where(mq.GroupID.Eq(groupID), mq.UserID.Eq(targetUserID)).First()
	if err != nil {
		return fmt.Errorf("target user not in group")
	}

	// 不能移除群主
	if targetMember.Role == RoleOwner {
		return fmt.Errorf("cannot remove owner")
	}

	// 删除成员
	_, err = mdo.Where(mq.GroupID.Eq(groupID), mq.UserID.Eq(targetUserID)).Delete()
	if err != nil {
		return err
	}

	// 更新群组成员数
	_, err = gdo.Where(gq.ID.Eq(groupID)).UpdateSimple(gq.MemberCount.Sub(1))
	if err != nil {
		return err
	}

	return nil
}

// JoinGroupByCode 通过邀请码加入群组
func (s *GroupService) JoinGroupByCode(ctx context.Context, userID string, inviteCode string) (*dto.JoinGroupByCodeResponse, error) {
	// 验证邀请码
	icQ := dao.Use(s.db).InvitationCode
	icDo := icQ.WithContext(ctx)

	invitationCode, err := icDo.Where(icQ.Code.Eq(inviteCode)).First()
	if err != nil {
		return nil, fmt.Errorf("invalid invite code")
	}

	groupID := invitationCode.UUID

	// 检查群组是否存在
	gq := dao.Use(s.db).Group
	gdo := gq.WithContext(ctx)

	group, err := gdo.Where(gq.ID.Eq(groupID)).First()
	if err != nil {
		return nil, fmt.Errorf("group not found")
	}

	// 检查是否已在群组中
	mq := dao.Use(s.db).GroupMember
	mdo := mq.WithContext(ctx)

	_, err = mdo.Where(mq.GroupID.Eq(groupID), mq.UserID.Eq(userID)).First()
	if err == nil {
		return nil, fmt.Errorf("already in group")
	}

	// 创建成员关系
	groupMember := model.GroupMember{
		GroupID: groupID,
		UserID:  userID,
		Role:    RoleMember,
	}

	if err := mdo.Create(&groupMember); err != nil {
		return nil, err
	}

	// 更新群组成员数
	_, err = gdo.Where(gq.ID.Eq(groupID)).UpdateSimple(gq.MemberCount.Add(1))
	if err != nil {
		return nil, err
	}

	return &dto.JoinGroupByCodeResponse{
		GroupID: group.ID,
		Name:    group.Name,
		Status:  "joined",
	}, nil
}
