#!/bin/bash
rm -rf ./api/proto
rm -rf ./api/tmp-swagger-gen

# IMPORTANT: These versions should match the go.mod!
buf export buf.build/bcp-innovations/hyperlane-cosmos:v1.0.1 --output api/proto
buf export buf.build/cosmos/cosmos-sdk:v0.50.0 --output api/proto
buf export buf.build/cosmos/ibc:e69c8f372b127401762cb251bb7b60371b4cef29 --output api/proto
buf export buf.build/noble-assets/aura:v2.0.0 --output api/proto
buf export buf.build/noble-assets/authority:v1.0.3 --output api/proto
buf export buf.build/noble-assets/cctp:4285c94ec19438ad1e05ba3e5106a5e7980cfffd --output api/proto
buf export buf.build/noble-assets/dollar:v2.2.0 --output api/proto
buf export buf.build/noble-assets/fiattokenfactory:5f9bd9dd2c5b5336b94bae4a47195bdf035f04af --output api/proto
buf export buf.build/noble-assets/florin:v2.0.0 --output api/proto
buf export buf.build/noble-assets/forwarding:v2.0.3 --output api/proto
buf export buf.build/noble-assets/globalfee:v1.0.1 --output api/proto
buf export buf.build/noble-assets/halo:v2.0.1 --output api/proto
buf export buf.build/noble-assets/orbiter:v2.0.0 --output api/proto
buf export buf.build/noble-assets/rate-limiting:v8.0.0 --output api/proto
buf export buf.build/noble-assets/swap:v1.0.2 --output api/proto
buf export buf.build/noble-assets/wormhole:v1.0.0 --output api/proto

buf generate
swagger-combine ./api/config.json -o ./api/gen/swagger.yaml -f yaml --includeDefinitions true
