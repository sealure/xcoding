#!/usr/bin/env python3
# -*- coding: utf-8 -*-

"""
 python3 deploy/python/executor_service/test_runner.py --skip-deploy --uses-ref RuningBird/actions-test@v17 --update-yaml
"""

import argparse
import json
import os
import subprocess
import sys
import time
import urllib.request
import urllib.error


def run_cmd(cmd, cwd=None, timeout=None):
    """运行命令并返回 (returncode, stdout, stderr)
    说明：不抛异常，便于脚本稳定执行；stderr 仅在失败时打印
    """
    try:
        p = subprocess.Popen(cmd, cwd=cwd, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        out, err = p.communicate(timeout=timeout)
        return p.returncode, out.decode("utf-8", errors="ignore"), err.decode("utf-8", errors="ignore")
    except subprocess.TimeoutExpired:
        try:
            p.kill()
        except Exception:
            pass
        return 124, "", "timeout"


def http_post_json(url, headers, payload):
    """使用内置库发送 JSON POST 请求并返回解析后的 JSON"""
    data = json.dumps(payload).encode("utf-8")
    req = urllib.request.Request(url=url, data=data, method="POST")
    for k, v in headers.items():
        req.add_header(k, v)
    req.add_header("Content-Type", "application/json")
    with urllib.request.urlopen(req, timeout=15) as resp:
        bs = resp.read()
        return json.loads(bs.decode("utf-8"))


def http_put_json(url, headers, payload):
    data = json.dumps(payload).encode("utf-8")
    req = urllib.request.Request(url=url, data=data, method="PUT")
    for k, v in headers.items():
        req.add_header(k, v)
    req.add_header("Content-Type", "application/json")
    with urllib.request.urlopen(req, timeout=15) as resp:
        bs = resp.read()
        return json.loads(bs.decode("utf-8"))


def deploy_executor(deploy_script_path):
    """执行器服务部署：调用现有部署脚本"""
    print(f"[deploy] 运行部署脚本: {deploy_script_path}")
    rc, out, err = run_cmd([sys.executable, deploy_script_path])
    if rc != 0:
        print("[deploy] 部署脚本执行失败")
        print(out)
        print(err)
        sys.exit(rc)
    print("[deploy] 部署脚本输出：")
    print(out)


def trigger_pipeline(api_base, pipeline_id, user_id, username):
    """触发一次构建并返回 build_id"""
    url = f"{api_base}/ci_service/api/v1/pipelines/{pipeline_id}/builds"
    headers = {
        "X-User-ID": str(user_id),
        "X-Username": username,
    }
    print(f"[trigger] POST {url}")
    resp = http_post_json(url, headers, {})
    print("[trigger] 响应：")
    print(json.dumps(resp, ensure_ascii=False, indent=2))
    build = resp.get("build") or {}
    bid = int(str(build.get("id", "0")))
    if bid <= 0:
        print("[trigger] 未获得有效 build id")
        sys.exit(2)
    return bid


def update_pipeline_yaml(api_base, pipeline_id, user_id, username, yaml_text):
    url = f"{api_base}/ci_service/api/v1/pipelines/{pipeline_id}"
    headers = {
        "X-User-ID": str(user_id),
        "X-Username": username,
    }
    payload = {
        "workflow_yaml": yaml_text,
    }
    print(f"[update] PUT {url}")
    try:
        resp = http_put_json(url, headers, payload)
        print("[update] 响应：")
        print(json.dumps(resp, ensure_ascii=False, indent=2))
    except urllib.error.HTTPError as e:
        body = e.read().decode("utf-8", errors="ignore")
        print(f"[update] 失败：HTTP {e.code}\n{body}")
        sys.exit(2)
    except Exception as e:
        print(f"[update] 失败：{e}")
        sys.exit(2)


def generate_yaml(uses_ref):
    parts = []
    parts.append("name: action1\n")
    parts.append("'on':\n")
    parts.append("  workflow_dispatch: {}\n")
    parts.append("jobs:\n")
    parts.append("  job-1:\n")
    parts.append("    name: Job 1\n")
    parts.append("    runs-on: ubuntu-latest\n")
    parts.append("    container: ipowerink/python-tree\n")
    parts.append("    env:\n")
    parts.append("      XC_ACTIONS_FORCE_REFRESH: 'true'\n")
    parts.append("    steps:\n")
    parts.append("      - name: github_action_demo\n")
    parts.append("        uses: RuningBird/actions-test@v21\n")
    parts.append("        with:\n")
    parts.append("          MESSAGE: demo\n")
    parts.append("      - name: echo hello\n")
    parts.append("        run: echo \"Hello1234\"\n")
    return "".join(parts)


def find_pod(namespace, build_id, job_name="job-1"):
    """通过标签 job-name 查找 Pod，优先使用选择器，失败时回退到全量匹配"""
    job_label = f"build-{build_id}-{job_name}"
    print(f"[k8s] 通过标签查找 Pod：job-name={job_label}")
    # 放宽等待时间，避免镜像拉取导致短时无 Pod
    for i in range(60):
        # 优先使用标签选择器精确匹配该 Job 的 Pod
        rc, out, err = run_cmd([
            "kubectl", "get", "pods", "-n", namespace,
            "-l", f"job-name={job_label}",
            "-o", "custom-columns=NAME:.metadata.name",
            "--no-headers",
        ])
        if rc == 0:
            lines = [s.strip() for s in out.splitlines() if s.strip()]
            if lines:
                name = lines[0]
                print(f"[k8s] 命中 Pod：{name}")
                return name
        # 回退：全量列出后按名称前缀匹配
        rc2, out2, err2 = run_cmd([
            "kubectl", "get", "pods", "-n", namespace,
            "-o", "custom-columns=NAME:.metadata.name",
            "--no-headers",
        ])
        if rc2 == 0:
            lines2 = [s.strip() for s in out2.splitlines() if s.strip()]
            for name in lines2:
                if name.startswith(job_label):
                    print(f"[k8s] 命中 Pod(回退)：{name}")
                    return name
        time.sleep(0.5)
    print("[k8s] 未找到 Pod，可能部署失败或 Job 未创建")
    return None


def stream_logs(namespace, pod_name, container="runner", duration_sec=30):
    """等待容器就绪后跟随日志，避免 ContainerCreating 误报"""
    print(f"[logs] 跟随容器日志：pod={pod_name} container={container}")
    # 先等待容器进入 Running/Ready 或已有日志可读
    for i in range(30):
        # 查询 Pod 状态
        rc, out, err = run_cmd([
            "kubectl", "get", "pod", pod_name, "-n", namespace, "-o", "json"
        ])
        if rc == 0 and out.strip():
            try:
                info = json.loads(out)
                cs = info.get("status", {}).get("containerStatuses", [])
                ready = False
                for it in cs:
                    if it.get("name") == container:
                        st = it.get("state", {})
                        if st.get("running") or it.get("ready"):
                            ready = True
                            break
                if ready:
                    break
            except Exception:
                pass
        time.sleep(1)
    try:
        # 使用 --follow 跟随，--timestamps 添加时间戳，便于排序与比对
        p = subprocess.Popen([
            "kubectl", "logs", pod_name, "-n", namespace, "-c", container, "--follow", "--timestamps"
        ], stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        start = time.time()
        while True:
            line = p.stdout.readline()
            if not line:
                # 若无输出且进程已结束，退出循环
                if p.poll() is not None:
                    break
                # 超时控制
                if time.time() - start > duration_sec:
                    try:
                        p.kill()
                    except Exception:
                        pass
                    break
                time.sleep(0.2)
                continue
            s = line.decode("utf-8", errors="ignore").rstrip()
            print(s)
            # 关键错误提示聚合
            t = s.strip()
            if ("action error:" in t) or ("download action tarball" in t) or ("nested uses not yet supported" in t):
                print(f"[hint] {t}")
        # 打印可能的错误
        err_out = p.stderr.read().decode("utf-8", errors="ignore").strip()
        if err_out:
            print(f"[logs] 错误输出：{err_out}")
    except Exception as e:
        print(f"[logs] 跟随失败：{e}")


def main():
    parser = argparse.ArgumentParser(description="CI 执行器远端 actions 测试脚本")
    parser.add_argument("--deploy-script", default=os.path.join("deploy", "python", "executor_service", "deploy.py"), help="执行器部署脚本路径")
    parser.add_argument("--api-base", default="http://localhost:31080", help="后端 API 基地址")
    parser.add_argument("--pipeline-id", type=int, default=166, help="流水线 ID")
    parser.add_argument("--user-id", type=int, default=1067, help="触发者用户ID")
    parser.add_argument("--username", default="user2", help="触发者用户名")
    parser.add_argument("--namespace", default=os.environ.get("POD_NAMESPACE", "xcoding"), help="K8s 命名空间")
    parser.add_argument("--skip-deploy", action="store_true", help="跳过部署执行器")
    parser.add_argument("--uses-ref", default="RuningBird/actions-test@v1", help="uses 引用（支持子路径）")
    parser.add_argument("--update-yaml", action="store_true", help="先更新流水线 YAML")
    args = parser.parse_args()

    if not args.skip_deploy:
        deploy_executor(args.deploy_script)

    if args.update_yaml:
        yaml_text = generate_yaml(args.uses_ref)
        update_pipeline_yaml(args.api_base, args.pipeline_id, args.user_id, args.username, yaml_text)

    build_id = trigger_pipeline(args.api_base, args.pipeline_id, args.user_id, args.username)
    print(f"[trigger] 新构建 ID：{build_id}")

    pod = find_pod(args.namespace, build_id, job_name="job-1")
    if not pod:
        sys.exit(3)

    stream_logs(args.namespace, pod, container="runner", duration_sec=40)
    print("[done] 日志拉取结束，请根据提示定位问题")


if __name__ == "__main__":
    main()
