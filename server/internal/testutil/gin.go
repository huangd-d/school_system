package testutil

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
)

// NewTestContext 创建一个 Gin 测试上下文。
//
// 用法：
//
//	c, w := testutil.NewTestContext("POST", "/api/v1/users", reqBody)
//	// 可选：设置 auth 中间件注入的上下文值
//	c.Set("user_id", uint(1))
//	h.Create(c)
//	assert.Equal(t, 200, w.Code)
func NewTestContext(method, path string, body interface{}) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	var bodyReader *bytes.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		bodyReader = bytes.NewReader(b)
	} else {
		bodyReader = bytes.NewReader(nil)
	}

	c.Request = httptest.NewRequest(method, path, bodyReader)
	c.Request.Header.Set("Content-Type", "application/json")

	return c, w
}

// ParseResponse 从 httptest.ResponseRecorder 解析出统一响应结构。
func ParseResponse(w *httptest.ResponseRecorder) (code int, message string, data json.RawMessage) {
	var resp struct {
		Code    int             `json:"code"`
		Message string          `json:"message"`
		Data    json.RawMessage `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	return resp.Code, resp.Message, resp.Data
}
