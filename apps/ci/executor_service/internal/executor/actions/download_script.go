package actions

import (
	"fmt"
	"strings"
	"xcoding/apps/ci/executor_service/internal/parser"
)

func DownloadUsesScript(step parser.Step) (string, error) {
	b := strings.Builder{}
	ref, err := ParseUsesRef(step.Uses)
	if err != nil {
	}
	//fmt.Fprintf(&b, "echo ------------下载插件[%v]start----------------------------------------------\n", step.Name)
	// 构造下载目录
	fmt.Fprintf(&b, "tmpdir=$(mktemp -d)\n")
	fmt.Fprintf(&b, "workdir=\"$tmpdir/action\"\n")
	fmt.Fprintf(&b, "mkdir -p \"$workdir\"\n")

	fmt.Fprintf(&b, "url=%s\n", shSingleQuote("https://api.github.com/repos/"+ref.Owner+"/"+ref.Name+"/tarball/"+ref.Version))
	fmt.Fprintf(&b, "out=\"$tmpdir/action.tgz\"\n")
	fmt.Fprintf(&b, "if command -v curl >/dev/null 2>&1; then\n")
	fmt.Fprintf(&b, "  if [ -n \"$XC_GITHUB_TOKEN\" ]; then curl -sSL -H \"Authorization: token $XC_GITHUB_TOKEN\" \"$url\" -o \"$out\"; else curl -sSL \"$url\" -o \"$out\"; fi\n")
	fmt.Fprintf(&b, "elif command -v wget >/dev/null 2>&1; then\n")
	fmt.Fprintf(&b, "  if [ -n \"$XC_GITHUB_TOKEN\" ]; then wget --header=\"Authorization: token $XC_GITHUB_TOKEN\" -qO \"$out\" \"$url\"; else wget -qO \"$out\" \"$url\"; fi\n")
	fmt.Fprintf(&b, "else\n")
	fmt.Fprintf(&b, "  python3 - \"$url\" \"$out\" <<'PY'\n")
	fmt.Fprintf(&b, "import sys, urllib.request\n")
	fmt.Fprintf(&b, "u=sys.argv[1]; o=sys.argv[2]\n")
	fmt.Fprintf(&b, "req=urllib.request.Request(u)\n")
	fmt.Fprintf(&b, "import os\n")
	fmt.Fprintf(&b, "tok=os.environ.get('XC_GITHUB_TOKEN','')\n")
	fmt.Fprintf(&b, "if tok: req.add_header('Authorization','token '+tok)\n")
	fmt.Fprintf(&b, "req.add_header('Accept','application/vnd.github+json')\n")
	fmt.Fprintf(&b, "with urllib.request.urlopen(req, timeout=30) as resp:\n")
	fmt.Fprintf(&b, "  open(o,'wb').write(resp.read())\n")
	fmt.Fprintf(&b, "PY\n")
	fmt.Fprintf(&b, "fi\n")
	fmt.Fprintf(&b, "if command -v tar >/dev/null 2>&1; then\n")
	fmt.Fprintf(&b, "  tar -xzf \"$out\" -C \"$workdir\" --strip-components=1\n")
	fmt.Fprintf(&b, "else\n")
	fmt.Fprintf(&b, "  echo 'tar not available'; exit 1\n")
	fmt.Fprintf(&b, "fi\n")
	//fmt.Fprintf(&b, "echo ------------下载插件[%v]end----------------------------------------------\n", step.Name)

	return b.String(), nil
}
