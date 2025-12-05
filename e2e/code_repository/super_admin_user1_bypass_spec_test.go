//go:build e2e
// +build e2e

package code_repository_test

import (
    "encoding/json"
    "fmt"
    "time"

    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
    e2ehttp "xcoding/e2e/helpers"
)

var _ = Describe("Super Admin User1 Bypass", func() {
	var (
		baseURL   string
		user1Token string
		projectID U64
		suf       int64
	)

    BeforeEach(func() {
        baseURL = e2ehttp.BaseURL()
        suf = time.Now().UnixNano()

		// 登录 user1 (超级管理员)
    status, body := DoRequest(baseURL, MethodPost, "/user_service/api/v1/users/login", map[string]any{
        "username": "user1",
        "password": "user123",
    }, nil)
    Expect(status).To(Equal(StatusOK), "login user1 should succeed, body: %s", string(body))

		var lg LoginResp
		Expect(json.Unmarshal(body, &lg)).To(Succeed(), "login response should be valid JSON")
		Expect(lg.Token).NotTo(BeEmpty(), "login token should not be empty")
		user1Token = lg.Token

		// 认证，确认 X-User-Role 为超级管理员
    status, body, respHeaders := DoRequestWithHeaders(baseURL, MethodGet, "/user_service/api/v1/auth", nil, map[string]string{"Authorization": "Bearer " + user1Token})
    Expect(status).To(Equal(StatusOK), "auth should succeed, body: %s", string(body))

		role := respHeaders["X-User-Role"]
		Expect(role).To(SatisfyAny(
			Equal("USER_ROLE_SUPER_ADMIN"),
			Equal("SUPER_ADMIN"),
			Equal("super_admin"),
		), "X-User-Role should indicate super admin")

		// 创建一个项目由非 user1 的普通用户拥有，确保 user1 不是成员
		ownerName := fmt.Sprintf("owner_%d", suf)
		_, _ = registerUser(ownerName, fmt.Sprintf("%s@example.com", ownerName))
		_, ownerToken := loginUser(ownerName)

    status, body = DoRequest(baseURL, MethodPost, "/project_service/api/v1/projects", map[string]any{
        "name":        fmt.Sprintf("proj_%d", suf),
        "description": "superadmin bypass project",
        "language":    "go",
        "framework":   "grpc-gateway",
        "is_public":   false,
    }, map[string]string{"Authorization": "Bearer " + ownerToken})
    Expect(status).To(Equal(StatusOK), "CreateProject should succeed, body: %s", string(body))

		var cp CreateProjectResp
		Expect(json.Unmarshal(body, &cp)).To(Succeed(), "CreateProject response should be valid JSON")
		projectID = cp.Project.ID
	})

	Context("when super admin accesses project resources", func() {
		It("should bypass permission checks and list repositories", func() {
			endpoint := fmt.Sprintf("/code_repository_service/api/v1/repositories?project_id=%d&page=1&page_size=10", projectID)
            status, body := DoRequest(baseURL, MethodGet, endpoint, nil, map[string]string{"Authorization": "Bearer " + user1Token})
            Expect(status).To(Equal(StatusOK), "SuperAdmin list repositories should succeed, body: %s", string(body))

			// 校验响应结构（允许为空集）
			var list ListRepositoriesResp
			Expect(json.Unmarshal(body, &list)).To(Succeed(), "ListRepositories response should be valid JSON")
		})
	})
})