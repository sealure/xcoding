package actions

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// CompositeStepMeta 精简的 composite 子步骤元数据
type CompositeStepMeta struct {
	Name  string            `yaml:"name"`
	Id    string            `yaml:"id"`
	Run   string            `yaml:"run"`
	Shell string            `yaml:"shell"`
	Uses  string            `yaml:"uses"`
	Env   map[string]string `yaml:"env"`
	With  map[string]string `yaml:"with"`
}

// LoadMetadata 从目标路径递归查找并解析 action.yml / action.yaml
func LoadMetadata(root string) (*ResolvedAction, error) {
	var found string
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d == nil || d.IsDir() {
			return nil
		}
		base := strings.ToLower(filepath.Base(path))
		if base == "action.yml" || base == "action.yaml" {
			found = path
			return io.EOF // 终止遍历（利用错误提前返回）
		}
		return nil
	})
	if strings.TrimSpace(found) == "" {
		return nil, os.ErrNotExist
	}
	var m struct {
		Runs struct {
			Using string              `yaml:"using"`
			Main  string              `yaml:"main"`
			Steps []CompositeStepMeta `yaml:"steps"`
		} `yaml:"runs"`
	}
	bs, err := os.ReadFile(found)
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(bs, &m); err != nil {
		return nil, err
	}
	using := strings.TrimSpace(strings.ToLower(m.Runs.Using))
	if using == "node12" || using == "node16" || using == "node20" {
		using = "node"
	}
	ra := &ResolvedAction{Using: using, Path: filepath.Dir(found)}
	if using == "composite" {
		ra.Composite = m.Runs.Steps
	} else if using == "node" {
		ra.Main = strings.TrimSpace(m.Runs.Main)
	}
	return ra, nil
}
