#!/usr/bin/env python3
import argparse
import subprocess
import json
import os
from datetime import datetime

TEMPLATES_DIR = os.path.abspath(os.path.join(os.path.dirname(__file__), '../../xcoding/templates/services/executor_service'))
CHART_DIR = os.path.abspath(os.path.join(os.path.dirname(__file__), '../../xcoding'))
DEPLOYMENT_FILE = os.path.join(TEMPLATES_DIR, 'deployment.yaml')
SERVICE_FILE = os.path.join(TEMPLATES_DIR, 'service.yaml')

REGISTRY = os.environ.get('REGISTRY', 'localhost:31500')
IMAGE_NAME = 'ci-executor-service'
DEFAULT_TAG = 'latest'

class ExecutorDeployer:
    def __init__(self, image_tag=None):
        self.registry = REGISTRY
        self.image_name = IMAGE_NAME
        self.image_tag = image_tag or DEFAULT_TAG
        self.full_image = f"{self.registry}/{self.image_name}:{self.image_tag}"
        self.release = 'xcoding'
        self.namespace = 'xcoding'

    def _run(self, cmd, cwd=None):
        print(f"$ {' '.join(cmd)}")
        subprocess.check_call(cmd, cwd=cwd)

    def update_deployment_image(self):
        with open(DEPLOYMENT_FILE, 'r') as f:
            content = f.read()
        updated = []
        for line in content.splitlines():
            if line.strip().startswith('image:'):
                updated.append(f"        image: {self.full_image}")
            else:
                updated.append(line)
        with open(DEPLOYMENT_FILE, 'w') as f:
            f.write('\n'.join(updated) + '\n')
        print(f"Updated executor deployment image to {self.full_image}")

    def build_and_push(self):
        root = os.path.abspath(os.path.join(os.path.dirname(__file__), '../../..'))
        dockerfile = os.path.abspath(os.path.join(root, 'apps', 'ci', 'executor_service', 'Dockerfile'))
        self._run(['docker', 'build', '-t', self.full_image, '-f', dockerfile, root])
        self._run(['docker', 'push', self.full_image])

    def helm_upgrade(self):
        self._run(['helm', 'upgrade', '--install', self.release, CHART_DIR, '--namespace', self.namespace])

    def status(self):
        self._run(['helm', 'status', self.release, '--namespace', self.namespace])

    def logs(self):
        cmd = ['kubectl', 'get', 'pods', '-n', self.namespace, '-l', 'app.kubernetes.io/component=ci-executor', '-o', 'json']
        out = subprocess.check_output(cmd)
        data = json.loads(out)
        items = data.get('items', [])
        if not items:
            print('No pods found for ci-executor')
            return
        name = items[0]['metadata']['name']
        self._run(['kubectl', 'logs', '-n', self.namespace, name])

def main():
    parser = argparse.ArgumentParser(description='Deploy executor service')
    parser.add_argument('--tag', help='Docker image tag', default=None)
    parser.add_argument('--release', help='Helm release name', default='xcoding')
    parser.add_argument('--namespace', help='Kubernetes namespace', default='xcoding')
    parser.add_argument('--action', choices=['deploy', 'status', 'logs'], default='deploy')
    args = parser.parse_args()

    try:
        from datetime import datetime, timezone
        tag = args.tag or datetime.now(timezone.utc).strftime('%Y%m%d%H%M%S')
    except Exception:
        tag = args.tag or 'latest'
    deployer = ExecutorDeployer(image_tag=tag)
    deployer.release = args.release
    deployer.namespace = args.namespace

    if args.action == 'deploy':
        deployer.update_deployment_image()
        deployer.build_and_push()
        deployer.helm_upgrade()
        deployer.status()
    elif args.action == 'status':
        deployer.status()
    elif args.action == 'logs':
        deployer.logs()

if __name__ == '__main__':
    main()
