FROM golang:1.25-alpine AS builder

WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN go mod tidy && CGO_ENABLED=0 GOOS=linux go build -o /out/demo-mcp .

FROM alpine:3.22

WORKDIR /app
COPY --from=builder /out/demo-mcp .
EXPOSE 3003
ENV ADDR=:3003
CMD ["./demo-mcp"]
