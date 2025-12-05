// Code Repository E2E helpers: shared constants and request wrappers.

package code_repository_test

import (
    "encoding/json"
    "fmt"

    . "github.com/onsi/ginkgo/v2"
    "xcoding/e2e/helpers"
)

// ---- 方法与状态码常量 ----
const (
    MethodGet    = helpers.MethodGet
    MethodPost   = helpers.MethodPost
    MethodPut    = helpers.MethodPut
    MethodDelete = helpers.MethodDelete

    StatusOK                  = helpers.StatusOK
    StatusCreated             = helpers.StatusCreated
    StatusNoContent           = helpers.StatusNoContent
    StatusBadRequest          = helpers.StatusBadRequest
    StatusUnauthorized        = helpers.StatusUnauthorized
    StatusForbidden           = helpers.StatusForbidden
    StatusNotFound            = helpers.StatusNotFound
    StatusMethodNotAllowed    = helpers.StatusMethodNotAllowed
    StatusConflict            = helpers.StatusConflict
    StatusInternalServerError = helpers.StatusInternalServerError
    StatusBadGateway          = helpers.StatusBadGateway
    StatusServiceUnavailable  = helpers.StatusServiceUnavailable
    StatusGatewayTimeout      = helpers.StatusGatewayTimeout
)

// ---- HTTP 请求助手 ----
// DoRequest 统一的 HTTP 请求助手（禁用重定向，默认 JSON/UA）
func DoRequest(baseURL, method, endpoint string, body any, headers map[string]string) (int, []byte) {
    return helpers.DoRequest(baseURL, method, endpoint, body, headers)
}

// DoRequestWithHeaders 返回响应头的版本
func DoRequestWithHeaders(baseURL, method, endpoint string, body any, headers map[string]string) (int, []byte, map[string]string) {
    return helpers.DoRequestWithHeaders(baseURL, method, endpoint, body, headers)
}

// 统一 BaseURL（直接使用共享 helpers.BaseURL）
// 移除多余包装，测试中直接调用 helpers.BaseURL()
// NOTE: 保留此函数可能导致重复定义，故删除并在测试中替换调用

// ---- 自定义类型 ----
// U64 支持字符串或数字的 uint64 类型
type U64 = helpers.FlexID

// ---- 数据模型 ----
type User struct {
	ID       U64    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

type RegisterResp struct {
	User User `json:"user"`
}

type LoginResp struct {
	User  User   `json:"user"`
	Token string `json:"token"`
}

type Project struct {
	ID          U64    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Language    string `json:"language"`
	Framework   string `json:"framework"`
	IsPublic    bool   `json:"is_public"`
}

type Pagination struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
	Total    int `json:"total"`
}

type Repository struct {
	ID          U64    `json:"id"`
	ProjectID   U64    `json:"project_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	GitURL      string `json:"git_url"`
	Branch      string `json:"branch"`
	AuthType    string `json:"auth_type"`
}

type CreateProjectResp struct {
	Project Project `json:"project"`
}

type AddMemberResp struct {
	Success bool `json:"success"`
}

type CreateRepositoryResp struct {
	Repository Repository `json:"repository"`
}

type GetRepositoryResp struct {
	Repository Repository `json:"repository"`
}

type UpdateRepositoryResp struct {
	Repository Repository `json:"repository"`
}

type ListRepositoriesResp struct {
	Data       []Repository `json:"data"`
	Pagination Pagination   `json:"pagination"`
}

type TestRepositoryConnectionResp struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type GetRepositoryBranchesResp struct {
	Data       []string   `json:"data"`
	Pagination Pagination `json:"pagination"`
}

// ---- 助手函数 ----
func registerUser(username, email string) (User, string) {
    // 保持原有行为：仅注册，返回空 token（部分测试用为空 token 验证未授权）
    status, body := helpers.DoRequest(helpers.BaseURL(), MethodPost, "/user_service/api/v1/users/register", map[string]any{
        "username": username,
        "email":    email,
        "password": "password123",
    }, nil)
    if status != StatusOK { Fail(fmt.Sprintf("RegisterUser failed: %d %s", status, string(body))) }
    var resp RegisterResp
    if err := json.Unmarshal(body, &resp); err != nil { Fail(fmt.Sprintf("RegisterUser json error: %v body=%s", err, string(body))) }
    return resp.User, ""
}

func loginUser(username string) (User, string) {
    status, body := helpers.DoRequest(helpers.BaseURL(), MethodPost, "/user_service/api/v1/users/login", map[string]any{
        "username": username,
        "password": "password123",
    }, nil)
    if status != StatusOK { Fail(fmt.Sprintf("LoginUser failed: %d %s", status, string(body))) }
    var resp LoginResp
    if err := json.Unmarshal(body, &resp); err != nil { Fail(fmt.Sprintf("LoginUser json error: %v body=%s", err, string(body))) }
    return resp.User, resp.Token
}

// 统一 registerAndLogin 包装（供未来测试使用）
func registerAndLogin(username, email string) (User, string) {
    uid, uname, token, err := helpers.RegisterAndLogin(username, email, "password123")
    if err != nil { Fail(fmt.Sprintf("RegisterAndLogin failed: %v", err)) }
    return User{ID: U64(uid), Username: uname, Email: email}, token
}
