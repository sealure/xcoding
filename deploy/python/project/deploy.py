import argparse
import subprocess
import json
import os
import sys
from datetime import datetime

TEMPLATES_DIR = os.path.abspath(os.path.join(os.path.dirname(__file__), '../../xcoding/templates/services/project'))
CHART_DIR = os.path.abspath(os.path.join(os.path.dirname(__file__), '../../xcoding'))
DEPLOYMENT_FILE = os.path.join(TEMPLATES_DIR, 'deployment.yaml')
SERVICE_FILE = os.path.join(TEMPLATES_DIR, 'service.yaml')

REGISTRY = os.environ.get('REGISTRY', 'localhost:31500')
IMAGE_NAME = 'project-service'
DEFAULT_TAG = 'latest'

class ProjectDeployer:
    def __init__(self, image_tag=None):
        self.registry = REGISTRY
        self.image_name = IMAGE_NAME
        self.image_tag = image_tag or DEFAULT_TAG
        self.full_image = f"{self.registry}/{self.image_name}:{self.image_tag}"

    def _run(self, cmd):
        print(f"$ {' '.join(cmd)}")
        subprocess.check_call(cmd)

    def update_deployment_image(self):
        # Update image in deployment.yaml
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
        print(f"Updated deployment image to {self.full_image}")

    def build_and_push(self):
        # Build docker image
        self._run(['docker', 'build', '-t', self.full_image, '-f', os.path.abspath(os.path.join(os.path.dirname(__file__), '../../..', 'apps', 'project', 'Dockerfile')), os.path.abspath(os.path.join(os.path.dirname(__file__), '../../..'))])
        # Push image
        self._run(['docker', 'push', self.full_image])

    def helm_upgrade(self, release='xcoding', namespace='default'):
        # Helm upgrade/install
        self._run(['helm', 'upgrade', '--install', release, CHART_DIR, '--namespace', namespace])

    def status(self, release='xcoding', namespace='default'):
        self._run(['helm', 'status', release, '--namespace', namespace])

    def logs(self, selector='app.kubernetes.io/component=project', namespace='default'):
        # Get pod name by selector
        cmd = ['kubectl', 'get', 'pods', '-n', namespace, '-l', selector, '-o', 'json']
        out = subprocess.check_output(cmd)
        data = json.loads(out)
        items = data.get('items', [])
        if not items:
            print('No pods found for selector:', selector)
            return
        name = items[0]['metadata']['name']
        self._run(['kubectl', 'logs', '-n', namespace, name])


def main():
    parser = argparse.ArgumentParser(description='Deploy project service')
    parser.add_argument('--tag', help='Docker image tag', default=None)
    parser.add_argument('--release', help='Helm release name', default='xcoding')
    parser.add_argument('--namespace', help='Kubernetes namespace', default='xcoding')
    parser.add_argument('--action', choices=['deploy', 'status', 'logs'], default='deploy')
    args = parser.parse_args()

    tag = args.tag or datetime.utcnow().strftime('%Y%m%d%H%M%S')

    deployer = ProjectDeployer(image_tag=tag)

    if args.action == 'deploy':
        deployer.update_deployment_image()
        deployer.build_and_push()
        deployer.helm_upgrade(release=args.release, namespace=args.namespace)
        deployer.status(release=args.release, namespace=args.namespace)
    elif args.action == 'status':
        deployer.status(release=args.release, namespace=args.namespace)
    elif args.action == 'logs':
        deployer.logs(namespace=args.namespace)

if __name__ == '__main__':
    main()