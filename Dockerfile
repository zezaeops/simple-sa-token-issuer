FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /simple-sa-token-issuer

FROM alpine:latest AS app

WORKDIR /

COPY --from=builder /simple-sa-token-issuer /simple-sa-token-issuer

ENTRYPOINT ["/simple-sa-token-issuer"]