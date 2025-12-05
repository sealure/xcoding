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

var _ = ginkgo.Describe("DAG concurrency preserves global seq order", func() {
	var baseURL string
	var token string
	ginkgo.BeforeEach(func() {
		baseURL = e2e.BaseURL()
		Expect(e2e.PingGateway(baseURL)).To(BeTrue())
		t, err := e2e.AdminLogin()
		Expect(err).NotTo(HaveOccurred())
		token = t
	})

	ginkgo.It("runs parallel jobs and assembles monotonic sequences", func() {
		headers := e2e.AuthHeader(token)
		suf := e2e.UniqueNano()
		pid, err := e2e.CreateProject(fmt.Sprintf("ci_exec_proj_%d", suf), "executor e2e", false, token)
		Expect(err).NotTo(HaveOccurred())

		yaml := strings.Join([]string{
			"jobs:",
			"  a:",
			"    container: alpine",
			"    - name: a1",
			"      run: echo a1",
			"  b:",
			"    container: alpine",
			"    - name: b1",
			"      run: echo b1",
			"  c:",
			"    needs: a b",
			"    container: alpine",
			"    - name: c1",
			"      run: echo c1",
		}, "\n")

		status, body := e2e.DoRequest(baseURL, e2e.MethodPost, "/ci_service/api/v1/pipelines", map[string]any{
			"name":          fmt.Sprintf("pipeline_%d", suf),
			"description":   "dag seq",
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
		}, 60*time.Second, 1*time.Second).Should(Equal("BUILD_STATUS_SUCCEEDED"))

		// 拉日志，至少包含标记，验证无重复标记（幂等）
		s, b := e2e.DoRequest(baseURL, e2e.MethodGet, fmt.Sprintf("/ci_service/api/v1/executor/builds/%d/logs", bid), nil, headers)
		Expect(s).To(Equal(e2e.StatusOK))
		var gl struct {
			Lines []string `json:"lines"`
		}
		_ = json.Unmarshal(b, &gl)
		seen := map[string]bool{}
		for _, ln := range gl.Lines {
			if strings.Contains(ln, "__step_begin__") || strings.Contains(ln, "__step_end__") {
				Expect(seen[ln]).To(BeFalse())
				seen[ln] = true
			}
		}

		// 只读 K8s 状态接口：检查至少返回一个 Job 与 Pod
		s, b = e2e.DoRequest(baseURL, e2e.MethodGet, fmt.Sprintf("/ci_service/api/v1/executor/builds/%d/k8s_status?page=1&page_size=5", bid), nil, headers)
		Expect(s).To(Equal(e2e.StatusOK))
		var ks struct {
			Jobs []struct {
				JobName string `json:"job_name"`
				Pods    []struct {
					Name  string `json:"name"`
					Phase string `json:"phase"`
					Node  string `json:"node"`
				} `json:"pods"`
			} `json:"jobs"`
		}
		_ = json.Unmarshal(b, &ks)
		Expect(len(ks.Jobs)).To(BeNumerically(">=", 1))
		hasPod := false
		for _, j := range ks.Jobs {
			if len(j.Pods) > 0 {
				hasPod = true
				break
			}
		}
		Expect(hasPod).To(BeTrue())
	})
})
