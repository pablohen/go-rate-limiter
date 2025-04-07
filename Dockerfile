FROM golang:1.24.2 AS builder

WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /rate-limiter ./cmd/server/main.go

FROM alpine:latest
COPY --from=builder /rate-limiter /rate-limiter
COPY --from=builder /app/cmd/server/.env /.env
EXPOSE 8080
CMD ["/rate-limiter"]