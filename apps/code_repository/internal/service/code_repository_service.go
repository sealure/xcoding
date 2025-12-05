package service

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"xcoding/apps/code_repository/internal/models"
	coderepositoryv1 "xcoding/gen/go/code_repository/v1"
	projectv1 "xcoding/gen/go/project/v1"
	"xcoding/pkg/auth"
)

// CodeRepositoryService 定义代码仓库的业务逻辑接口。
// CRUD 方法在项目作用域内工作。
type CodeRepositoryService interface {
	CreateRepository(ctx context.Context, projectID uint64, name, description, gitURL, branch string, authType coderepositoryv1.RepositoryAuthType, gitUsername, gitPassword, gitSSHKey string) (*coderepositoryv1.Repository, error)
	GetRepository(ctx context.Context, projectID, repositoryID uint64) (*coderepositoryv1.Repository, error)
	ListRepositories(ctx context.Context, projectID uint64, page, pageSize int32) ([]*coderepositoryv1.Repository, int32, int32, error)
	UpdateRepository(ctx context.Context, projectID, repositoryID uint64, updates map[string]interface{}) (*coderepositoryv1.Repository, error)
	DeleteRepository(ctx context.Context, projectID, repositoryID uint64) error
	TestRepositoryConnection(ctx context.Context, projectID, repositoryID uint64) (bool, string, error)
	GetRepositoryBranches(ctx context.Context, projectID, repositoryID uint64, page, pageSize int32) ([]string, int32, int32, error)

	// Branch CRUD
	CreateBranch(ctx context.Context, projectID, repositoryID uint64, name string, isDefault bool) (*coderepositoryv1.Branch, error)
	GetBranch(ctx context.Context, projectID, repositoryID, branchID uint64) (*coderepositoryv1.Branch, error)
	ListBranches(ctx context.Context, projectID, repositoryID uint64, page, pageSize int32) ([]*coderepositoryv1.Branch, int32, int32, error)
	UpdateBranch(ctx context.Context, projectID, repositoryID, branchID uint64, name string, isDefault bool) (*coderepositoryv1.Branch, error)
	DeleteBranch(ctx context.Context, projectID, repositoryID, branchID uint64) error

	// Commit CRUD
	CreateCommit(ctx context.Context, projectID, repositoryID, branchID uint64, hash, message, authorName, authorEmail string, authoredAt *time.Time, committerName, committerEmail string, committedAt *time.Time) (*coderepositoryv1.Commit, error)
	GetCommit(ctx context.Context, projectID, repositoryID, commitID uint64) (*coderepositoryv1.Commit, error)
	ListCommits(ctx context.Context, projectID, repositoryID, branchID uint64, page, pageSize int32) ([]*coderepositoryv1.Commit, int32, int32, error)
	UpdateCommit(ctx context.Context, projectID, repositoryID, commitID uint64, message string) (*coderepositoryv1.Commit, error)
	DeleteCommit(ctx context.Context, projectID, repositoryID, commitID uint64) error
}

type codeRepositoryService struct {
	db            *gorm.DB
	projectClient projectv1.ProjectServiceClient
}

func NewCodeRepositoryService(db *gorm.DB, projectClient projectv1.ProjectServiceClient) CodeRepositoryService {
	return &codeRepositoryService{db: db, projectClient: projectClient}
}

func (s *codeRepositoryService) CreateRepository(ctx context.Context, projectID uint64, name, description, gitURL, branch string, authType coderepositoryv1.RepositoryAuthType, gitUsername, gitPassword, gitSSHKey string) (*coderepositoryv1.Repository, error) {
	// 权限：仅项目所有者或管理员可创建
	actorID, err := getUserIDFromCtx(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "missing user: %v", err)
	}
	if err := s.ensureOwnerOrAdmin(ctx, projectID, actorID); err != nil {
		return nil, err
	}

	// 唯一性：同一项目下仓库名称唯一
	var existing models.Repository
	if err := s.db.WithContext(ctx).Where("project_id = ? AND name = ?", projectID, name).First(&existing).Error; err == nil {
		return nil, status.Errorf(codes.AlreadyExists, "repository name already exists in this project")
	} else if err != gorm.ErrRecordNotFound {
		return nil, status.Errorf(codes.Internal, "failed to check existing repository: %v", err)
	}

	repo := models.Repository{
		ProjectID:   projectID,
		Name:        name,
		Description: description,
		GitURL:      gitURL,
		AuthType:    int32(authType),
		GitUsername: gitUsername,
		GitPassword: gitPassword,
		GitSSHKey:   gitSSHKey,
		IsActive:    true,
		SyncStatus:  int32(coderepositoryv1.RepositorySyncStatus_REPOSITORY_SYNC_STATUS_PENDING),
	}

	if err := s.db.WithContext(ctx).Create(&repo).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create repository: %v", err)
	}

	// 在分支表创建默认分支记录
	if branch != "" {
		if err := s.db.WithContext(ctx).Create(&models.RepositoryBranch{RepositoryID: repo.ID, Name: branch, IsDefault: true}).Error; err != nil {
			return nil, status.Errorf(codes.Internal, "failed to create default branch: %v", err)
		}
	}

	// 转换前清理敏感字段（避免明文返回）
	repo.GitPassword = ""
	repo.GitSSHKey = ""
	return models.ToProtoWithBranch(&repo, branch), nil
}

func (s *codeRepositoryService) GetRepository(ctx context.Context, projectID, repositoryID uint64) (*coderepositoryv1.Repository, error) {
	// 权限：超级管理员或项目成员（含所有者/管理员）可查看
	actorID, err := getUserIDFromCtx(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "missing user: %v", err)
	}
	if !isUserRoleSuperAdmin(ctx) {
		ok, perr := s.isMemberOrHigher(ctx, projectID, actorID)
		if perr != nil {
			return nil, perr
		}
		if !ok {
			return nil, status.Errorf(codes.PermissionDenied, "only project members can view repository")
		}
	}

	var repo models.Repository
	if err := s.db.WithContext(ctx).Where("id = ? AND project_id = ?", repositoryID, projectID).First(&repo).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "repository not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get repository: %v", err)
	}

	// 查询默认分支
	branchName := ""
	var def models.RepositoryBranch
	if err := s.db.WithContext(ctx).Where("repository_id = ? AND is_default = ?", repo.ID, true).First(&def).Error; err == nil {
		branchName = def.Name
	}

	repo.GitPassword = ""
	repo.GitSSHKey = ""
	return models.ToProtoWithBranch(&repo, branchName), nil
}

func (s *codeRepositoryService) UpdateRepository(ctx context.Context, projectID, repositoryID uint64, updates map[string]interface{}) (*coderepositoryv1.Repository, error) {
	// 权限：仅项目所有者或管理员可更新
	actorID, err := getUserIDFromCtx(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "missing user: %v", err)
	}
	if err := s.ensureOwnerOrAdmin(ctx, projectID, actorID); err != nil {
		return nil, err
	}

	var repo models.Repository
	if err := s.db.WithContext(ctx).Where("id = ? AND project_id = ?", repositoryID, projectID).First(&repo).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "repository not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get repository: %v", err)
	}

	if name, ok := updates["name"].(string); ok && name != "" && name != repo.Name {
		var existing models.Repository
		if err := s.db.WithContext(ctx).Where("project_id = ? AND name = ? AND id != ?", projectID, name, repositoryID).First(&existing).Error; err == nil {
			return nil, status.Errorf(codes.AlreadyExists, "repository name already exists in this project")
		}
	}

	if v, ok := updates["name"].(string); ok && v != "" {
		repo.Name = v
	}
	if v, ok := updates["description"].(string); ok {
		repo.Description = v
	}
	if v, ok := updates["git_url"].(string); ok && v != "" {
		repo.GitURL = v
	}
	if v, ok := updates["auth_type"].(coderepositoryv1.RepositoryAuthType); ok && v != coderepositoryv1.RepositoryAuthType_REPOSITORY_AUTH_TYPE_UNSPECIFIED {
		repo.AuthType = int32(v)
	}
	if v, ok := updates["git_username"].(string); ok {
		repo.GitUsername = v
	}
	if v, ok := updates["git_password"].(string); ok {
		repo.GitPassword = v
	}
	if v, ok := updates["git_ssh_key"].(string); ok {
		repo.GitSSHKey = v
	}
	if v, ok := updates["is_active"].(bool); ok {
		repo.IsActive = v
	}

	// 更新默认分支（由分支表维护）
	if v, ok := updates["branch"].(string); ok && v != "" {
		// 重置所有默认分支标记
		if err := s.db.WithContext(ctx).Model(&models.RepositoryBranch{}).Where("repository_id = ?", repo.ID).Update("is_default", false).Error; err != nil {
			return nil, status.Errorf(codes.Internal, "failed to reset default branch: %v", err)
		}
		// 设置指定分支为默认（存在则更新，不存在则创建）
		var br models.RepositoryBranch
		if err := s.db.WithContext(ctx).Where("repository_id = ? AND name = ?", repo.ID, v).First(&br).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				br = models.RepositoryBranch{RepositoryID: repo.ID, Name: v, IsDefault: true}
				if err := s.db.WithContext(ctx).Create(&br).Error; err != nil {
					return nil, status.Errorf(codes.Internal, "failed to set default branch: %v", err)
				}
			} else {
				return nil, status.Errorf(codes.Internal, "failed to get branch: %v", err)
			}
		} else {
			br.IsDefault = true
			if err := s.db.WithContext(ctx).Save(&br).Error; err != nil {
				return nil, status.Errorf(codes.Internal, "failed to set default branch: %v", err)
			}
		}
	}

	if err := s.db.WithContext(ctx).Save(&repo).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update repository: %v", err)
	}

	repo.GitPassword = ""
	repo.GitSSHKey = ""
	// 回填默认分支
	branchName := ""
	var def models.RepositoryBranch
	if err := s.db.WithContext(ctx).Where("repository_id = ? AND is_default = ?", repo.ID, true).First(&def).Error; err == nil {
		branchName = def.Name
	}
	return models.ToProtoWithBranch(&repo, branchName), nil
}

func (s *codeRepositoryService) DeleteRepository(ctx context.Context, projectID, repositoryID uint64) error {
	// 权限：仅项目所有者或管理员可删除
	actorID, err := getUserIDFromCtx(ctx)
	if err != nil {
		return status.Errorf(codes.Unauthenticated, "missing user: %v", err)
	}
	if err := s.ensureOwnerOrAdmin(ctx, projectID, actorID); err != nil {
		return err
	}

	var repo models.Repository
	if err := s.db.WithContext(ctx).Where("id = ? AND project_id = ?", repositoryID, projectID).First(&repo).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return status.Errorf(codes.NotFound, "repository not found")
		}
		return status.Errorf(codes.Internal, "failed to get repository: %v", err)
	}
	if err := s.db.WithContext(ctx).Delete(&repo).Error; err != nil {
		return status.Errorf(codes.Internal, "failed to delete repository: %v", err)
	}
	return nil
}

// ==== 权限相关辅助函数 ====
func getUserIDFromCtx(ctx context.Context) (uint64, error) { return auth.GetUserIDFromCtx(ctx) }
func isUserRoleSuperAdmin(ctx context.Context) bool        { return auth.IsUserRoleSuperAdmin(ctx) }

// isMemberOrHigher 检查操作者是否为项目所有者、管理员或成员
func (s *codeRepositoryService) isMemberOrHigher(ctx context.Context, projectID uint64, actorID uint64) (bool, error) {
	resp, err := s.projectClient.GetProject(ctx, &projectv1.GetProjectRequest{ProjectId: projectID})
	if err != nil {
		return false, status.Errorf(codes.Internal, "failed to get project: %v", err)
	}

	p := resp.GetProject()
	if p == nil {
		return false, status.Errorf(codes.NotFound, "project not found")
	}

	if p.OwnerId == actorID {
		return true, nil
	}
	members, err := s.projectClient.ListProjectMembers(ctx, &projectv1.ListProjectMembersRequest{ProjectId: projectID})
	if err != nil {
		return false, status.Errorf(codes.Internal, "failed to list project members: %v", err)
	}
	for _, m := range members.GetData() {
		if m.GetUserId() == actorID {
			return true, nil
		}
	}
	return false, nil
}

func (s *codeRepositoryService) ensureOwnerOrAdmin(ctx context.Context, projectID uint64, actorID uint64) error {
	if isUserRoleSuperAdmin(ctx) {
		return nil
	}
	resp, err := s.projectClient.GetProject(ctx, &projectv1.GetProjectRequest{ProjectId: projectID})
	if err != nil {
		return status.Errorf(codes.Internal, "failed to get project: %v", err)
	}
	p := resp.GetProject()
	if p == nil {
		return status.Errorf(codes.NotFound, "project not found")
	}
	if p.OwnerId == actorID {
		return nil
	}
	members, err := s.projectClient.ListProjectMembers(ctx, &projectv1.ListProjectMembersRequest{ProjectId: projectID})
	if err != nil {
		return status.Errorf(codes.Internal, "failed to list project members: %v", err)
	}
	for _, m := range members.GetData() {
		if m.GetUserId() == actorID {
			role := m.GetRole()
			if role == projectv1.ProjectMemberRole_PROJECT_MEMBER_ROLE_OWNER || role == projectv1.ProjectMemberRole_PROJECT_MEMBER_ROLE_ADMIN {
				return nil
			}
		}
	}
	return status.Errorf(codes.PermissionDenied, "only owner or admin can perform this action")
}

func (s *codeRepositoryService) ListRepositories(ctx context.Context, projectID uint64, page, pageSize int32) ([]*coderepositoryv1.Repository, int32, int32, error) {
	// 服务层分页参数兜底与限制，避免绕过处理器导致大查询
	const maxPageSize int32 = 100
	if page < 1 {
		page = 1
	} // page 从 1 开始
	if pageSize < 1 {
		pageSize = 10
	} // 默认每页 10
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	} // 上限保护

	// Permission: project members can list. Super admin bypasses membership check.
	if !isUserRoleSuperAdmin(ctx) {
		actorID, err := getUserIDFromCtx(ctx)
		if err != nil {
			return nil, 0, 0, status.Errorf(codes.Unauthenticated, "missing user: %v", err)
		}
		ok, perr := s.isMemberOrHigher(ctx, projectID, actorID)
		if perr != nil {
			return nil, 0, 0, perr
		}
		if !ok {
			return nil, 0, 0, status.Errorf(codes.PermissionDenied, "only project members can list repositories")
		}
	}

	// 计算偏移量
	offset := (page - 1) * pageSize

	// 统计总数
	var total int64
	if err := s.db.WithContext(ctx).Model(&models.Repository{}).Where("project_id = ?", projectID).Count(&total).Error; err != nil {
		return nil, 0, 0, status.Errorf(codes.Internal, "failed to count repositories: %v", err)
	}

	// 分页查询
	var repos []models.Repository
	if err := s.db.WithContext(ctx).Where("project_id = ?", projectID).Offset(int(offset)).Limit(int(pageSize)).Find(&repos).Error; err != nil {
		return nil, 0, 0, status.Errorf(codes.Internal, "failed to list repositories: %v", err)
	}

	// 转 proto 前清理敏感字段（不返回密码/私钥）并填充默认分支
	result := make([]*coderepositoryv1.Repository, 0, len(repos))
	for i := range repos {
		repos[i].GitPassword = ""
		repos[i].GitSSHKey = ""
		branchName := ""
		var def models.RepositoryBranch
		if err := s.db.WithContext(ctx).Where("repository_id = ? AND is_default = ?", repos[i].ID, true).First(&def).Error; err == nil {
			branchName = def.Name
		}
		result = append(result, models.ToProtoWithBranch(&repos[i], branchName))
	}

	// 计算总页数
	totalItems := int32(total)
	totalPages := int32((total + int64(pageSize) - 1) / int64(pageSize))
	return result, totalItems, totalPages, nil
}

func (s *codeRepositoryService) TestRepositoryConnection(ctx context.Context, projectID, repositoryID uint64) (bool, string, error) {
	var repo models.Repository
	if err := s.db.WithContext(ctx).Where("id = ? AND project_id = ?", repositoryID, projectID).First(&repo).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, "repository not found", status.Errorf(codes.NotFound, "repository not found")
		}
		return false, "failed to get repository", status.Errorf(codes.Internal, "failed to get repository: %v", err)
	}

	_, err := runGitLsRemote(ctx, repo.GitURL, coderepositoryv1.RepositoryAuthType(repo.AuthType), repo.GitUsername, repo.GitPassword, repo.GitSSHKey)
	if err != nil {
		return false, fmt.Sprintf("connection failed: %v", err), nil
	}
	return true, "connection successful", nil
}

func (s *codeRepositoryService) GetRepositoryBranches(ctx context.Context, projectID, repositoryID uint64, page, pageSize int32) ([]string, int32, int32, error) {
	// 服务层分页参数兜底与限制，避免绕过处理器导致大查询
	const maxPageSize int32 = 100
	if page < 1 {
		page = 1
	} // page 从 1 开始
	if pageSize < 1 {
		pageSize = 10
	} // 默认每页 10
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	} // 上限保护

	// 获取仓库信息
	var repo models.Repository
	if err := s.db.WithContext(ctx).Where("id = ? AND project_id = ?", repositoryID, projectID).First(&repo).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, 0, 0, status.Errorf(codes.NotFound, "repository not found")
		}
		return nil, 0, 0, status.Errorf(codes.Internal, "failed to get repository: %v", err)
	}

	// 通过 git ls-remote 获取分支；失败时使用默认分支兜底
	out, err := runGitLsRemote(ctx, repo.GitURL, coderepositoryv1.RepositoryAuthType(repo.AuthType), repo.GitUsername, repo.GitPassword, repo.GitSSHKey)
	if err != nil {
		branches := []string{}
		var def models.RepositoryBranch
		if err2 := s.db.WithContext(ctx).Where("repository_id = ? AND is_default = ?", repo.ID, true).First(&def).Error; err2 == nil && def.Name != "" {
			branches = []string{def.Name}
		}
		// 分页裁剪（兜底场景）
		totalItems := int32(len(branches))
		totalPages := int32((totalItems + pageSize - 1) / pageSize)
		start := int((page - 1) * pageSize)
		if start > int(totalItems) {
			return []string{}, totalItems, totalPages, nil
		}
		end := start + int(pageSize)
		if end > int(totalItems) {
			end = int(totalItems)
		}
		return branches[start:end], totalItems, totalPages, nil
	}

	allBranches := parseBranches(out)
	if len(allBranches) == 0 {
		var def models.RepositoryBranch
		if err2 := s.db.WithContext(ctx).Where("repository_id = ? AND is_default = ?", repo.ID, true).First(&def).Error; err2 == nil && def.Name != "" {
			allBranches = []string{def.Name}
		}
	}

	// 分页裁剪（正常场景）
	totalItems := int32(len(allBranches))
	totalPages := int32((totalItems + pageSize - 1) / pageSize)
	start := int((page - 1) * pageSize)
	if start > int(totalItems) {
		return []string{}, totalItems, totalPages, nil
	}
	end := start + int(pageSize)
	if end > int(totalItems) {
		end = int(totalItems)
	}
	return allBranches[start:end], totalItems, totalPages, nil
}

func runGitLsRemote(ctx context.Context, gitURL string, authType coderepositoryv1.RepositoryAuthType, username, password, sshKey string) (string, error) {
	// 根据认证类型构造 git 调用参数与环境变量
	// - 密码认证：将用户名和密码注入到 URL（仅内存使用，不持久化）
	// - SSH 认证：写入临时私钥文件，并通过 GIT_SSH_COMMAND 指定；StrictHostKeyChecking=no 避免首次连接失败
	// 注意：私钥仅写入临时文件，函数返回后即删除，避免泄露
	cmdArgs := []string{"ls-remote", "--heads"}
	var env []string
	urlToUse := gitURL

	switch authType {
	case coderepositoryv1.RepositoryAuthType_REPOSITORY_AUTH_TYPE_NONE:
		// no change
	case coderepositoryv1.RepositoryAuthType_REPOSITORY_AUTH_TYPE_PASSWORD:
		if username != "" && password != "" {
			u, err := url.Parse(gitURL)
			if err == nil {
				u.User = url.UserPassword(username, password)
				urlToUse = u.String()
			}
		}
	case coderepositoryv1.RepositoryAuthType_REPOSITORY_AUTH_TYPE_SSH:
		if sshKey != "" {
			// write temp private key
			tmpFile, err := os.CreateTemp("", "git-ssh-key-*")
			if err != nil {
				return "", fmt.Errorf("create temp ssh key failed: %w", err)
			}
			defer os.Remove(tmpFile.Name())
			if _, err := tmpFile.WriteString(sshKey); err != nil {
				return "", fmt.Errorf("write ssh key failed: %w", err)
			}
			_ = tmpFile.Close()
			_ = os.Chmod(tmpFile.Name(), 0600)
			sshCmd := fmt.Sprintf("ssh -i %s -o StrictHostKeyChecking=no", tmpFile.Name())
			env = append(os.Environ(), "GIT_SSH_COMMAND="+sshCmd)
		}
	}

	cmdArgs = append(cmdArgs, urlToUse)
	cmd := exec.CommandContext(ctx, "git", cmdArgs...)
	if len(env) > 0 {
		cmd.Env = env
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git ls-remote failed: %v, output: %s", err, string(output))
	}
	return string(output), nil
}

func parseBranches(output string) []string {
	// 解析 git ls-remote 输出，提取分支名称（refs/heads/<branch>），并过滤空行
	lines := strings.Split(output, "\n")
	branches := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 2 && strings.HasPrefix(parts[1], "refs/heads/") {
			branches = append(branches, strings.TrimPrefix(parts[1], "refs/heads/"))
		}
	}
	return branches
}

// ==== Branch CRUD ====
func (s *codeRepositoryService) CreateBranch(ctx context.Context, projectID, repositoryID uint64, name string, isDefault bool) (*coderepositoryv1.Branch, error) {
	actorID, err := getUserIDFromCtx(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "missing user: %v", err)
	}
	if err := s.ensureOwnerOrAdmin(ctx, projectID, actorID); err != nil {
		return nil, err
	}

	var repo models.Repository
	if err := s.db.WithContext(ctx).Where("id = ? AND project_id = ?", repositoryID, projectID).First(&repo).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "repository not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get repository: %v", err)
	}
	var existing models.RepositoryBranch
	if err := s.db.WithContext(ctx).Where("repository_id = ? AND name = ?", repositoryID, name).First(&existing).Error; err == nil {
		return nil, status.Errorf(codes.AlreadyExists, "branch name already exists in this repository")
	} else if err != gorm.ErrRecordNotFound {
		return nil, status.Errorf(codes.Internal, "failed to check existing branch: %v", err)
	}
	br := models.RepositoryBranch{RepositoryID: repositoryID, Name: name, IsDefault: false}
	if isDefault {
		if err := s.db.WithContext(ctx).Model(&models.RepositoryBranch{}).Where("repository_id = ?", repositoryID).Update("is_default", false).Error; err != nil {
			return nil, status.Errorf(codes.Internal, "failed to reset default branch: %v", err)
		}
		br.IsDefault = true
	}
	if err := s.db.WithContext(ctx).Create(&br).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create branch: %v", err)
	}
	return models.ToProtoBranch(&br), nil
}

func (s *codeRepositoryService) GetBranch(ctx context.Context, projectID, repositoryID, branchID uint64) (*coderepositoryv1.Branch, error) {
	actorID, err := getUserIDFromCtx(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "missing user: %v", err)
	}
	if !isUserRoleSuperAdmin(ctx) {
		ok, perr := s.isMemberOrHigher(ctx, projectID, actorID)
		if perr != nil {
			return nil, perr
		}
		if !ok {
			return nil, status.Errorf(codes.PermissionDenied, "only project members can view branch")
		}
	}
	var br models.RepositoryBranch
	if err := s.db.WithContext(ctx).Where("id = ? AND repository_id = ?", branchID, repositoryID).First(&br).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "branch not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get branch: %v", err)
	}
	var repo models.Repository
	if err := s.db.WithContext(ctx).Where("id = ? AND project_id = ?", repositoryID, projectID).First(&repo).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "repository not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get repository: %v", err)
	}
	return models.ToProtoBranch(&br), nil
}

func (s *codeRepositoryService) ListBranches(ctx context.Context, projectID, repositoryID uint64, page, pageSize int32) ([]*coderepositoryv1.Branch, int32, int32, error) {
	const maxPageSize int32 = 100
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	actorID, err := getUserIDFromCtx(ctx)
	if err != nil {
		return nil, 0, 0, status.Errorf(codes.Unauthenticated, "missing user: %v", err)
	}
	if !isUserRoleSuperAdmin(ctx) {
		ok, perr := s.isMemberOrHigher(ctx, projectID, actorID)
		if perr != nil {
			return nil, 0, 0, perr
		}
		if !ok {
			return nil, 0, 0, status.Errorf(codes.PermissionDenied, "only project members can list branches")
		}
	}
	var repo models.Repository
	if err := s.db.WithContext(ctx).Where("id = ? AND project_id = ?", repositoryID, projectID).First(&repo).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, 0, 0, status.Errorf(codes.NotFound, "repository not found")
		}
		return nil, 0, 0, status.Errorf(codes.Internal, "failed to get repository: %v", err)
	}
	offset := (page - 1) * pageSize
	var total int64
	if err := s.db.WithContext(ctx).Model(&models.RepositoryBranch{}).Where("repository_id = ?", repositoryID).Count(&total).Error; err != nil {
		return nil, 0, 0, status.Errorf(codes.Internal, "failed to count branches: %v", err)
	}
	var items []models.RepositoryBranch
	if err := s.db.WithContext(ctx).Where("repository_id = ?", repositoryID).Offset(int(offset)).Limit(int(pageSize)).Find(&items).Error; err != nil {
		return nil, 0, 0, status.Errorf(codes.Internal, "failed to list branches: %v", err)
	}
	result := make([]*coderepositoryv1.Branch, 0, len(items))
	for i := range items {
		result = append(result, models.ToProtoBranch(&items[i]))
	}
	return result, int32(total), int32((total + int64(pageSize) - 1) / int64(pageSize)), nil
}

func (s *codeRepositoryService) UpdateBranch(ctx context.Context, projectID, repositoryID, branchID uint64, name string, isDefault bool) (*coderepositoryv1.Branch, error) {
	actorID, err := getUserIDFromCtx(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "missing user: %v", err)
	}
	if err := s.ensureOwnerOrAdmin(ctx, projectID, actorID); err != nil {
		return nil, err
	}
	var br models.RepositoryBranch
	if err := s.db.WithContext(ctx).Where("id = ? AND repository_id = ?", branchID, repositoryID).First(&br).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "branch not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get branch: %v", err)
	}
	if name != "" && name != br.Name {
		var existing models.RepositoryBranch
		if err := s.db.WithContext(ctx).Where("repository_id = ? AND name = ? AND id != ?", repositoryID, name, branchID).First(&existing).Error; err == nil {
			return nil, status.Errorf(codes.AlreadyExists, "branch name already exists in this repository")
		}
		br.Name = name
	}
	if isDefault {
		if err := s.db.WithContext(ctx).Model(&models.RepositoryBranch{}).Where("repository_id = ?", repositoryID).Update("is_default", false).Error; err != nil {
			return nil, status.Errorf(codes.Internal, "failed to reset default branch: %v", err)
		}
		br.IsDefault = true
	} else {
		br.IsDefault = false
	}
	if err := s.db.WithContext(ctx).Save(&br).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update branch: %v", err)
	}
	return models.ToProtoBranch(&br), nil
}

func (s *codeRepositoryService) DeleteBranch(ctx context.Context, projectID, repositoryID, branchID uint64) error {
	actorID, err := getUserIDFromCtx(ctx)
	if err != nil {
		return status.Errorf(codes.Unauthenticated, "missing user: %v", err)
	}
	if err := s.ensureOwnerOrAdmin(ctx, projectID, actorID); err != nil {
		return err
	}
	var br models.RepositoryBranch
	if err := s.db.WithContext(ctx).Where("id = ? AND repository_id = ?", branchID, repositoryID).First(&br).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return status.Errorf(codes.NotFound, "branch not found")
		}
		return status.Errorf(codes.Internal, "failed to get branch: %v", err)
	}
	if err := s.db.WithContext(ctx).Where("branch_id = ?", br.ID).Delete(&models.Commit{}).Error; err != nil {
		return status.Errorf(codes.Internal, "failed to delete commits: %v", err)
	}
	if err := s.db.WithContext(ctx).Delete(&br).Error; err != nil {
		return status.Errorf(codes.Internal, "failed to delete branch: %v", err)
	}
	return nil
}

// ==== Commit CRUD ====
func (s *codeRepositoryService) CreateCommit(ctx context.Context, projectID, repositoryID, branchID uint64, hash, message, authorName, authorEmail string, authoredAt *time.Time, committerName, committerEmail string, committedAt *time.Time) (*coderepositoryv1.Commit, error) {
	actorID, err := getUserIDFromCtx(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "missing user: %v", err)
	}
	if err := s.ensureOwnerOrAdmin(ctx, projectID, actorID); err != nil {
		return nil, err
	}
	var repo models.Repository
	if err := s.db.WithContext(ctx).Where("id = ? AND project_id = ?", repositoryID, projectID).First(&repo).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "repository not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get repository: %v", err)
	}
	var br models.RepositoryBranch
	if err := s.db.WithContext(ctx).Where("id = ? AND repository_id = ?", branchID, repositoryID).First(&br).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "branch not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get branch: %v", err)
	}
	var existing models.Commit
	if err := s.db.WithContext(ctx).Where("branch_id = ? AND hash = ?", branchID, hash).First(&existing).Error; err == nil {
		return nil, status.Errorf(codes.AlreadyExists, "commit already exists for this branch")
	} else if err != gorm.ErrRecordNotFound {
		return nil, status.Errorf(codes.Internal, "failed to check existing commit: %v", err)
	}
	cm := models.Commit{BranchID: branchID, Hash: hash, Message: message, AuthorName: authorName, AuthorEmail: authorEmail, CommitterName: committerName, CommitterEmail: committerEmail}
	if authoredAt != nil {
		cm.AuthoredAt = authoredAt
	}
	if committedAt != nil {
		cm.CommittedAt = committedAt
	}
	if err := s.db.WithContext(ctx).Create(&cm).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create commit: %v", err)
	}
	return models.ToProtoCommit(&cm), nil
}

func (s *codeRepositoryService) GetCommit(ctx context.Context, projectID, repositoryID, commitID uint64) (*coderepositoryv1.Commit, error) {
	actorID, err := getUserIDFromCtx(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "missing user: %v", err)
	}
	if !isUserRoleSuperAdmin(ctx) {
		ok, perr := s.isMemberOrHigher(ctx, projectID, actorID)
		if perr != nil {
			return nil, perr
		}
		if !ok {
			return nil, status.Errorf(codes.PermissionDenied, "only project members can view commit")
		}
	}
	var cm models.Commit
	if err := s.db.WithContext(ctx).Where("id = ?", commitID).First(&cm).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "commit not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get commit: %v", err)
	}
	var br models.RepositoryBranch
	if err := s.db.WithContext(ctx).Where("id = ? AND repository_id = ?", cm.BranchID, repositoryID).First(&br).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.PermissionDenied, "commit does not belong to repository")
		}
		return nil, status.Errorf(codes.Internal, "failed to verify commit branch: %v", err)
	}
	return models.ToProtoCommit(&cm), nil
}

func (s *codeRepositoryService) ListCommits(ctx context.Context, projectID, repositoryID, branchID uint64, page, pageSize int32) ([]*coderepositoryv1.Commit, int32, int32, error) {
	const maxPageSize int32 = 100
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	actorID, err := getUserIDFromCtx(ctx)
	if err != nil {
		return nil, 0, 0, status.Errorf(codes.Unauthenticated, "missing user: %v", err)
	}
	if !isUserRoleSuperAdmin(ctx) {
		ok, perr := s.isMemberOrHigher(ctx, projectID, actorID)
		if perr != nil {
			return nil, 0, 0, perr
		}
		if !ok {
			return nil, 0, 0, status.Errorf(codes.PermissionDenied, "only project members can list commits")
		}
	}
	var repo models.Repository
	if err := s.db.WithContext(ctx).Where("id = ? AND project_id = ?", repositoryID, projectID).First(&repo).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, 0, 0, status.Errorf(codes.NotFound, "repository not found")
		}
		return nil, 0, 0, status.Errorf(codes.Internal, "failed to get repository: %v", err)
	}
	var br models.RepositoryBranch
	if err := s.db.WithContext(ctx).Where("id = ? AND repository_id = ?", branchID, repositoryID).First(&br).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, 0, 0, status.Errorf(codes.NotFound, "branch not found")
		}
		return nil, 0, 0, status.Errorf(codes.Internal, "failed to get branch: %v", err)
	}
	offset := (page - 1) * pageSize
	var total int64
	if err := s.db.WithContext(ctx).Model(&models.Commit{}).Where("branch_id = ?", branchID).Count(&total).Error; err != nil {
		return nil, 0, 0, status.Errorf(codes.Internal, "failed to count commits: %v", err)
	}
	var items []models.Commit
	if err := s.db.WithContext(ctx).Where("branch_id = ?", branchID).Offset(int(offset)).Limit(int(pageSize)).Order("id DESC").Find(&items).Error; err != nil {
		return nil, 0, 0, status.Errorf(codes.Internal, "failed to list commits: %v", err)
	}
	result := make([]*coderepositoryv1.Commit, 0, len(items))
	for i := range items {
		result = append(result, models.ToProtoCommit(&items[i]))
	}
	return result, int32(total), int32((total + int64(pageSize) - 1) / int64(pageSize)), nil
}

func (s *codeRepositoryService) UpdateCommit(ctx context.Context, projectID, repositoryID, commitID uint64, message string) (*coderepositoryv1.Commit, error) {
	actorID, err := getUserIDFromCtx(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "missing user: %v", err)
	}
	if err := s.ensureOwnerOrAdmin(ctx, projectID, actorID); err != nil {
		return nil, err
	}
	var cm models.Commit
	if err := s.db.WithContext(ctx).Where("id = ?", commitID).First(&cm).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "commit not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get commit: %v", err)
	}
	var br models.RepositoryBranch
	if err := s.db.WithContext(ctx).Where("id = ? AND repository_id = ?", cm.BranchID, repositoryID).First(&br).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.PermissionDenied, "commit does not belong to repository")
		}
		return nil, status.Errorf(codes.Internal, "failed to verify commit branch: %v", err)
	}
	if message != "" {
		cm.Message = message
	}
	if err := s.db.WithContext(ctx).Save(&cm).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update commit: %v", err)
	}
	return models.ToProtoCommit(&cm), nil
}

func (s *codeRepositoryService) DeleteCommit(ctx context.Context, projectID, repositoryID, commitID uint64) error {
	actorID, err := getUserIDFromCtx(ctx)
	if err != nil {
		return status.Errorf(codes.Unauthenticated, "missing user: %v", err)
	}
	if err := s.ensureOwnerOrAdmin(ctx, projectID, actorID); err != nil {
		return err
	}
	var cm models.Commit
	if err := s.db.WithContext(ctx).Where("id = ?", commitID).First(&cm).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return status.Errorf(codes.NotFound, "commit not found")
		}
		return status.Errorf(codes.Internal, "failed to get commit: %v", err)
	}
	var br models.RepositoryBranch
	if err := s.db.WithContext(ctx).Where("id = ? AND repository_id = ?", cm.BranchID, repositoryID).First(&br).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return status.Errorf(codes.PermissionDenied, "commit does not belong to repository")
		}
		return status.Errorf(codes.Internal, "failed to verify commit branch: %v", err)
	}
	if err := s.db.WithContext(ctx).Delete(&cm).Error; err != nil {
		return status.Errorf(codes.Internal, "failed to delete commit: %v", err)
	}
	return nil
}
