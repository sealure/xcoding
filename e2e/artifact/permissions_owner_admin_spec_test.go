//go:build e2e
// +build e2e

package artifact_e2e

import (
    "encoding/json"
    "fmt"
    "time"

    "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"

    "xcoding/e2e/helpers"
)

var _ = ginkgo.Describe("Permissions Owner/Admin", func() {
    var baseURL string

    ginkgo.BeforeEach(func() {
        baseURL = helpers.BaseURL()
    })

    ginkgo.It("validates owner/admin permissions for private resources", func() {
        var status int
        var body []byte

        suf := time.Now().UnixNano()

        // 使用共享权限场景构建器：Owner+Admin+Super，私有资源与私有仓库
        scen, err := helpers.BuildPermissionScenarioAdmin()
        Expect(err).NotTo(HaveOccurred())

        headersOwner := map[string]string{"Authorization": "Bearer " + scen.OwnerToken}
        headersAdmin := map[string]string{"Authorization": "Bearer " + scen.AdminToken}

        // Admin 在私有仓库写 Tag（允许）
        status, body = helpers.DoRequest(baseURL, helpers.MethodPost, "/artifact_service/api/v1/tags", map[string]any{
            "name":          fmt.Sprintf("v%d", suf),
            "digest":        fmt.Sprintf("sha256:%x", suf),
            "manifest":      "{}",
            "repository_id": scen.RepositoryID,
            "size_bytes":    1,
        }, headersAdmin)
        Expect(status).To(Equal(helpers.StatusOK), fmt.Sprintf("CreateTag(admin) status=%d body=%s", status, string(body)))

        // Owner 读取私有 Repository（允许）
        status, body = helpers.DoRequest(baseURL, helpers.MethodGet, fmt.Sprintf("/artifact_service/api/v1/repositories/%d", scen.RepositoryID), nil, headersOwner)
        Expect(status).To(Equal(helpers.StatusOK), fmt.Sprintf("GetRepository(owner) status=%d body=%s", status, string(body)))

        // 如果有生成的 Tag，尝试删除
        var tag struct { Tag struct { ID uint64 `json:"id,string"` } `json:"tag"` }
        _ = json.Unmarshal(body, &tag)
        if tag.Tag.ID > 0 {
            status, body = helpers.DoRequest(baseURL, helpers.MethodPost, "/artifact_service/api/v1/tags/delete", map[string]any{
                "id": tag.Tag.ID,
            }, headersAdmin)
            Expect(status).To(Equal(helpers.StatusOK))
        }
    })
})