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

var _ = ginkgo.Describe("Repository CRUD", func() {
	var baseURL string

	ginkgo.BeforeEach(func() {
		baseURL = helpers.BaseURL()
	})

	ginkgo.It("creates, gets, updates, and deletes a repository", func() {
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

		// 使用共享构建器：项目、注册表、命名空间一键生成
		_, _, nid, err := helpers.BuildProjectRegistryNamespace(jwt, fmt.Sprintf("proj_%d", helpers.UniqueNano()))
		Expect(err).NotTo(HaveOccurred())
		Expect(nid).ToNot(BeZero())

		// 使用共享仓库构建器
		repoID, err := helpers.BuildRepository(jwt, helpers.FlexID(nid), fmt.Sprintf("repo_%d", helpers.UniqueNano()), false)
		Expect(err).NotTo(HaveOccurred())
		Expect(repoID).ToNot(BeZero())

		status, body = helpers.DoRequest(baseURL, helpers.MethodGet, fmt.Sprintf("/artifact_service/api/v1/repositories/%d", uint64(repoID)), nil, authHeaders)
		Expect(status).To(Equal(helpers.StatusOK))
		var gr struct {
			Repository struct {
				ID uint64 `json:"id,string"`
			} `json:"repository"`
		}
		Expect(json.Unmarshal(body, &gr)).To(Succeed())
		Expect(gr.Repository.ID).To(Equal(uint64(repoID)))

		status, body = helpers.DoRequest(baseURL, helpers.MethodPut, fmt.Sprintf("/artifact_service/api/v1/repositories/%d", uint64(repoID)), map[string]any{
			"description": "updated repo",
		}, authHeaders)
		Expect(status).To(Equal(helpers.StatusOK))

		status, body = helpers.DoRequest(baseURL, helpers.MethodDelete, fmt.Sprintf("/artifact_service/api/v1/repositories/%d", uint64(repoID)), nil, authHeaders)
		Expect(status).To(Equal(helpers.StatusOK))
		// 删除返回 200 即视为成功（与各模块统一）
		Expect(status).To(Equal(helpers.StatusOK))
	})
})
