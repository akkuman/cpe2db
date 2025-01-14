FROM mcr.microsoft.com/devcontainers/go:1-1.23-bullseye

ARG DEBIAN_FRONTEND=noninteractive
ARG USER=vscode

RUN sed -i "s|http://deb.debian.org/debian|http://mirror.sjtu.edu.cn/debian|g" /etc/apt/sources.list && \
    apt-get update && \
    apt-get install -y bash-completion direnv jq

USER $USER
ARG HOME="/home/$USER"

ENV LANG=C.UTF-8 \
    TZ=Asia/Shanghai \
    PATH="${HOME}/.local/bin:$PATH"

RUN go env -w GO111MODULE=on && \
    go env -w GOPROXY=https://goproxy.cn,direct
