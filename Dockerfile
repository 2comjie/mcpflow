# ---- 后端构建 ----
FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o mcpflow ./app/server

# ---- 运行 ----
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/mcpflow .
COPY configs/config.yml configs/config.yml

EXPOSE 8080
CMD ["./mcpflow"]
