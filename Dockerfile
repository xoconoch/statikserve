FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o server main.go

FROM gcr.io/distroless/base:nonroot

WORKDIR /app

COPY --from=builder --chown=65532:65532 /app/server .

COPY --from=builder --chown=65532:65532 /app/site /app/site

EXPOSE 80

CMD ["./server"]
