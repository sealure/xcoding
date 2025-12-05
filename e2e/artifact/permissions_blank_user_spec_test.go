//go:build e2e
// +build e2e

package artifact_e2e

import (
    "fmt"
    "time"

    "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"

    "xcoding/e2e/helpers"
)

var _ = ginkgo.Describe("Permissions Blank User", func() {
	var baseURL string

	ginkgo.BeforeEach(func() {
		baseURL = helpers.BaseURL()
	})

    ginkgo.It("denies write and read for unauthenticated user on private resources", func() {
        var status int
        var body []byte

        suf := time.Now().UnixNano()

        // 使用共享权限场景构建器：仅超管创建私有资源
        scen, err := helpers.BuildPermissionScenarioBlank()
        Expect(err).NotTo(HaveOccurred())

        // 未认证用户读取与写入应被拒绝（严格断言，不跳过）
        status, body = helpers.DoRequest(baseURL, helpers.MethodGet, fmt.Sprintf("/artifact_service/api/v1/namespaces/%d", scen.NamespaceID), nil, nil)
        Expect(status).NotTo(Equal(helpers.StatusOK), fmt.Sprintf("blank user should be denied read, got status=%d body=%s", status, string(body)))

        status, body = helpers.DoRequest(baseURL, helpers.MethodPost, "/artifact_service/api/v1/repositories", map[string]any{
            "name":         fmt.Sprintf("repo_%d", suf),
            "description":  "private repo",
            "namespace_id": scen.NamespaceID,
            "is_public":    false,
        }, nil)
        Expect(status).NotTo(Equal(helpers.StatusOK), fmt.Sprintf("blank user should be denied write, got status=%d body=%s", status, string(body)))
    })
})