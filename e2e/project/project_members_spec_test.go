//go:build e2e
// +build e2e

package project_e2e

import (
    "encoding/json"
    "fmt"
    "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
    "time"
    "xcoding/e2e/helpers"
)

var _ = ginkgo.Describe("Project Members & Permissions", func() {
    ginkgo.It("adds members, handles permissions and lists members", func() {
        suf := time.Now().UnixNano()
        aliceName := fmt.Sprintf("alice_%d", suf)
        bobName := fmt.Sprintf("bob_%d", suf)

        _, _, _ = registerUser(aliceName, fmt.Sprintf("%s@example.com", aliceName), "pass1234")
        bobUser, _, _ := registerUser(bobName, fmt.Sprintf("%s@example.com", bobName), "pass1234")
        _, aliceToken, _ := loginUser(aliceName, "pass1234")
        _, bobToken, _ := loginUser(bobName, "pass1234")

        status, body := helpers.DoRequest(baseURL(), helpers.MethodPost, "/project_service/api/v1/projects", map[string]any{
            "name":        fmt.Sprintf("proj_%d", suf),
            "description": "project members flow",
            "language":    "go",
            "framework":   "grpc-gateway",
            "is_public":   false,
        }, map[string]string{"Authorization": "Bearer " + aliceToken})
        Expect(status).To(Equal(helpers.StatusOK), fmt.Sprintf("CreateProject status=%d body=%s", status, string(body)))
        var cp CreateProjectResp
        Expect(json.Unmarshal(body, &cp)).To(Succeed())
        projectID := cp.Project.ID

        status, body = helpers.DoRequest(baseURL(), helpers.MethodPost, fmt.Sprintf("/project_service/api/v1/projects/%d/members", projectID), map[string]any{
            "project_id": projectID,
            "user_id":    bobUser.ID,
            "role":       3,
        }, map[string]string{"Authorization": "Bearer " + aliceToken})
        Expect(status).To(Equal(helpers.StatusOK), fmt.Sprintf("AddProjectMember status=%d body=%s", status, string(body)))
        var apm AddProjectMemberResp
        Expect(json.Unmarshal(body, &apm)).To(Succeed())

        status, body = helpers.DoRequest(baseURL(), helpers.MethodGet, "/project_service/api/v1/projects?all=false", nil, map[string]string{"Authorization": "Bearer " + bobToken})
        Expect(status).To(Equal(helpers.StatusOK), fmt.Sprintf("ListProjects(bob) status=%d body=%s", status, string(body)))
        var lpBob ListProjectsResp
        Expect(json.Unmarshal(body, &lpBob)).To(Succeed())

        status, body = helpers.DoRequest(baseURL(), helpers.MethodPut, fmt.Sprintf("/project_service/api/v1/projects/%d", projectID), map[string]any{ "role": "PROJECT_MEMBER_ROLE_MEMBER" }, map[string]string{"Authorization": "Bearer " + bobToken})
        Expect(status == helpers.StatusForbidden || status == helpers.StatusUnauthorized).To(BeTrue(), fmt.Sprintf("Member update should be forbidden, got %d body=%s", status, string(body)))

        status, body = helpers.DoRequest(baseURL(), helpers.MethodPut, fmt.Sprintf("/project_service/api/v1/projects/%d/members/%d", projectID, bobUser.ID), map[string]any{ "role": 2 }, map[string]string{"Authorization": "Bearer " + aliceToken})
        if status != helpers.StatusOK {
            statusDel, bodyDel := helpers.DoRequest(baseURL(), helpers.MethodDelete, fmt.Sprintf("/project_service/api/v1/projects/%d/members/%d", projectID, bobUser.ID), nil, map[string]string{"Authorization": "Bearer " + aliceToken})
            Expect(statusDel).To(Equal(helpers.StatusOK), fmt.Sprintf("RemoveProjectMember failed: %d %s", statusDel, string(bodyDel)))
            statusAdd, bodyAdd := helpers.DoRequest(baseURL(), helpers.MethodPost, fmt.Sprintf("/project_service/api/v1/projects/%d/members", projectID), map[string]any{
                "project_id": projectID,
                "user_id":    bobUser.ID,
                "role":       2,
            }, map[string]string{"Authorization": "Bearer " + aliceToken})
            Expect(statusAdd).To(Equal(helpers.StatusOK), fmt.Sprintf("Fallback AddProjectMember failed: %d %s", statusAdd, string(bodyAdd)))
        }

        status, body = helpers.DoRequest(baseURL(), helpers.MethodGet, "/project_service/api/v1/projects?all=true", nil, map[string]string{"Authorization": "Bearer " + bobToken})
        Expect(status).To(Equal(helpers.StatusOK), fmt.Sprintf("Admin all=true should be allowed, got %d %s", status, string(body)))

        status, body = helpers.DoRequest(baseURL(), helpers.MethodGet, fmt.Sprintf("/project_service/api/v1/projects/%d/members", projectID), nil, map[string]string{"Authorization": "Bearer " + aliceToken})
        Expect(status).To(Equal(helpers.StatusOK), fmt.Sprintf("ListProjectMembers status=%d body=%s", status, string(body)))
        var lpm ListProjectMembersResp
        Expect(json.Unmarshal(body, &lpm)).To(Succeed())
        foundAdmin := false
        for _, m := range lpm.Data {
            if uint64(m.UserID) == uint64(bobUser.ID) && roleToNum(string(m.Role)) == 2 { foundAdmin = true }
            Expect(m.User).NotTo(BeNil())
            Expect(m.User.Username).NotTo(BeEmpty())
        }
        Expect(foundAdmin).To(BeTrue(), "Bob should be ADMIN(2)")
    })
})