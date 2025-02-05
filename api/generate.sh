rm -rf api/proto
rm -rf api/tmp-swagger-gen

# IMPORTANT: These versions should match the go.mod!
buf export buf.build/noble-assets/wormhole:v1.0.0-alpha.2 --output api/proto
buf export buf.build/noble-assets/dollar:v1.0.0-alpha.1 --output api/proto
buf export buf.build/noble-assets/swap:v1.0.0-alpha.3 --output api/proto

buf generate
swagger-combine ./api/config.json -o ./api/gen/swagger.yaml -f yaml
