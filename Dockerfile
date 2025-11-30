# 第一阶段：构建阶段
# 使用具体版本的Go镜像，确保与go.mod中的版本兼容
FROM golang:1.21-alpine AS builder

# 设置Go环境变量
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# 安装构建所需的依赖
RUN apk add --no-cache git

# 创建工作目录
WORKDIR /build

# 复制go.mod和go.sum文件
COPY go.mod ./

# 打印Go版本信息用于调试
RUN go version

# 下载依赖（利用缓存层）
RUN echo "正在下载依赖..."
RUN go mod download
RUN go mod tidy

# 复制所有源代码
COPY . .

# 编译应用程序，禁用CGO以支持alpine基础镜像
RUN echo "正在编译应用..."
RUN go build -v -ldflags="-s -w" -o javbus-api .

# 第二阶段：运行阶段 - 使用alpine作为基础镜像
FROM alpine:latest

# 添加时区包和证书
RUN apk --no-cache add tzdata ca-certificates

# 设置工作目录
WORKDIR /app

# 从构建阶段复制编译好的二进制文件
COPY --from=builder /build/javbus-api /app/

# 复制必要的配置文件
COPY config.toml /app/
COPY .env.sample /app/.env

# 创建必要的目录
RUN mkdir -p /app/data

# 设置可执行权限
RUN chmod +x /app/javbus-api

# 声明暴露的端口
EXPOSE 3000

# 定义环境变量
ENV GIN_MODE=release

# 启动命令
CMD ["./javbus-api"]