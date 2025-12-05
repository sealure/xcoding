//go:build e2e
// +build e2e

package user_e2e

import (
    "fmt"
    "encoding/json"
    "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
    "xcoding/e2e/helpers"
)

var _ = ginkgo.Describe("API Tokens", func() {
    // 移除 ping() 跳过逻辑，直接执行测试以观察真实失败

    ginkgo.It("creates, lists and deletes tokens", func() {
        _, _, jwt, err := registerAndLogin()
        Expect(err).NotTo(HaveOccurred())
        headers := map[string]string{"Authorization": "Bearer " + jwt}

        // Create
        createReq := map[string]any{
            "name":        "goe2e_token",
            "description": "go e2e token",
            "expires_in":  "TOKEN_EXPIRATION_ONE_MONTH",
            "scopes":      []string{"SCOPE_READ", "SCOPE_WRITE"},
        }
        status, body := do(helpers.MethodPost, "/user_service/api/v1/tokens", createReq, headers)
        if status != helpers.StatusOK { ginkgo.GinkgoWriter.Println(fmt.Sprintf("create token resp: status=%d body=%s", status, string(body))) }
        Expect(status).To(Equal(helpers.StatusOK))
        var created APIToken
        Expect(json.Unmarshal(body, &created)).To(Succeed())
        Expect(created.ID).NotTo(BeZero())
        Expect(created.Token).NotTo(BeEmpty())

        // List
        status, body = do(helpers.MethodGet, "/user_service/api/v1/tokens", nil, headers)
        if status != helpers.StatusOK { ginkgo.GinkgoWriter.Println(fmt.Sprintf("list token resp: status=%d body=%s", status, string(body))) }
        Expect(status).To(Equal(helpers.StatusOK))
        var tokens ListAPITokensResp
        Expect(json.Unmarshal(body, &tokens)).To(Succeed())
        Expect(len(tokens.Tokens)).To(BeNumerically(">", 0))

        // Delete
        status, body = do(helpers.MethodDelete, "/user_service/api/v1/tokens/"+fmtUint(created.ID), nil, headers)
        if status != helpers.StatusOK { ginkgo.GinkgoWriter.Println(fmt.Sprintf("delete token resp: status=%d body=%s", status, string(body))) }
        Expect(status).To(Equal(helpers.StatusOK))
        var del struct{ Success bool `json:"success"` }
        Expect(json.Unmarshal(body, &del)).To(Succeed())
        Expect(del.Success).To(BeTrue())
    })
})