#!/usr/bin/env python3
"""
CI Pipeline 服务部署脚本
功能：读取Deployment YAML、修改镜像标签、构建镜像、推送和使用Helm部署到Kubernetes
说明：该脚本风格对齐 deploy/python/artifact/deploy.py。
"""

import os
import sys
import subprocess
from time import sleep
import yaml
import argparse
import random
import string
from typing import Dict, Any


class CIDeployer:
    def __init__(self, project_root: str = "/home/hr/xcoding"):
        self.project_root = project_root
        self.deployment_yaml_path = os.path.join(project_root, "deploy/xcoding/templates/services/ci/deployment.yaml")
        self.chart_path = os.path.join(project_root, "deploy/xcoding")
        self.namespace = "xcoding"
        self.release_name = "xcoding"

    def generate_random_tag(self, length: int = 8) -> str:
        return ''.join(random.choices(string.ascii_lowercase + string.digits, k=length))

    def load_deployment_yaml(self) -> Dict[str, Any]:
        try:
            with open(self.deployment_yaml_path, 'r') as file:
                return yaml.safe_load(file)
        except FileNotFoundError:
            print(f"未找到 {self.deployment_yaml_path}，请先在 Helm 模板中添加 CI 的 Deployment。")
            sys.exit(1)
        except Exception as e:
            print(f"加载deployment.yaml失败: {e}")
            sys.exit(1)

    def save_deployment_yaml(self, data: Dict[str, Any]) -> None:
        try:
            with open(self.deployment_yaml_path, 'w') as file:
                yaml.dump(data, file, default_flow_style=False, sort_keys=False)
            print("deployment.yaml已更新")
        except Exception as e:
            print(f"保存deployment.yaml失败: {e}")
            sys.exit(1)

    def update_image_tag(self, tag: str) -> None:
        target_image = f"localhost:31500/ci-pipeline-service:{tag}"
        print(f"更新 CI Pipeline 服务镜像标签为: {target_image}")
        data = self.load_deployment_yaml()

        containers = data['spec']['template']['spec']['containers']
        updated = False
        for container in containers:
            if container.get('name') == 'ci-pipeline':
                current_image = container.get('image')
                container['image'] = target_image
                print(f"镜像已从 {current_image} 更新为 {container['image']}")
                updated = True
                break

        if not updated:
            print("未在 deployment.yaml 中找到容器名为 'ci-pipeline' 的配置，请检查模板")
            sys.exit(1)

        self.save_deployment_yaml(data)

    def build_image(self, tag: str) -> bool:
        image = f"localhost:31500/ci-pipeline-service:{tag}"
        print(f"构建 CI Pipeline 服务镜像: {image}")
        cmd = ["docker", "build", "-t", image, "-f", "apps/ci/pipeline_service/Dockerfile", "."]
        result = subprocess.run(cmd, cwd=self.project_root)
        if result.returncode != 0:
            print("镜像构建失败")
            return False
        print("镜像构建成功")
        return True

    def push_image(self, tag: str) -> bool:
        image = f"localhost:31500/ci-pipeline-service:{tag}"
        print(f"推送 CI Pipeline 服务镜像: {image}")
        cmd = ["docker", "push", image]
        result = subprocess.run(cmd)
        if result.returncode != 0:
            print("镜像推送失败")
            return False
        print("镜像推送成功")
        return True

    def deploy(self) -> bool:
        print("使用Helm部署 CI Pipeline 服务...")

        ns_check_cmd = ["kubectl", "get", "namespace", self.namespace]
        ns_result = subprocess.run(ns_check_cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        if ns_result.returncode != 0:
            print(f"命名空间 {self.namespace} 不存在，将创建...")
            create_ns_cmd = ["kubectl", "create", "namespace", self.namespace]
            create_result = subprocess.run(create_ns_cmd)
            if create_result.returncode != 0:
                print(f"创建命名空间 {self.namespace} 失败")
                return False

        status_cmd = ["helm", "status", self.release_name, "-n", self.namespace]
        status_result = subprocess.run(status_cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        if status_result.returncode == 0:
            print("检测到已安装的版本，执行升级操作...")
            cmd = ["helm", "upgrade", self.release_name, self.chart_path, "-n", self.namespace]
            operation = "升级"
        else:
            print("未检测到已安装的版本，执行安装操作...")
            cmd = ["helm", "install", self.release_name, self.chart_path, "-n", self.namespace]
            operation = "安装"

        result = subprocess.run(cmd)
        if result.returncode != 0:
            print(f"Helm{operation}失败")
            return False
        print(f"Helm{operation}成功")
        return True

    def check_deployment_status(self) -> None:
        print("\n检查部署状态...")
        print("\nCI Pipeline 服务Pod状态:")
        pods_cmd = ["kubectl", "get", "pods", "-n", self.namespace, "-l", "app.kubernetes.io/component=ci-pipeline"]
        subprocess.run(pods_cmd)
        print("\n服务列表:")
        svc_cmd = ["kubectl", "get", "services", "-n", self.namespace]
        subprocess.run(svc_cmd)

    def get_logs(self) -> None:
        sleep(3)
        print("\n获取 CI Pipeline 服务日志...")
        logs_cmd = ["kubectl", "logs", "-n", self.namespace, "-l", "app.kubernetes.io/component=ci-pipeline", "--tail=20"]
        subprocess.run(logs_cmd)

    def deploy_ci_service(self, tag: str, push_image: bool = True) -> bool:
        print(f"开始 CI Pipeline 服务部署流程，镜像标签: {tag}")
        self.update_image_tag(tag)
        if not self.build_image(tag):
            return False
        if push_image:
            if not self.push_image(tag):
                return False
        if not self.deploy():
            return False
        self.check_deployment_status()
        self.get_logs()
        print("\nCI Pipeline 服务部署流程完成！")
        return True


def main():
    parser = argparse.ArgumentParser(description='Deploy CI Pipeline service')
    parser.add_argument('--tag', help='Docker image tag', default=None)
    parser.add_argument('--release', help='Helm release name', default='xcoding')
    parser.add_argument('--namespace', help='Kubernetes namespace', default='xcoding')
    parser.add_argument('--action', choices=['deploy', 'status', 'logs'], default='deploy')
    args = parser.parse_args()

    tag = args.tag or ''.join(random.choices(string.ascii_lowercase + string.digits, k=8))

    deployer = CIDeployer()
    deployer.release_name = args.release
    deployer.namespace = args.namespace

    if args.action == 'deploy':
        deployer.deploy_ci_service(tag)
    elif args.action == 'status':
        deployer.check_deployment_status()
    elif args.action == 'logs':
        deployer.get_logs()


if __name__ == "__main__":
    main()