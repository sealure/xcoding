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

var _ = Describe("Repository Update", func() {
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

		// 添加 bob 为 MEMBER
    status, body = DoRequest(baseURL, MethodPost, fmt.Sprintf("/project_service/api/v1/projects/%d/members", projectID), map[string]any{
        "project_id": projectID,
        "user_id":    bobUser.ID,
        "role":       1,
    }, map[string]string{"Authorization": "Bearer " + aliceToken})
    Expect(status).To(Equal(StatusOK), "AddMember should succeed, body: %s", string(body))

		// 创建代码仓库
    status, body = DoRequest(baseURL, MethodPost, "/code_repository_service/api/v1/repositories", map[string]any{
        "project_id":  projectID,
        "name":        "repo_update",
        "description": "repo for update",
        "git_url":     "https://example.invalid/repo.git",
        "branch":      "main",
        "auth_type":   "REPOSITORY_AUTH_TYPE_NONE",
    }, map[string]string{"Authorization": "Bearer " + aliceToken})
    Expect(status).To(Equal(StatusOK), "CreateRepository should succeed, body: %s", string(body))

		var cr CreateRepositoryResp
		Expect(json.Unmarshal(body, &cr)).To(Succeed(), "CreateRepository response should be valid JSON")
		repoID = cr.Repository.ID
	})

	Context("when promoting member to admin", func() {
		It("should allow admin to update repository", func() {
			// 提升 bob 为 ADMIN（失败则 remove+add）
            status, body := DoRequest(baseURL, MethodPut, fmt.Sprintf("/project_service/api/v1/projects/%d/members/%d", projectID, bobUser.ID), map[string]any{
                "role": 2,
            }, map[string]string{"Authorization": "Bearer " + aliceToken})

            if status != StatusOK {
                // 如果直接更新失败，尝试删除后重新添加
                statusDel, bodyDel := DoRequest(baseURL, MethodDelete, fmt.Sprintf("/project_service/api/v1/projects/%d/members/%d", projectID, bobUser.ID), nil, map[string]string{"Authorization": "Bearer " + aliceToken})
                Expect(statusDel).To(Equal(StatusOK), "RemoveMember should succeed, body: %s", string(bodyDel))

                statusAdd, bodyAdd := DoRequest(baseURL, MethodPost, fmt.Sprintf("/project_service/api/v1/projects/%d/members", projectID), map[string]any{
                    "user_id": bobUser.ID,
                    "role":    2,
                }, map[string]string{"Authorization": "Bearer " + aliceToken})
                Expect(statusAdd).To(Equal(StatusOK), "Fallback add member should succeed, body: %s", string(bodyAdd))
            }

			// 同步用户权限，确保服务侧角色缓存刷新
            status, body = DoRequest(baseURL, MethodPost, fmt.Sprintf("/project_service/api/v1/projects/%d/sync-permissions", projectID), map[string]any{
                "project_id": projectID,
            }, map[string]string{"Authorization": "Bearer " + aliceToken})
            Expect(status).To(Equal(StatusOK), "SyncUserPermissions should succeed, body: %s", string(body))

			// 管理员更新仓库
            status, body = DoRequest(baseURL, MethodPut, fmt.Sprintf("/code_repository_service/api/v1/repositories/%d", repoID), map[string]any{
                "project_id":  projectID,
                "description": "updated-by-admin",
                "is_active":   true,
            }, map[string]string{"Authorization": "Bearer " + bobToken})
            Expect(status).To(Equal(StatusOK), "Admin update should succeed, body: %s", string(body))

			var ur UpdateRepositoryResp
			Expect(json.Unmarshal(body, &ur)).To(Succeed(), "UpdateRepository response should be valid JSON")
			Expect(ur.Repository.Description).To(Equal("updated-by-admin"), "UpdateRepository should persist description")
		})
	})
})