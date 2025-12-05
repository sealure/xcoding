//go:build e2e

package executor_e2e

import (
	"encoding/json"
	"fmt"
	e2e "xcoding/e2e/helpers"
)

func baseURL() string                     { return e2e.BaseURL() }
func ping() bool                          { return e2e.PingGateway(baseURL()) }
func auth(token string) map[string]string { return e2e.AuthHeader(token) }
func adminLogin() (string, error)         { return e2e.AdminLogin() }
func uniqueNano() int64                   { return e2e.UniqueNano() }
func createProject(name, description string, isPublic bool, token string) (e2e.FlexID, error) {
	return e2e.CreateProject(name, description, isPublic, token)
}

func do(method, endpoint string, body any, headers map[string]string) (int, []byte) {
	return e2e.DoRequest(baseURL(), method, endpoint, body, headers)
}

func doWithHeaders(method, endpoint string, body any, headers map[string]string) (int, []byte, map[string]string) {
	return e2e.DoRequestWithHeaders(baseURL(), method, endpoint, body, headers)
}

type Pipeline struct {
	ID e2e.FlexID `json:"id"`
}
type Build struct {
	ID         e2e.FlexID `json:"id"`
	PipelineID e2e.FlexID `json:"pipeline_id,omitempty"`
	Status     string     `json:"status"`
}
type CreatePipelineResp struct {
	Pipeline Pipeline `json:"pipeline"`
}
type StartBuildResp struct {
	Build Build `json:"build"`
}
type GetBuildResp struct {
	Build Build `json:"build"`
}
type GetLogsResp struct {
	Lines []string `json:"lines"`
}

func createPipeline(token string, projectID e2e.FlexID, name, description, yaml string) (e2e.FlexID, error) {
	url := "/ci_service/api/v1/pipelines"
	status, body := do("POST", url, map[string]any{
		"name":          name,
		"description":   description,
		"project_id":    projectID,
		"workflow_yaml": yaml,
		"is_active":     true,
	}, auth(token))
	if status != e2e.StatusOK && status != e2e.StatusCreated {
		return 0, &e2e.HTTPError{Status: status, Body: string(body)}
	}
	var resp CreatePipelineResp
	if err := json.Unmarshal(body, &resp); err != nil {
		return 0, err
	}
	fmt.Printf("func createPipeline,request url:%v, response pipelineID:%v\n", url, resp.Pipeline.ID)
	return resp.Pipeline.ID, nil
}

func startBuild(token string, pipelineID e2e.FlexID, branch, triggeredBy string) (e2e.FlexID, error) {
	url := fmt.Sprintf("/ci_service/api/v1/pipelines/%d/builds", uint64(pipelineID))
	status, body := do("POST", url, map[string]any{
		"branch":       branch,
		"triggered_by": triggeredBy,
	}, auth(token))
	if status != e2e.StatusOK && status != e2e.StatusCreated {
		return 0, &e2e.HTTPError{Status: status, Body: string(body)}
	}
	var resp StartBuildResp
	if err := json.Unmarshal(body, &resp); err != nil {
		return 0, err
	}
	fmt.Printf("func startBuild,request url:%v, response buildID:%v\n", url, resp.Build.ID)
	return resp.Build.ID, nil
}

func getBuildStatus(token string, buildID e2e.FlexID) (string, error) {
	url := fmt.Sprintf("/ci_service/api/v1/builds/%d", uint64(buildID))
	status, body := do("GET", url, nil, auth(token))
	if status != e2e.StatusOK {
		return "", &e2e.HTTPError{Status: status, Body: string(body)}
	}
	var resp GetBuildResp
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", err
	}
	fmt.Printf("func getBuildStatus,request url:%v, response status:%v\n", url, resp.Build.Status)
	return resp.Build.Status, nil
}

func getExecutorLogs(token string, buildID e2e.FlexID) ([]string, error) {
	url := fmt.Sprintf("/ci_service/api/v1/executor/builds/%d/logs", uint64(buildID))
	status, body := do("GET", url, nil, auth(token))
	if status != e2e.StatusOK {
		return nil, &e2e.HTTPError{Status: status, Body: string(body)}
	}
	var resp GetLogsResp
	_ = json.Unmarshal(body, &resp)
	fmt.Printf("func getExecutorLogs,request url:%v, response lines:%v\n", url, resp.Lines)
	return resp.Lines, nil
}
