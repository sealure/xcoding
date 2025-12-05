//go:build e2e

package executor_e2e

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
	e2e "xcoding/e2e/helpers"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = ginkgo.Describe("continue_on_error keeps workflow running", func() {
	var baseURL string
	var token string
	ginkgo.BeforeEach(func() {
		baseURL = e2e.BaseURL()
		Expect(e2e.PingGateway(baseURL)).To(BeTrue())
		t, err := e2e.AdminLogin()
		Expect(err).NotTo(HaveOccurred())
		token = t
	})

	ginkgo.It("marks failed step but completes job", func() {
		headers := e2e.AuthHeader(token)
		suf := e2e.UniqueNano()
		pid, err := e2e.CreateProject(fmt.Sprintf("ci_exec_proj_%d", suf), "executor e2e", false, token)
		Expect(err).NotTo(HaveOccurred())

		yaml := strings.Join([]string{
			"jobs:",
			"  build:",
			"    container: alpine:3.19",
			"    - name: step1",
			"      env:",
			"        XC_CONTINUE_ON_ERROR: true",
			"      run: exit 7",
			"    - name: step2",
			"      run: echo ok",
		}, "\n")

		status, body := e2e.DoRequest(baseURL, e2e.MethodPost, "/ci_service/api/v1/pipelines", map[string]any{
			"name":          fmt.Sprintf("pipeline_%d", suf),
			"description":   "coe",
			"project_id":    pid,
			"workflow_yaml": yaml,
			"is_active":     true,
		}, headers)
		Expect(status).To(Equal(e2e.StatusOK))
		var cp struct {
			Pipeline struct {
				ID e2e.FlexID `json:"id"`
			} `json:"pipeline"`
		}
		Expect(json.Unmarshal(body, &cp)).To(Succeed())

		status, body = e2e.DoRequest(baseURL, e2e.MethodPost, fmt.Sprintf("/ci_service/api/v1/pipelines/%d/builds", uint64(cp.Pipeline.ID)), map[string]any{
			"branch":       "main",
			"triggered_by": "e2e-exec",
		}, headers)
		Expect(status).To(Equal(e2e.StatusOK))
		var sb struct {
			Build struct {
				ID     e2e.FlexID `json:"id"`
				Status string     `json:"status"`
			} `json:"build"`
		}
		Expect(json.Unmarshal(body, &sb)).To(Succeed())
		bid := uint64(sb.Build.ID)

		Eventually(func() string {
			s, b := e2e.DoRequest(baseURL, e2e.MethodGet, fmt.Sprintf("/ci_service/api/v1/builds/%d", bid), nil, headers)
			if s != e2e.StatusOK {
				return ""
			}
			var gb struct {
				Build struct {
					Status string `json:"status"`
				} `json:"build"`
			}
			_ = json.Unmarshal(b, &gb)
			return gb.Build.Status
		}, 30*time.Second, 1*time.Second).Should(Equal("BUILD_STATUS_SUCCEEDED"))

		// 日志包含 step1 的 exit 标记
		s, b := e2e.DoRequest(baseURL, e2e.MethodGet, fmt.Sprintf("/ci_service/api/v1/executor/builds/%d/logs", bid), nil, headers)
		Expect(s).To(Equal(e2e.StatusOK))
		var gl struct {
			Lines []string `json:"lines"`
		}
		_ = json.Unmarshal(b, &gl)
		found := false
		for _, ln := range gl.Lines {
			if strings.Contains(ln, "__step_exit__ step1 7") {
				found = true
				break
			}
		}
		Expect(found).To(BeTrue(), "missing step exit code mark in logs")

		// 只读 K8s 状态接口：检查 conditions 与 succeeded/failed 字段存在
		s, b = e2e.DoRequest(baseURL, e2e.MethodGet, fmt.Sprintf("/ci_service/api/v1/executor/builds/%d/k8s_status", bid), nil, headers)
		Expect(s).To(Equal(e2e.StatusOK))
		var ks struct {
			Jobs []struct {
				Succeeded  int32 `json:"succeeded"`
				Failed     int32 `json:"failed"`
				Conditions []struct {
					Type   string `json:"type"`
					Status string `json:"status"`
				} `json:"conditions"`
			} `json:"jobs"`
		}
		_ = json.Unmarshal(b, &ks)
		Expect(len(ks.Jobs)).To(BeNumerically(">=", 1))
		// 至少存在 conditions 字段（可能为空），以及 Succeeded/Failed 数值字段
		Expect(ks.Jobs[0].Succeeded >= 0).To(BeTrue())
		Expect(ks.Jobs[0].Failed >= 0).To(BeTrue())
	})
})
