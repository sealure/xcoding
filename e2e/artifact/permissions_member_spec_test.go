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

var _ = ginkgo.Describe("Permissions Member", func() {
    var baseURL string

    ginkgo.BeforeEach(func() {
        baseURL = helpers.BaseURL()
    })

    ginkgo.It("validates member read permissions and denies writes", func() {
        var status int
        var body []byte

        suf := time.Now().UnixNano()

        // 使用共享权限场景构建器：Owner+Member+Super，私有资源与私有仓库
        scen, err := helpers.BuildPermissionScenarioMember()
        Expect(err).NotTo(HaveOccurred())

        // 成员尝试创建 Tag（应拒绝）
        headersMember := map[string]string{"Authorization": "Bearer " + scen.MemberToken}
        status, body = helpers.DoRequest(baseURL, helpers.MethodPost, "/artifact_service/api/v1/tags", map[string]any{
            "name":          fmt.Sprintf("v%d", suf),
            "digest":        fmt.Sprintf("sha256:%x", suf),
            "manifest":      "{}",
            "repository_id": scen.RepositoryID,
            "size_bytes":    1,
        }, headersMember)
        Expect(status).NotTo(Equal(helpers.StatusOK), fmt.Sprintf("Member should be denied to create tag, got status=%d body=%s", status, string(body)))

        // 成员读取私有 Repository（允许）
        status, body = helpers.DoRequest(baseURL, helpers.MethodGet, fmt.Sprintf("/artifact_service/api/v1/repositories/%d", scen.RepositoryID), nil, headersMember)
        Expect(status).To(Equal(helpers.StatusOK), fmt.Sprintf("GetRepository(member) status=%d body=%s", status, string(body)))
    })
})