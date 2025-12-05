// User E2E helpers: base URL, request wrappers, and models.

package user_e2e

import (
    "fmt"
    "github.com/onsi/ginkgo/v2"
    "time"

    "xcoding/e2e/helpers"
)

// 统一 BaseURL
func baseURL() string { return helpers.BaseURL() }

// 简化请求封装
func do(method, endpoint string, body any, headers map[string]string) (int, []byte) {
    return helpers.DoRequest(baseURL(), method, endpoint, body, headers)
}

func doWithHeaders(method, endpoint string, body any, headers map[string]string) (int, []byte, map[string]string) {
    return helpers.DoRequestWithHeaders(baseURL(), method, endpoint, body, headers)
}

func route(method, endpoint string) bool { return helpers.RouteExistsWithMethod(baseURL(), method, endpoint) }
func ping() bool { return helpers.PingGateway(baseURL()) }

// 结构体与响应模型
type User struct {
    ID       helpers.FlexID `json:"id,string"`
    Username string `json:"username"`
    Email    string `json:"email"`
    Role     string `json:"role"`
    IsActive bool   `json:"is_active"`
}

type RegisterResp struct { User User `json:"user"`; Token string `json:"token"` }
type LoginResp struct { User User `json:"user"`; Token string `json:"token"` }
type UpdateUserResp struct { User User `json:"user"` }
type ListUsersResp struct {
    Data       []User `json:"data"`
    Pagination struct{
        Page int `json:"page"`
        PageSize int `json:"page_size"`
        TotalItems int `json:"total_items"`
        TotalPages int `json:"total_pages"`
    } `json:"pagination"`
}
type DeleteUserResp struct { Success bool `json:"success"` }

type APIToken struct {
    ID          helpers.FlexID   `json:"id,string"`
    Name        string   `json:"name"`
    Token       string   `json:"token"`
    ExpiresAt   string   `json:"expires_at"`
    Description string   `json:"description"`
    Scopes      []string `json:"scopes"`
    CreatedAt   string   `json:"created_at"`
}
type ListAPITokensResp struct { Tokens []APIToken `json:"tokens"` }

type AuthResp struct {
    Authenticated bool              `json:"authenticated"`
    User          *User             `json:"user"`
    Reason        string            `json:"reason"`
    Headers       map[string]string `json:"headers"`
    ExpiresAt     string            `json:"expires_at"`
}

// 工具函数
func uniqueName() string { return fmt.Sprintf("goe2e_%d", time.Now().UnixNano()) }
func uniqueEmail() string { return fmt.Sprintf("%s@example.com", uniqueName()) }
func fmtUint(v helpers.FlexID) string { return fmt.Sprintf("%d", v) }

// 注册并登录，返回 uid/username/token
func registerAndLogin() (helpers.FlexID, string, string, error) {
    uid, uname, token, err := helpers.RegisterAndLogin(uniqueName(), uniqueEmail(), "testpassword123")
    if err != nil { ginkgo.GinkgoWriter.Println(fmt.Sprintf("registerAndLogin error: %v", err)) }
    return helpers.FlexID(uid), uname, token, err
}