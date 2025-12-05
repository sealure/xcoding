// Artifact E2E helpers: build common resources (project, registry, namespace).

package helpers

import (
    "encoding/json"
    "fmt"
)

// CreateRegistryResp 统一解析结构
type CreateRegistryResp struct { Registry struct { ID FlexID `json:"id"` } `json:"registry"` }
type CreateNamespaceResp struct { Namespace struct { ID FlexID `json:"id"` } `json:"namespace"` }

// BuildProjectRegistryNamespace 以给定 token 构建项目/注册表/命名空间，返回三者 ID
// - 由调用者控制注册表是否 public
// - 命名空间默认 private（便于权限测试）
func BuildProjectRegistryNamespace(token string, projectName string) (FlexID, FlexID, FlexID, error) {
    // 1) 创建项目
    pid, err := CreateProject(projectName, "e2e project", false, token)
    if err != nil { return 0, 0, 0, err }

    // 2) 创建注册表（默认私有）
    status, body := DoRequest(BaseURL(), MethodPost, "/artifact_service/api/v1/registries", map[string]any{
        "name":        fmt.Sprintf("reg_%d", UniqueNano()),
        "url":         "https://registry.example.com",
        "description": "e2e registry",
        "is_public":   false,
        "username":    "u",
        "password":    "p",
        "project_id":  pid,
        // 默认类型配置，覆盖常用字段
        "artifact_type":  "ARTIFACT_TYPE_GENERIC_FILE",
        "artifact_source": "ARTIFACT_SOURCE_XCODING_REGISTRY",
    }, AuthHeader(token))
    if status != StatusOK { return 0, 0, 0, &HTTPError{Status: status, Body: string(body)} }
    var cr CreateRegistryResp
    if err := json.Unmarshal(body, &cr); err != nil { return 0, 0, 0, err }
    rid := cr.Registry.ID

    // 3) 创建命名空间（默认私有）
    status, body = DoRequest(BaseURL(), MethodPost, "/artifact_service/api/v1/namespaces", map[string]any{
        "registry_id": rid,
        "name":        fmt.Sprintf("ns_%d", UniqueNano()),
        "description": "e2e namespace",
        "is_public":   false,
    }, AuthHeader(token))
    if status != StatusOK { return 0, 0, 0, &HTTPError{Status: status, Body: string(body)} }
    var cn CreateNamespaceResp
    if err := json.Unmarshal(body, &cn); err != nil { return 0, 0, 0, err }
    nid := cn.Namespace.ID

    return pid, rid, nid, nil
}

// BuildRegistry 在指定项目下创建注册表（默认私有）
func BuildRegistry(token string, projectID FlexID) (FlexID, error) {
    status, body := DoRequest(BaseURL(), MethodPost, "/artifact_service/api/v1/registries", map[string]any{
        "name":        fmt.Sprintf("reg_%d", UniqueNano()),
        "url":         "https://registry.example.com",
        "description": "e2e registry",
        "is_public":   false,
        "username":    "u",
        "password":    "p",
        "project_id":  projectID,
    }, AuthHeader(token))
    if status != StatusOK { return 0, &HTTPError{Status: status, Body: string(body)} }
    var cr CreateRegistryResp
    if err := json.Unmarshal(body, &cr); err != nil { return 0, err }
    return cr.Registry.ID, nil
}

// BuildRegistryWithPublic 在指定项目下创建注册表，允许指定 is_public
func BuildRegistryWithPublic(token string, projectID FlexID, isPublic bool) (FlexID, error) {
    status, body := DoRequest(BaseURL(), MethodPost, "/artifact_service/api/v1/registries", map[string]any{
        "name":        fmt.Sprintf("reg_%d", UniqueNano()),
        "url":         "https://registry.example.com",
        "description": "e2e registry",
        "is_public":   isPublic,
        "username":    "u",
        "password":    "p",
        "project_id":  projectID,
    }, AuthHeader(token))
    if status != StatusOK { return 0, &HTTPError{Status: status, Body: string(body)} }
    var cr CreateRegistryResp
    if err := json.Unmarshal(body, &cr); err != nil { return 0, err }
    return cr.Registry.ID, nil
}

// BuildNamespace 在指定注册表下创建命名空间（默认私有）
func BuildNamespace(token string, registryID FlexID, name string) (FlexID, error) {
    if name == "" { name = fmt.Sprintf("ns_%d", UniqueNano()) }
    status, body := DoRequest(BaseURL(), MethodPost, "/artifact_service/api/v1/namespaces", map[string]any{
        "registry_id": registryID,
        "name":        name,
        "description": "e2e namespace",
        "is_public":   false,
    }, AuthHeader(token))
    if status != StatusOK { return 0, &HTTPError{Status: status, Body: string(body)} }
    var cn CreateNamespaceResp
    if err := json.Unmarshal(body, &cn); err != nil { return 0, err }
    return cn.Namespace.ID, nil
}

// BuildProjectAndRegistry 构建项目与注册表，不创建命名空间
func BuildProjectAndRegistry(token string, projectName string) (FlexID, FlexID, error) {
    // 1) 创建项目
    pid, err := CreateProject(projectName, "e2e project", false, token)
    if err != nil { return 0, 0, err }

    // 2) 创建注册表（默认私有）
    status, body := DoRequest(BaseURL(), MethodPost, "/artifact_service/api/v1/registries", map[string]any{
        "name":        fmt.Sprintf("reg_%d", UniqueNano()),
        "url":         "https://registry.example.com",
        "description": "e2e registry",
        "is_public":   false,
        "username":    "u",
        "password":    "p",
        "project_id":  pid,
    }, AuthHeader(token))
    if status != StatusOK { return 0, 0, &HTTPError{Status: status, Body: string(body)} }
    var cr CreateRegistryResp
    if err := json.Unmarshal(body, &cr); err != nil { return 0, 0, err }
    rid := cr.Registry.ID

    return pid, rid, nil
}

// BuildRepository 在指定命名空间下创建仓库
func BuildRepository(token string, namespaceID FlexID, name string, isPublic bool) (FlexID, error) {
    return BuildRepositoryWithPath(token, namespaceID, name, isPublic, "")
}

// BuildRepositoryWithPath 在指定命名空间下创建仓库，支持可选 path 字段
func BuildRepositoryWithPath(token string, namespaceID FlexID, name string, isPublic bool, path string) (FlexID, error) {
    payload := map[string]any{
        "namespace_id": namespaceID,
        "name":          name,
        "description":   "e2e repo",
        "is_public":     isPublic,
    }
    if path != "" {
        payload["path"] = path
    }
    status, body := DoRequest(BaseURL(), MethodPost, "/artifact_service/api/v1/repositories", payload, AuthHeader(token))
    if status != StatusOK { return 0, &HTTPError{Status: status, Body: string(body)} }
    var rr struct { Repository struct { ID FlexID `json:"id"` } `json:"repository"` }
    if err := json.Unmarshal(body, &rr); err != nil { return 0, err }
    return rr.Repository.ID, nil
}

// BuildTag 在指定仓库下创建标签
func BuildTag(token string, repositoryID FlexID, name, digest, manifest string, sizeBytes uint64) (FlexID, error) {
    payload := map[string]any{
        "name":          name,
        "digest":        digest,
        "manifest":      manifest,
        "repository_id": repositoryID,
        "size_bytes":    sizeBytes,
    }
    status, body := DoRequest(BaseURL(), MethodPost, "/artifact_service/api/v1/tags", payload, AuthHeader(token))
    if status != StatusOK { return 0, &HTTPError{Status: status, Body: string(body)} }
    var tr struct { Tag struct { ID FlexID `json:"id"` } `json:"tag"` }
    if err := json.Unmarshal(body, &tr); err != nil { return 0, err }
    return tr.Tag.ID, nil
}

// ==== 权限场景构建器 ====

// PermissionScenario 描述权限测试所需的关键参与者与资源
type PermissionScenario struct {
    // 用户与令牌
    OwnerID    uint64
    OwnerToken string
    MemberID   uint64
    MemberToken string
    AdminID    uint64
    AdminToken string
    SuperToken string
    // 资源 ID
    ProjectID   FlexID
    RegistryID  FlexID
    NamespaceID FlexID
    RepositoryID FlexID
}

// BuildPermissionScenarioBlank 构建仅由超级管理员创建的私有资源（项目/注册表/命名空间），用于空用户访问被拒的测试
func BuildPermissionScenarioBlank() (*PermissionScenario, error) {
    suf := UniqueNano()
    super, err := AdminLogin()
    if err != nil { return nil, err }

    // 使用超管创建项目
    pid, err := CreateProject(fmt.Sprintf("proj_%d", suf), "private e2e project", false, super)
    if err != nil { return nil, err }
    // 使用超管创建私有 registry 与 namespace
    rid, err := BuildRegistry(super, pid)
    if err != nil { return nil, err }
    nid, err := BuildNamespace(super, rid, fmt.Sprintf("ns_%d", suf))
    if err != nil { return nil, err }

    return &PermissionScenario{SuperToken: super, ProjectID: pid, RegistryID: rid, NamespaceID: nid}, nil
}

// BuildPermissionScenarioMember 构建：Owner 创建项目并添加 Member，超管创建私有 Registry/Namespace，Owner 创建私有 Repo
func BuildPermissionScenarioMember() (*PermissionScenario, error) {
    suf := UniqueNano()
    // 注册并登录 Owner 与 Member
    ownerID, _, ownerToken, err := RegisterAndLogin(fmt.Sprintf("owner_%d", suf), fmt.Sprintf("owner_%d@example.com", suf), "testpassword123")
    if err != nil { return nil, err }
    memberID, _, memberToken, err := RegisterAndLogin(fmt.Sprintf("member_%d", suf), fmt.Sprintf("member_%d@example.com", suf), "testpassword123")
    if err != nil { return nil, err }
    super, err := AdminLogin()
    if err != nil { return nil, err }

    // Owner 创建项目
    pid, err := CreateProject(fmt.Sprintf("proj_%d", suf), "artifact member permissions", false, ownerToken)
    if err != nil { return nil, err }
    // 添加 member 为 MEMBER 角色（3）
    if err := AddProjectMember(pid, FlexID(memberID), 3, ownerToken); err != nil { return nil, err }
    // 可选：同步权限（若服务端支持）
    _ = SyncProjectPermissions(BaseURL(), pid, ownerToken)

    // 超管创建私有 Registry 与 Namespace
    rid, err := BuildRegistry(super, pid)
    if err != nil { return nil, err }
    nid, err := BuildNamespace(super, rid, fmt.Sprintf("ns_%d", suf))
    if err != nil { return nil, err }

    // Owner 在命名空间下创建私有仓库
    repoID, err := BuildRepository(ownerToken, nid, fmt.Sprintf("repo_%d", suf), false)
    if err != nil { return nil, err }

    return &PermissionScenario{
        OwnerID: ownerID, OwnerToken: ownerToken,
        MemberID: memberID, MemberToken: memberToken,
        SuperToken: super,
        ProjectID: pid, RegistryID: rid, NamespaceID: nid, RepositoryID: repoID,
    }, nil
}

// BuildPermissionScenarioAdmin 构建：Owner 创建项目并添加 Admin，超管创建私有 Registry/Namespace，Owner 创建私有 Repo
func BuildPermissionScenarioAdmin() (*PermissionScenario, error) {
    suf := UniqueNano()
    // 注册并登录 Owner 与 Admin
    ownerID, _, ownerToken, err := RegisterAndLogin(fmt.Sprintf("owner_%d", suf), fmt.Sprintf("owner_%d@example.com", suf), "testpassword123")
    if err != nil { return nil, err }
    adminID, _, adminToken, err := RegisterAndLogin(fmt.Sprintf("admin_%d", suf), fmt.Sprintf("admin_%d@example.com", suf), "testpassword123")
    if err != nil { return nil, err }
    super, err := AdminLogin()
    if err != nil { return nil, err }

    // Owner 创建项目
    pid, err := CreateProject(fmt.Sprintf("proj_%d", suf), "artifact owner/admin permissions", false, ownerToken)
    if err != nil { return nil, err }
    // 添加 admin 为 ADMIN 角色（2）
    if err := AddProjectMember(pid, FlexID(adminID), 2, ownerToken); err != nil { return nil, err }
    // 可选：同步权限
    _ = SyncProjectPermissions(BaseURL(), pid, ownerToken)

    // 超管创建私有 Registry 与 Namespace
    rid, err := BuildRegistry(super, pid)
    if err != nil { return nil, err }
    nid, err := BuildNamespace(super, rid, fmt.Sprintf("ns_%d", suf))
    if err != nil { return nil, err }

    // Owner 在命名空间下创建私有仓库
    repoID, err := BuildRepository(ownerToken, nid, fmt.Sprintf("repo_%d", suf), false)
    if err != nil { return nil, err }

    return &PermissionScenario{
        OwnerID: ownerID, OwnerToken: ownerToken,
        AdminID: adminID, AdminToken: adminToken,
        SuperToken: super,
        ProjectID: pid, RegistryID: rid, NamespaceID: nid, RepositoryID: repoID,
    }, nil
}

// BuildPermissionScenarioPublicRegistry 构建：超管创建项目与公开 Registry、私有 Namespace 与私有 Repo
// 用于验证匿名/非成员可读取（因为 Registry 公开），但写入仍被拒绝
func BuildPermissionScenarioPublicRegistry() (*PermissionScenario, error) {
    suf := UniqueNano()
    super, err := AdminLogin()
    if err != nil { return nil, err }

    // 超管创建项目
    pid, err := CreateProject(fmt.Sprintf("proj_%d", suf), "artifact public registry permissions", false, super)
    if err != nil { return nil, err }
    // 创建公开 Registry 与私有 Namespace
    rid, err := BuildRegistryWithPublic(super, pid, true)
    if err != nil { return nil, err }
    nid, err := BuildNamespace(super, rid, fmt.Sprintf("ns_%d", suf))
    if err != nil { return nil, err }
    // 创建私有仓库（挂在公开 Registry 下）
    repoID, err := BuildRepository(super, nid, fmt.Sprintf("repo_%d", suf), false)
    if err != nil { return nil, err }

    return &PermissionScenario{ SuperToken: super, ProjectID: pid, RegistryID: rid, NamespaceID: nid, RepositoryID: repoID }, nil
}

// BuildPermissionScenarioPublicRepository 构建：超管创建项目与私有 Registry、私有 Namespace，仓库公开
// 用于验证匿名/非成员可读取（因为 Repository 公开），但写入仍被拒绝
func BuildPermissionScenarioPublicRepository() (*PermissionScenario, error) {
    suf := UniqueNano()
    super, err := AdminLogin()
    if err != nil { return nil, err }

    // 超管创建项目
    pid, err := CreateProject(fmt.Sprintf("proj_%d", suf), "artifact public repository permissions", false, super)
    if err != nil { return nil, err }
    // 创建私有 Registry 与私有 Namespace
    rid, err := BuildRegistry(super, pid)
    if err != nil { return nil, err }
    nid, err := BuildNamespace(super, rid, fmt.Sprintf("ns_%d", suf))
    if err != nil { return nil, err }
    // 创建公开仓库（挂在私有 Registry 下）
    repoID, err := BuildRepository(super, nid, fmt.Sprintf("repo_%d", suf), true)
    if err != nil { return nil, err }

    return &PermissionScenario{ SuperToken: super, ProjectID: pid, RegistryID: rid, NamespaceID: nid, RepositoryID: repoID }, nil
}