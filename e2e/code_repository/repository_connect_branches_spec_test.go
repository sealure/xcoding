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

var _ = Describe("Repository Connect and Branches", func() {
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

		// 在进行仓库连接测试前同步权限
    status, body = DoRequest(baseURL, MethodPost, fmt.Sprintf("/project_service/api/v1/projects/%d/sync-permissions", projectID), map[string]any{
        "project_id": projectID,
    }, map[string]string{"Authorization": "Bearer " + aliceToken})
    Expect(status).To(Equal(StatusOK), "SyncUserPermissions should succeed, body: %s", string(body))

		// 创建代码仓库
    status, body = DoRequest(baseURL, MethodPost, "/code_repository_service/api/v1/repositories", map[string]any{
        "project_id":  projectID,
        "name":        "repo_connect",
        "description": "repo for connection",
        "git_url":     "https://example.invalid/repo.git",
        "branch":      "main",
        "auth_type":   "REPOSITORY_AUTH_TYPE_NONE",
    }, map[string]string{"Authorization": "Bearer " + aliceToken})
    Expect(status).To(Equal(StatusOK), "CreateRepository should succeed, body: %s", string(body))

		var cr CreateRepositoryResp
		Expect(json.Unmarshal(body, &cr)).To(Succeed(), "CreateRepository response should be valid JSON")
		repoID = cr.Repository.ID
	})

	Context("when testing repository connection", func() {
		It("should test repository connection successfully", func() {
            status, body := DoRequest(baseURL, MethodPost, fmt.Sprintf("/code_repository_service/api/v1/repositories/%d/test", repoID), map[string]any{
                "project_id": projectID,
            }, map[string]string{"Authorization": "Bearer " + bobToken})
            Expect(status).To(Equal(StatusOK), "TestRepositoryConnection should succeed, body: %s", string(body))

			var tr TestRepositoryConnectionResp
			Expect(json.Unmarshal(body, &tr)).To(Succeed(), "TestRepositoryConnection response should be valid JSON")
			Expect(tr.Success || tr.Message != "").To(BeTrue(), "TestRepositoryConnection should return either success or message")
		})
	})

	Context("when getting repository branches", func() {
		It("should get branches list with main branch", func() {
            status, body := DoRequest(baseURL, MethodGet, fmt.Sprintf("/code_repository_service/api/v1/repositories/%d/branches?page=1&page_size=10&project_id=%d", repoID, projectID), nil, map[string]string{"Authorization": "Bearer " + bobToken})
            Expect(status).To(Equal(StatusOK), "GetRepositoryBranches should succeed, body: %s", string(body))

			var br GetRepositoryBranchesResp
			Expect(json.Unmarshal(body, &br)).To(Succeed(), "GetRepositoryBranches response should be valid JSON")
			Expect(len(br.Data)).To(BeNumerically(">=", 1), "Branches list should contain at least default branch")

			foundMain := false
			for _, b := range br.Data {
				if b == "main" {
					foundMain = true
					break
				}
			}
			Expect(foundMain).To(BeTrue(), "Default branch 'main' should be present in branches")
			Expect(br.Pagination.Page).To(Equal(1), "Pagination page should be 1")
			Expect(br.Pagination.PageSize).To(SatisfyAny(Equal(10), Equal(0)), "Pagination page size should be 10 or 0")
		})
	})
})