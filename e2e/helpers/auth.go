// User auth helpers: register and login for tests.

package helpers

import (
    "encoding/json"
    "fmt"
    "strconv"
)

// RegisterAndLogin 在用户服务中注册并登录，返回用户ID、用户名与token
// password: 使用的明文密码；如果为空则使用默认值
func RegisterAndLogin(username, email, password string) (uint64, string, string, error) {
    if password == "" { password = "testpassword123" }

    status, body := DoRequest(BaseURL(), MethodPost, "/user_service/api/v1/users/register", map[string]any{
        "username": username,
        "email":    email,
        "password": password,
    }, nil)
    if status != StatusOK {
        return 0, "", "", &HTTPError{Status: status, Body: string(body)}
    }
    var reg struct{ User struct{ ID uint64 `json:"id,string"`; Username string `json:"username"` } `json:"user"` }
    if err := json.Unmarshal(body, &reg); err != nil { return 0, "", "", err }

    status, body = DoRequest(BaseURL(), MethodPost, "/user_service/api/v1/users/login", map[string]any{
        "username": reg.User.Username,
        "password": password,
    }, nil)
    if status != StatusOK {
        return 0, "", "", &HTTPError{Status: status, Body: string(body)}
    }
    var lg struct{ User struct{ ID uint64 `json:"id,string"`; Username string `json:"username"` } `json:"user"`; Token string `json:"token"` }
    if err := json.Unmarshal(body, &lg); err != nil { return 0, "", "", err }
    return lg.User.ID, lg.User.Username, lg.Token, nil
}

// HTTPError 用于统一错误信息格式
type HTTPError struct {
    Status int
    Body   string
}

func (e *HTTPError) Error() string { return fmt.Sprintf("http status=%s body=%s", strconv.Itoa(e.Status), e.Body) }

// AdminLogin 登录内置的超级管理员（user1/user123），返回其 JWT token
func AdminLogin() (string, error) {
    status, body := DoRequest(BaseURL(), MethodPost, "/user_service/api/v1/users/login", map[string]any{
        "username": "user1",
        "password": "user123",
    }, nil)
    if status != StatusOK { return "", &HTTPError{Status: status, Body: string(body)} }
    var lg struct{ Token string `json:"token"` }
    if err := json.Unmarshal(body, &lg); err != nil { return "", err }
    return lg.Token, nil
}