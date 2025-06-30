FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /simple-sa-token-issuer

FROM alpine:latest AS app

WORKDIR /

RUN addgroup -g 1001 golang && \
    adduser --shell /sbin/nologin --disabled-password \
    --no-create-home --uid 1001 --ingroup golang golang

COPY --from=builder /simple-sa-token-issuer /simple-sa-token-issuer

USER golang

ENTRYPOINT ["/simple-sa-token-issuer"]