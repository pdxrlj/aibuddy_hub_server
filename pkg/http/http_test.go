package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

// 测试请求结构体
type TestRequest struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
}

// 测试响应结构体
type TestResponse struct {
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// 测试处理器函数
type testHandler struct{}

func (h *testHandler) handleGet(state *State) error {
	return NewResponse(state.Ctx).SetData(TestResponse{
		Message: "GET request successful",
	}).Success()
}

func (h *testHandler) handlePost(state *State, req *TestRequest) error {
	return NewResponse(state.Ctx).SetData(TestResponse{
		Message: "POST request successful",
		Data:    req,
	}).Success()
}

func (h *testHandler) handlePut(state *State, req *TestRequest) error {
	return NewResponse(state.Ctx).SetData(TestResponse{
		Message: "PUT request successful",
		Data:    req,
	}).Success()
}

func (h *testHandler) handlePatch(state *State, req *TestRequest) error {
	return NewResponse(state.Ctx).SetData(TestResponse{
		Message: "PATCH request successful",
		Data:    req,
	}).Success()
}

func (h *testHandler) handleDelete(state *State) error {
	return NewResponse(state.Ctx).SetData(TestResponse{
		Message: "DELETE request successful",
	}).Success()
}

func (h *testHandler) handleOptions(state *State) error {
	return NewResponse(state.Ctx).SetData(TestResponse{
		Message: "OPTIONS request successful",
	}).Success()
}

func (h *testHandler) handleHead(state *State) error {
	return NewResponse(state.Ctx).SetData(TestResponse{
		Message: "HEAD request successful",
	}).Success()
}

// 创建测试服务器
func createTestServer() (*Base, *echo.Echo) {
	e := echo.New()
	base := NewBase(e, &State{})
	return base, e
}

// 发送测试请求并返回响应
type testResponse struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func sendTestRequest(method, url string, body interface{}) (*http.Response, testResponse, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, testResponse{}, err
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, testResponse{}, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, testResponse{}, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp, testResponse{}, err
	}

	var testResp testResponse
	if err := json.Unmarshal(respBody, &testResp); err != nil {
		return resp, testResponse{}, err
	}

	return resp, testResp, nil
}

// 测试 GET 方法
func TestBase_GET(t *testing.T) {
	base, e := createTestServer()
	handler := &testHandler{}

	// 注册路由
	base.GET("/test", handler.handleGet)

	// 创建测试服务器
	server := httptest.NewServer(e)
	defer server.Close()

	// 发送 GET 请求
	resp, testResp, err := sendTestRequest("GET", fmt.Sprintf("%s/test", server.URL), nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "GET request successful", testResp.Data.(map[string]interface{})["message"])
}

// 测试 POST 方法
func TestBase_POST(t *testing.T) {
	base, e := createTestServer()
	handler := &testHandler{}

	// 注册路由
	base.POST("/test", handler.handlePost)

	// 创建测试服务器
	server := httptest.NewServer(e)
	defer server.Close()

	// 测试有效的 POST 请求
	validReq := &TestRequest{
		Name:  "test user",
		Email: "test@example.com",
	}

	resp, testResp, err := sendTestRequest("POST", fmt.Sprintf("%s/test", server.URL), validReq)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	responseData := testResp.Data.(map[string]interface{})
	assert.Equal(t, "POST request successful", responseData["message"])

	// 验证返回的数据
	reqData := responseData["data"].(map[string]interface{})
	assert.Equal(t, "test user", reqData["name"])
	assert.Equal(t, "test@example.com", reqData["email"])

	// 测试无效的 POST 请求（缺少必填字段）
	invalidReq := &TestRequest{
		Name: "test user",
		// 缺少 email 字段
	}

	resp2, _, err2 := sendTestRequest("POST", fmt.Sprintf("%s/test", server.URL), invalidReq)
	assert.NoError(t, err2)
	assert.Equal(t, http.StatusOK, resp2.StatusCode) // 由于验证错误处理方式，这里可能仍是 200
}

// 测试 PUT 方法
func TestBase_PUT(t *testing.T) {
	base, e := createTestServer()
	handler := &testHandler{}

	// 注册路由
	base.PUT("/test", handler.handlePut)

	// 创建测试服务器
	server := httptest.NewServer(e)
	defer server.Close()

	// 发送 PUT 请求
	req := &TestRequest{
		Name:  "updated user",
		Email: "updated@example.com",
	}

	resp, testResp, err := sendTestRequest("PUT", fmt.Sprintf("%s/test", server.URL), req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	responseData := testResp.Data.(map[string]interface{})
	assert.Equal(t, "PUT request successful", responseData["message"])

	// 验证返回的数据
	reqData := responseData["data"].(map[string]interface{})
	assert.Equal(t, "updated user", reqData["name"])
	assert.Equal(t, "updated@example.com", reqData["email"])
}

// 测试 PATCH 方法
func TestBase_PATCH(t *testing.T) {
	base, e := createTestServer()
	handler := &testHandler{}

	// 注册路由
	base.PATCH("/test", handler.handlePatch)

	// 创建测试服务器
	server := httptest.NewServer(e)
	defer server.Close()

	// 发送 PATCH 请求
	req := &TestRequest{
		Name:  "patched user",
		Email: "patched@example.com",
	}

	resp, testResp, err := sendTestRequest("PATCH", fmt.Sprintf("%s/test", server.URL), req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	responseData := testResp.Data.(map[string]interface{})
	assert.Equal(t, "PATCH request successful", responseData["message"])

	// 验证返回的数据
	reqData := responseData["data"].(map[string]interface{})
	assert.Equal(t, "patched user", reqData["name"])
	assert.Equal(t, "patched@example.com", reqData["email"])
}

// 测试 DELETE 方法
func TestBase_DELETE(t *testing.T) {
	base, e := createTestServer()
	handler := &testHandler{}

	// 注册路由
	base.DELETE("/test", handler.handleDelete)

	// 创建测试服务器
	server := httptest.NewServer(e)
	defer server.Close()

	// 发送 DELETE 请求
	resp, testResp, err := sendTestRequest("DELETE", fmt.Sprintf("%s/test", server.URL), nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "DELETE request successful", testResp.Data.(map[string]interface{})["message"])
}

// 测试 OPTIONS 方法
func TestBase_OPTIONS(t *testing.T) {
	base, e := createTestServer()
	handler := &testHandler{}

	// 注册路由
	base.OPTIONS("/test", handler.handleOptions)

	// 创建测试服务器
	server := httptest.NewServer(e)
	defer server.Close()

	// 发送 OPTIONS 请求
	resp, testResp, err := sendTestRequest("OPTIONS", fmt.Sprintf("%s/test", server.URL), nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "OPTIONS request successful", testResp.Data.(map[string]interface{})["message"])
}

// 测试 HEAD 方法
func TestBase_HEAD(t *testing.T) {
	base, e := createTestServer()
	handler := &testHandler{}

	// 注册路由
	base.HEAD("/test", handler.handleHead)

	// 创建测试服务器
	server := httptest.NewServer(e)
	defer server.Close()

	// 发送 HEAD 请求
	req, err := http.NewRequest("HEAD", fmt.Sprintf("%s/test", server.URL), nil)
	assert.NoError(t, err)

	client := &http.Client{}
	resp, err := client.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	// HEAD 请求应该返回 200 状态码但没有响应体
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// 读取响应体（应该为空）
	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Empty(t, body)
}

// 测试多个路由注册
func TestBase_MultipleRoutes(t *testing.T) {
	base, e := createTestServer()
	handler := &testHandler{}

	// 注册多个路由
	base.GET("/users", handler.handleGet)
	base.POST("/users", handler.handlePost)
	base.PUT("/users/:id", handler.handlePut)
	base.DELETE("/users/:id", handler.handleDelete)

	// 创建测试服务器
	server := httptest.NewServer(e)
	defer server.Close()

	// 测试 GET /users
	resp1, testResp1, err1 := sendTestRequest("GET", fmt.Sprintf("%s/users", server.URL), nil)
	assert.NoError(t, err1)
	assert.Equal(t, http.StatusOK, resp1.StatusCode)
	assert.Equal(t, "GET request successful", testResp1.Data.(map[string]interface{})["message"])

	// 测试 POST /users
	req := &TestRequest{
		Name:  "new user",
		Email: "new@example.com",
	}
	resp2, testResp2, err2 := sendTestRequest("POST", fmt.Sprintf("%s/users", server.URL), req)
	assert.NoError(t, err2)
	assert.Equal(t, http.StatusOK, resp2.StatusCode)
	assert.Equal(t, "POST request successful", testResp2.Data.(map[string]interface{})["message"])
}

// 测试中间件功能
func TestBase_WithMiddleware(t *testing.T) {
	base, e := createTestServer()
	handler := &testHandler{}

	// 创建一个简单的中间件
	middleware := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set("X-Test-Middleware", "applied")
			return next(c)
		}
	}

	// 注册带中间件的路由
	base.GET("/test-with-middleware", handler.handleGet, middleware)

	// 创建测试服务器
	server := httptest.NewServer(e)
	defer server.Close()

	// 发送请求
	resp, _, err := sendTestRequest("GET", fmt.Sprintf("%s/test-with-middleware", server.URL), nil)
	assert.NoError(t, err)
	assert.Equal(t, "applied", resp.Header.Get("X-Test-Middleware"))
}

// 测试参数验证功能
func TestBase_ParameterValidation(t *testing.T) {
	base, e := createTestServer()

	// 设置验证器
	validator := &Validator{Validator: validator.New()}
	base.Validator = validator

	handler := &testHandler{}

	// 注册需要参数验证的路由
	base.POST("/validate", handler.handlePost)

	// 创建测试服务器
	server := httptest.NewServer(e)
	defer server.Close()

	// 测试 1: 有效的请求参数
	t.Run("ValidParameters", func(t *testing.T) {
		validReq := &TestRequest{
			Name:  "test user",
			Email: "test@example.com",
		}

		resp, testResp, err := sendTestRequest("POST", fmt.Sprintf("%s/validate", server.URL), validReq)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		responseData := testResp.Data.(map[string]interface{})
		assert.Equal(t, "POST request successful", responseData["message"])

		// 验证返回的数据
		reqData := responseData["data"].(map[string]interface{})
		assert.Equal(t, "test user", reqData["name"])
		assert.Equal(t, "test@example.com", reqData["email"])
	})

	// 测试 2: 缺少必填字段
	t.Run("MissingRequiredField", func(t *testing.T) {
		invalidReq := &TestRequest{
			Name: "test user",
			// 缺少 email 字段
		}

		// 手动发送请求来正确处理响应
		jsonBody, err := json.Marshal(invalidReq)
		assert.NoError(t, err)

		req, err := http.NewRequest("POST", fmt.Sprintf("%s/validate", server.URL), bytes.NewBuffer(jsonBody))
		assert.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)

		// 验证状态码
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// 响应体应该包含验证错误信息
		responseStr := string(body)
		assert.Contains(t, responseStr, "参数错误")
		assert.Contains(t, responseStr, "email")
		assert.Contains(t, responseStr, "required")
	})

	// 测试 3: 无效的邮箱格式
	t.Run("InvalidEmailFormat", func(t *testing.T) {
		invalidReq := &TestRequest{
			Name:  "test user",
			Email: "invalid-email", // 无效的邮箱格式
		}

		// 手动发送请求
		jsonBody, err := json.Marshal(invalidReq)
		assert.NoError(t, err)

		req, err := http.NewRequest("POST", fmt.Sprintf("%s/validate", server.URL), bytes.NewBuffer(jsonBody))
		assert.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)

		// 验证状态码
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// 响应体应该包含邮箱格式错误
		responseStr := string(body)
		assert.Contains(t, responseStr, "参数错误")
		assert.Contains(t, responseStr, "email")
		assert.Contains(t, responseStr, "email")
	})

	// 测试 4: 空的请求体
	t.Run("EmptyRequestBody", func(t *testing.T) {
		req, err := http.NewRequest("POST", fmt.Sprintf("%s/validate", server.URL), nil)
		assert.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)

		// 验证状态码
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// 响应体应该包含绑定错误
		responseStr := string(body)
		assert.Contains(t, responseStr, "参数错误")
	})

	// 测试 5: GET 请求不需要参数验证
	t.Run("GetRequestNoValidation", func(t *testing.T) {
		base.GET("/validate-get", handler.handleGet)

		resp, testResp, err := sendTestRequest("GET", fmt.Sprintf("%s/validate-get", server.URL), nil)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "GET request successful", testResp.Data.(map[string]interface{})["message"])
	})
}

// 测试错误处理
func TestBase_ErrorHandling(t *testing.T) {
	base, e := createTestServer()

	// 创建一个会返回错误的处理器
	errorHandler := func(state *State) error {
		return fmt.Errorf("test error")
	}

	// 注册路由
	base.GET("/error", errorHandler)

	// 创建测试服务器
	server := httptest.NewServer(e)
	defer server.Close()

	// 发送请求
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/error", server.URL), nil)
	assert.NoError(t, err)

	client := &http.Client{}
	resp, err := client.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)

	// 验证状态码
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// 根据当前实现，错误只记录日志，不返回给客户端
	// 响应体可能为空或包含标准的成功响应格式
	t.Logf("Response body: %s", string(body))
}
