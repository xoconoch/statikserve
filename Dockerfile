FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY main.go .
RUN CGO_ENABLED=0 GOOS=linux go build -o server main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /home/appuser

COPY --from=builder /app/server .

RUN mkdir -p /home/appuser/site && \
    chown -R appuser:appgroup /home/appuser && \
    chmod -R 755 /home/appuser

USER appuser

EXPOSE 80

CMD ["./server"]
