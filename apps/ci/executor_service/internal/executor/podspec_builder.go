package executor

import (
	"xcoding/apps/ci/executor_service/internal/config"

	corev1 "k8s.io/api/core/v1"
)

// BuildPodSpec 构建 Pod 规范：单容器 runner，执行脚本
// 说明：使用 /bin/sh -c 运行合成脚本；重启策略 Never，便于 Job 控制
func BuildPodSpec(image string, script string, envs []corev1.EnvVar) corev1.PodSpec {
	ru := int64(0) // runner 用户 ID
	rg := int64(0) // runner 组 ID
	fg := int64(0) // 挂载卷的组所有权
	//rootUser := int64(0)
	return corev1.PodSpec{
		RestartPolicy:   corev1.RestartPolicyNever,
		SecurityContext: &corev1.PodSecurityContext{FSGroup: &fg},
		Volumes: []corev1.Volume{{
			Name:         "workspace",
			VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
		}},
		//InitContainers: []corev1.Container{{
		//	Name:            "installer",
		//	Image:           image,
		//	SecurityContext: &corev1.SecurityContext{RunAsUser: &rootUser},
		//	Command:         []string{"/bin/bash", "-c"},
		//	Args:            []string{"apt update && apt install -y tree"},
		//}},
		Containers: []corev1.Container{{
			Name:            "runner",
			Image:           image,
			SecurityContext: &corev1.SecurityContext{RunAsUser: &ru, RunAsGroup: &rg},
			//WorkingDir:      "/workspace",
			WorkingDir: config.WORKDIR,
			Command:    []string{"/bin/bash", "-c"},
			Args:       []string{script},
			Env:        envs,
			VolumeMounts: []corev1.VolumeMount{{
				Name:      "workspace",
				MountPath: "/workspace",
			}},
		}},
	}
}
