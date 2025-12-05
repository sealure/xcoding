// Project helpers: create project, add member, sync permissions.

package helpers

import (
    "encoding/json"
    "fmt"
)

type CreateProjectResp struct { Project struct { ID FlexID `json:"id"` } `json:"project"` }

// CreateProject 创建项目并返回ID
func CreateProject(name, description string, isPublic bool, token string) (FlexID, error) {
    status, body := DoRequest(BaseURL(), MethodPost, "/project_service/api/v1/projects", map[string]any{
        "name":        name,
        "description": description,
        "language":    "go",
        "framework":   "grpc-gateway",
        "is_public":   isPublic,
    }, AuthHeader(token))
    if status != StatusOK { return 0, &HTTPError{Status: status, Body: string(body)} }
    var cp CreateProjectResp
    if err := json.Unmarshal(body, &cp); err != nil { return 0, err }
    return cp.Project.ID, nil
}

// AddProjectMember 添加项目成员
func AddProjectMember(projectID, userID FlexID, role int, token string) error {
    endpoint := fmt.Sprintf("/project_service/api/v1/projects/%d/members", projectID)
    status, body := DoRequest(BaseURL(), MethodPost, endpoint, map[string]any{
        "project_id": projectID,
        "user_id":    userID,
        "role":       role,
    }, AuthHeader(token))
    if status != StatusOK { return &HTTPError{Status: status, Body: string(body)} }
    return nil
}

// SyncProjectPermissions 同步项目成员权限
func SyncProjectPermissions(baseURL string, projectID FlexID, token string) error {
    endpoint := fmt.Sprintf("/project_service/api/v1/projects/%d/sync-permissions", projectID)
    status, body := DoRequest(baseURL, MethodPost, endpoint, map[string]any{
        "project_id": projectID,
    }, AuthHeader(token))
    if status != StatusOK { return &HTTPError{Status: status, Body: string(body)} }
    return nil
}