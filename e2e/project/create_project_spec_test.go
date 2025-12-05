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

var _ = ginkgo.Describe("Create Project With User", func() {
    ginkgo.It("creates project with user and gets it", func() {
        suf := time.Now().UnixNano()
        aliceName := fmt.Sprintf("alice_%d", suf)
        bobName := fmt.Sprintf("bob_%d", suf)

        _, _ , _ = registerUser(aliceName, fmt.Sprintf("%s@example.com", aliceName), "pass1234")
        _, _, _ = registerUser(bobName, fmt.Sprintf("%s@example.com", bobName), "pass1234")
        _, aliceToken, err := loginUser(aliceName, "pass1234")
        Expect(err).NotTo(HaveOccurred())

        status, body := helpers.DoRequest(baseURL(), helpers.MethodPost, "/project_service/api/v1/projects/create-with-user", map[string]any{
            "user": map[string]any{"username": aliceName},
            "project": ProjectInfo{
                Name:        fmt.Sprintf("proj_%d", suf),
                Description: "extended e2e project",
                Language:    "go",
                Framework:   "grpc-gateway",
                IsPublic:    false,
            },
            "create_repository": false,
        }, map[string]string{"Authorization": "Bearer " + aliceToken})
        Expect(status == helpers.StatusOK || status == helpers.StatusCreated).To(BeTrue(), "CreateProjectWithUser status=%d body=%s", status, string(body))

        var cpwu CreateProjectWithUserResp
        Expect(json.Unmarshal(body, &cpwu)).To(Succeed())
        projectID := cpwu.Project.ID

        status, body = helpers.DoRequest(baseURL(), helpers.MethodGet, fmt.Sprintf("/project_service/api/v1/projects/%d", projectID), nil, map[string]string{"Authorization": "Bearer " + aliceToken})
        Expect(status).To(Equal(helpers.StatusOK))
        var gp GetProjectResp
        Expect(json.Unmarshal(body, &gp)).To(Succeed())
        Expect(gp.Project.ID).To(Equal(projectID))
    })
})