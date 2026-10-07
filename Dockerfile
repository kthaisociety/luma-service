FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o luma-service ./cmd/api

FROM alpine:latest

RUN apk --no-cache add ca-certificates \
    && adduser -D -u 10001 app

WORKDIR /app
COPY --from=builder /app/luma-service .

ENV HOST=0.0.0.0 \
    PORT=8000

USER app
EXPOSE 8000

CMD ["./luma-service"]
