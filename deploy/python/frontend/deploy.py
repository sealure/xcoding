#!/usr/bin/env python3
"""
前端服务部署脚本
功能：读取Helm模板、修改镜像标签、构建镜像、推送和部署到Kubernetes
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

class FrontendDeployer:
    def __init__(self, project_root: str = "/home/hr/xcoding"):
        self.project_root = project_root
        self.values_yaml_path = os.path.join(project_root, "deploy/xcoding/values.yaml")
        self.chart_path = os.path.join(project_root, "deploy/xcoding")
        self.namespace = "xcoding"
        self.release_name = "xcoding"
        self.container_name = "frontend"
        self.registry_image = "localhost:31500/frontend"

    def generate_random_tag(self, length: int = 8) -> str:
        return ''.join(random.choices(string.ascii_lowercase + string.digits, k=length))

    def load_yaml(self) -> Dict[str, Any]:
        try:
            with open(self.values_yaml_path, 'r') as file:
                return yaml.safe_load(file)
        except Exception as e:
            print(f"加载values.yaml失败: {e}")
            sys.exit(1)

    def save_yaml(self, data: Dict[str, Any]) -> None:
        try:
            with open(self.values_yaml_path, 'w') as file:
                yaml.dump(data, file, default_flow_style=False, sort_keys=False)
            print("values.yaml已更新")
        except Exception as e:
            print(f"保存values.yaml失败: {e}")
            sys.exit(1)

    def update_image_tag(self, tag: str) -> None:
        print(f"更新前端镜像标签为: {tag}")
        data = self.load_yaml()
        fe = data.get('frontend', {})
        # 启用前端并设置镜像
        fe['enabled'] = True
        image = fe.get('image', {})
        image['repository'] = self.registry_image
        image['tag'] = tag
        image['pullPolicy'] = image.get('pullPolicy', 'IfNotPresent')
        fe['image'] = image
        # 服务端口配置
        service = fe.get('service', {})
        service['containerPort'] = service.get('containerPort', 80)
        service['port'] = service.get('port', 80)
        fe['service'] = service
        # 其他默认值
        fe['replicaCount'] = fe.get('replicaCount', 1)
        fe['resources'] = fe.get('resources', {})
        data['frontend'] = fe
        self.save_yaml(data)

    def build_image(self, tag: str) -> bool:
        print(f"构建前端镜像: {self.registry_image}:{tag}")
        cmd = ["docker", "build", "-t", f"{self.registry_image}:{tag}", "-f", "apps/frontend/Dockerfile", "."]
        result = subprocess.run(cmd, cwd=self.project_root)
        if result.returncode != 0:
            print("镜像构建失败")
            return False
        print("镜像构建成功")
        return True

    def push_image(self, tag: str) -> bool:
        print(f"推送前端镜像: {self.registry_image}:{tag}")
        cmd = ["docker", "push", f"{self.registry_image}:{tag}"]
        result = subprocess.run(cmd)
        if result.returncode != 0:
            print("镜像推送失败")
            return False
        print("镜像推送成功")
        return True

    def deploy(self) -> bool:
        print("使用Helm部署前端服务...")
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
        print("\n前端服务Pod状态:")
        pods_cmd = ["kubectl", "get", "pods", "-n", self.namespace, "-l", "app.kubernetes.io/component=frontend"]
        subprocess.run(pods_cmd)
        print("\n服务状态:")
        svc_cmd = ["kubectl", "get", "services", "-n", self.namespace]
        subprocess.run(svc_cmd)

    def get_logs(self) -> None:
        sleep(3)
        print("\n获取前端服务日志...")
        logs_cmd = ["kubectl", "logs", "-n", self.namespace, "-l", "app.kubernetes.io/component=frontend", "--tail=20"]
        subprocess.run(logs_cmd)

    def deploy_frontend(self, tag: str, push_image: bool = True) -> bool:
        print(f"开始前端服务部署流程，镜像标签: {tag}")
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
        print("\n前端服务部署流程完成！")
        return True


def main():
    parser = argparse.ArgumentParser(description="前端服务部署脚本")
    parser.add_argument("--tag", help="镜像标签（可选，未提供时将自动生成随机值）")
    parser.add_argument("--no-push", action="store_true", help="跳过镜像推送")

    args = parser.parse_args()

    deployer = FrontendDeployer()
    tag = args.tag or deployer.generate_random_tag()
    print(f"使用镜像标签: {tag}")

    success = deployer.deploy_frontend(tag, push_image=not args.no_push)
    if not success:
        sys.exit(1)


if __name__ == "__main__":
    main()