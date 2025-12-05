package models

import (
	artifactv1 "xcoding/gen/go/artifact/v1"
)

// ArtifactType 定义内部使用的制品类型枚举，值与 proto 保持一致
// 这样在模型和服务层代码中更易读，并提供到/从 proto 的转换方法
type ArtifactType int32

const (
	ArtifactTypeUnspecified ArtifactType = 0
	ArtifactTypeDocker      ArtifactType = 1
	ArtifactTypeGenericFile ArtifactType = 2
)

func (t ArtifactType) ToProto() artifactv1.ArtifactType            { return artifactv1.ArtifactType(t) }
func ArtifactTypeFromProto(v artifactv1.ArtifactType) ArtifactType { return ArtifactType(v) }
func (t ArtifactType) String() string                              { return artifactv1.ArtifactType(t).String() }

// ArtifactSource 定义内部使用的来源枚举，值与 proto 保持一致
// 提供到/从 proto 的转换方法以便模型与 RPC 层互通
type ArtifactSource int32

const (
	ArtifactSourceUnspecified     ArtifactSource = 0
	ArtifactSourceXCodingRegistry ArtifactSource = 1
	ArtifactSourceAliRegistry     ArtifactSource = 2
	ArtifactSourceSMB             ArtifactSource = 10
	ArtifactSourceFTP             ArtifactSource = 11
)

func (s ArtifactSource) ToProto() artifactv1.ArtifactSource              { return artifactv1.ArtifactSource(s) }
func ArtifactSourceFromProto(v artifactv1.ArtifactSource) ArtifactSource { return ArtifactSource(v) }
func (s ArtifactSource) String() string                                  { return artifactv1.ArtifactSource(s).String() }
