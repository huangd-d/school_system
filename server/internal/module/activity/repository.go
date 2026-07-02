package activity

import (
	"context"
	"school-system/internal/model"

	"gorm.io/gorm"
)

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository { return &repository{db: db} }

func (r *repository) FindAll(ctx context.Context) ([]model.Activity, error) {
	var activities []model.Activity
	err := r.db.WithContext(ctx).Find(&activities).Error
	return activities, err
}

func (r *repository) FindByCampusID(ctx context.Context, campusID uint) ([]model.Activity, error) {
	var activities []model.Activity
	err := r.db.WithContext(ctx).Where("campus_id = ?", campusID).Find(&activities).Error
	return activities, err
}

func (r *repository) FindByContactUserID(ctx context.Context, userID uint) ([]model.Activity, error) {
	var activities []model.Activity
	err := r.db.WithContext(ctx).
		Joins("JOIN activity_contacts ON activity_contacts.activity_id = activities.id").
		Where("activity_contacts.user_id = ?", userID).
		Find(&activities).Error
	return activities, err
}

func (r *repository) FindByIDs(ctx context.Context, ids []uint) ([]model.Activity, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var activities []model.Activity
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&activities).Error
	return activities, err
}

func (r *repository) FindByID(ctx context.Context, id uint) (*model.Activity, error) {
	var activity model.Activity
	err := r.db.WithContext(ctx).First(&activity, id).Error
	if err != nil {
		return nil, err
	}
	return &activity, nil
}

func (r *repository) Create(ctx context.Context, activity *model.Activity) error {
	return r.db.WithContext(ctx).Create(activity).Error
}

func (r *repository) Update(ctx context.Context, activity *model.Activity) error {
	return r.db.WithContext(ctx).Save(activity).Error
}

// ---- 联系人 ----
func (r *repository) SetContacts(ctx context.Context, activityID uint, userIDs []uint) error {
	r.db.WithContext(ctx).Where("activity_id = ?", activityID).Delete(&model.ActivityContact{})
	for _, uid := range userIDs {
		r.db.WithContext(ctx).Create(&model.ActivityContact{ActivityID: activityID, UserID: uid})
	}
	return nil
}

func (r *repository) FindContactIDs(ctx context.Context, activityID uint) ([]uint, error) {
	var contacts []model.ActivityContact
	if err := r.db.WithContext(ctx).Where("activity_id = ?", activityID).Find(&contacts).Error; err != nil {
		return nil, err
	}
	ids := make([]uint, len(contacts))
	for i, c := range contacts {
		ids[i] = c.UserID
	}
	return ids, nil
}

// ---- 执行记录 ----
func (r *repository) CreateExecution(ctx context.Context, exec *model.ExecutionRecord) error {
	return r.db.WithContext(ctx).Create(exec).Error
}

func (r *repository) FindExecutions(ctx context.Context, activityID uint) ([]model.ExecutionRecord, error) {
	var records []model.ExecutionRecord
	err := r.db.WithContext(ctx).Where("activity_id = ?", activityID).Find(&records).Error
	return records, err
}

func (r *repository) SumExecutions(ctx context.Context, activityID uint) (int, error) {
	var total int64
	err := r.db.WithContext(ctx).Model(&model.ExecutionRecord{}).
		Where("activity_id = ?", activityID).
		Select("COALESCE(SUM(count), 0)").Scan(&total).Error
	return int(total), err
}
