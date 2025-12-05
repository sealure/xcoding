FROM python:3.11

# 设置不交互与无缓存，安装 tree 工具并清理 apt 缓存
RUN apt-get update \
    && apt install -y tree 


# 缺省工作目录
WORKDIR /workspace
