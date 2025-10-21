FROM --platform=$BUILDPLATFORM golang:1.24-alpine AS nobled-builder

RUN apk add --no-cache \
    build-base \
    git \
    linux-headers

WORKDIR /src

COPY go.mod go.sum ./

RUN go mod download

COPY . .

# https://www.docker.com/blog/faster-multi-platform-builds-dockerfile-cross-compilation-guide
ARG TARGETOS TARGETARCH

RUN GOOS=$TARGETOS GOARCH=$TARGETARCH make build LDFLAGS='-linkmode external -extldflags "-static"'

FROM alpine:3

WORKDIR /root

COPY --from=nobled-builder /src/build/nobled /usr/bin/nobled
