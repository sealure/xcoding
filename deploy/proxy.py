import subprocess
def prepare():
    subprocess.Popen(["kubectl", "port-forward", "service/postgresql", "--address=localhost", "5432:5432", "--namespace=xcoding"], start_new_session=True)
    subprocess.Popen(["kubectl", "port-forward", "service/docker-registry", "--address=localhost", "31500:5000", "--namespace=xcoding"], start_new_session=True)
    subprocess.Popen(["kubectl", "port-forward", "service/apisix-gateway", "--address=localhost", "31080:80", "--namespace=xcoding"], start_new_session=True)
    subprocess.Popen(["kubectl", "port-forward", "service/rabbitmq", "--address=localhost", "5672:5672", "--namespace=xcoding"], start_new_session=True)
    subprocess.Popen(["kubectl", "port-forward", "service/rabbitmq", "--address=localhost", "15672:15672", "--namespace=xcoding"], start_new_session=True)


def main():
    #
    prepare()

if __name__ == "__main__":
    main()