package executor

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type JobPodStatus struct {
	JobName    string         `json:"job_name"`
	Succeeded  int32          `json:"succeeded"`
	Failed     int32          `json:"failed"`
	Conditions []JobCondition `json:"conditions"`
	Pods       []PodStatus    `json:"pods"`
}

type PodStatus struct {
	Name   string          `json:"name"`
	Phase  corev1.PodPhase `json:"phase"`
	Node   string          `json:"node"`
	Reason string          `json:"reason"`
}

type JobCondition struct {
	Type    string `json:"type"`
	Status  string `json:"status"`
	Reason  string `json:"reason"`
	Message string `json:"message"`
}

// ListBuildJobPodStatus 返回与构建有关的 Job 的 Pod 状态
func (e *K8sEnv) ListBuildJobPodStatus(ctx context.Context, buildID uint64) ([]JobPodStatus, error) {
	ns := e.Namespace
	jobs, err := e.Clientset.BatchV1().Jobs(ns).List(ctx, metav1.ListOptions{LabelSelector: "xcoding.io/build-id=" + fmt.Sprintf("%d", buildID)})
	if err != nil {
		return nil, err
	}
	out := make([]JobPodStatus, 0, len(jobs.Items))
	for i := range jobs.Items {
		j := jobs.Items[i]
		pods, err := e.Clientset.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{LabelSelector: "job-name=" + j.Name})
		if err != nil {
			return nil, err
		}
		ps := make([]PodStatus, 0, len(pods.Items))
		for _, p := range pods.Items {
			node := p.Spec.NodeName
			phase := p.Status.Phase
			reason := ""
			if len(p.Status.Conditions) > 0 {
				reason = string(p.Status.Conditions[len(p.Status.Conditions)-1].Reason)
			}
			ps = append(ps, PodStatus{Name: p.Name, Phase: phase, Node: node, Reason: reason})
		}
		conds := make([]JobCondition, 0, len(j.Status.Conditions))
		for _, c := range j.Status.Conditions {
			conds = append(conds, JobCondition{Type: string(c.Type), Status: string(c.Status), Reason: c.Reason, Message: c.Message})
		}
		out = append(out, JobPodStatus{JobName: j.Name, Succeeded: j.Status.Succeeded, Failed: j.Status.Failed, Conditions: conds, Pods: ps})
	}
	return out, nil
}
