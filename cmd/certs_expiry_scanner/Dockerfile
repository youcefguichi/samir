FROM golang:1.23-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o certs_expiry_scanner main.go certs_expiry_scanner.go


FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/certs_expiry_scanner .

EXPOSE 3005

CMD ["./certs_expiry_scanner"]
