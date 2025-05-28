FROM node:21 AS node_build

WORKDIR /go/src/github.com/siyuan-note/siyuan/
ADD . /go/src/github.com/siyuan-note/siyuan/

# 安装jq和构建前端
RUN apt-get update && \
    apt-get install -y jq && \
    cd app && \
    packageManager=$(jq -r '.packageManager' package.json) && \
    if [ -n "$packageManager" ]; then \
        npm install -g $packageManager; \
    else \
        echo "No packageManager field found in package.json"; \
        npm install -g pnpm; \
    fi && \
    pnpm install --registry=http://registry.npmjs.org/ --silent && \
    pnpm run build && \
    cd .. && \
    apt-get purge -y jq && \
    apt-get autoremove -y && \
    rm -rf /var/lib/apt/lists/*

FROM golang:1.24-alpine AS go_build
WORKDIR /go/src/github.com/siyuan-note/siyuan/
COPY --from=node_build /go/src/github.com/siyuan-note/siyuan/ /go/src/github.com/siyuan-note/siyuan/

ENV GO111MODULE=on
ENV CGO_ENABLED=1

# install dependencies
RUN apk add --no-cache gcc musl-dev

# build application
RUN cd kernel && \
    echo "Building Go application..." && \
    go build --tags fts5 -v -ldflags "-s -w" && \
    echo "Go build completed" && \
    ls -la

# prepare dirs
RUN mkdir -p /opt/siyuan/ && \
    echo "Moving files..." && \
    test -d /go/src/github.com/siyuan-note/siyuan/app/appearance/ && \
        mv /go/src/github.com/siyuan-note/siyuan/app/appearance/ /opt/siyuan/ || \
        echo "Warning: appearance directory not found" && \
    test -d /go/src/github.com/siyuan-note/siyuan/app/stage/ && \
        mv /go/src/github.com/siyuan-note/siyuan/app/stage/ /opt/siyuan/ || \
        echo "Warning: stage directory not found" && \
    test -d /go/src/github.com/siyuan-note/siyuan/app/guide/ && \
        mv /go/src/github.com/siyuan-note/siyuan/app/guide/ /opt/siyuan/ || \
        echo "Warning: guide directory not found" && \
    test -d /go/src/github.com/siyuan-note/siyuan/app/changelogs/ && \
        mv /go/src/github.com/siyuan-note/siyuan/app/changelogs/ /opt/siyuan/ || \
        echo "Warning: changelogs directory not found" && \
    test -f /go/src/github.com/siyuan-note/siyuan/kernel/kernel && \
        mv /go/src/github.com/siyuan-note/siyuan/kernel/kernel /opt/siyuan/ || \
        echo "Error: kernel binary not found" && \
    test -f /go/src/github.com/siyuan-note/siyuan/kernel/entrypoint.sh && \
        mv /go/src/github.com/siyuan-note/siyuan/kernel/entrypoint.sh /opt/siyuan/entrypoint.sh || \
        echo "Error: entrypoint.sh not found"

# remove git
RUN find /opt/siyuan/ -name .git -type d -exec rm -rf {} + || true

# verify build
RUN ls -la /opt/siyuan/ && \
    echo "Build verification completed"

FROM alpine:latest
LABEL maintainer="Liang Ding<845765@qq.com>"

WORKDIR /opt/siyuan/
COPY --from=go_build /opt/siyuan/ /opt/siyuan/

RUN apk add --no-cache ca-certificates tzdata su-exec && \
    chmod +x /opt/siyuan/entrypoint.sh

ENV TZ=Asia/Shanghai
ENV HOME=/home/siyuan
ENV RUN_IN_CONTAINER=true
EXPOSE 6806

ENTRYPOINT ["/opt/siyuan/entrypoint.sh"]
CMD ["/opt/siyuan/kernel"]