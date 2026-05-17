# ---------- Build ----------
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

RUN apk add --no-cache \
    ca-certificates \
    dumb-init

# nginx config
RUN rm /etc/nginx/conf.d/default.conf
COPY nginx.conf /etc/nginx/nginx.conf
COPY site.conf /etc/nginx/conf.d/site.conf

# app binary
COPY --from=builder /app/uploader /usr/local/bin/uploader

# writable directories for arbitrary UID
RUN mkdir -p \
    /var/www/site \
    /tmp/nginx \
    /var/cache/nginx \
    /var/run/nginx && \
    chgrp -R 0 \
    /var/www \
    /tmp/nginx \
    /var/cache/nginx \
    /var/run/nginx && \
    chmod -R g=u \
    /var/www \
    /tmp/nginx \
    /var/cache/nginx \
    /var/run/nginx

# non-root default user
# runtime may override with any UID
USER 65532:0

EXPOSE 8080

ENTRYPOINT ["dumb-init", "--"]

COPY docker-entrypoint.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/docker-entrypoint.sh

ENTRYPOINT ["dumb-init", "--", "/usr/local/bin/docker-entrypoint.sh"]
