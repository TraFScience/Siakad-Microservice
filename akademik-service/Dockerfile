# Stage 1: Build
FROM golang:1.24-alpine AS builder
WORKDIR /app
RUN apk add --no-cache ca-certificates tzdata
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o service main.go

# Stage 2: Runtime
FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata
RUN adduser -D -u 10001 appuser
WORKDIR /app
COPY --from=builder /app/service .
USER appuser
EXPOSE 8080
ENTRYPOINT ["./service"]
