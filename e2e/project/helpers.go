//go:build e2e
// +build e2e

package project_e2e

import (
	"encoding/json"
	"fmt"
	"strconv"
	"xcoding/e2e/helpers"
)

// BaseURL（支持环境变量覆盖）
func baseURL() string { return helpers.GetBaseURLOrDefault("http://xcoding.local:31080") }

// 简化请求封装
func do(method, endpoint string, body any, headers map[string]string) (int, []byte) {
	return helpers.DoRequest(baseURL(), method, endpoint, body, headers)
}

// ---- 轻量模型 & 助手（从 apps/project/e2e 迁移）----

// Flexible numeric type: accepts JSON number or string
type U64 uint64

func (u *U64) UnmarshalJSON(b []byte) error {
	if len(b) == 0 {
		*u = 0
		return nil
	}
	if b[0] == '"' {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		if s == "" {
			*u = 0
			return nil
		}
		v, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return err
		}
		*u = U64(v)
		return nil
	}
	var v uint64
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	*u = U64(v)
	return nil
}

// Flexible string type: accepts JSON string or number, stores as string
type StringOrInt string

func (s *StringOrInt) UnmarshalJSON(b []byte) error {
	if len(b) == 0 {
		*s = ""
		return nil
	}
	if b[0] == '"' {
		var v string
		if err := json.Unmarshal(b, &v); err != nil {
			return err
		}
		*s = StringOrInt(v)
		return nil
	}
	var iv int
	if err := json.Unmarshal(b, &iv); err == nil {
		*s = StringOrInt(strconv.Itoa(iv))
		return nil
	}
	var uv uint64
	if err := json.Unmarshal(b, &uv); err == nil {
		*s = StringOrInt(strconv.FormatUint(uv, 10))
		return nil
	}
	return fmt.Errorf("invalid StringOrInt")
}

type User struct {
	ID       U64    `json:"id"`
	Username string `json:"username"`
}

type Project struct {
	ID          U64    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type ProjectInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Language    string `json:"language"`
	Framework   string `json:"framework"`
	IsPublic    bool   `json:"is_public"`
}

type ProjectMember struct {
	UserID    U64         `json:"user_id"`
	ProjectID U64         `json:"project_id"`
	Role      StringOrInt `json:"role"`
	User      *User       `json:"user"`
}

type Pagination struct {
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	TotalItems int `json:"total_items"`
	TotalPages int `json:"total_pages"`
}

type CreateProjectWithUserResp struct {
	User    User    `json:"user"`
	Project Project `json:"project"`
}

type CreateProjectResp struct {
	Project Project `json:"project"`
}
type GetProjectResp struct {
	Project Project `json:"project"`
}
type ListProjectsResp struct {
	Data       []Project  `json:"data"`
	Pagination Pagination `json:"pagination"`
}
type AddProjectMemberResp struct {
	Member ProjectMember `json:"member"`
}
type UpdateProjectMemberResp struct {
	Member ProjectMember `json:"member"`
}
type ListProjectMembersResp struct {
	Data       []ProjectMember `json:"data"`
	Pagination Pagination      `json:"pagination"`
}
type SyncUserPermissionsResp struct {
	Members []ProjectMember `json:"members"`
}
type UpdateProjectResp struct {
	Project Project `json:"project"`
}
type DeleteProjectResp struct {
	Success bool `json:"success"`
}
type RegisterResp struct {
	User  User   `json:"user"`
	Token string `json:"token"`
}
type LoginResp struct {
	User  User   `json:"user"`
	Token string `json:"token"`
}

// 用户注册/登录（不做路由预检与跳过）
func registerUser(username, email, password string) (User, string, error) {
	status, body := helpers.DoRequest(baseURL(), helpers.MethodPost, "/user_service/api/v1/users/register", map[string]any{
		"username": username,
		"email":    email,
		"password": password,
	}, nil)
	if status != helpers.StatusOK {
		return User{}, "", fmt.Errorf("register status=%d body=%s", status, string(body))
	}
	var reg RegisterResp
	if err := json.Unmarshal(body, &reg); err != nil {
		return User{}, "", fmt.Errorf("register json error: %v body=%s", err, string(body))
	}
	return reg.User, reg.Token, nil
}

func loginUser(username, password string) (User, string, error) {
	status, body := helpers.DoRequest(baseURL(), helpers.MethodPost, "/user_service/api/v1/users/login", map[string]any{
		"username": username,
		"password": password,
	}, nil)
	if status != helpers.StatusOK {
		return User{}, "", fmt.Errorf("login status=%d body=%s", status, string(body))
	}
	var lg LoginResp
	if err := json.Unmarshal(body, &lg); err != nil {
		return User{}, "", fmt.Errorf("login json error: %v body=%s", err, string(body))
	}
	return lg.User, lg.Token, nil
}

func roleToNum(role string) int {
	switch role {
	case "PROJECT_MEMBER_ROLE_OWNER", "OWNER":
		return 1
	case "PROJECT_MEMBER_ROLE_ADMIN", "ADMIN":
		return 2
	case "PROJECT_MEMBER_ROLE_MEMBER", "MEMBER":
		return 3
	case "PROJECT_MEMBER_ROLE_GUEST", "GUEST":
		return 4
	default:
		var x int
		_, _ = fmt.Sscanf(role, "%d", &x)
		return x
	}
}
