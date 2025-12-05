//go:build e2e
// +build e2e

package ci_e2e

import (
    "encoding/json"
    "fmt"
    "os"

    "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"

    e2e "xcoding/e2e/helpers"
)

// 最小响应解析结构（避免与其它用例重名）
type createPipelinePendingResp struct { Pipeline struct { ID e2e.FlexID `json:"id"` } `json:"pipeline"` }
type startBuildPendingResp struct { Build struct { ID e2e.FlexID `json:"id"`; PipelineID e2e.FlexID `json:"pipeline_id"`; Status string `json:"status"` } `json:"build"` }
type getBuildPendingResp struct { Build struct { ID e2e.FlexID `json:"id"`; Status string `json:"status"` } `json:"build"` }

var _ = ginkgo.Describe("StartPipelineBuild returns PENDING", func() {
    var baseURL string

    ginkgo.BeforeEach(func() {
        baseURL = e2e.BaseURL()
        // 网关可达性与至少一个路由存在
        Expect(e2e.PingGateway(baseURL)).To(BeTrue(), fmt.Sprintf("gateway not reachable: %s", baseURL))
        Expect(e2e.RouteExistsWithMethod(baseURL, e2e.MethodPost, "/ci_service/api/v1/pipelines")).To(BeTrue(), "CreatePipeline route not available at gateway")
    })

    ginkgo.It("triggers build and asserts PENDING status", func() {
        // 超管登录，便于跨项目与权限校验
        token, err := e2e.AdminLogin()
        Expect(err).NotTo(HaveOccurred())
        headers := e2e.AuthHeader(token)

        suf := e2e.UniqueNano()

        // 创建项目（供流水线归属）
        pid, err := e2e.CreateProject(fmt.Sprintf("ci_proj_%d", suf), "ci e2e project for pending", false, token)
        Expect(err).NotTo(HaveOccurred())
        Expect(pid).ToNot(BeZero())

        // 读取示例流水线 YAML，与仓库中的 .workflows/example_pipeline.yml 保持一致
        var workflowYaml string
        candidates := []string{"../../.workflows/example_pipeline.yml", ".workflows/example_pipeline.yml"}
        for _, p := range candidates {
            if b, err := os.ReadFile(p); err == nil {
                workflowYaml = string(b)
                break
            }
        }
        Expect(workflowYaml).NotTo(BeEmpty(), fmt.Sprintf("cannot read example workflow yaml from paths: %v", candidates))

        // 创建流水线
        status, body := e2e.DoRequest(baseURL, e2e.MethodPost, "/ci_service/api/v1/pipelines", map[string]any{
            "name":          fmt.Sprintf("pipeline_%d", suf),
            "description":   "pending assertion",
            "project_id":    pid,
            "workflow_yaml": workflowYaml,
            "is_active":     true,
        }, headers)
        Expect(status).To(Equal(e2e.StatusOK), "CreatePipeline failed: %s", string(body))
        var cp createPipelinePendingResp
        Expect(json.Unmarshal(body, &cp)).To(Succeed())
        pipelineID := cp.Pipeline.ID
        Expect(pipelineID).ToNot(BeZero())

        // 触发构建（proto: StartPipelineBuild）
        status, body = e2e.DoRequest(baseURL, e2e.MethodPost, fmt.Sprintf("/ci_service/api/v1/pipelines/%d/builds", pipelineID), map[string]any{
            "branch":       "main",
            "triggered_by": "e2e-pending",
            "variables":    map[string]string{"k": "v"},
        }, headers)
        Expect(status).To(Equal(e2e.StatusOK), "StartPipelineBuild failed: %s", string(body))
        var sb startBuildPendingResp
        Expect(json.Unmarshal(body, &sb)).To(Succeed())
        Expect(sb.Build.ID).ToNot(BeZero())
        // 立即返回应为 PENDING（服务创建构建后入队）
        Expect(sb.Build.Status).To(Equal("BUILD_STATUS_PENDING"))

        // 再次通过 GetBuild 验证 PENDING
        status, body = e2e.DoRequest(baseURL, e2e.MethodGet, fmt.Sprintf("/ci_service/api/v1/builds/%d", sb.Build.ID), nil, headers)
        Expect(status).To(Equal(e2e.StatusOK))
        var gb getBuildPendingResp
        Expect(json.Unmarshal(body, &gb)).To(Succeed())
        Expect(uint64(gb.Build.ID)).To(Equal(uint64(sb.Build.ID)))
        Expect(gb.Build.Status).To(Equal("BUILD_STATUS_PENDING"))
    })
})