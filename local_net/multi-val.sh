#!/bin/bash

source ./utils.sh

alias nobled=../build/nobled

for arg in "$@"
do
    case $arg in
        -r|--reset)
        rm -rf .duke
        shift
        ;;
    esac
done

HOME1=.duke/val1
HOME2=.duke/val2
HOME3=.duke/val3

P2P1=0.0.0.0:26656
P2P2=0.0.0.0:36656
P2P3=0.0.0.0:46656

 # if private validator file does not exist, create a new network
if ! [ -f .duke/data/priv_validator_state.json ]; then
  nobled init val1 --chain-id "duke-1" --home $HOME1 &> /dev/null
  nobled init val2 --chain-id "duke-1" --home $HOME2 &> /dev/null
  nobled init val3 --chain-id "duke-1" --home $HOME3 &> /dev/null

  # Create keys
  nobled keys add val --keyring-backend test --home $HOME1 &> /dev/null
  nobled keys add val --keyring-backend test --home $HOME2 &> /dev/null
  nobled keys add val --keyring-backend test --home $HOME3 &> /dev/null

  # Add genesis accounts from each validator
  nobled genesis add-genesis-account val 1000000ustake --home $HOME1 --keyring-backend test
  nobled genesis add-genesis-account val 1000000ustake --home $HOME2 --keyring-backend test
  nobled genesis add-genesis-account val 1000000ustake --home $HOME3 --keyring-backend test
  # Add genesis accounts to validator 1 who will be collecting genesis
  nobled genesis add-genesis-account "$(nobled keys show val -a --keyring-backend test --home $HOME2)" 1000000ustake --home $HOME1 
  nobled genesis add-genesis-account "$(nobled keys show val -a --keyring-backend test --home $HOME3)" 1000000ustake --home $HOME1 

  # Create genesis transaction's
  nobled genesis gentx val 1000000ustake --chain-id "duke-1" --keyring-backend test --home $HOME1 
  nobled genesis gentx val 1000000ustake --chain-id "duke-1" --output-document $HOME1/config/gentx/val2.json --keyring-backend test --home $HOME2
  nobled genesis gentx val 1000000ustake --chain-id "duke-1" --output-document $HOME1/config/gentx/val3.json --keyring-backend test --home $HOME3 

  # Collect the gentx and finalize genesis
  nobled genesis collect-gentxs --home $HOME1 &> /dev/null

  AUTHORITY=$(nobled keys add authority --home $HOME1 --keyring-backend test --output json | jq .address)
  nobled genesis add-genesis-account authority 4000000ustake --home $HOME1 --keyring-backend test

  update_genesis $HOME1 $AUTHORITY

  # Copy genesis to val 2 and 3
  cp "$HOME1/config/genesis.json" "$HOME2/config/genesis.json"
  cp "$HOME1/config/genesis.json" "$HOME3/config/genesis.json"

  # Configure config.toml setting not available in a flag
  sed -i '' 's|addr_book_strict = true|addr_book_strict = false|' $HOME1/config/config.toml
  sed -i '' 's|addr_book_strict = true|addr_book_strict = false|' $HOME2/config/config.toml
  sed -i '' 's|addr_book_strict = true|addr_book_strict = false|' $HOME3/config/config.toml

  sed -i '' 's|allow_duplicate_ip = false|allow_duplicate_ip = true|' $HOME1/config/config.toml
  sed -i '' 's|allow_duplicate_ip = false|allow_duplicate_ip = true|' $HOME2/config/config.toml
  sed -i '' 's|allow_duplicate_ip = false|allow_duplicate_ip = true|' $HOME3/config/config.toml
fi

# Get persistent peers
NODE_ID1=$(nobled tendermint show-node-id --home "$HOME1")
PP1="$NODE_ID1@$P2P1"

NODE_ID2=$(nobled tendermint show-node-id --home "$HOME2")
PP2="$NODE_ID2@$P2P2"

NODE_ID3=$(nobled tendermint show-node-id --home "$HOME3")
PP3="$NODE_ID3@$P2P3"

# Start tmux session
SESSION="3v-network"
tmux new-session -d -s "$SESSION"

tmux split-window -h -t "$SESSION"
tmux split-window -h -t "$SESSION"

# Send start command
# Note: C-m is equivalent to pressing Enter
tmux send-keys -t "$SESSION:0.0" "../build/nobled start --api.enable false --home $HOME1 > $HOME1/logs.log 2>&1 &" C-m
tmux send-keys -t "$SESSION:0.1" "../build/nobled start --api.enable false --rpc.laddr tcp://127.0.0.1:36657 --rpc.pprof_laddr localhost:6061 --grpc.address localhost:9092 --p2p.laddr tcp://$P2P2 --p2p.persistent_peers $PP1,$PP3  --home $HOME2 > $HOME2/logs.log 2>&1 &" C-m
tmux send-keys -t "$SESSION:0.2" "../build/nobled start --api.enable false --rpc.laddr tcp://127.0.0.1:46657 --rpc.pprof_laddr localhost:6062 --grpc.address localhost:9093 --p2p.laddr tcp://$P2P3 --p2p.persistent_peers $PP1,$PP2  --home $HOME3 > $HOME3/logs.log 2>&1 &" C-m

# Watch logs
tmux send-keys -t "$SESSION:0.0" "tail -f $HOME1/logs.log" C-m
tmux send-keys -t "$SESSION:0.1" "tail -f $HOME2/logs.log" C-m
tmux send-keys -t "$SESSION:0.2" "tail -f $HOME3/logs.log" C-m

# bring up session
tmux attach-session -t "$SESSION"
