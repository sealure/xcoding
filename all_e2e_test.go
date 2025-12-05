package xcoding_test

import (
	"context"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// parseTestFlagsFromArgs maps root test binary flags (-test.*) to go test CLI flags.
func parseTestFlagsFromArgs(args []string) []string {
	var out []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "-test.v") {
			out = append(out, "-v")
		} else if strings.HasPrefix(arg, "-test.short") && strings.Contains(arg, "=true") {
			out = append(out, "-short")
		} else if strings.HasPrefix(arg, "-test.failfast") && strings.Contains(arg, "=true") {
			out = append(out, "-failfast")
		} else if strings.HasPrefix(arg, "-test.run=") {
			pat := strings.TrimPrefix(arg, "-test.run=")
			out = append(out, "-run", pat)
		} else if arg == "-test.run" && i+1 < len(args) {
			out = append(out, "-run", args[i+1])
			i++
		} else if strings.HasPrefix(arg, "-test.count=") {
			out = append(out, "-count", strings.TrimPrefix(arg, "-test.count="))
		}
	}
	return out
}

// discoverE2EPackages returns all packages under ./apps/... or ./e2e/... whose import path ends with /e2e or is inside ./e2e/*.
func discoverE2EPackages(t *testing.T) []string {
	t.Helper()
	root := rootDir()
	var pkgs []string
	// apps/*/e2e
	_ = filepath.WalkDir(filepath.Join(root, "apps"), func(path string, d fs.DirEntry, err error) error {
		if err != nil { return nil }
		if d == nil || !d.IsDir() { return nil }
		if filepath.Base(path) != "e2e" { return nil }
		entries, e := os.ReadDir(path)
		if e != nil { return nil }
		for _, de := range entries {
			if !de.IsDir() && strings.HasSuffix(de.Name(), ".go") {
				rel, _ := filepath.Rel(root, path)
				pkgs = append(pkgs, "./"+rel)
				break
			}
		}
		return nil
	})
	// top-level ./e2e/* (Ginkgo suites)
	top := filepath.Join(root, "e2e")
	_ = filepath.WalkDir(top, func(path string, d fs.DirEntry, err error) error {
		if err != nil { return nil }
		if d == nil || !d.IsDir() { return nil }
		if path == top { return nil }
		entries, e := os.ReadDir(path)
		if e != nil { return nil }
		for _, de := range entries {
			if !de.IsDir() && strings.HasSuffix(de.Name(), ".go") {
				rel, _ := filepath.Rel(root, path)
				pkgs = append(pkgs, "./"+rel)
				break
			}
		}
		return nil
	})
	if len(pkgs) == 0 {
		t.Skip("no e2e packages discovered under ./apps or ./e2e")
	}
	return pkgs
}

// runGoTestPkg runs `go test <pkg> <flags>` as a sub-process and fails on non-zero exit.
func runGoTestPkg(t *testing.T, pkg string, extraArgs ...string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()
	args := []string{"test", pkg, "-tags", "e2e"}
	args = append(args, extraArgs...)
	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Env = os.Environ()
	cmd.Dir = rootDir()
	out, err := cmd.CombinedOutput()
	t.Logf("go %s\n%s", strings.Join(args, " "), string(out))
	if ctx.Err() == context.DeadlineExceeded {
		t.Fatalf("timeout running: go %s", strings.Join(args, " "))
	}
	if err != nil {
		t.Fatalf("go test failed for %s: %v", pkg, err)
	}
}

func rootDir() string {
	// Ensure we run from repository root even if test binary CWD differs
	wd, _ := os.Getwd()
	return filepath.Clean(wd)
}

// TestAllE2E dynamically discovers and runs all ./apps/*/e2e and ./e2e/* test packages.
func TestAllE2E(t *testing.T) {
	pkgs := discoverE2EPackages(t)
	flags := parseTestFlagsFromArgs(os.Args)
	for _, p := range pkgs {
		p := p // capture
		name := strings.TrimPrefix(p, "./")
		t.Run(name, func(t *testing.T) {
			runGoTestPkg(t, p, flags...)
		})
	}
}