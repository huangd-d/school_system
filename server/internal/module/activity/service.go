package activity

import (
	"context"
	"school-system/internal/model"
	"school-system/pkg/apperror"
	"time"
)

// Repository 活动数据访问接口
type Repository interface {
	FindAll(ctx context.Context) ([]model.Activity, error)
	FindByCampusID(ctx context.Context, campusID uint) ([]model.Activity, error)
	FindByContactUserID(ctx context.Context, userID uint) ([]model.Activity, error)
	FindByID(ctx context.Context, id uint) (*model.Activity, error)
	FindByIDs(ctx context.Context, ids []uint) ([]model.Activity, error)
	Create(ctx context.Context, activity *model.Activity) error
	Update(ctx context.Context, activity *model.Activity) error
	// 联系人
	SetContacts(ctx context.Context, activityID uint, userIDs []uint) error
	FindContactIDs(ctx context.Context, activityID uint) ([]uint, error)
	// 执行记录
	CreateExecution(ctx context.Context, exec *model.ExecutionRecord) error
	FindExecutions(ctx context.Context, activityID uint) ([]model.ExecutionRecord, error)
	SumExecutions(ctx context.Context, activityID uint) (int, error)
}

// CampusLookup 校区查询接口（由 campus 模块满足，避免跨模块直接依赖）
type CampusLookup interface {
	FindByID(ctx context.Context, id uint) (*model.Campus, error)
}

// UserLookup 用户查询接口（由 user 模块满足）
type UserLookup interface {
	FindByID(ctx context.Context, id uint) (*model.User, error)
}

// ActivityWithSummary 活动列表项（含联系人和执行汇总）
type ActivityWithSummary struct {
	model.Activity
	ContactIDs    []uint
	TotalExecuted int
}

// ActivityDetail 活动详情
type ActivityDetail struct {
	model.Activity
	Contacts      []model.User
	Executions    []model.ExecutionRecord
	TotalExecuted int
}

// Service 活动业务逻辑
type Service struct {
	repo     Repository
	campuses CampusLookup
	users    UserLookup
}

func NewService(repo Repository, campuses CampusLookup, users UserLookup) *Service {
	return &Service{repo: repo, campuses: campuses, users: users}
}

// List 查询活动列表（按角色限定数据范围）
func (s *Service) List(ctx context.Context, userID, campusID uint, role string) ([]ActivityWithSummary, error) {
	var activities []model.Activity
	var err error

	switch role {
	case model.RoleHQAdmin:
		activities, err = s.repo.FindAll(ctx)
	case model.RoleCampusOperator:
		activities, err = s.repo.FindByCampusID(ctx, campusID)
	case model.RoleActivityContact:
		activities, err = s.repo.FindByContactUserID(ctx, userID)
	default:
		return nil, apperror.ErrActivityPermissionDenied
	}

	if err != nil {
		return nil, apperror.Newf(apperror.ErrInternal.Code, "查询活动列表失败: %v", err)
	}

	result := make([]ActivityWithSummary, 0, len(activities))
	for _, a := range activities {
		s.autoTransition(ctx, &a)
		contactIDs, _ := s.repo.FindContactIDs(ctx, a.ID)
		total, _ := s.repo.SumExecutions(ctx, a.ID)
		result = append(result, ActivityWithSummary{
			Activity:      a,
			ContactIDs:    contactIDs,
			TotalExecuted: total,
		})
	}

	return result, nil
}

// Create 新建活动
func (s *Service) Create(ctx context.Context, name string, campusID uint, contactIDs []uint,
	plannedExec int, startDate, endDate string, createdBy uint,
	creatorRole string, creatorCampusID uint) (*model.Activity, error) {

	// 名称校验
	if name == "" {
		return nil, apperror.ErrActivityNameEmpty
	}
	if len(name) > 200 {
		return nil, apperror.ErrActivityNameTooLong
	}

	// 计划次数校验
	if plannedExec <= 0 {
		return nil, apperror.ErrActivityPlannedExecInvalid
	}

	// 角色权限校验
	if creatorRole != model.RoleHQAdmin && creatorRole != model.RoleCampusOperator {
		return nil, apperror.ErrActivityPermissionDenied
	}

	// 校区操作员只能创建本校区的活动
	if creatorRole == model.RoleCampusOperator && creatorCampusID != campusID {
		return nil, apperror.ErrActivityCampusMismatch
	}

	// 校区存在性校验
	if _, err := s.campuses.FindByID(ctx, campusID); err != nil {
		return nil, apperror.ErrActivityCampusNotFound
	}

	// 日期校验
	st, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return nil, apperror.ErrActivityDateInvalid
	}
	ed, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return nil, apperror.ErrActivityDateInvalid
	}
	if !ed.After(st) {
		return nil, apperror.ErrActivityDateInvalid
	}

	// 联系人校验
	if len(contactIDs) > 0 {
		if err := s.validateContacts(ctx, contactIDs, campusID); err != nil {
			return nil, err
		}
	}

	activity := &model.Activity{
		Name:              name,
		CampusID:          campusID,
		PlannedExecutions: plannedExec,
		StartDate:         st,
		EndDate:           ed,
		Status:            model.ActivityNotStarted,
		CreatedBy:         createdBy,
	}

	if err := s.repo.Create(ctx, activity); err != nil {
		return nil, apperror.Newf(apperror.ErrInternal.Code, "创建活动失败: %v", err)
	}

	// 设置联系人
	if err := s.repo.SetContacts(ctx, activity.ID, contactIDs); err != nil {
		return nil, apperror.Newf(apperror.ErrInternal.Code, "设置活动联系人失败: %v", err)
	}

	return activity, nil
}

// Update 编辑活动（支持部分更新）
func (s *Service) Update(ctx context.Context, id uint, name string, contactIDs []uint,
	plannedExec int, operatorRole string, operatorCampusID uint) (*model.Activity, error) {

	// 角色权限校验
	if operatorRole != model.RoleHQAdmin && operatorRole != model.RoleCampusOperator {
		return nil, apperror.ErrActivityPermissionDenied
	}

	activity, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, apperror.ErrActivityNotFound
	}

	// 校区操作员只能修改本校区的活动
	if operatorRole == model.RoleCampusOperator && operatorCampusID != activity.CampusID {
		return nil, apperror.ErrActivityPermissionDenied
	}

	// 已归档的活动不可修改
	if activity.Status == model.ActivityArchived {
		return nil, apperror.ErrActivityArchivedCannotModify
	}

	// 名称更新（非空时才更新）
	if name != "" {
		if len(name) > 200 {
			return nil, apperror.ErrActivityNameTooLong
		}
		activity.Name = name
	}

	// 计划次数更新（>0 时才更新）
	if plannedExec > 0 {
		total, err := s.repo.SumExecutions(ctx, id)
		if err != nil {
			return nil, apperror.Newf(apperror.ErrInternal.Code, "查询执行记录失败: %v", err)
		}
		if plannedExec < total {
			return nil, apperror.ErrActivityPlannedExecBelowExecuted
		}
		activity.PlannedExecutions = plannedExec
	}

	// 联系人更新（nil 表示不修改；空切片表示清空）
	if contactIDs != nil {
		if len(contactIDs) > 0 {
			if err := s.validateContacts(ctx, contactIDs, activity.CampusID); err != nil {
				return nil, err
			}
		}
		if err := s.repo.SetContacts(ctx, id, contactIDs); err != nil {
			return nil, apperror.Newf(apperror.ErrInternal.Code, "更新活动联系人失败: %v", err)
		}
	}

	if err := s.repo.Update(ctx, activity); err != nil {
		return nil, apperror.Newf(apperror.ErrInternal.Code, "更新活动失败: %v", err)
	}

	return activity, nil
}

// Detail 获取活动详情
func (s *Service) Detail(ctx context.Context, id uint) (*ActivityDetail, error) {
	activity, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, apperror.ErrActivityNotFound
	}

	// 惰性状态推进
	s.autoTransition(ctx, activity)

	// 联系人
	contactIDs, err := s.repo.FindContactIDs(ctx, id)
	if err != nil {
		return nil, apperror.Newf(apperror.ErrInternal.Code, "查询活动联系人失败: %v", err)
	}

	contacts := make([]model.User, 0, len(contactIDs))
	for _, cid := range contactIDs {
		u, err := s.users.FindByID(ctx, cid)
		if err != nil {
			// 已删除的用户静默跳过
			continue
		}
		contacts = append(contacts, *u)
	}

	// 执行记录
	executions, err := s.repo.FindExecutions(ctx, id)
	if err != nil {
		return nil, apperror.Newf(apperror.ErrInternal.Code, "查询执行记录失败: %v", err)
	}

	total, err := s.repo.SumExecutions(ctx, id)
	if err != nil {
		return nil, apperror.Newf(apperror.ErrInternal.Code, "查询执行汇总失败: %v", err)
	}

	return &ActivityDetail{
		Activity:      *activity,
		Contacts:      contacts,
		Executions:    executions,
		TotalExecuted: total,
	}, nil
}

// AddExecution 录入执行次数
func (s *Service) AddExecution(ctx context.Context, activityID uint, count int,
	recordedBy uint, operatorRole string) error {

	// 次数校验
	if count <= 0 {
		return apperror.ErrActivityExecCountInvalid
	}

	activity, err := s.repo.FindByID(ctx, activityID)
	if err != nil {
		return apperror.ErrActivityNotFound
	}

	// 状态校验：仅未开始和进行中可录入
	if activity.Status != model.ActivityNotStarted && activity.Status != model.ActivityInProgress {
		return apperror.ErrActivityStatusNoExec
	}

	// 权限校验
	if operatorRole != model.RoleHQAdmin {
		contactIDs, err := s.repo.FindContactIDs(ctx, activityID)
		if err != nil {
			return apperror.Newf(apperror.ErrInternal.Code, "查询活动联系人失败: %v", err)
		}
		isContact := false
		for _, cid := range contactIDs {
			if cid == recordedBy {
				isContact = true
				break
			}
		}
		if !isContact {
			return apperror.ErrActivityNotContactPerson
		}
	}

	// 累计不超计划
	total, err := s.repo.SumExecutions(ctx, activityID)
	if err != nil {
		return apperror.Newf(apperror.ErrInternal.Code, "查询执行汇总失败: %v", err)
	}
	if total+count > activity.PlannedExecutions {
		return apperror.ErrActivityExecExceedPlanned
	}

	exec := &model.ExecutionRecord{
		ActivityID: activityID,
		Count:      count,
		RecordedBy: recordedBy,
	}

	if err := s.repo.CreateExecution(ctx, exec); err != nil {
		return apperror.Newf(apperror.ErrInternal.Code, "记录执行失败: %v", err)
	}

	return nil
}

// Archive 归档活动（仅已结算状态可归档）
func (s *Service) Archive(ctx context.Context, id uint, operatorRole string, operatorCampusID uint) error {
	activity, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return apperror.ErrActivityNotFound
	}

	// 角色权限校验
	if operatorRole != model.RoleHQAdmin && operatorRole != model.RoleCampusOperator {
		return apperror.ErrActivityPermissionDenied
	}

	// 校区操作员只能归档本校区的活动
	if operatorRole == model.RoleCampusOperator && operatorCampusID != activity.CampusID {
		return apperror.ErrActivityPermissionDenied
	}

	// 仅已结算可归档
	if activity.Status != model.ActivitySettled {
		return apperror.ErrActivityNotSettled
	}

	activity.Status = model.ActivityArchived
	if err := s.repo.Update(ctx, activity); err != nil {
		return apperror.Newf(apperror.ErrInternal.Code, "归档活动失败: %v", err)
	}

	return nil
}

// ---- 辅助方法 ----

// validateContacts 校验联系人：存在且属于同一校区
func (s *Service) validateContacts(ctx context.Context, contactIDs []uint, campusID uint) error {
	for _, cid := range contactIDs {
		u, err := s.users.FindByID(ctx, cid)
		if err != nil {
			return apperror.ErrActivityContactsNotFound
		}
		if u.CampusID != campusID {
			return apperror.ErrActivityContactsCampusMismatch
		}
	}
	return nil
}

// autoTransition 根据当前时间自动推进活动状态（best-effort）
func (s *Service) autoTransition(ctx context.Context, activity *model.Activity) {
	now := time.Now()
	updated := false

	if activity.Status == model.ActivityNotStarted && !activity.StartDate.After(now) {
		activity.Status = model.ActivityInProgress
		updated = true
	}
	if activity.Status == model.ActivityInProgress && activity.EndDate.Before(now) {
		activity.Status = model.ActivityEnded
		updated = true
	}

	if updated {
		_ = s.repo.Update(ctx, activity) // best-effort，忽略错误
	}
}
