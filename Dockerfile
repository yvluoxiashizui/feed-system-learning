# 构建阶段：编译 API 与 Worker 两个二进制
FROM golang:1.26-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o /out/api ./cmd/api \
 && CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o /out/worker ./cmd/worker

# 运行阶段
FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /out/api /app/api
COPY --from=builder /out/worker /app/worker
COPY preview.html /app/preview.html
EXPOSE 8080
ENTRYPOINT ["/app/api"]
