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

var _ = ginkgo.Describe("User CRUD", func() {
    // 移除 ping() 跳过逻辑，直接执行测试以观察真实失败

    ginkgo.It("creates, gets, updates, lists and deletes self", func() {

        uid, _, jwt, err := registerAndLogin()
        Expect(err).NotTo(HaveOccurred())
        headers := map[string]string{"Authorization": "Bearer " + jwt}

        // Get self
        status, body := do(helpers.MethodGet, "/user_service/api/v1/users/"+fmtUint(uid), nil, headers)
        if status != helpers.StatusOK { ginkgo.GinkgoWriter.Println(fmt.Sprintf("get self resp: status=%d body=%s", status, string(body))) }
        Expect(status).To(Equal(helpers.StatusOK))
        var me struct{ User User `json:"user"` }
        Expect(json.Unmarshal(body, &me)).To(Succeed())
        Expect(me.User.ID).NotTo(BeZero())

        // Update email
        updateReq := map[string]any{ "email": "updated_" + me.User.Email }
        status, body = do(helpers.MethodPut, "/user_service/api/v1/users/"+fmtUint(uid), updateReq, headers)
        if status != helpers.StatusOK { ginkgo.GinkgoWriter.Println(fmt.Sprintf("update resp: status=%d body=%s", status, string(body))) }
        Expect(status).To(Equal(helpers.StatusOK))
        var updated UpdateUserResp
        Expect(json.Unmarshal(body, &updated)).To(Succeed())
        Expect(updated.User.Email).To(Equal("updated_" + me.User.Email))

        // List users
        status, body = do(helpers.MethodGet, "/user_service/api/v1/users", nil, headers)
        if status != helpers.StatusOK { ginkgo.GinkgoWriter.Println(fmt.Sprintf("list resp: status=%d body=%s", status, string(body))) }
        Expect(status).To(Equal(helpers.StatusOK))
        var list ListUsersResp
        Expect(json.Unmarshal(body, &list)).To(Succeed())
        Expect(len(list.Data)).To(BeNumerically(">", 0))

        // Delete self
        status, body = do(helpers.MethodDelete, "/user_service/api/v1/users/"+fmtUint(uid), nil, headers)
        if status != helpers.StatusOK { ginkgo.GinkgoWriter.Println(fmt.Sprintf("delete resp: status=%d body=%s", status, string(body))) }
        Expect(status).To(Equal(helpers.StatusOK))
        var del DeleteUserResp
        Expect(json.Unmarshal(body, &del)).To(Succeed())
        Expect(del.Success).To(BeTrue())
    })
})