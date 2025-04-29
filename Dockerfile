FROM golang:1.23.1-alpine AS builder
WORKDIR /pong_src

# dependencies
RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY cmd cmd/
COPY internal internal/
COPY web web/

# build
RUN CGO_ENABLED=0 go build -o bin/app ./cmd/server/main.go

FROM alpine:latest
WORKDIR /pong_src

# bin
COPY --from=builder /pong_src/bin/app /pong_src/app
COPY --from=builder /pong_src/web/templates/html /pong_src/web/templates/html

EXPOSE 8080
CMD ["/pong_src/app"]