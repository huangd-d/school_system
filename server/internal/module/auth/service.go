package auth

import (
	"school-system/internal/config"
	"school-system/internal/model"
	"school-system/pkg/apperror"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Service 认证业务逻辑
type Service struct {
	db  *gorm.DB
	cfg *config.JWTConfig
}

// NewService 创建 Service
func NewService(db *gorm.DB, cfg *config.JWTConfig) *Service {
	return &Service{db: db, cfg: cfg}
}

// Login 用户名密码登录，返回 JWT Token
func (s *Service) Login(username, password string) (token string, user *model.User, err error) {
	// 查询用户
	if err := s.db.Where("username = ?", username).First(&user).Error; err != nil {
		return "", nil, apperror.ErrUserLoginFailed
	}

	// 检查账户状态
	if user.Status == model.UserStatusDisabled {
		return "", nil, apperror.ErrUserDisabled
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", nil, apperror.ErrUserLoginFailed
	}

	// 签发 JWT
	token, err = s.signToken(user)
	if err != nil {
		return "", nil, apperror.Newf(apperror.ErrInternal.Code, "签发令牌失败: %v", err)
	}

	return token, user, nil
}

// Refresh 刷新 Token（接受已过期的 Token，只要签名有效）
func (s *Service) Refresh(oldToken string) (newToken string, err error) {
	// 解析旧 Token（跳过过期校验）
	claims := jwt.MapClaims{}
	parser := jwt.NewParser()
	_, _, err = parser.ParseUnverified(oldToken, claims)
	if err != nil {
		return "", apperror.ErrUnauthorized
	}

	// 检查是否在允许的刷新窗口内（过期后 7 天内）
	if expClaim, ok := claims["exp"]; ok {
		if expFloat, ok := expClaim.(float64); ok {
			expTime := time.Unix(int64(expFloat), 0)
			if time.Since(expTime) > 7*24*time.Hour {
				return "", apperror.ErrUnauthorized
			}
		}
	}

	// 签发新 Token
	newToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(s.cfg.Secret))
	if err != nil {
		return "", apperror.Newf(apperror.ErrInternal.Code, "签发令牌失败: %v", err)
	}

	return newToken, nil
}

// signToken 签发 JWT
func (s *Service) signToken(user *model.User) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"user_id":   user.ID,
		"campus_id": user.CampusID,
		"role":      user.Role,
		"iat":       now.Unix(),
		"exp":       now.Add(time.Duration(s.cfg.ExpireHour) * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.Secret))
}
