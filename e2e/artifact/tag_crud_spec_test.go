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

var _ = ginkgo.Describe("Tag CRUD", func() {
	var baseURL string

	ginkgo.BeforeEach(func() {
		baseURL = helpers.BaseURL()
	})

	ginkgo.It("creates, gets, updates, and deletes a tag", func() {
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

        // 创建仓库
        // 使用共享仓库构建器
        repoID, err := helpers.BuildRepository(jwt, helpers.FlexID(nid), fmt.Sprintf("repo_%d", helpers.UniqueNano()), false)
        Expect(err).NotTo(HaveOccurred())
        Expect(repoID).ToNot(BeZero())

        // 使用共享 Tag 构建器
        tagID, err := helpers.BuildTag(jwt, repoID, "v1", "sha256:deadbeef", "{}", 0)
        Expect(err).NotTo(HaveOccurred())
        Expect(tagID).ToNot(BeZero())

        status, body = helpers.DoRequest(baseURL, helpers.MethodGet, fmt.Sprintf("/artifact_service/api/v1/tags/%d", uint64(tagID)), nil, authHeaders)
        Expect(status).To(Equal(helpers.StatusOK))
        var gt struct {
            Tag struct {
                ID uint64 `json:"id,string"`
            } `json:"tag"`
        }
        Expect(json.Unmarshal(body, &gt)).To(Succeed())
        Expect(gt.Tag.ID).To(Equal(uint64(tagID)))

        status, body = helpers.DoRequest(baseURL, helpers.MethodPut, fmt.Sprintf("/artifact_service/api/v1/tags/%d", uint64(tagID)), map[string]any{
            "name": "v2",
        }, authHeaders)
        Expect(status).To(Equal(helpers.StatusOK))

        status, body = helpers.DoRequest(baseURL, helpers.MethodDelete, fmt.Sprintf("/artifact_service/api/v1/tags/%d", uint64(tagID)), nil, authHeaders)
        Expect(status).To(Equal(helpers.StatusOK))
        var dt struct {
            Success bool `json:"success"`
        }
        Expect(json.Unmarshal(body, &dt)).To(Succeed())
        Expect(dt.Success).To(BeTrue())
	})
})
