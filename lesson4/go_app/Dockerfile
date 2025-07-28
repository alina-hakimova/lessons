FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY main.go .
RUN go build -o app main.go

FROM alpine:3.20
WORKDIR /app
ARG UID
ARG GID
RUN addgroup -g $GID appgroup && adduser -D -u $UID -G appgroup appuser
COPY --from=builder /app/app .
USER appuser
EXPOSE 8003
HEALTHCHECK --interval=30s --timeout=10s --retries=3 CMD wget -qO- http://localhost:8003/ || exit 1
CMD ["./app"]
