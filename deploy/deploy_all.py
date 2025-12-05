import subprocess


def main():
    subprocess.run(["python", "deploy/python/user/deploy.py"])
    subprocess.run(["python", "deploy/python/project/deploy.py"])
    subprocess.run(["python", "deploy/python/code_repository/deploy.py"])
    subprocess.run(["python", "deploy/python/artifact/deploy.py"])
    subprocess.run(["python", "deploy/python/ci/deploy.py"])
    

if __name__ == "__main__":
    main()