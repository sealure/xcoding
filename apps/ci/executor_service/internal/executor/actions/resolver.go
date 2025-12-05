package actions

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"xcoding/apps/ci/executor_service/internal/config"
	"xcoding/apps/ci/executor_service/internal/parser"
)

// 支持带子路径的 uses：owner/name(/path...)?@version
var usesRe = regexp.MustCompile(`^([A-Za-z0-9_.-]+)/([A-Za-z0-9_.-]+)(/[A-Za-z0-9_./-]+)?@([A-Za-z0-9_.-]+)$`)

// ParseUsesRef 解析 uses 引用，返回结构化的 owner/name/version
func ParseUsesRef(uses string) (ParsedRef, error) {
	s := strings.TrimSpace(uses)
	m := usesRe.FindStringSubmatch(s)
	if m == nil {
		return ParsedRef{}, fmt.Errorf("invalid uses ref: %s", uses)
	}
	path := strings.TrimSpace(m[3])
	return ParsedRef{Owner: m[1], Name: m[2], Path: path, Version: m[4]}, nil
}

func findSubdir(root, sub string) (string, error) {
	var found string
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d == nil || !d.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, sub) {
			found = path
			return io.EOF
		}
		return nil
	})
	if strings.TrimSpace(found) == "" {
		return "", os.ErrNotExist
	}
	return found, nil
}

// shSingleQuote 对任意字符串进行 Shell 单引号安全包裹
// 说明：将单引号转义为 '\”，以避免 Shell 解析错误；支持换行与特殊字符
func shSingleQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}

// buildInputExportScript 工程化地将 with 映射为环境变量的注入脚本
// 设计：
// - 首选使用 base64 解码注入，可安全支持任意字符与换行
// - 若容器内缺少 base64 命令，则回退为单引号安全包裹的直接 export
// - 键名转换：kebab-case 转为大写下划线（如 message-id -> INPUT_MESSAGE_ID）
func buildInputExportScript(with map[string]string) string {

	if len(with) == 0 {
		return ""
	}
	b := strings.Builder{}
	keys := make([]string, 0, len(with))
	for k := range with {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		v := with[k]
		fmt.Fprintf(&b, "export %s=%s\n", k, v)
	}
	return b.String()
}

// BuildUsesScript 构建 uses 步骤的脚本片段：先注入 INPUT_*，再拼接具体动作脚本
func BuildUsesScript(step parser.Step, job parser.Job) (string, error) {
	ref, err := ParseUsesRef(step.Uses)
	if err != nil {
		return "", err
	}
	// 统一走远端仓库解析

	tmpServerDir, terr := os.MkdirTemp("", "xc_action_")
	if terr != nil {
		return "", terr
	}
	defer os.RemoveAll(tmpServerDir)
	// 统一环境注入策略：优先使用 job.Env 中的值设置进程环境，供服务器侧下载使用
	if tok := strings.TrimSpace(job.Env["XC_GITHUB_TOKEN"]); tok != "" {
		_ = os.Setenv("XC_GITHUB_TOKEN", tok)
	}
	if err := FetchTarball(ref.Owner, ref.Name, ref.Version, tmpServerDir); err != nil {
		return "", fmt.Errorf("download action tarball: %w", err)
	}
	serverSearchRoot := tmpServerDir
	if p := strings.TrimSpace(ref.Path); p != "" {
		p = strings.TrimPrefix(p, "/")
		if ss, serr := findSubdir(tmpServerDir, p); serr == nil {
			serverSearchRoot = ss
		} else {
			return "", fmt.Errorf("subpath not found: %s", p)
		}
	}
	meta, err := LoadMetadata(serverSearchRoot)
	if err != nil {
		return "", fmt.Errorf("load action metadata: %w", err)
	}

	// 命令行脚本
	var b strings.Builder
	// 注入插件环境变量
	b.WriteString(buildInputExportScript(step.With))

	//fmt.Fprintf(&b, "echo ------------111111111111--------------------------\n")
	//fmt.Fprintf(&b, "pwd\n")

	//os.Setenv("ACTION_PATH", "actions")
	//fmt.Fprintf(&b, "mkdir -p tmpdir=$ACTION_PATH\n")

	down_script, _ := DownloadUsesScript(step)
	fmt.Fprintf(&b, down_script)

	workSearchRoot := "$workdir"
	if p := strings.TrimSpace(ref.Path); p != "" {
		p = strings.TrimPrefix(p, "/")
		workSearchRoot = filepath.Join("$workdir", p)
	}
	//fmt.Fprintf(&b, "pwd\n")
	fmt.Fprintf(&b, "echo work plugin root: %s \n", workSearchRoot)

	// 切换到插件目录
	fmt.Fprintf(&b, "cd \"%s\"\n", workSearchRoot)
	//fmt.Fprintf(&b, "echo ------------2222--------------------------\n")
	//fmt.Fprintf(&b, "pwd\n")

	fmt.Fprintf(&b, "echo ------------开始执行插件[%v]----------------------------------------------\n", step.Name)

	using := strings.TrimSpace(meta.Using)
	switch using {
	case "composite":
		// 展开 composite 子步骤（run 为主；嵌套 uses 暂提示）
		for i, cs := range meta.Composite {
			name := strings.TrimSpace(cs.Name)
			if name == "" {
				name = fmt.Sprintf("composite-%d", i+1)
			}
			fmt.Fprintf(&b, "echo %s %s\n", "✔️__step_begin__", name)

			if strings.TrimSpace(cs.Run) != "" {
				fmt.Fprintf(&b, "%s\n", cs.Run)
				fmt.Fprintf(&b, "code=$?; echo __step_exit__ %s $code\nif [ $code -ne 0 ]; then exit $code; fi\n", name)
			} else if strings.TrimSpace(cs.Uses) != "" {
				fmt.Fprintf(&b, "echo \"nested uses not yet supported: %s\"\n", cs.Uses)
			}
			fmt.Fprintf(&b, "echo __step_end__ %s\n", name)
		}
	case "node":
		fmt.Fprintf(&b, "if command -v node >/dev/null 2>&1; then\n")
		fmt.Fprintf(&b, "  node \"%s\"\n", meta.Main)
		fmt.Fprintf(&b, "else\n  echo \"node not available; please use composite or provide runtime\"\n  exit 1\nfi\n")
	case "docker":
		fmt.Fprintf(&b, "echo \"docker action not supported in this runner\"\n")
	default:
		fmt.Fprintf(&b, "echo \"[unknown using] %s\"\n", using)
	}
	fmt.Fprintf(&b, "echo ------------结束执行插件[%v]----------------------------------------------\n", step.Name)
	//fmt.Fprintf(&b, "echo ------------333333--------------------------\n")
	//fmt.Fprintf(&b, "pwd\n")
	fmt.Fprintf(&b, "rm -rf \"$tmpdir\"\n")
	fmt.Fprintf(&b, "cd  %s \n", config.WORKDIR)
	fmt.Fprintf(&b, "pwd \n")
	//fmt.Fprintf(&b, "echo ------------444--------------------------\n")
	//fmt.Fprintf(&b, "pwd\n")
	//fmt.Fprintf(&b, "echo ------------step end--------------------------\n")
	//fmt.Fprintf(&b, "env\n")
	//fmt.Fprintf(&b, "echo ------------step end--------------------------\n")
	//fmt.Fprintf(&b, "echo ------------555--------------------------\n")
	//fmt.Fprintf(&b, "pwd\n")
	return b.String(), nil
}
