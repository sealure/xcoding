package service

import (
	"context"
	"xcoding/apps/ci/executor_service/internal/executor"
	civ1 "xcoding/gen/go/ci/v1"
)

func (s *ExecutorService) GetK8SStatus(ctx context.Context, req *civ1.GetK8SStatusRequest) (*civ1.GetK8SStatusResponse, error) {
	env, err := executor.NewK8sEnv()
	if err != nil {
		return nil, err
	}
	items, err := env.ListBuildJobPodStatus(ctx, req.GetBuildId())
	if err != nil {
		return nil, err
	}
	prefix := req.GetJobNamePrefix()
	filtered := make([]executor.JobPodStatus, 0, len(items))
	for _, it := range items {
		if prefix == "" || (len(it.JobName) >= len(prefix) && it.JobName[:len(prefix)] == prefix) {
			filtered = append(filtered, it)
		}
	}
	page := req.GetPage()
	if page <= 0 {
		page = 1
	}
	size := req.GetPageSize()
	if size <= 0 {
		size = 20
	}
	total := int32(len(filtered))
	totalPages := int32((total + int32(size) - 1) / int32(size))
	start := int((page - 1) * size)
	if start > len(filtered) {
		start = len(filtered)
	}
	end := start + int(size)
	if end > len(filtered) {
		end = len(filtered)
	}
	sliced := filtered[start:end]
	out := make([]*civ1.K8SJobStatus, 0, len(sliced))
	for _, it := range sliced {
		pods := make([]*civ1.K8SPodStatus, 0, len(it.Pods))
		for _, p := range it.Pods {
			pods = append(pods, &civ1.K8SPodStatus{Name: p.Name, Phase: string(p.Phase), Node: p.Node, Reason: p.Reason})
		}
		conds := make([]*civ1.K8SJobCondition, 0, len(it.Conditions))
		for _, c := range it.Conditions {
			conds = append(conds, &civ1.K8SJobCondition{Type: c.Type, Status: c.Status, Reason: c.Reason, Message: c.Message})
		}
		out = append(out, &civ1.K8SJobStatus{JobName: it.JobName, Pods: pods, Succeeded: it.Succeeded, Failed: it.Failed, Conditions: conds})
	}
	return &civ1.GetK8SStatusResponse{Jobs: out, Pagination: &civ1.GetK8SStatusResponse_Pagination{Page: page, PageSize: size, TotalItems: total, TotalPages: totalPages}}, nil
}
