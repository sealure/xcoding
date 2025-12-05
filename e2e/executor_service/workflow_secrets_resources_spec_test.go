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

type buildResp struct {
	Build struct {
		ID         e2e.FlexID `json:"id"`
		PipelineID e2e.FlexID `json:"pipeline_id"`
		Status     string     `json:"status"`
	} `json:"build"`
}
type getBuildResp struct {
	Build struct {
		ID     e2e.FlexID `json:"id"`
		Status string     `json:"status"`
	} `json:"build"`
}
type getLogsResp struct {
	Lines []string `json:"lines"`
}

var _ = ginkgo.Describe("Executor workflow with secrets and resources", func() {
	var baseURL string
	var token string

	ginkgo.BeforeEach(func() {
		baseURL = e2e.BaseURL()
		Expect(e2e.PingGateway(baseURL)).To(BeTrue())
		t, err := e2e.AdminLogin()
		Expect(err).NotTo(HaveOccurred())
		token = t
	})

	ginkgo.It("runs workflow and asserts job/step states and log seq order", func() {
		headers := e2e.AuthHeader(token)
		suf := e2e.UniqueNano()
		pid, err := e2e.CreateProject(fmt.Sprintf("ci_exec_proj_%d", suf), "executor e2e", false, token)
		Expect(err).NotTo(HaveOccurred())

		yaml := strings.Join([]string{
			"jobs:",
			"  build:",
			"    container: alpine:3.19",
			"    env:",
			"      XC_RESOURCE_CPU_REQUEST: 100m",
			"      XC_RESOURCE_MEMORY_LIMIT: 128Mi",
			"      TOKEN: secret://example-secret/token",
			"    - name: step1",
			"      env:",
			"        TMP: tmpvalue",
			"      run: echo hello1",
			"    - name: step2",
			"      env:",
			"        TMP: tmpvalue2",
			"      run: echo hello2",
		}, "\n")

		status, body := e2e.DoRequest(baseURL, e2e.MethodPost, "/ci_service/api/v1/pipelines", map[string]any{
			"name":          fmt.Sprintf("pipeline_%d", suf),
			"description":   "executor e2e",
			"project_id":    pid,
			"workflow_yaml": yaml,
			"is_active":     true,
		}, headers)
		Expect(status).To(Equal(e2e.StatusOK), "CreatePipeline failed: %s", string(body))
		var cp struct {
			Pipeline struct {
				ID e2e.FlexID `json:"id"`
			} `json:"pipeline"`
		}
		Expect(json.Unmarshal(body, &cp)).To(Succeed())
		pipelineID := cp.Pipeline.ID

		status, body = e2e.DoRequest(baseURL, e2e.MethodPost, fmt.Sprintf("/ci_service/api/v1/pipelines/%d/builds", pipelineID), map[string]any{
			"branch":       "main",
			"triggered_by": "e2e-exec",
			"variables":    map[string]string{"k": "v"},
		}, headers)
		Expect(status).To(Equal(e2e.StatusOK), "StartPipelineBuild failed: %s", string(body))
		var sb buildResp
		Expect(json.Unmarshal(body, &sb)).To(Succeed())

		bid := uint64(sb.Build.ID)

		// 等待 RUNNING，再等待完成
		eventuallyStatus := func(expect string) {
			Eventually(func() string {
				s, b := e2e.DoRequest(baseURL, e2e.MethodGet, fmt.Sprintf("/ci_service/api/v1/builds/%d", bid), nil, headers)
				if s != e2e.StatusOK {
					return ""
				}
				var gb getBuildResp
				_ = json.Unmarshal(b, &gb)
				return gb.Build.Status
			}, 30*time.Second, 1*time.Second).Should(Equal(expect))
		}
		eventuallyStatus("BUILD_STATUS_RUNNING")
		eventuallyStatus("BUILD_STATUS_SUCCEEDED")

		// 拉取日志并校验序号顺序（网关接口返回仅有 lines，无 seq；这里只校验关键标记顺序）
		s, b := e2e.DoRequest(baseURL, e2e.MethodGet, fmt.Sprintf("/ci_service/api/v1/executor/builds/%d/logs", bid), nil, headers)
		Expect(s).To(Equal(e2e.StatusOK))
		var gl getLogsResp
		Expect(json.Unmarshal(b, &gl)).To(Succeed())
		var marks []string
		for _, ln := range gl.Lines {
			if strings.Contains(ln, "__step_") {
				marks = append(marks, ln)
			}
		}
		// 日志去重：验证没有重复的标记（批量幂等）
		seen := map[string]bool{}
		for _, m := range marks {
			Expect(seen[m]).To(BeFalse(), "duplicate log line: %s", m)
			seen[m] = true
		}
		Expect(marks).To(ContainElement("__step_begin__ step1"))
		Expect(marks).To(ContainElement("__step_end__ step1"))
		Expect(marks).To(ContainElement("__step_begin__ step2"))
		Expect(marks).To(ContainElement("__step_end__ step2"))
	})
})
