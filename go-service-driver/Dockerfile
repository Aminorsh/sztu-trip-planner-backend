FROM m.daocloud.io/docker.io/library/golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go env -w GOPROXY=https://goproxy.cn,direct && go mod download

COPY . .
# RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=vendor -o server cmd/server/main.go
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server cmd/main.go

FROM m.daocloud.io/docker.io/library/alpine:latest

WORKDIR /app
COPY --from=builder /app/server .
COPY .env .

EXPOSE 8080

CMD ["./server"]