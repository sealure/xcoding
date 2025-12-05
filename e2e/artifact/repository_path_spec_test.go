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

var _ = ginkgo.Describe("Repository Path", func() {
    var baseURL string

    ginkgo.BeforeEach(func() {
        baseURL = helpers.BaseURL()
    })

    ginkgo.It("creates repository with path and updates it", func() {
        var status int
        var body []byte

        // 登录超级管理员用于创建 Registry（需要更高权限）
        jwt, err := adminLogin()
        Expect(err).NotTo(HaveOccurred())
        authHeaders := map[string]string{
            "Content-Type": "application/json",
            "User-Agent":   "xcoding-e2e",
            "Authorization": fmt.Sprintf("Bearer %s", jwt),
        }

        // 共享构建器：项目、注册表、命名空间一键生成
        _, _, nid, err := helpers.BuildProjectRegistryNamespace(jwt, fmt.Sprintf("proj_%d", helpers.UniqueNano()))
        Expect(err).NotTo(HaveOccurred())
        Expect(nid).ToNot(BeZero())

        // 使用共享仓库构建器（带 path）
        repoID, err := helpers.BuildRepositoryWithPath(jwt, helpers.FlexID(nid), fmt.Sprintf("repo_%d", helpers.UniqueNano()), false, "/e2e/path")
        Expect(err).NotTo(HaveOccurred())
        Expect(repoID).ToNot(BeZero())

        status, body = helpers.DoRequest(baseURL, helpers.MethodGet, fmt.Sprintf("/artifact_service/api/v1/repositories/%d", uint64(repoID)), nil, authHeaders)
        Expect(status).To(Equal(helpers.StatusOK))
        var gr struct{ Repository struct{ ID uint64 `json:"id,string"`; Path string `json:"path"` } `json:"repository"` }
        Expect(json.Unmarshal(body, &gr)).To(Succeed())
        Expect(gr.Repository.ID).To(Equal(uint64(repoID)))
        Expect(gr.Repository.Path).NotTo(BeEmpty())

        status, body = helpers.DoRequest(baseURL, helpers.MethodPut, fmt.Sprintf("/artifact_service/api/v1/repositories/%d", uint64(repoID)), map[string]any{
            "path": "/e2e/updated",
        }, authHeaders)
        Expect(status).To(Equal(helpers.StatusOK))

        status, body = helpers.DoRequest(baseURL, helpers.MethodGet, fmt.Sprintf("/artifact_service/api/v1/repositories/%d", uint64(repoID)), nil, authHeaders)
        Expect(status).To(Equal(helpers.StatusOK))
        var gr2 struct{ Repository struct{ ID uint64 `json:"id,string"`; Path string `json:"path"` } `json:"repository"` }
        Expect(json.Unmarshal(body, &gr2)).To(Succeed())
        Expect(gr2.Repository.Path).To(Equal("/e2e/updated"))
    })
})