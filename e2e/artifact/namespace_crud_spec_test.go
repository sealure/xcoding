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

var _ = ginkgo.Describe("Namespace CRUD", func() {
	var baseURL string

	ginkgo.BeforeEach(func() {
		baseURL = helpers.BaseURL()
	})

	ginkgo.It("creates, gets, updates, and deletes a namespace", func() {
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

		// 使用共享构建器创建项目与注册表
		_, rid, err := helpers.BuildProjectAndRegistry(jwt, fmt.Sprintf("proj_%d", helpers.UniqueNano()))
		Expect(err).NotTo(HaveOccurred())
		Expect(rid).ToNot(BeZero())

		// 创建命名空间
		status, body = helpers.DoRequest(baseURL, helpers.MethodPost, "/artifact_service/api/v1/namespaces", map[string]any{
			"registry_id": uint64(rid),
			"name":        fmt.Sprintf("ns_%d", helpers.UniqueNano()),
			"description": "e2e namespace",
			"is_public":   false,
		}, authHeaders)
		Expect(status).To(Equal(helpers.StatusOK))
		var ns struct {
			Namespace struct {
				ID uint64 `json:"id,string"`
			} `json:"namespace"`
		}
		Expect(json.Unmarshal(body, &ns)).To(Succeed())
		Expect(ns.Namespace.ID).ToNot(BeZero())

		// 获取命名空间
		status, body = helpers.DoRequest(baseURL, helpers.MethodGet, fmt.Sprintf("/artifact_service/api/v1/namespaces/%d", ns.Namespace.ID), nil, authHeaders)
		Expect(status).To(Equal(helpers.StatusOK))
		var gn struct {
			Namespace struct {
				ID uint64 `json:"id,string"`
			} `json:"namespace"`
		}
		Expect(json.Unmarshal(body, &gn)).To(Succeed())
		Expect(gn.Namespace.ID).To(Equal(ns.Namespace.ID))

		// 更新 name
		status, body = helpers.DoRequest(baseURL, helpers.MethodPut, fmt.Sprintf("/artifact_service/api/v1/namespaces/%d", ns.Namespace.ID), map[string]any{
			"name": fmt.Sprintf("ns_%d_updated", helpers.UniqueNano()),
		}, authHeaders)
		Expect(status).To(Equal(helpers.StatusOK))

		// 删除命名空间
		status, body = helpers.DoRequest(baseURL, helpers.MethodDelete, fmt.Sprintf("/artifact_service/api/v1/namespaces/%d", ns.Namespace.ID), nil, authHeaders)
		Expect(status).To(Equal(helpers.StatusOK))
		var dn struct {
			Success bool `json:"success"`
		}
		Expect(json.Unmarshal(body, &dn)).To(Succeed())
		Expect(dn.Success).To(BeTrue())
	})
})
