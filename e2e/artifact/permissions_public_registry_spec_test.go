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

// 验证：当 Registry 公开时，即便 Namespace/Repository 私有，匿名用户可读取仓库，但写入（创建 Tag）仍被拒绝
var _ = ginkgo.Describe("Permissions Public Registry", func() {
    var baseURL string

    ginkgo.BeforeEach(func() {
        baseURL = helpers.BaseURL()
    })

    ginkgo.It("allows read on private repo under public registry and denies writes", func() {
        var status int
        var body []byte

        suf := time.Now().UnixNano()

        // 构建公开 Registry 场景（Namespace/Repository 私有）
        scen, err := helpers.BuildPermissionScenarioPublicRegistry()
        Expect(err).NotTo(HaveOccurred())

        // 注册一个与项目无关的普通用户，用于模拟“非成员但已认证”的访问
        _, _, otherToken, err := helpers.RegisterAndLogin(fmt.Sprintf("visitor_%d", suf), fmt.Sprintf("visitor_%d@example.com", suf), "testpassword123")
        Expect(err).NotTo(HaveOccurred())
        headers := map[string]string{"Authorization": "Bearer " + otherToken}

        // 非成员用户读取私有 Repository（因为 Registry 公开，允许）
        status, body = helpers.DoRequest(baseURL, helpers.MethodGet, fmt.Sprintf("/artifact_service/api/v1/repositories/%d", scen.RepositoryID), nil, headers)
        Expect(status).To(Equal(helpers.StatusOK), fmt.Sprintf("GetRepository(blank, public registry) status=%d body=%s", status, string(body)))

        // 非成员用户创建 Tag（应被拒绝）
        status, body = helpers.DoRequest(baseURL, helpers.MethodPost, "/artifact_service/api/v1/tags", map[string]any{
            "name":          fmt.Sprintf("v%d", suf),
            "digest":        fmt.Sprintf("sha256:%x", suf),
            "manifest":      "{}",
            "repository_id": scen.RepositoryID,
            "size_bytes":    1,
        }, headers)
        Expect(status).NotTo(Equal(helpers.StatusOK), fmt.Sprintf("CreateTag(blank) should be denied, got status=%d body=%s", status, string(body)))
    })
})