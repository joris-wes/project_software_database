FROM golang:1.19-alpine AS builder

WORKDIR /app

ENV CGO_ENABLED 0
ENV GOPATH /go
ENV GOCACHE /go-build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o bin/listener .

CMD ["/app/bin/listener"]