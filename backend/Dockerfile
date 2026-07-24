FROM golang:1.26-alpine AS builder

WORKDIR /equislice-backend

COPY go.mod go.sum ./
RUN go mod download

COPY  . .

RUN GOOS=linux GOARCH=amd64 go build -o /equislice-package ./cmd/main.go

FROM alpine:3.21

WORKDIR /root/

COPY --from=builder /equislice-package /equislice-package

EXPOSE 3000

ENTRYPOINT ["/equislice-package"]