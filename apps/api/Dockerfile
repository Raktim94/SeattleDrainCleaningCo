FROM golang:1.24-alpine AS builder
WORKDIR /src
COPY . .
# Transitive deps (e.g. via gin) may require Go 1.23+; image must match.
RUN go mod tidy && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/submify-api ./cmd/server

FROM alpine:3.20
WORKDIR /app
RUN adduser -D -g '' submify
COPY --from=builder /out/submify-api /app/submify-api
USER submify
EXPOSE 8080
ENTRYPOINT ["/app/submify-api"]
