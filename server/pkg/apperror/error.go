package apperror

import "fmt"

// AppError 应用错误
type AppError struct {
	Code    int    // 业务错误码
	Message string // 错误描述
}

func (e *AppError) Error() string {
	return e.Message
}

// New 创建应用错误
func New(code int, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

// Newf 格式化创建
func Newf(code int, format string, args ...interface{}) *AppError {
	return &AppError{Code: code, Message: fmt.Sprintf(format, args...)}
}
