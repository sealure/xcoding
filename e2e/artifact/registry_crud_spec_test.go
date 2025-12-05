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

var _ = ginkgo.Describe("Registry CRUD", func() {
    var baseURL string

    ginkgo.BeforeEach(func() {
        baseURL = helpers.BaseURL()
    })

    ginkgo.It("creates, gets, updates, and deletes a registry", func() {
        var status int
        var body []byte

        // 登录超级管理员用于创建 Registry（需要更高权限）
        jwt, err := adminLogin()
        Expect(err).NotTo(HaveOccurred())

		// 带上认证头
		authHeaders := map[string]string{
			"Content-Type":  "application/json",
			"User-Agent":    "xcoding-e2e",
            "Authorization": fmt.Sprintf("Bearer %s", jwt),
        }

		// 使用共享构建器创建项目与注册表
		_, rid, err := helpers.BuildProjectAndRegistry(jwt, fmt.Sprintf("proj_%d", helpers.UniqueNano()))
		Expect(err).NotTo(HaveOccurred())
		Expect(rid).ToNot(BeZero())

		// 获取注册表
		status, body = helpers.DoRequest(baseURL, helpers.MethodGet, fmt.Sprintf("/artifact_service/api/v1/registries/%d", rid), nil, authHeaders)
		Expect(status).To(Equal(helpers.StatusOK))
        var gr struct {
            Registry struct {
                ID       uint64 `json:"id,string"`
                IsPublic bool   `json:"is_public"`
            } `json:"registry"`
        }
		Expect(json.Unmarshal(body, &gr)).To(Succeed())
		Expect(gr.Registry.ID).To(Equal(uint64(rid)))

		// 更新注册表 is_public
		status, body = helpers.DoRequest(baseURL, helpers.MethodPut, fmt.Sprintf("/artifact_service/api/v1/registries/%d", rid), map[string]any{
			"id":        uint64(rid),
			"is_public": true,
		}, authHeaders)
		Expect(status).To(Equal(helpers.StatusOK))
        var ur struct {
            Registry struct {
                ID       uint64 `json:"id,string"`
                IsPublic bool   `json:"is_public"`
            } `json:"registry"`
        }
		Expect(json.Unmarshal(body, &ur)).To(Succeed())
		Expect(ur.Registry.IsPublic).To(BeTrue())

		// 删除注册表
		status, body = helpers.DoRequest(baseURL, helpers.MethodDelete, fmt.Sprintf("/artifact_service/api/v1/registries/%d", rid), nil, authHeaders)
		Expect(status).To(Equal(helpers.StatusOK))
		var dr struct {
			Success bool `json:"success"`
		}
		Expect(json.Unmarshal(body, &dr)).To(Succeed())
		Expect(dr.Success).To(BeTrue())
	})
})