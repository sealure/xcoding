package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"crypto/sha256"
	execmodels "xcoding/apps/ci/executor_service/models"
	"xcoding/apps/ci/pipeline_service/internal/models"
	civ1 "xcoding/gen/go/ci/v1"
)

// ==== 构建相关操作 ====
func (s *pipelineService) StartPipelineBuild(ctx context.Context, req *civ1.StartPipelineBuildRequest) (*civ1.StartPipelineBuildResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "request nil")
	}
	var p models.Pipeline
	if err := s.db.WithContext(ctx).First(&p, req.GetPipelineId()).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "pipeline not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get pipeline: %v", err)
	}
	actorID, err := getUserIDFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	if !isUserRoleSuperAdmin(ctx) { // 允许项目成员触发构建
		ok, perr := s.isMemberOrHigher(ctx, p.ProjectID, actorID)
		if perr != nil {
			return nil, perr
		}
		if !ok {
			return nil, status.Errorf(codes.PermissionDenied, "not allowed to start build")
		}
	}

	// 校验变量：仅允许非空键与字符串值；限制映射大小
	validatedVars, verr := validateBuildVariables(req.GetVariables())
	if verr != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid variables: %v", verr)
	}

	// 直接在数据库创建构建记录（不再调用 Executor RPC）
	now := time.Now()
	b := execmodels.Build{
		PipelineID:  p.ID,
		Name:        p.Name,
		Status:      int32(civ1.BuildStatus_BUILD_STATUS_PENDING),
		TriggeredBy: req.GetTriggeredBy(),
		CommitSHA:   req.GetCommitSha(),
		Branch:      req.GetBranch(),
		CreatedAt:   now,
	}
	if b.TriggeredBy == "" {
		username, err := getUsernameFromCtx(ctx)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to get username: %v", err)
		}
		b.TriggeredBy = username
	}

	if err := s.db.WithContext(ctx).Create(&b).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create build: %v", err)
	}

	snap := execmodels.BuildSnapshot{
		BuildID:      b.ID,
		PipelineID:   b.PipelineID,
		Name:         b.Name,
		WorkflowYAML: p.WorkflowYAML,
		YamlSHA256:   sha256Hex(p.WorkflowYAML),
		CreatedAt:    now,
	}
	if err := s.db.WithContext(ctx).Create(&snap).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create snapshot: %v", err)
	}

	created := b.ToProto()

	q := getBuildQueue()
	if q == nil {
		return nil, status.Errorf(codes.FailedPrecondition, "build queue not configured")
	}
	if e := q.Enqueue(ctx, BuildJob{
		BuildID:    created.GetId(),
		PipelineID: p.ID,
		ProjectID:  p.ProjectID,
		CommitSHA:  created.GetCommitSha(),
		Branch:     created.GetBranch(),
		Variables:  validatedVars,
	}); e != nil {
		return nil, status.Errorf(codes.Internal, "enqueue failed: %v", e)
	}
	return &civ1.StartPipelineBuildResponse{Build: created}, nil
}

// validateBuildVariables 保证变量为字符串并控制合理上限
func validateBuildVariables(in map[string]string) (map[string]string, error) {
	if len(in) == 0 {
		return nil, nil
	}
	if len(in) > 100 {
		return nil, errors.New("too many variables; max 100")
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		if k == "" {
			return nil, errors.New("empty variable key")
		}
		if len(k) > 128 {
			return nil, errors.New("variable key too long")
		}
		if len(v) > 4096 {
			return nil, errors.New("variable value too long")
		}
		out[k] = v
	}
	return out, nil
}

// 队列与执行器最小接口及访问方法
type BuildJob struct {
	BuildID    uint64
	PipelineID uint64
	ProjectID  uint64
	CommitSHA  string
	Branch     string
	Variables  map[string]string
}

type BuildQueue interface {
	Enqueue(ctx context.Context, job BuildJob) error
}

var buildQueue BuildQueue

func getBuildQueue() BuildQueue  { return buildQueue }
func SetBuildQueue(q BuildQueue) { buildQueue = q }

// sha256Hex 计算内容的 SHA256 十六进制字符串（用于快照校验/去重提示）
func sha256Hex(s string) string {
	if s == "" {
		return ""
	}
	h := sha256.Sum256([]byte(s))
	return fmt.Sprintf("%x", h[:])
}

// GetWorkflowSnapshotByBuildID 便捷函数：按 build_id 获取 YAML 快照
// 注意：不暴露到 proto，供内部执行器或调试使用。
func (s *pipelineService) GetWorkflowSnapshotByBuildID(ctx context.Context, buildID uint64) (*execmodels.BuildSnapshot, error) {
	var snap execmodels.BuildSnapshot
	if err := s.db.WithContext(ctx).Where("build_id = ?", buildID).First(&snap).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "snapshot not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get snapshot: %v", err)
	}
	return &snap, nil
}
