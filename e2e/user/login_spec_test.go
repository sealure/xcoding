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

var _ = ginkgo.Describe("Login", func() {
    // 移除 ping() 跳过逻辑，直接执行测试以观察真实失败

    ginkgo.It("registers then logs in", func() {
        status, body := do(helpers.MethodPost, "/user_service/api/v1/users/register", map[string]any{
            "username": uniqueName(),
            "email":    uniqueEmail(),
            "password": "testpassword123",
        }, nil)
        if status != helpers.StatusOK { ginkgo.GinkgoWriter.Println(fmt.Sprintf("register resp: status=%d body=%s", status, string(body))) }
        Expect(status).To(Equal(helpers.StatusOK))
        var reg RegisterResp
        Expect(json.Unmarshal(body, &reg)).To(Succeed())
        status, body = do(helpers.MethodPost, "/user_service/api/v1/users/login", map[string]any{
            "username": reg.User.Username,
            "password": "testpassword123",
        }, nil)
        if status != helpers.StatusOK { ginkgo.GinkgoWriter.Println(fmt.Sprintf("login resp: status=%d body=%s", status, string(body))) }
        Expect(status).To(Equal(helpers.StatusOK))
        var lg LoginResp
        Expect(json.Unmarshal(body, &lg)).To(Succeed())
        Expect(lg.Token).NotTo(BeEmpty())
        Expect(lg.User.ID).NotTo(BeZero())
    })
})