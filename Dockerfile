# ==================== 前端构建 ====================
FROM node:20-alpine AS frontend
WORKDIR /app/web
COPY web/package.json web/package-lock.json* ./
RUN npm install
COPY web/ .
RUN npm run build

# ==================== 后端构建 ====================
FROM golang:1.26-alpine AS backend
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o server ./app/server/

# ==================== 最终镜像 ====================
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app

COPY --from=backend /app/server .
COPY --from=frontend /app/web/dist ./web/dist
COPY configs/config.yml ./configs/config.yml

ENV GIN_MODE=release
EXPOSE 8080

CMD ["./server"]
