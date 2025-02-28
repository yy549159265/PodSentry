# 使用官方的Go镜像作为构建环境
FROM golang:1.23 AS builder

# 设置工作目录
WORKDIR /app

# 复制go.mod和go.sum文件
COPY ../kubeconfig-monitor/go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY .. .

# 构建应用程序
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# 使用轻量级的Alpine镜像作为运行环境
FROM alpine:latest

# 设置工作目录
WORKDIR /root/

# 从构建阶段复制二进制文件
COPY --from=builder /app/main .



# 运行应用程序
CMD ["./main"]
