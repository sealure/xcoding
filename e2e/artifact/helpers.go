//go:build e2e
// +build e2e

package artifact_e2e

import (
    "fmt"
    "strconv"
    "time"

    "xcoding/e2e/helpers"
)

// Base URL（统一使用共享 BaseURL）
func baseURL() string { return helpers.BaseURL() }

// Simple request wrappers
func do(method, endpoint string, body any, headers map[string]string) (int, []byte) {
    return helpers.DoRequest(baseURL(), method, endpoint, body, headers)
}

func doWithHeaders(method, endpoint string, body any, headers map[string]string) (int, []byte, map[string]string) {
    return helpers.DoRequestWithHeaders(baseURL(), method, endpoint, body, headers)
}

// Utils
func uniqueName() string { return fmt.Sprintf("goe2e_%d", time.Now().UnixNano()) }
func uniqueEmail() string { return fmt.Sprintf("%s@example.com", uniqueName()) }
func fmtUint(v uint64) string { return strconv.FormatUint(v, 10) }

// Common user flow: register and login, return uid/username/jwt
type userRegisterResp struct { User struct{ ID uint64 `json:"id,string"`; Username string `json:"username"` } `json:"user"`; Token string `json:"token"` }
type userLoginResp struct { User struct{ ID uint64 `json:"id,string"`; Username string `json:"username"` } `json:"user"`; Token string `json:"token"` }

func registerAndLogin() (uint64, string, string, error) {
    // 直接使用共享 RegisterAndLogin，保持密码一致
    return helpers.RegisterAndLogin(uniqueName(), uniqueEmail(), "testpassword123")
}

// Login seeded super admin (user1/user123) and return JWT
func adminLogin() (string, error) { return helpers.AdminLogin() }