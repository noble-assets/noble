FROM bufbuild/buf:latest as BUILDER
FROM golang:1.20-alpine

RUN apk add --no-cache \
  git \
  make \
  nodejs \
  npm

RUN go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway@latest
RUN go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger@latest

RUN git clone https://github.com/regen-network/cosmos-proto.git; \
  cd cosmos-proto/protoc-gen-gocosmos; \
  go install .; \
  cd ..

RUN git clone https://github.com/regen-network/protobuf.git; \
  cd protobuf; \
  go mod download; \
  make install; \
  cd ..

RUN npm install -g swagger-combine

COPY --from=BUILDER /usr/local/bin /usr/local/bin
