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

var _ = Describe("Repository Delete", func() {
    var (
        baseURL    string
        aliceToken string
        bobUser    User
        bobToken   string
        projectID  U64
        repoID     U64
        suf        int64
    )

    BeforeEach(func() {
        baseURL = e2ehttp.BaseURL()
        suf = time.Now().UnixNano()

		// 注册与登录用户
		aliceName := fmt.Sprintf("alice_%d", suf)
		bobName := fmt.Sprintf("bob_%d", suf)

		_, _ = registerUser(aliceName, fmt.Sprintf("%s@example.com", aliceName))
		bobUser, _ = registerUser(bobName, fmt.Sprintf("%s@example.com", bobName))

    _, aliceToken = loginUser(aliceName)
		bobUser, bobToken = loginUser(bobName)

		// 创建项目
    status, body := DoRequest(baseURL, MethodPost, "/project_service/api/v1/projects", map[string]any{
        "name":        fmt.Sprintf("proj_%d", suf),
        "description": "code-repo e2e project",
        "language":    "go",
        "framework":   "grpc-gateway",
        "is_public":   false,
    }, map[string]string{"Authorization": "Bearer " + aliceToken})
    Expect(status).To(Equal(StatusOK), "CreateProject should succeed, body: %s", string(body))

		var cp CreateProjectResp
		Expect(json.Unmarshal(body, &cp)).To(Succeed(), "CreateProject response should be valid JSON")
		projectID = cp.Project.ID

		// 添加 bob 为 ADMIN
    status, body = DoRequest(baseURL, MethodPost, fmt.Sprintf("/project_service/api/v1/projects/%d/members", projectID), map[string]any{
        "user_id": bobUser.ID,
        "role":    2,
    }, map[string]string{"Authorization": "Bearer " + aliceToken})
    Expect(status).To(Equal(StatusOK), "AddMember(admin) should succeed, body: %s", string(body))

		// 同步权限，使管理员角色立即生效
    status, body = DoRequest(baseURL, MethodPost, fmt.Sprintf("/project_service/api/v1/projects/%d/sync-permissions", projectID), map[string]any{
        "project_id": projectID,
    }, map[string]string{"Authorization": "Bearer " + aliceToken})
    Expect(status).To(Equal(StatusOK), "SyncUserPermissions should succeed, body: %s", string(body))

		// 创建代码仓库
    status, body = DoRequest(baseURL, MethodPost, "/code_repository_service/api/v1/repositories", map[string]any{
        "project_id":  projectID,
        "name":        "repo_delete",
        "description": "repo for delete",
        "git_url":     "https://example.invalid/repo.git",
        "branch":      "main",
        "auth_type":   "REPOSITORY_AUTH_TYPE_NONE",
    }, map[string]string{"Authorization": "Bearer " + aliceToken})
    Expect(status).To(Equal(StatusOK), "CreateRepository should succeed, body: %s", string(body))

		var cr CreateRepositoryResp
		Expect(json.Unmarshal(body, &cr)).To(Succeed(), "CreateRepository response should be valid JSON")
		repoID = cr.Repository.ID
	})

	Context("when deleting a repository", func() {
		It("should delete repository successfully and return 404 on subsequent get", func() {
			// 删除仓库
            status, body := DoRequest(baseURL, MethodDelete, fmt.Sprintf("/code_repository_service/api/v1/repositories/%d?project_id=%d", repoID, projectID), nil, map[string]string{"Authorization": "Bearer " + bobToken})
            Expect(status).To(Equal(StatusOK), "Admin delete repo should succeed, body: %s", string(body))

			// 再次获取应返回 404
            status, body = DoRequest(baseURL, MethodGet, fmt.Sprintf("/code_repository_service/api/v1/repositories/%d?project_id=%d", repoID, projectID), nil, map[string]string{"Authorization": "Bearer " + aliceToken})
            Expect(status).To(Equal(StatusNotFound), "Get deleted repository should return 404, body: %s", string(body))
		})
	})
})