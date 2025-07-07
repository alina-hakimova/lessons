FROM golang:1.22.2-alpine
WORKDIR /app
COPY go.mod .
COPY main.go .
RUN go mod download
COPY . .
RUN go build -o notes .
CMD ["./notes"]
