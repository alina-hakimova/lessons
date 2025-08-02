FROM golang:1.19 AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main main.go
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o goworker worker.go

FROM golang:1.19-alpine
WORKDIR /goapp
COPY --from=build /src/main ./
COPY --from=build /src/goworker ./

CMD ["./main"]
