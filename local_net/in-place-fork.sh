#!/bin/bash

BIN=../build/nobled

cleanup() {
  echo "Stopping processes..."
  kill "$NOBLED_PID" "$TAIL_PID" "$NOBLED_PID2" 2>/dev/null
  exit 0
}
trap cleanup SIGINT SIGTERM

HOME1=.duke

CHAINID="noble-1"
PEERS="4f9df51e0800e79e0d45fd376c11236b99be4e12@99.79.58.157:26656,3402e50ad4d838b26f8341a956c7b4b8a3c61ee5@65.109.93.44:21556"
SNAP_RPC="https://noble-rpc.polkachu.com:443"

UPGRADE_ARG=""

for arg in "$@"; do
  case $arg in
  -r | --reset)
    rm -rf "$HOME1"
    ;;
  -t | --testnet)
    CHAINID="grand-1"
    PEERS="f2067cc7a23a4b2525f5f98430797b1e5c92e3aa@35.183.110.236:26656,8b22414f37d381a99ba99cd1edc5b884d43b7e53@65.109.23.114:21556"
    SNAP_RPC="https://noble-testnet-rpc.polkachu.com:443"
    ;;
  -u | --trigger-testnet-upgrade)
    UPGRADE_ARG="$2"
    shift
    ;;
  esac
  shift
done

$BIN init in-place --chain-id $CHAINID --home $HOME1

LATEST_HEIGHT=$(curl -s $SNAP_RPC/block | jq -r .result.block.header.height)
BLOCK_HEIGHT=$((LATEST_HEIGHT - 2000))
TRUST_HASH=$(curl -s "$SNAP_RPC/block?height=$BLOCK_HEIGHT" | jq -r .result.block_id.hash)

sed -i.bak -E "
s|^(enable[[:space:]]+=[[:space:]]+).*$|\1true| ;
s|^(rpc_servers[[:space:]]+=[[:space:]]+).*$|\1\"$SNAP_RPC,$SNAP_RPC\"| ;
s|^(trust_height[[:space:]]+=[[:space:]]+).*$|\1$BLOCK_HEIGHT| ;
s|^(trust_hash[[:space:]]+=[[:space:]]+).*$|\1\"$TRUST_HASH\"| ;
s|^persistent_peers *=.*|persistent_peers = \"$PEERS\"|" $HOME1/config/config.toml

$BIN start --halt-height $LATEST_HEIGHT --home "$HOME1" >"$HOME1/logs.log" 2>&1 &
NOBLED_PID=$!

tail -f "$HOME1/logs.log" &
TAIL_PID=$!

# Wait for node to halt because of `--halt-height` flag
wait $NOBLED_PID
echo "Node is synced! Preparing for in-place testnet..."
sleep 2

# Create operator address that will control the chain
OPERATOR=$($BIN keys add operator --home $HOME1 --keyring-backend test --output json | jq -r .address)

if [[ -n "$UPGRADE_ARG" ]]; then
  printf 'y\n' | $BIN in-place-testnet inPlace "$OPERATOR" --trigger-testnet-upgrade "$UPGRADE_ARG" --home "$HOME1" >>"$HOME1/logs.log" 2>&1 &
else
  printf 'y\n' | $BIN in-place-testnet inPlace "$OPERATOR" --home "$HOME1" >>"$HOME1/logs.log" 2>&1 &
fi

NOBLED_PID2=$!

cat <<'EOF'
######################################################################
#                                                                    #
#                     STARTING IN-PLACE TESTNET                      #
#                                                                    #
######################################################################
EOF
echo 👑 Operator: $OPERATOR

# Keep tailing logs in the foreground to prevent script from exiting
wait "$NOBLED_PID2"
