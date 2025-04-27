FROM golang:1.23.1-alpine AS builder
WORKDIR /usr/local/src

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
WORKDIR /usr/local/src

# bin
COPY --from=builder /usr/local/src/bin/app /usr/local/src/app
COPY --from=builder /usr/local/src/web/templates/html /usr/local/src/web/templates/html

EXPOSE 8080
CMD ["/usr/local/src/app"]