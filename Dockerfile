FROM golang:1.25.4

WORKDIR ${GOPATH}/pr-assignment-service/

COPY go.mod go.sum ./
RUN go mod download && \
    go install github.com/pressly/goose/v3/cmd/goose@latest

COPY . .

RUN go build -o bin/pr-assignment-service ./cmd/pr-assignment-service

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=1m --start-period=30s --start-interval=10s --retries=2 \
  CMD curl -f http://localhost:8080/ping

ENTRYPOINT ["./bin/pr-assignment-service"]
