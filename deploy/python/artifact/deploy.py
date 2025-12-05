#!/usr/bin/env python3
"""
artifact 服务部署脚本
功能：读取Deployment YAML、修改镜像标签、构建镜像、推送和使用Helm部署到Kubernetes
说明：该脚本风格对齐 deploy/python/code_repository/deploy.py。
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

class ArtifactDeployer:
    def __init__(self, project_root: str = "/home/hr/xcoding"):
        # 项目根目录
        self.project_root = project_root
        # 目标 deployment.yaml 路径（约定容器名为 artifact）
        self.deployment_yaml_path = os.path.join(project_root, "deploy/xcoding/templates/services/artifact/deployment.yaml")
        # Helm chart 根目录
        self.chart_path = os.path.join(project_root, "deploy/xcoding")
        # 部署命名空间及 release
        self.namespace = "xcoding"
        self.release_name = "xcoding"

    def generate_random_tag(self, length: int = 8) -> str:
        """生成随机标签"""
        return ''.join(random.choices(string.ascii_lowercase + string.digits, k=length))

    def load_deployment_yaml(self) -> Dict[str, Any]:
        """加载deployment.yaml文件（中文提示）"""
        try:
            with open(self.deployment_yaml_path, 'r') as file:
                return yaml.safe_load(file)
        except FileNotFoundError:
            print(f"未找到 {self.deployment_yaml_path}，请先在 Helm 模板中添加 artifact 的 Deployment。")
            sys.exit(1)
        except Exception as e:
            print(f"加载deployment.yaml失败: {e}")
            sys.exit(1)

    def save_deployment_yaml(self, data: Dict[str, Any]) -> None:
        """保存deployment.yaml文件"""
        try:
            with open(self.deployment_yaml_path, 'w') as file:
                yaml.dump(data, file, default_flow_style=False, sort_keys=False)
            print("deployment.yaml已更新")
        except Exception as e:
            print(f"保存deployment.yaml失败: {e}")
            sys.exit(1)

    def update_image_tag(self, tag: str) -> None:
        """更新 artifact 服务的镜像标签（镜像名：localhost:31500/artifact-service:<tag>）"""
        target_image = f"localhost:31500/artifact-service:{tag}"
        print(f"更新 artifact 服务镜像标签为: {target_image}")
        data = self.load_deployment_yaml()

        # 获取容器配置
        containers = data['spec']['template']['spec']['containers']
        updated = False
        for container in containers:
            # 约定容器名为 artifact
            if container.get('name') == 'artifact':
                current_image = container.get('image')
                container['image'] = target_image
                print(f"镜像已从 {current_image} 更新为 {container['image']}")
                updated = True
                break

        if not updated:
            print("未在 deployment.yaml 中找到容器名为 'artifact' 的配置，请检查模板")
            sys.exit(1)

        # 保存文件
        self.save_deployment_yaml(data)

    def build_image(self, tag: str) -> bool:
        """构建Docker镜像（使用 apps/artifact/Dockerfile）"""
        image = f"localhost:31500/artifact-service:{tag}"
        print(f"构建 artifact 服务镜像: {image}")
        cmd = ["docker", "build", "-t", image, "-f", "apps/artifact/Dockerfile", "."]
        result = subprocess.run(cmd, cwd=self.project_root)

        if result.returncode != 0:
            print("镜像构建失败")
            return False

        print("镜像构建成功")
        return True

    def push_image(self, tag: str) -> bool:
        """推送Docker镜像"""
        image = f"localhost:31500/artifact-service:{tag}"
        print(f"推送 artifact 服务镜像: {image}")
        cmd = ["docker", "push", image]
        result = subprocess.run(cmd)

        if result.returncode != 0:
            print("镜像推送失败")
            return False

        print("镜像推送成功")
        return True

    def deploy(self) -> bool:
        """使用Helm部署或升级服务（流程与 code_repository 对齐）"""
        print("使用Helm部署 artifact 服务...")

        # 检查命名空间是否存在
        ns_check_cmd = ["kubectl", "get", "namespace", self.namespace]
        ns_result = subprocess.run(ns_check_cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE)

        if ns_result.returncode != 0:
            print(f"命名空间 {self.namespace} 不存在，将创建...")
            create_ns_cmd = ["kubectl", "create", "namespace", self.namespace]
            create_result = subprocess.run(create_ns_cmd)
            if create_result.returncode != 0:
                print(f"创建命名空间 {self.namespace} 失败")
                return False

        # 检查是否已安装
        status_cmd = ["helm", "status", self.release_name, "-n", self.namespace]
        status_result = subprocess.run(status_cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE)

        if status_result.returncode == 0:
            # 已安装，执行升级
            print("检测到已安装的版本，执行升级操作...")
            cmd = ["helm", "upgrade", self.release_name, self.chart_path, "-n", self.namespace]
            operation = "升级"
        else:
            # 未安装，执行安装
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
        """检查部署状态（按标签查询 artifact 组件）"""
        print("\n检查部署状态...")

        # 检查Pod状态
        print("\nartifact 服务Pod状态:")
        pods_cmd = ["kubectl", "get", "pods", "-n", self.namespace, "-l", "app.kubernetes.io/component=artifact"]
        subprocess.run(pods_cmd)

        # 检查服务状态
        print("\n服务列表:")
        svc_cmd = ["kubectl", "get", "services", "-n", self.namespace]
        subprocess.run(svc_cmd)

    def get_logs(self) -> None:
        """获取 artifact 服务日志"""
        sleep(3)
        print("\n获取 artifact 服务日志...")
        logs_cmd = ["kubectl", "logs", "-n", self.namespace, "-l", "app.kubernetes.io/component=artifact", "--tail=20"]
        subprocess.run(logs_cmd)

    def deploy_artifact_service(self, tag: str, push_image: bool = True) -> bool:
        """完整的 artifact 服务部署流程（与代码仓库一致的6步）"""
        print(f"开始 artifact 服务部署流程，镜像标签: {tag}")

        # 步骤1: 更新镜像标签
        self.update_image_tag(tag)

        # 步骤2: 构建镜像
        if not self.build_image(tag):
            return False

        # 步骤3: 推送镜像（如果需要）
        if push_image:
            if not self.push_image(tag):
                return False

        # 步骤4: 部署到Kubernetes
        if not self.deploy():
            return False

        # 步骤5: 检查部署状态
        self.check_deployment_status()

        # 步骤6: 获取日志
        self.get_logs()

        print("\nartifact 服务部署流程完成！")
        return True


def main():
    parser = argparse.ArgumentParser(description="artifact 服务部署脚本")
    parser.add_argument("--tag", help="镜像标签（可选，未提供时自动生成随机值）")
    args = parser.parse_args()

    deployer = ArtifactDeployer()

    # 如果没有提供tag，则生成随机值
    if args.tag:
        tag = args.tag
        print(f"使用提供的镜像标签: {tag}")
    else:
        tag = deployer.generate_random_tag()
        print(f"未提供镜像标签，自动生成随机标签: {tag}")

    success = deployer.deploy_artifact_service(tag)

    if not success:
        sys.exit(1)


if __name__ == "__main__":
    main()