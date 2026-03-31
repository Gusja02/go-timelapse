FROM golang:1.26-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY main.go .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o timelapse main.go

FROM alpine:latest
RUN apk add --no-cache ffmpeg ca-certificates tzdata
WORKDIR /root/
COPY --from=builder /app/timelapse .

CMD ["./timelapse"]