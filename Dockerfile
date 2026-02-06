# 构建阶段
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

# 复制依赖文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY src/ ./src/

# 编译
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o emqgodb ./src

# 运行阶段
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

# 创建非root用户
RUN adduser -D -g '' appuser

WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /build/emqgodb .

# 创建配置目录
RUN mkdir -p /app/config && chown -R appuser:appuser /app

USER appuser

ENTRYPOINT ["./emqgodb"]
CMD ["-config", "/app/config/config.yaml"]