package service

import (
	"context"
	"fmt"
	"time"
	"xcoding/apps/ci/executor_service/internal/executor"
	"xcoding/apps/ci/executor_service/models"
	civ1 "xcoding/gen/go/ci/v1"

	"gorm.io/gorm"
)

type ExecutorService struct {
	civ1.UnimplementedExecutorServiceServer
	db *gorm.DB
}

// New 创建执行器服务实例
func New(db *gorm.DB) *ExecutorService { return &ExecutorService{db: db} }

// GetBuild 按 ID 查询执行器侧的构建（包含起止时间与状态）
func (s *ExecutorService) GetBuild(ctx context.Context, req *civ1.GetExecutorBuildRequest) (*civ1.GetExecutorBuildResponse, error) {
	var b models.Build
	if err := s.db.First(&b, req.GetBuildId()).Error; err != nil {
		return nil, err
	}

	// 加载 BuildSnapshot
	var snap models.BuildSnapshot
	pb := b.ToProto()
	if err := s.db.Where("build_id = ?", b.ID).First(&snap).Error; err == nil {
		// 找到了 snapshot，将 YAML 加入响应
		pb.Snapshot = snap.WorkflowYAML
	}

	return &civ1.GetExecutorBuildResponse{Build: pb}, nil
}

// ListBuilds 列出指定流水线的执行器构建列表（分页）
func (s *ExecutorService) ListBuilds(ctx context.Context, req *civ1.ListExecutorBuildsRequest) (*civ1.ListExecutorBuildsResponse, error) {
	page := req.GetPage()
	if page <= 0 {
		page = 1
	}
	size := req.GetPageSize()
	if size <= 0 {
		size = 20
	}
	offset := (page - 1) * size
	var total int64
	q := s.db.Model(&models.Build{}).Where("pipeline_id = ?", req.GetPipelineId())
	if err := q.Count(&total).Error; err != nil {
		return nil, err
	}
	var items []models.Build
	if err := q.Offset(int(offset)).Limit(int(size)).Order("id DESC").Find(&items).Error; err != nil {
		return nil, err
	}
	data := make([]*civ1.Build, len(items))
	for i := range items {
		data[i] = items[i].ToProto()
	}
	totalPages := int32((total + int64(size) - 1) / int64(size))
	return &civ1.ListExecutorBuildsResponse{Data: data, Pagination: &civ1.ListExecutorBuildsResponse_Pagination{Page: page, PageSize: size, TotalItems: int32(total), TotalPages: totalPages}}, nil
}

// GetBuildLogs 获取构建日志（偏移/限制），返回纯文本行数组
func (s *ExecutorService) GetBuildLogs(ctx context.Context, req *civ1.GetBuildLogsRequest) (*civ1.GetBuildLogsResponse, error) {
	offset := int(req.GetOffset())
	limit := int(req.GetLimit())
	if limit <= 0 {
		limit = 100
	}

	// 查询 BuildStepLogChunk，通过 JOIN BuildStep 过滤 build_id
	type LogResult struct {
		Content string
	}
	var results []LogResult
	err := s.db.Table("build_step_log_chunk c").
		Select("c.content").
		Joins("JOIN build_step s ON c.build_step_id = s.id").
		Where("s.build_id = ?", req.GetBuildId()).
		Order("c.id ASC").
		Offset(offset).
		Limit(limit).
		Find(&results).Error
	if err != nil {
		return nil, err
	}

	lines := make([]string, 0, len(results))
	for i := range results {
		lines = append(lines, results[i].Content)
	}
	next := req.GetOffset() + uint64(len(results))
	return &civ1.GetBuildLogsResponse{Lines: lines, NextOffset: next}, nil
}

// CancelBuild 取消未完成的构建：删除 K8s Job 并标记为 CANCELLED
func (s *ExecutorService) CancelBuild(ctx context.Context, req *civ1.CancelExecutorBuildRequest) (*civ1.CancelExecutorBuildResponse, error) {
	var b models.Build
	if err := s.db.First(&b, req.GetBuildId()).Error; err != nil {
		return nil, err
	}
	if civ1.BuildStatus(b.Status) == civ1.BuildStatus_BUILD_STATUS_SUCCEEDED || civ1.BuildStatus(b.Status) == civ1.BuildStatus_BUILD_STATUS_FAILED || civ1.BuildStatus(b.Status) == civ1.BuildStatus_BUILD_STATUS_CANCELLED {
		return &civ1.CancelExecutorBuildResponse{Success: false, Build: b.ToProto()}, nil
	}
	now := time.Now()
	// 删除 K8s Job
	// 注意：删除容忍 Job 不存在错误
	if err := s.deleteJob(ctx, b.ID); err != nil {
		fmt.Printf("cancel: delete job error: %v\n", err)
	}
	b.Status = int32(civ1.BuildStatus_BUILD_STATUS_CANCELLED)
	b.FinishedAt = &now
	if err := s.db.Save(&b).Error; err != nil {
		return nil, err
	}
	return &civ1.CancelExecutorBuildResponse{Success: true, Build: b.ToProto()}, nil
}

// deleteJob 尝试删除与构建关联的 K8s Job（命名规则：build-<id>）
func (s *ExecutorService) deleteJob(ctx context.Context, buildID uint64) error {
	env, err := executor.NewK8sEnv()
	if err != nil {
		return err
	}
	return env.CancelBuild(ctx, buildID)
}
