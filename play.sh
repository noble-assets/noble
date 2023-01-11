#!/bin/bash

# This script is meant for quick experimentation with the Tokenfactory. After running this script, 
# you can recover the seed phrase of the "owner" account from the keys_seed.json file from the "play_sh" 
# folder. This will allow  you to create the other privledged accounts.

# PLEASE ONLY RUN THIS SCRIPT IN A TESTING ENVIRONMENT 



BINARY="nobled"
CHAINID="noble-1"
CHAINDIR="play_sh"
RPCPORT="26657"
P2PPORT="26656"
PROFPORT="6060"
GRPCPORT="9090"
DENOM="token"
BASEDENOM="utokens"

KEYRING=--keyring-backend="test"
SILENT=1

redirect() {
  if [ "$SILENT" -eq 1 ]; then
    "$@" > /dev/null 2>&1
  else
    "$@"
  fi
}

# Add dir for chain, exit if error
if ! mkdir -p $CHAINDIR/$CHAINID 2>/dev/null; then
    echo "Failed to create chain folder. Aborting..."
    exit 1
fi

# Build genesis file incl account for passed address
coins="100000000000$DENOM,100000000000$BASEDENOM"
delegate="100000000000$DENOM"

redirect $BINARY --home $CHAINDIR/$CHAINID --chain-id $CHAINID init $CHAINID 
sleep 1
$BINARY --home $CHAINDIR/$CHAINID keys add validator $KEYRING --output json > $CHAINDIR/$CHAINID/validator_seed.json 2>&1
sleep 1
$BINARY --home $CHAINDIR/$CHAINID keys add owner $KEYRING --output json > $CHAINDIR/$CHAINID/key_seed.json 2>&1
sleep 1
redirect $BINARY --home $CHAINDIR/$CHAINID add-genesis-account $($BINARY --home $CHAINDIR/$CHAINID keys $KEYRING show owner -a) $coins 
sleep 1
redirect $BINARY --home $CHAINDIR/$CHAINID add-genesis-account $($BINARY --home $CHAINDIR/$CHAINID keys $KEYRING show validator -a) $coins 
sleep 1
redirect $BINARY --home $CHAINDIR/$CHAINID gentx validator $delegate $KEYRING --chain-id $CHAINID
sleep 1
redirect $BINARY --home $CHAINDIR/$CHAINID collect-gentxs 
sleep 1

OWNER=$(jq '.address' $CHAINDIR/$CHAINID/key_seed.json)

# Check platform
platform='unknown'
unamestr=`uname`
if [ "$unamestr" = 'Linux' ]; then
   platform='linux'
fi

# Set proper defaults and change ports (use a different sed for Mac or Linux)
echo "Change settings in config.toml and genesis.json files..."
if [ $platform = 'linux' ]; then
  sed -i 's#"tcp://127.0.0.1:26657"#"tcp://0.0.0.0:'"$RPCPORT"'"#g' $CHAINDIR/$CHAINID/config/config.toml
  sed -i 's#"tcp://0.0.0.0:26656"#"tcp://0.0.0.0:'"$P2PPORT"'"#g' $CHAINDIR/$CHAINID/config/config.toml
  sed -i 's#"localhost:6060"#"localhost:'"$P2PPORT"'"#g' $CHAINDIR/$CHAINID/config/config.toml
  sed -i 's/timeout_commit = "5s"/timeout_commit = "1s"/g' $CHAINDIR/$CHAINID/config/config.toml
  sed -i 's/timeout_propose = "3s"/timeout_propose = "1s"/g' $CHAINDIR/$CHAINID/config/config.toml
  sed -i 's/index_all_keys = false/index_all_keys = true/g' $CHAINDIR/$CHAINID/config/config.toml
  sed -i 's/owner": null/owner": { "address": '"$OWNER"' }/g' $CHAINDIR/$CHAINID/config/genesis.json

  # sed -i '' 's#index-events = \[\]#index-events = \["message.action","send_packet.packet_src_channel","send_packet.packet_sequence"\]#g' $CHAINDIR/$CHAINID/config/app.toml
else
  sed -i '' 's#"tcp://127.0.0.1:26657"#"tcp://0.0.0.0:'"$RPCPORT"'"#g' $CHAINDIR/$CHAINID/config/config.toml
  sed -i '' 's#"tcp://0.0.0.0:26656"#"tcp://0.0.0.0:'"$P2PPORT"'"#g' $CHAINDIR/$CHAINID/config/config.toml
  sed -i '' 's#"localhost:6060"#"localhost:'"$P2PPORT"'"#g' $CHAINDIR/$CHAINID/config/config.toml
  sed -i '' 's/timeout_commit = "5s"/timeout_commit = "1s"/g' $CHAINDIR/$CHAINID/config/config.toml
  sed -i '' 's/timeout_propose = "3s"/timeout_propose = "1s"/g' $CHAINDIR/$CHAINID/config/config.toml
  sed -i '' 's/index_all_keys = false/index_all_keys = true/g' $CHAINDIR/$CHAINID/config/config.toml
  sed -i '' 's/owner": null/owner": { "address": '"$OWNER"' }/g' $CHAINDIR/$CHAINID/config/genesis.json
  # sed -i '' 's#index-events = \[\]#index-events = \["message.action","send_packet.packet_src_channel","send_packet.packet_sequence"\]#g' $CHAINDIR/$CHAINID/config/app.toml
fi

redirect $BINARY --home $CHAINDIR/$CHAINID add-consumer-section


# Start
$BINARY --home $CHAINDIR/$CHAINID start --pruning=nothing --grpc-web.enable=false --grpc.address="0.0.0.0:$GRPCPORT" > $CHAINDIR/$CHAINID.log 2>&1 &


# sleep 2
# nobled tx tokenfactory update-master-minter $(nobled keys show masterminter -a) --from owner -y
# sleep 2
# nobled tx tokenfactory configure-minter-controller $(nobled keys show mintercontroller -a) $(nobled keys show minter -a) --from masterminter -y
# sleep 2
# nobled tx tokenfactory configure-minter $(nobled keys show minter -a) 1000uusdc --from mintercontroller -y
# sleep 2
# nobled tx tokenfactory mint $(nobled keys show user -a) 100uusdc --from minter -y
# sleep 2
# nobled q bank balances $(nobled keys show user -a)
# sleep 2
# nobled tx tokenfactory update-blacklister $(nobled keys show blacklister -a) --from owner -y
# sleep 2
# nobled tx tokenfactory blacklist $(nobled keys show user -a) --from blacklister -y
# sleep 2
# nobled tx tokenfactory mint $(nobled keys show user -a) 100uusdc --from minter -y
# sleep 2
# nobled tx tokenfactory unblacklist $(nobled keys show user -a) --from blacklister -y
# sleep 2
# nobled tx tokenfactory mint $(nobled keys show user -a) 100uusdc --from minter -y
# sleep 2
# nobled tx tokenfactory update-pauser $(nobled keys show pauser -a) --from owner -y
# sleep 2
# nobled tx tokenfactory pause --from pauser -y
# sleep 2
# nobled tx tokenfactory mint $(nobled keys show user -a) 100uusdc --from minter -y
# sleep 2
# nobled tx bank send $(nobled keys show user -a) $(nobled keys show alice -a) 100uusdc --from user -y
# sleep 2
# nobled tx tokenfactory unpause --from pauser -y
# sleep 2
# nobled tx bank send $(nobled keys show user -a) $(nobled keys show alice -a) 100uusdc --from user -y

# nobled tx tokenfactory mint $(nobled keys show user -a) 100uusdc --from minter -y
# sleep 2
# nobled q bank balances $(nobled keys show user -a)
# sleep 2
# nobled tx tokenfactory mint $(nobled keys show minter -a) 100uusdc --from minter -y
# sleep 2
# nobled tx tokenfactory burn 100uusdc --from minter -y
# sleep 2
# nobled q bank balances $(nobled keys show user -a)
