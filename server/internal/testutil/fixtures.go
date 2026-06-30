package testutil

import (
	"testing"

	"school-system/internal/model"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// CreateTestCampus 插入一个校区并返回。
func CreateTestCampus(t *testing.T, db *gorm.DB, name, campusType string) *model.Campus {
	t.Helper()
	c := &model.Campus{Name: name, Type: campusType}
	require.NoError(t, db.Create(c).Error, "创建测试校区失败")
	return c
}

// CreateTestUser 插入一个用户（密码哈希使用 MinCost 以加快测试速度）并返回。
func CreateTestUser(t *testing.T, db *gorm.DB, username, role string, campusID uint) *model.User {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.MinCost)
	require.NoError(t, err, "bcrypt 哈希失败")
	u := &model.User{
		Username:     username,
		PasswordHash: string(hash),
		Phone:        "13800138000",
		Role:         role,
		CampusID:     campusID,
		Status:       model.UserStatusActive,
	}
	require.NoError(t, db.Create(u).Error, "创建测试用户失败")
	return u
}

// FastHash 使用 MinCost 生成 bcrypt 哈希（仅用于测试夹具，速度 ~5ms，比 DefaultCost 快约 20 倍）。
func FastHash(password string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		panic("FastHash: " + err.Error())
	}
	return string(hash)
}
