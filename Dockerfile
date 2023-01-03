#https://docs.docker.com/build/building/multi-stage/
FROM golang:1.19.4-alpine AS builder
RUN  apk add git
WORKDIR /logunifier
ENV CGO_ENABLED=0
ENV GOCACHE=/root/.cache/go-build
ENV GOMODCACHE=/root/.cache/go-build
ADD go.* .
RUN  --mount=type=cache,target=/root/.cache/go-build  go mod download -x



COPY . .
RUN  --mount=type=cache,target=/root/.cache/go-build  go mod tidy
WORKDIR /logunifier/cmd/logunifier
RUN  --mount=type=cache,target=/root/.cache/go-build \
GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /out/logunifier .
###docker build --target builder
FROM alpine:3.16.2
COPY --from=builder /out/logunifier /logunifier
COPY --from=builder /logunifier/internal/config/local.cfg /cfg/local.cfg
ENTRYPOINT [ "/logunifier","-config" ,"/cfg/local.cfg" ]


