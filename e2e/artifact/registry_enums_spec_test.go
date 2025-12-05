//go:build e2e
// +build e2e

package artifact_e2e

import (
	"encoding/json"
	"fmt"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"xcoding/e2e/helpers"
)

var _ = ginkgo.Describe("Registry Enums", func() {
    var baseURL string

    ginkgo.BeforeEach(func() {
        baseURL = helpers.BaseURL()
    })

    ginkgo.It("creates registry with enums and updates them", func() {
        var status int
        var body []byte

        // 登录超级管理员用于创建 Registry（需要更高权限）
        jwt, err := adminLogin()
        Expect(err).NotTo(HaveOccurred())
        authHeaders := map[string]string{
            "Content-Type":  "application/json",
            "User-Agent":    "xcoding-e2e",
            "Authorization": fmt.Sprintf("Bearer %s", jwt),
        }

		// 创建项目
        status, body = helpers.DoRequest(baseURL, helpers.MethodPost, "/project_service/api/v1/projects", map[string]any{
            "name":        fmt.Sprintf("proj_%d", helpers.UniqueNano()),
            "description": "e2e project",
            "language":    "go",
            "framework":   "grpc",
            "is_public":   false,
        }, authHeaders)
		Expect(status).To(Equal(helpers.StatusOK))
		var pr struct {
			Project struct {
				ID uint64 `json:"id,string"`
			} `json:"project"`
		}
		Expect(json.Unmarshal(body, &pr)).To(Succeed())
		Expect(pr.Project.ID).ToNot(BeZero())

		// 创建注册表，携带 enums
		status, body = helpers.DoRequest(baseURL, helpers.MethodPost, "/artifact_service/api/v1/registries", map[string]any{
			"name":            fmt.Sprintf("reg_%d", helpers.UniqueNano()),
			"url":             "https://registry.example.com",
			"description":     "e2e registry",
			"is_public":       false,
			"username":        "u",
			"password":        "p",
			"project_id":      pr.Project.ID,
			"artifact_type":   "ARTIFACT_TYPE_GENERIC_FILE",
			"artifact_source": "ARTIFACT_SOURCE_XCODING_REGISTRY",
		}, authHeaders)
		Expect(status).To(Equal(helpers.StatusOK))
        var cr struct {
            Registry struct {
                ID uint64 `json:"id,string"`
            } `json:"registry"`
        }
		Expect(json.Unmarshal(body, &cr)).To(Succeed())
		Expect(cr.Registry.ID).ToNot(BeZero())

		// 获取并断言 enums 存在
		status, body = helpers.DoRequest(baseURL, helpers.MethodGet, fmt.Sprintf("/artifact_service/api/v1/registries/%d", cr.Registry.ID), nil, authHeaders)
		Expect(status).To(Equal(helpers.StatusOK))
        var gr struct {
            Registry struct {
                ID             uint64 `json:"id,string"`
                ArtifactType   string `json:"artifact_type"`
                ArtifactSource string `json:"artifact_source"`
            } `json:"registry"`
        }
		Expect(json.Unmarshal(body, &gr)).To(Succeed())
		Expect(gr.Registry.ID).To(Equal(cr.Registry.ID))
		Expect(gr.Registry.ArtifactType).NotTo(BeEmpty())
		Expect(gr.Registry.ArtifactSource).NotTo(BeEmpty())

		// 更新 enums
		status, body = helpers.DoRequest(baseURL, helpers.MethodPut, fmt.Sprintf("/artifact_service/api/v1/registries/%d", cr.Registry.ID), map[string]any{
			"id":              cr.Registry.ID,
			"artifact_type":   "ARTIFACT_TYPE_DOCKER",
			"artifact_source": "ARTIFACT_SOURCE_SMB",
			"project_id":      pr.Project.ID,
		}, authHeaders)
		Expect(status).To(Equal(helpers.StatusOK))

	status, body = helpers.DoRequest(baseURL, helpers.MethodGet, fmt.Sprintf("/artifact_service/api/v1/registries/%d", cr.Registry.ID), nil, authHeaders)
	Expect(status).To(Equal(helpers.StatusOK))
        var gr2 struct {
            Registry struct {
                ID             uint64 `json:"id,string"`
                ArtifactType   string `json:"artifact_type"`
                ArtifactSource string `json:"artifact_source"`
            } `json:"registry"`
        }
	Expect(json.Unmarshal(body, &gr2)).To(Succeed())
	// 断言更新后的枚举值
	Expect(gr2.Registry.ArtifactType).To(Equal("ARTIFACT_TYPE_DOCKER"))
	Expect(gr2.Registry.ArtifactSource).To(Equal("ARTIFACT_SOURCE_SMB"))
	})
})