//go:build e2e
// +build e2e

package user_e2e

import (
    "encoding/json"
    "fmt"
    "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
    "xcoding/e2e/helpers"
)

var _ = ginkgo.Describe("Auth Endpoint", func() {
    // 移除 ping() 跳过逻辑，直接执行测试以观察真实失败

    ginkgo.It("authenticates via JWT and API token, handles invalid", func() {
        id, username, jwt, err := registerAndLogin()
        Expect(err).NotTo(HaveOccurred())
        headers := map[string]string{"Authorization": "Bearer " + jwt}

        // JWT
        status, body, respHeaders := doWithHeaders(helpers.MethodGet, "/user_service/api/v1/auth", nil, headers)
        if status != helpers.StatusOK { ginkgo.GinkgoWriter.Println(fmt.Sprintf("auth via jwt resp: status=%d body=%s", status, string(body))) }
        Expect(status).To(Equal(helpers.StatusOK))
        var auth AuthResp
        Expect(json.Unmarshal(body, &auth)).To(Succeed())
        Expect(auth.Authenticated).To(BeTrue())
        Expect(auth.User).NotTo(BeNil())
        Expect(respHeaders["X-User-ID"]).To(Equal(fmt.Sprintf("%d", id)))
        Expect(respHeaders["X-Username"]).To(Equal(username))
        Expect(respHeaders["X-User-Role"]).To(Equal("USER_ROLE_USER"))
        Expect(auth.Headers).NotTo(BeNil())
        Expect(auth.Headers["X-User-ID"]).To(Equal(respHeaders["X-User-ID"]))
        Expect(auth.Headers["X-Username"]).To(Equal(respHeaders["X-Username"]))
        Expect(auth.Headers["X-User-Role"]).To(Equal(respHeaders["X-User-Role"]))

        // Create API token
        createReq := map[string]any{
            "name":        "goe2e_auth_token",
            "description": "auth e2e token",
            "expires_in":  "TOKEN_EXPIRATION_ONE_MONTH",
            "scopes":      []string{"SCOPE_READ", "SCOPE_WRITE"},
        }
        status, body = do(helpers.MethodPost, "/user_service/api/v1/tokens", createReq, headers)
        if status != helpers.StatusOK { ginkgo.GinkgoWriter.Println(fmt.Sprintf("create token resp: status=%d body=%s", status, string(body))) }
        Expect(status).To(Equal(helpers.StatusOK))
        var created APIToken
        Expect(json.Unmarshal(body, &created)).To(Succeed())
        Expect(created.Token).NotTo(BeEmpty())

        status, body, respHeaders = doWithHeaders(helpers.MethodGet, "/user_service/api/v1/auth", nil, map[string]string{"Authorization": "Bearer " + created.Token})
        if status != helpers.StatusOK { ginkgo.GinkgoWriter.Println(fmt.Sprintf("auth via api-token resp: status=%d body=%s", status, string(body))) }
        Expect(status).To(Equal(helpers.StatusOK))
        Expect(json.Unmarshal(body, &auth)).To(Succeed())
        Expect(auth.Authenticated).To(BeTrue())
        Expect(auth.User).NotTo(BeNil())
        Expect(respHeaders["X-User-ID"]).To(Equal(fmt.Sprintf("%d", id)))
        Expect(respHeaders["X-Username"]).To(Equal(username))
        Expect(respHeaders["X-User-Role"]).To(Equal("USER_ROLE_USER"))
        Expect(respHeaders["X-Scopes"]).To(Equal("SCOPE_READ,SCOPE_WRITE"))
        Expect(auth.Headers["X-Scopes"]).To(Equal(respHeaders["X-Scopes"]))

        // Invalid token
        status, body = do(helpers.MethodGet, "/user_service/api/v1/auth", nil, map[string]string{"Authorization": "Bearer invalid-token-test"})
        if status != helpers.StatusOK { ginkgo.GinkgoWriter.Println(fmt.Sprintf("auth via invalid token resp: status=%d body=%s", status, string(body))) }
        if status == helpers.StatusOK {
            var invalid AuthResp
            _ = json.Unmarshal(body, &invalid)
            Expect(invalid.Authenticated).To(BeFalse())
        }
    })
})