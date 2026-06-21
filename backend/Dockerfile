FROM golang:1.25-alpine AS builder
RUN apk add --no-cache gcc musl-dev
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o keep-server ./cmd/keep_server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /build/keep-server /usr/local/bin/keep-server
EXPOSE 8080
ENTRYPOINT ["keep-server"]
