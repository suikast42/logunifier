#https://docs.docker.com/build/building/multi-stage/
FROM golang:1.19.4-alpine AS builder

WORKDIR /logunifier
COPY . .
RUN go mod tidy
WORKDIR /logunifier/cmd/logunifier
RUN  GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o logunifier
##docker build --target builder
FROM alpine:3.16.2
COPY --from=builder /logunifier/cmd/logunifier/ /logunifier
COPY --from=builder /logunifier/internal/config/local.cfg /cfg/local.cfg
ENTRYPOINT [ "/logunifier --config /cfg/local.cfg" ]