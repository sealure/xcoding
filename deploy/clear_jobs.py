#!/usr/bin/env python3
"""
清空xcoding命名空间下所有Jobs的脚本
功能：删除指定命名空间中的所有Jobs和CronJobs
"""

import subprocess
import sys
import argparse
from typing import List, Tuple


class JobCleaner:
    def __init__(self, namespace: str = "xcoding", dry_run: bool = False):
        self.namespace = namespace
        self.dry_run = dry_run
        self.kubectl_cmd = ["kubectl"]
        
    def run_command(self, cmd: List[str], capture_output: bool = True) -> Tuple[int, str, str]:
        """执行命令并返回结果"""
        if capture_output:
            result = subprocess.run(cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE, 
                                  text=True, encoding='utf-8')
            return result.returncode, result.stdout, result.stderr
        else:
            result = subprocess.run(cmd)
            return result.returncode, "", ""

    def check_namespace_exists(self) -> bool:
        """检查命名空间是否存在"""
        cmd = self.kubectl_cmd + ["get", "namespace", self.namespace]
        code, stdout, stderr = self.run_command(cmd)
        return code == 0

    def get_jobs(self) -> List[str]:
        """获取命名空间下所有的Jobs"""
        cmd = self.kubectl_cmd + ["get", "jobs", "-n", self.namespace, 
                                "-o", "jsonpath={.items[*].metadata.name}"]
        code, stdout, stderr = self.run_command(cmd)
        
        if code != 0:
            print(f"获取jobs失败: {stderr}")
            return []
        
        jobs = stdout.strip().split() if stdout.strip() else []
        return jobs

    def get_cronjobs(self) -> List[str]:
        """获取命名空间下所有的CronJobs"""
        cmd = self.kubectl_cmd + ["get", "cronjobs", "-n", self.namespace,
                                "-o", "jsonpath={.items[*].metadata.name}"]
        code, stdout, stderr = self.run_command(cmd)
        
        if code != 0:
            print(f"获取cronjobs失败: {stderr}")
            return []
            
        cronjobs = stdout.strip().split() if stdout.strip() else []
        return cronjobs

    def delete_job(self, job_name: str) -> bool:
        """删除单个Job"""
        if self.dry_run:
            print(f"[DRY RUN] 将删除Job: {job_name}")
            return True
            
        cmd = self.kubectl_cmd + ["delete", "job", job_name, "-n", self.namespace]
        code, stdout, stderr = self.run_command(cmd)
        
        if code == 0:
            print(f"✅ Job删除成功: {job_name}")
            return True
        else:
            print(f"❌ Job删除失败: {job_name} - {stderr}")
            return False

    def delete_cronjob(self, cronjob_name: str) -> bool:
        """删除单个CronJob"""
        if self.dry_run:
            print(f"[DRY RUN] 将删除CronJob: {cronjob_name}")
            return True
            
        cmd = self.kubectl_cmd + ["delete", "cronjob", cronjob_name, "-n", self.namespace]
        code, stdout, stderr = self.run_command(cmd)
        
        if code == 0:
            print(f"✅ CronJob删除成功: {cronjob_name}")
            return True
        else:
            print(f"❌ CronJob删除失败: {cronjob_name} - {stderr}")
            return False

    def delete_job_pods(self) -> bool:
        """清理相关的Pods（包含job标签的Pods）"""
        if self.dry_run:
            print(f"[DRY RUN] 将删除相关Pods")
            return True
            
        cmd = self.kubectl_cmd + ["delete", "pods", "-n", self.namespace, 
                                "-l", "job-name"]
        code, stdout, stderr = self.run_command(cmd)
        
        if code == 0:
            print(f"✅ 相关Pods删除成功")
            return True
        else:
            print(f"❌ 相关Pods删除失败: {stderr}")
            return False

    def clear_all_jobs(self) -> bool:
        """清空所有jobs和cronjobs"""
        print(f"开始清空命名空间 '{self.namespace}' 下的所有Jobs...")
        
        # 检查命名空间是否存在
        if not self.check_namespace_exists():
            print(f"❌ 命名空间 '{self.namespace}' 不存在")
            return False
        
        # 获取所有Jobs
        jobs = self.get_jobs()
        cronjobs = self.get_cronjobs()
        
        print(f"找到 {len(jobs)} 个Jobs，{len(cronjobs)} 个CronJobs")
        
        if not jobs and not cronjobs:
            print("✅ 没有找到需要删除的Jobs或CronJobs")
            return True
        
        success_count = 0
        total_count = len(jobs) + len(cronjobs)
        
        # 删除Jobs
        for job in jobs:
            if self.delete_job(job):
                success_count += 1
        
        # 删除CronJobs
        for cronjob in cronjobs:
            if self.delete_cronjob(cronjob):
                success_count += 1
        
        # 清理相关Pods
        if jobs or cronjobs:
            print("\n清理相关Pods...")
            self.delete_job_pods()
        
        # 输出结果统计
        print(f"\n📊 清理统计:")
        print(f"   总计: {total_count} 个资源")
        print(f"   成功: {success_count} 个")
        print(f"   失败: {total_count - success_count} 个")
        
        if success_count == total_count:
            print("✅ 所有Jobs和CronJobs清理完成！")
            return True
        else:
            print("⚠️  部分Jobs清理失败，请检查上面的错误信息")
            return False

    def show_current_jobs(self):
        """显示当前命名空间下的所有Jobs"""
        print(f"当前命名空间 '{self.namespace}' 下的Jobs状态:")
        
        if not self.check_namespace_exists():
            print(f"❌ 命名空间 '{self.namespace}' 不存在")
            return
        
        # 显示Jobs
        print("\n📋 Jobs:")
        cmd = self.kubectl_cmd + ["get", "jobs", "-n", self.namespace]
        subprocess.run(cmd)
        
        # 显示CronJobs
        print("\n📋 CronJobs:")
        cmd = self.kubectl_cmd + ["get", "cronjobs", "-n", self.namespace]
        subprocess.run(cmd)
        
        # 显示相关Pods
        print("\n📋 相关Pods:")
        cmd = self.kubectl_cmd + ["get", "pods", "-n", self.namespace, "-l", "job-name"]
        subprocess.run(cmd)


def main():
    parser = argparse.ArgumentParser(description='清空指定命名空间下的所有Jobs')
    parser.add_argument('--namespace', '-n', 
                       default='xcoding', 
                       help='目标命名空间 (默认: xcoding)')
    parser.add_argument('--dry-run', 
                       action='store_true', 
                       help='干运行模式，仅显示将要删除的资源，不实际删除')
    parser.add_argument('--show', 
                       action='store_true', 
                       help='仅显示当前命名空间下的Jobs状态，不执行清理')
    
    args = parser.parse_args()
    
    cleaner = JobCleaner(namespace=args.namespace, dry_run=args.dry_run)
    
    if args.show:
        cleaner.show_current_jobs()
    else:
        print("=" * 60)
        if args.dry_run:
            print("🔍 干运行模式 - 不会实际删除任何资源")
            print("=" * 60)
        
        success = cleaner.clear_all_jobs()
        
        if success:
            print("\n✅ 清理任务完成!")
            sys.exit(0)
        else:
            print("\n❌ 清理任务失败!")
            sys.exit(1)


if __name__ == "__main__":
    main()