FROM golang:1.23.1-alpine AS builder
WORKDIR /messanger_src

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
WORKDIR /messanger_src

# bin
COPY --from=builder /messanger_src/bin/app /messanger_src/app
COPY --from=builder /messanger_src/web/templates/html /messanger_src/web/templates/html

EXPOSE 8080
CMD ["/messanger_src/app"]