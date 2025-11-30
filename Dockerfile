# 第一阶段：构建阶段
FROM golang:1.24-alpine AS builder

# 设置Go环境变量
# 注意：如果你是在 M1/M2 Mac 上构建并部署到 Linux 服务器，保持 GOARCH=amd64 是对的
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# 安装构建所需的依赖
RUN apk add --no-cache git

# 创建工作目录
WORKDIR /build
# 3. 复制所有源代码 (这一步必须在 tidy 之前)
COPY . .

# 4. 【关键修正】源码复制进去后，再执行 tidy
# 确保所有依赖都被正确解析并补全
RUN go mod tidy && echo "正在编译应用..." && go build -v -ldflags="-s -w" -o javbus-api .

# 第二阶段：运行阶段
FROM alpine:latest

# 添加时区包和证书
RUN apk --no-cache add tzdata ca-certificates

# 设置工作目录
WORKDIR /app

# 从构建阶段复制编译好的二进制文件
COPY --from=builder /build/javbus-api /app/

# 复制必要的配置文件
# 注意：请确保 config.toml 和 .env.sample 真的存在于你本地目录，否则会报错
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