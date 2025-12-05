package parser

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

type Workflow struct {
	Name string            `yaml:"name"`
	Env  map[string]string `yaml:"env"`
	Jobs map[string]Job    `yaml:"jobs"`
}

type Job struct {
	Name      string            `yaml:"name"`
	Needs     StringOrSlice     `yaml:"needs"`
	Container string            `yaml:"container"`
	Env       map[string]string `yaml:"env"`
	Steps     []Step            `yaml:"steps"`
}

type Step struct {
    Name string            `yaml:"name"`
    Run  string            `yaml:"run"`
    Uses string            `yaml:"uses"`
    With map[string]string `yaml:"with"`
    Env  map[string]string `yaml:"env"`
}

// StringOrSlice 处理可以是单个字符串或字符串列表的 YAML 字段。
// 它还支持空格分隔的字符串以实现向后兼容。
type StringOrSlice []string

func (s *StringOrSlice) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind == yaml.ScalarNode {
		// 处理空格分隔的字符串以实现向后兼容（例如 "job1 job2"）
		*s = strings.Fields(value.Value)
		return nil
	}
	if value.Kind == yaml.SequenceNode {
		var temp []string
		if err := value.Decode(&temp); err != nil {
			return err
		}
		*s = temp
		return nil
	}
	return fmt.Errorf("expected string or list of strings")
}

func ParseWorkflowYAML(content string) (*Workflow, error) {
	var wf Workflow
	if err := yaml.Unmarshal([]byte(content), &wf); err != nil {
		return nil, fmt.Errorf("parse yaml: %w", err)
	}
	return &wf, nil
}
