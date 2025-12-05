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

var _ = ginkgo.Describe("Job timeout and node selector", func() {
	var baseURL string
	var token string
	ginkgo.BeforeEach(func() {
		baseURL = e2e.BaseURL()
		Expect(e2e.PingGateway(baseURL)).To(BeTrue())
		t, err := e2e.AdminLogin()
		Expect(err).NotTo(HaveOccurred())
		token = t
	})

	ginkgo.It("applies node selector and respects timeout", func() {
		headers := e2e.AuthHeader(token)
		suf := e2e.UniqueNano()
		pid, err := e2e.CreateProject(fmt.Sprintf("ci_exec_proj_%d", suf), "executor e2e", false, token)
		Expect(err).NotTo(HaveOccurred())

		yaml := strings.Join([]string{
			"jobs:",
			"  build:",
			"    container: alpine",
			"    env:",
			"      XC_JOB_TIMEOUT_SECONDS: 5",
			"      XC_NODE_SELECTOR_KEY: kubernetes.io/os",
			"      XC_NODE_SELECTOR_VALUE: linux",
			"    - name: hold",
			"      run: sleep 10",
		}, "\n")

		status, body := e2e.DoRequest(baseURL, e2e.MethodPost, "/ci_service/api/v1/pipelines", map[string]any{
			"name":          fmt.Sprintf("pipeline_%d", suf),
			"description":   "timeout & ns",
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
		}, 30*time.Second, 1*time.Second).Should(Equal("BUILD_STATUS_FAILED"))
	})

	ginkgo.It("node selector mismatch yields unschedulable and is observable via readonly API", func() {
		headers := e2e.AuthHeader(token)
		suf := e2e.UniqueNano()
		pid, err := e2e.CreateProject(fmt.Sprintf("ci_exec_proj_%d", suf), "executor e2e", false, token)
		Expect(err).NotTo(HaveOccurred())

		yaml := strings.Join([]string{
			"jobs:",
			"  build:",
			"    container: alpine",
			"    env:",
			"      XC_JOB_TIMEOUT_SECONDS: 5",
			"      XC_NODE_SELECTOR_KEY: kubernetes.io/os",
			"      XC_NODE_SELECTOR_VALUE: windows",
			"    - name: hold",
			"      run: sleep 20",
		}, "\n")

		status, body := e2e.DoRequest(baseURL, e2e.MethodPost, "/ci_service/api/v1/pipelines", map[string]any{
			"name":          fmt.Sprintf("pipeline_%d", suf),
			"description":   "ns mismatch",
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

		Eventually(func() bool {
			s, b := e2e.DoRequest(baseURL, e2e.MethodGet, fmt.Sprintf("/ci_service/api/v1/executor/builds/%d/k8s_status?job_name_prefix=%s", bid, fmt.Sprintf("build-%d-", bid)), nil, headers)
			if s != e2e.StatusOK {
				return false
			}
			var resp struct {
				Jobs []struct {
					JobName    string `json:"job_name"`
					Succeeded  int32  `json:"succeeded"`
					Failed     int32  `json:"failed"`
					Conditions []struct {
						Type    string `json:"type"`
						Status  string `json:"status"`
						Reason  string `json:"reason"`
						Message string `json:"message"`
					} `json:"conditions"`
					Pods []struct {
						Name   string `json:"name"`
						Phase  string `json:"phase"`
						Node   string `json:"node"`
						Reason string `json:"reason"`
					} `json:"pods"`
				} `json:"jobs"`
				Pagination struct {
					TotalItems int32 `json:"total_items"`
				} `json:"pagination"`
			}
			_ = json.Unmarshal(b, &resp)
			if len(resp.Jobs) == 0 {
				return false
			}
			found := false
			for _, j := range resp.Jobs {
				for _, p := range j.Pods {
					if p.Reason == "Unschedulable" || p.Node == "" {
						found = true
						break
					}
				}
				if found {
					break
				}
			}
			return found
		}, 30*time.Second, 1*time.Second).Should(BeTrue())

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
		}, 30*time.Second, 1*time.Second).Should(Equal("BUILD_STATUS_FAILED"))
	})
})
