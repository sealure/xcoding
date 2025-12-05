//go:build e2e
// +build e2e

package ci_e2e

import (
    "encoding/json"
    "fmt"

    "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"

    e2e "xcoding/e2e/helpers"
)

// 响应解析结构（与服务返回字段保持最小依赖）
type createPipelineResp struct { Pipeline struct { ID e2e.FlexID `json:"id"` } `json:"pipeline"` }
type getPipelineResp struct { Pipeline struct { ID e2e.FlexID `json:"id"`; Name string `json:"name"` } `json:"pipeline"` }
type listPipelinesResp struct { Data []struct { ID e2e.FlexID `json:"id"` } `json:"data"` }

type createScheduleResp struct { Schedule struct { ID e2e.FlexID `json:"id"` } `json:"schedule"` }
type listSchedulesResp struct { Data []struct { ID e2e.FlexID `json:"id"` } `json:"data"` }
type updateScheduleResp struct { Schedule struct { ID e2e.FlexID `json:"id"`; Enabled bool `json:"enabled"` } `json:"schedule"` }
type deleteScheduleResp struct { Success bool `json:"success"` }

type startBuildResp struct { Build struct { ID e2e.FlexID `json:"id"`; PipelineID e2e.FlexID `json:"pipeline_id"` } `json:"build"` }
type getBuildResp struct { Build struct { ID e2e.FlexID `json:"id"`; Status string `json:"status"` } `json:"build"` }
type listBuildsResp struct { Data []struct { ID e2e.FlexID `json:"id"` } `json:"data"` }
type cancelBuildResp struct { Success bool `json:"success"` }

var _ = ginkgo.Describe("CI Pipeline/Schedule/Build", func() {
    var baseURL string

    ginkgo.BeforeEach(func() {
        baseURL = e2e.BaseURL()
        // 路由探测（网关与生成路径），环境未就绪时直接失败以暴露问题
        Expect(e2e.PingGateway(baseURL)).To(BeTrue(), fmt.Sprintf("gateway not reachable: %s", baseURL))
        // 至少确认一个路径存在以快速失败
        Expect(e2e.RouteExistsWithMethod(baseURL, e2e.MethodPost, "/ci_service/api/v1/pipelines")).To(BeTrue(), "CreatePipeline route not available at gateway")
    })

    ginkgo.It("creates pipeline, manages schedules, and handles builds", func() {
        // 1) 超管登录，方便跨项目操作与权限校验
        token, err := e2e.AdminLogin()
        Expect(err).NotTo(HaveOccurred())
        headers := e2e.AuthHeader(token)

        suf := e2e.UniqueNano()

        // 2) 创建项目（供流水线归属）
        pid, err := e2e.CreateProject(fmt.Sprintf("ci_proj_%d", suf), "ci e2e project", false, token)
        Expect(err).NotTo(HaveOccurred())
        Expect(pid).ToNot(BeZero())

        // 3) 创建流水线
        status, body := e2e.DoRequest(baseURL, e2e.MethodPost, "/ci_service/api/v1/pipelines", map[string]any{
            "name":          fmt.Sprintf("pipeline_%d", suf),
            "description":   "e2e pipeline",
            "project_id":    pid,
            "workflow_yaml": "name: E2E\nsteps:\n  - run: echo hello",
            "is_active":     true,
        }, headers)
        Expect(status).To(Equal(e2e.StatusOK), "CreatePipeline failed: %s", string(body))
        var cp createPipelineResp
        Expect(json.Unmarshal(body, &cp)).To(Succeed())
        pipelineID := cp.Pipeline.ID
        Expect(pipelineID).ToNot(BeZero())

        // 4) 获取流水线详情
        status, body = e2e.DoRequest(baseURL, e2e.MethodGet, fmt.Sprintf("/ci_service/api/v1/pipelines/%d", pipelineID), nil, headers)
        Expect(status).To(Equal(e2e.StatusOK))
        var gp getPipelineResp
        Expect(json.Unmarshal(body, &gp)).To(Succeed())
        Expect(uint64(gp.Pipeline.ID)).To(Equal(uint64(pipelineID)))

        // 5) 列出流水线
        status, body = e2e.DoRequest(baseURL, e2e.MethodGet, "/ci_service/api/v1/pipelines", nil, headers)
        Expect(status).To(Equal(e2e.StatusOK))
        var lp listPipelinesResp
        Expect(json.Unmarshal(body, &lp)).To(Succeed())
        Expect(len(lp.Data)).To(BeNumerically(">=", 1))

        // ==== 调度管理 ====
        // 6) 创建调度
        status, body = e2e.DoRequest(baseURL, e2e.MethodPost, fmt.Sprintf("/ci_service/api/v1/pipelines/%d/schedules", pipelineID), map[string]any{
            "cron":     "0 2 1 * *",
            "timezone": "UTC",
            "enabled":  true,
        }, headers)
        Expect(status).To(Equal(e2e.StatusOK), "CreateSchedule failed: %s", string(body))
        var cs createScheduleResp
        Expect(json.Unmarshal(body, &cs)).To(Succeed())
        scheduleID := cs.Schedule.ID
        Expect(scheduleID).ToNot(BeZero())

        // 7) 列出调度
        status, body = e2e.DoRequest(baseURL, e2e.MethodGet, fmt.Sprintf("/ci_service/api/v1/pipelines/%d/schedules", pipelineID), nil, headers)
        Expect(status).To(Equal(e2e.StatusOK))
        var ls listSchedulesResp
        Expect(json.Unmarshal(body, &ls)).To(Succeed())
        Expect(len(ls.Data)).To(BeNumerically(">=", 1))

        // 8) 更新调度（禁用）
        status, body = e2e.DoRequest(baseURL, e2e.MethodPut, fmt.Sprintf("/ci_service/api/v1/pipelines/%d/schedules/%d", pipelineID, scheduleID), map[string]any{
            "enabled": false,
        }, headers)
        Expect(status).To(Equal(e2e.StatusOK))
        var us updateScheduleResp
        Expect(json.Unmarshal(body, &us)).To(Succeed())
        Expect(bool(us.Schedule.Enabled)).To(BeFalse())

        // 9) 删除调度
        status, body = e2e.DoRequest(baseURL, e2e.MethodDelete, fmt.Sprintf("/ci_service/api/v1/pipelines/%d/schedules/%d", pipelineID, scheduleID), nil, headers)
        Expect(status).To(SatisfyAny(Equal(e2e.StatusOK), Equal(e2e.StatusNoContent)))
        // 响应可能为空或 {success:true}
        if len(body) > 0 {
            var ds deleteScheduleResp
            _ = json.Unmarshal(body, &ds)
        }

        // ==== 构建触发与查询 ====
        // 10) 触发构建
        status, body = e2e.DoRequest(baseURL, e2e.MethodPost, fmt.Sprintf("/ci_service/api/v1/pipelines/%d/builds", pipelineID), map[string]any{
            "branch":       "main",
            "triggered_by": "e2e",
            "variables":    map[string]string{"key": "value"},
        }, headers)
        Expect(status).To(Equal(e2e.StatusOK), "StartBuild failed: %s", string(body))
        var sb startBuildResp
        Expect(json.Unmarshal(body, &sb)).To(Succeed())
        buildID := sb.Build.ID
        Expect(buildID).ToNot(BeZero())

        // 11) 获取构建详情
        status, body = e2e.DoRequest(baseURL, e2e.MethodGet, fmt.Sprintf("/ci_service/api/v1/builds/%d", buildID), nil, headers)
        Expect(status).To(Equal(e2e.StatusOK))
        var gb getBuildResp
        Expect(json.Unmarshal(body, &gb)).To(Succeed())
        Expect(uint64(gb.Build.ID)).To(Equal(uint64(buildID)))

        // 12) 列出构建
        status, body = e2e.DoRequest(baseURL, e2e.MethodGet, fmt.Sprintf("/ci_service/api/v1/pipelines/%d/builds", pipelineID), nil, headers)
        Expect(status).To(Equal(e2e.StatusOK))
        var lb listBuildsResp
        Expect(json.Unmarshal(body, &lb)).To(Succeed())
        Expect(len(lb.Data)).To(BeNumerically(">=", 1))

        // 13) 取消构建（可选）
        status, body = e2e.DoRequest(baseURL, e2e.MethodPost, fmt.Sprintf("/ci_service/api/v1/builds/%d/cancel", buildID), nil, headers)
        Expect(status).To(SatisfyAny(Equal(e2e.StatusOK), Equal(e2e.StatusNoContent)))
        if len(body) > 0 {
            var cb cancelBuildResp
            _ = json.Unmarshal(body, &cb)
        }

        // 14) 删除流水线（清理）
        status, body = e2e.DoRequest(baseURL, e2e.MethodDelete, fmt.Sprintf("/ci_service/api/v1/pipelines/%d", pipelineID), nil, headers)
        Expect(status).To(SatisfyAny(Equal(e2e.StatusOK), Equal(e2e.StatusNoContent)))
        _ = body
    })
})