package actions

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// FetchTarball 下载 GitHub 仓库 tarball 并解压到目标目录
// 说明：支持匿名下载；如需私仓，读取环境变量 XC_GITHUB_TOKEN 注入 Authorization
func FetchTarball(owner, name, ref, dest string) error {
	if err := os.MkdirAll(dest, 0o755); err != nil {
		return err
	}
    tries := []string{ref}
	var lastErr error
	for _, r := range tries {
		url := fmt.Sprintf("https://api.github.com/repos/%s/%s/tarball/%s", owner, name, r)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			lastErr = err
			continue
		}
		if tok := os.Getenv("XC_GITHUB_TOKEN"); tok != "" {
			req.Header.Set("Authorization", "token "+tok)
		}
		req.Header.Set("Accept", "application/vnd.github+json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("download tarball: %s", resp.Status)
			resp.Body.Close()
			continue
		}
		// 解压 tar.gz 到目标目录
		gz, err := gzip.NewReader(resp.Body)
		if err != nil {
			lastErr = err
			resp.Body.Close()
			continue
		}
		tr := tar.NewReader(gz)
		var topPrefix string
		for {
			hdr, er := tr.Next()
			if er == io.EOF {
				break
			}
			if er != nil {
				lastErr = er
				break
			}
			rel := filepath.Clean(hdr.Name)
			if topPrefix == "" {
				// GitHub tarball 的首层通常为 <owner>-<name>-<sha>/...
				// 记录该前缀并剥离，便于按照仓库根相对路径展开
				i := strings.IndexByte(rel, os.PathSeparator)
				if i > 0 {
					topPrefix = rel[:i]
				} else {
					topPrefix = rel
				}
			}
			if strings.HasPrefix(rel, topPrefix) {
				rel = strings.TrimPrefix(rel, topPrefix)
				rel = strings.TrimPrefix(rel, string(os.PathSeparator))
			}
			if rel == "." || rel == ".." {
				continue
			}
			outPath := filepath.Join(dest, rel)
			switch hdr.Typeflag {
			case tar.TypeDir:
				if err := os.MkdirAll(outPath, 0o755); err != nil {
					lastErr = err
					break
				}
			case tar.TypeReg:
				if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
					lastErr = err
					break
				}
				f, err := os.OpenFile(outPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
				if err != nil {
					lastErr = err
					break
				}
				if _, err := io.Copy(f, tr); err != nil {
					_ = f.Close()
					lastErr = err
					break
				}
				_ = f.Close()
			}
		}
		gz.Close()
		resp.Body.Close()
		if lastErr == nil {
			return nil
		}
		// 尝试下一个 ref
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("unknown error downloading tarball")
	}
	return lastErr
}
