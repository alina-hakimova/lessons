FROM golang:1.21-alpine
 WORKDIR /app
 COPY main.go .
 RUN go build -o goapp main.go
 CMD ["./goapp"]