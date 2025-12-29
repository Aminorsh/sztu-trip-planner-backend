FROM m.daocloud.io/docker.io/library/golang:1.24-alpine AS builder

WORKDIR /app
COPY go-service-driver/go.mod go-service-driver/go.sum ./
RUN go env -w GOPROXY=https://goproxy.cn,direct && go mod download

COPY go-service-driver/ .
RUN go build -o server cmd/server/main.go

FROM m.daocloud.io/docker.io/library/alpine:3.18

WORKDIR /app
COPY --from=builder /app/server .
COPY .env .

EXPOSE 8080

CMD ["./server"]