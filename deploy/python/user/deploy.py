#!/usr/bin/env python3
"""
用户服务部署脚本
功能：读取YAML配置、修改镜像标签、构建镜像、推送和部署到Kubernetes
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

class UserDeployer:
    def __init__(self, project_root: str = "/home/hr/xcoding"):
        self.project_root = project_root
        self.deployment_yaml_path = os.path.join(project_root, "deploy/xcoding/templates/services/user/deployment.yaml")
        self.chart_path = os.path.join(project_root, "deploy/xcoding")
        self.namespace = "xcoding"
        self.release_name = "xcoding"
        
    def generate_random_tag(self, length: int = 8) -> str:
        """生成随机标签"""
        return ''.join(random.choices(string.ascii_lowercase + string.digits, k=length))
    
    def load_deployment_yaml(self) -> Dict[str, Any]:
        """加载deployment.yaml文件"""
        try:
            with open(self.deployment_yaml_path, 'r') as file:
                return yaml.safe_load(file)
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
        """更新用户服务的镜像标签"""
        print(f"更新用户服务镜像标签为: {tag}")
        data = self.load_deployment_yaml()
        
        # 获取容器配置
        containers = data['spec']['template']['spec']['containers']
        for container in containers:
            if container['name'] == 'user':
                # 更新镜像标签
                current_image = container['image']
                # 使用我们的私有registry
                container['image'] = f"localhost:31500/user-service:{tag}"
                print(f"镜像已从 {current_image} 更新为 {container['image']}")
                break
        
        # 保存文件
        self.save_deployment_yaml(data)
    
    def build_image(self, tag: str) -> bool:
        """构建Docker镜像"""
        print(f"构建用户服务镜像: localhost:31500/user-service:{tag}")
        cmd = ["docker", "build", "-t", f"localhost:31500/user-service:{tag}", "-f", "apps/user/Dockerfile", "."]
        result = subprocess.run(cmd, cwd=self.project_root)
        
        if result.returncode != 0:
            print("镜像构建失败")
            return False
        
        print("镜像构建成功")
        return True
    
    def push_image(self, tag: str) -> bool:
        """推送Docker镜像"""
        print(f"推送用户服务镜像: localhost:31500/user-service:{tag}")
        cmd = ["docker", "push", f"localhost:31500/user-service:{tag}"]
        result = subprocess.run(cmd)
        
        if result.returncode != 0:
            print("镜像推送失败")
            return False
        
        print("镜像推送成功")
        return True
    
    def deploy(self) -> bool:
        """使用Helm部署或升级服务"""
        print("使用Helm部署用户服务...")
        
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
        """检查部署状态"""
        print("\n检查部署状态...")
        
        # 检查Pod状态
        print("\n用户服务Pod状态:")
        pods_cmd = ["kubectl", "get", "pods", "-n", self.namespace, "-l", "app.kubernetes.io/component=user"]
        subprocess.run(pods_cmd)
        
        # 检查服务状态
        print("\n用户服务状态:")
        svc_cmd = ["kubectl", "get", "services", "-n", self.namespace]
        subprocess.run(svc_cmd)
    
    def get_logs(self) -> None:
        """获取用户服务日志"""
        sleep(3)
        print("\n获取用户服务日志...")
        logs_cmd = ["kubectl", "logs", "-n", self.namespace, "-l", "app.kubernetes.io/component=user", "--tail=20"]
        subprocess.run(logs_cmd)
    
    def deploy_user_service(self, tag: str, push_image: bool = True) -> bool:
        """完整的用户服务部署流程"""
        print(f"开始用户服务部署流程，镜像标签: {tag}")
        
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
        
        print("\n用户服务部署流程完成！")
        return True


def main():
    parser = argparse.ArgumentParser(description="用户服务部署脚本")
    parser.add_argument("--tag", help="镜像标签（可选，未提供时将自动生成随机值）")
    
    args = parser.parse_args()
    
    deployer = UserDeployer()
    
    # 如果没有提供tag，则生成随机值
    if args.tag:
        tag = args.tag
        print(f"使用提供的镜像标签: {tag}")
    else:
        tag = deployer.generate_random_tag()
        print(f"未提供镜像标签，自动生成随机标签: {tag}")
    
    success = deployer.deploy_user_service(tag)
    
    if not success:
        sys.exit(1)


if __name__ == "__main__":
    main()