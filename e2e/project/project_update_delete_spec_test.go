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

var _ = ginkgo.Describe("Project Update/Delete Flow", func() {
    ginkgo.It("updates and deletes a project with role changes", func() {
        suf := time.Now().UnixNano()
        aliceName := fmt.Sprintf("alice_%d", suf)
        bobName := fmt.Sprintf("bob_%d", suf)

        _, _, _ = registerUser(aliceName, fmt.Sprintf("%s@example.com", aliceName), "pass1234")
        bobUser, _, _ := registerUser(bobName, fmt.Sprintf("%s@example.com", bobName), "pass1234")
        _, aliceToken, _ := loginUser(aliceName, "pass1234")
        _, bobToken, _ := loginUser(bobName, "pass1234")

        status, body := helpers.DoRequest(baseURL(), helpers.MethodPost, "/project_service/api/v1/projects", map[string]any{
            "name":        fmt.Sprintf("proj_%d", suf),
            "description": "project update delete flow",
            "language":    "go",
            "framework":   "grpc-gateway",
            "is_public":   false,
        }, map[string]string{"Authorization": "Bearer " + aliceToken})
        Expect(status).To(Equal(helpers.StatusOK), fmt.Sprintf("CreateProject status=%d body=%s", status, string(body)))
        var cp CreateProjectResp
        Expect(json.Unmarshal(body, &cp)).To(Succeed())
        projectID := cp.Project.ID

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

        status, body = helpers.DoRequest(baseURL(), helpers.MethodPost, fmt.Sprintf("/project_service/api/v1/projects/%d/sync-permissions", projectID), map[string]any{ "project_id": projectID }, map[string]string{"Authorization": "Bearer " + aliceToken})
        Expect(status).To(Equal(helpers.StatusOK), fmt.Sprintf("SyncUserPermissions failed: %d %s", status, string(body)))

        status, body = helpers.DoRequest(baseURL(), helpers.MethodPut, fmt.Sprintf("/project_service/api/v1/projects/%d", projectID), map[string]any{ "status": "archived", "is_public": true }, map[string]string{"Authorization": "Bearer " + aliceToken})
        Expect(status).To(Equal(helpers.StatusOK), fmt.Sprintf("Owner update failed: %d %s", status, string(body)))

        status, body = helpers.DoRequest(baseURL(), helpers.MethodPut, fmt.Sprintf("/project_service/api/v1/projects/%d", projectID), map[string]any{ "description": "updated-by-admin" }, map[string]string{"Authorization": "Bearer " + bobToken})
        Expect(status).To(Equal(helpers.StatusOK), fmt.Sprintf("Admin update failed: %d %s", status, string(body)))

        status, body = helpers.DoRequest(baseURL(), helpers.MethodDelete, fmt.Sprintf("/project_service/api/v1/projects/%d", projectID), nil, map[string]string{"Authorization": "Bearer " + bobToken})
        Expect(status).To(Equal(helpers.StatusOK), fmt.Sprintf("Admin delete failed: %d %s", status, string(body)))
    })
})