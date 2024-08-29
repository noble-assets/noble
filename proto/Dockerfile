FROM bufbuild/buf:1.39 as BUILDER
FROM golang:1.21-alpine

RUN apk add --no-cache \
  nodejs \
  npm \
  git \
  make

RUN go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway@latest

RUN git clone https://github.com/regen-network/protobuf.git; \
  cd protobuf; \
  go mod download; \
  make install; \
  cd ..

RUN git clone https://github.com/regen-network/cosmos-proto.git; \
  cd cosmos-proto/protoc-gen-gocosmos; \
  go install .; \
  cd ..

COPY --from=BUILDER /usr/local/bin /usr/local/bin
