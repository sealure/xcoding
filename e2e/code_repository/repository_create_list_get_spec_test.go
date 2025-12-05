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

var _ = Describe("Repository Create List Get", func() {
	var (
		baseURL     string
		aliceToken  string
		bobUser     User
		bobToken    string
		charlieToken string
		projectID   U64
		repoID      U64
		suf         int64
	)

    BeforeEach(func() {
        baseURL = e2ehttp.BaseURL()
        suf = time.Now().UnixNano()

		// 注册与登录用户
		aliceName := fmt.Sprintf("alice_%d", suf)
		bobName := fmt.Sprintf("bob_%d", suf)
		charlieName := fmt.Sprintf("charlie_%d", suf)

		_, _ = registerUser(aliceName, fmt.Sprintf("%s@example.com", aliceName))
		bobUser, _ = registerUser(bobName, fmt.Sprintf("%s@example.com", bobName))
		_, charlieToken = registerUser(charlieName, fmt.Sprintf("%s@example.com", charlieName))

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
            "role":       3,
        }, map[string]string{"Authorization": "Bearer " + aliceToken})
        Expect(status).To(Equal(StatusOK), "AddMember should succeed, body: %s", string(body))

		// 同步权限
		status, body = DoRequest(baseURL, MethodPost, fmt.Sprintf("/project_service/api/v1/projects/%d/sync-permissions", projectID), map[string]any{
			"project_id": projectID,
		}, map[string]string{"Authorization": "Bearer " + aliceToken})
		Expect(status).To(Equal(StatusOK), "SyncUserPermissions should succeed, body: %s", string(body))
	})

	Context("when creating a repository", func() {
		It("should create repository successfully", func() {
			repoName := fmt.Sprintf("repo_%d", suf)
            status, body := DoRequest(baseURL, MethodPost, "/code_repository_service/api/v1/repositories", map[string]any{
                "project_id":  projectID,
                "name":        repoName,
                "description": "first repo",
                "git_url":     "https://example.invalid/repo.git",
                "branch":      "main",
                "auth_type":   "REPOSITORY_AUTH_TYPE_NONE",
            }, map[string]string{"Authorization": "Bearer " + aliceToken})
            Expect(status).To(Equal(StatusOK), "CreateRepository should succeed, body: %s", string(body))

			var cr CreateRepositoryResp
			Expect(json.Unmarshal(body, &cr)).To(Succeed(), "CreateRepository response should be valid JSON")
			repoID = cr.Repository.ID
			Expect(repoID).NotTo(BeZero(), "Repository ID should not be zero")
		})
	})

	Context("when listing repositories", func() {
		BeforeEach(func() {
			// 创建仓库用于测试
			repoName := fmt.Sprintf("repo_%d", suf)
            status, body := DoRequest(baseURL, MethodPost, "/code_repository_service/api/v1/repositories", map[string]any{
                "project_id":  projectID,
                "name":        repoName,
                "description": "first repo",
                "git_url":     "https://example.invalid/repo.git",
                "branch":      "main",
                "auth_type":   "REPOSITORY_AUTH_TYPE_NONE",
            }, map[string]string{"Authorization": "Bearer " + aliceToken})
            Expect(status).To(Equal(StatusOK))

			var cr CreateRepositoryResp
			Expect(json.Unmarshal(body, &cr)).To(Succeed())
			repoID = cr.Repository.ID
		})

		It("should list repositories for owner", func() {
            status, body := DoRequest(baseURL, MethodGet, fmt.Sprintf("/code_repository_service/api/v1/repositories?project_id=%d", projectID), nil, map[string]string{"Authorization": "Bearer " + aliceToken})
            Expect(status).To(Equal(StatusOK), "ListRepositories(owner) should succeed, body: %s", string(body))

			var lr ListRepositoriesResp
			Expect(json.Unmarshal(body, &lr)).To(Succeed(), "ListRepositories(owner) response should be valid JSON")

			found := false
			for _, r := range lr.Data {
				if r.ID == repoID {
					found = true
					break
				}
			}
			Expect(found).To(BeTrue(), "ListRepositories should include created repo")
		})

		It("should handle pagination defaults correctly", func() {
            status, body := DoRequest(baseURL, MethodGet, fmt.Sprintf("/code_repository_service/api/v1/repositories?page=0&page_size=0&project_id=%d", projectID), nil, map[string]string{"Authorization": "Bearer " + aliceToken})
            Expect(status).To(Equal(StatusOK), "ListRepositories(defaults) should succeed, body: %s", string(body))

			var lr ListRepositoriesResp
			Expect(json.Unmarshal(body, &lr)).To(Succeed(), "ListRepositories(defaults) response should be valid JSON")
			Expect(lr.Pagination.Page).To(Equal(1), "Default page should be 1")
			Expect(lr.Pagination.PageSize).To(SatisfyAny(Equal(10), Equal(0)), "Default page size should be 10 or 0")
		})

		It("should allow member to list repositories", func() {
            status, body := DoRequest(baseURL, MethodGet, fmt.Sprintf("/code_repository_service/api/v1/repositories?project_id=%d", projectID), nil, map[string]string{"Authorization": "Bearer " + bobToken})
            Expect(status).To(Equal(StatusOK), "ListRepositories(bob) should succeed, body: %s", string(body))

			var lr ListRepositoriesResp
			Expect(json.Unmarshal(body, &lr)).To(Succeed(), "ListRepositories(bob) response should be valid JSON")

			found := false
			for _, r := range lr.Data {
				if r.ID == repoID {
					found = true
					break
				}
			}
			Expect(found).To(BeTrue(), "Bob (member) should be able to see repositories")
		})

		It("should deny non-member access to list repositories", func() {
            status, _ := DoRequest(baseURL, MethodGet, fmt.Sprintf("/code_repository_service/api/v1/repositories?project_id=%d", projectID), nil, map[string]string{"Authorization": "Bearer " + charlieToken})
            Expect(status).To(SatisfyAny(Equal(StatusForbidden), Equal(StatusUnauthorized)), "Non-member should be denied list access")
		})
	})

	Context("when getting repository details", func() {
		BeforeEach(func() {
			// 创建仓库用于测试
			repoName := fmt.Sprintf("repo_%d", suf)
            status, body := DoRequest(baseURL, MethodPost, "/code_repository_service/api/v1/repositories", map[string]any{
                "project_id":  projectID,
                "name":        repoName,
                "description": "first repo",
                "git_url":     "https://example.invalid/repo.git",
                "branch":      "main",
                "auth_type":   "REPOSITORY_AUTH_TYPE_NONE",
            }, map[string]string{"Authorization": "Bearer " + aliceToken})
            Expect(status).To(Equal(StatusOK))

			var cr CreateRepositoryResp
			Expect(json.Unmarshal(body, &cr)).To(Succeed())
			repoID = cr.Repository.ID
		})

		It("should allow member to get repository details", func() {
            status, body := DoRequest(baseURL, MethodGet, fmt.Sprintf("/code_repository_service/api/v1/repositories/%d?project_id=%d", repoID, projectID), nil, map[string]string{"Authorization": "Bearer " + bobToken})
            Expect(status).To(Equal(StatusOK), "GetRepository(bob) should succeed, body: %s", string(body))

			var gr GetRepositoryResp
			Expect(json.Unmarshal(body, &gr)).To(Succeed(), "GetRepository(bob) response should be valid JSON")
			Expect(gr.Repository.ID).To(Equal(repoID), "GetRepository should return correct repository")
		})

		It("should deny non-member access to get repository details", func() {
            status, _ := DoRequest(baseURL, MethodGet, fmt.Sprintf("/code_repository_service/api/v1/repositories/%d?project_id=%d", repoID, projectID), nil, map[string]string{"Authorization": "Bearer " + charlieToken})
            Expect(status).To(SatisfyAny(Equal(StatusForbidden), Equal(StatusUnauthorized)), "Non-member should be denied get access")
		})
	})
})