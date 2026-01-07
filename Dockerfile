# ---------- Build Go binary ----------
FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /app
COPY go.mod ./
COPY main.go ./

RUN CGO_ENABLED=0 \
    GOOS=$TARGETOS \
    GOARCH=$TARGETARCH \
    go build -o uploader

# ---------- Runtime ----------
FROM nginx:1.25-alpine

RUN apk add --no-cache ca-certificates

RUN rm /etc/nginx/conf.d/default.conf
COPY nginx.conf /etc/nginx/conf.d/site.conf

COPY --from=builder /app/uploader /usr/local/bin/uploader

RUN mkdir -p /var/www/site

EXPOSE 80

CMD ["/bin/sh", "-c", "uploader & nginx -g 'daemon off;'"]
