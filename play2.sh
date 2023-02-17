#!/bin/bash

# !! PLEASE ONLY RUN THIS SCRIPT IN A TESTING ENVIRONMENT !!
#   THIS SCRIPT: 
#     - Stops/starts noble binary
#     - Adds/deletes noble keys
#     

# This script is meant for quick experimentation of the rupeefactory Module and Noble Chain functionality. 
#   - Starts Noble chain
#   - Delagates and funds all privledged accounts relating to the rupeefactory Module
#   - The "owner" account is both the rupeefactory Owner AND the Param Authority

killall nobled
[ -d "play_sh" ] && rm -r "play_sh"


BINARY="nobled"
CHAINID="noble-1"
CHAINDIR="play_sh"
RPCPORT="26657"
P2PPORT="26656"
PROFPORT="6060"
GRPCPORT="9090"
DENOM="stake"
BASEDENOM="ustake"

MINTING_BASEDENOM="urupee"

KEYRING="--keyring-backend test"
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
# ownerGenesis="1000000$DENOM,1000000$BASEDENOM"
# valGenesis="0$DENOM,1$BASEDENOM"
# valGenTx="1$BASEDENOM"
# delegate="1$BASEDENOM"

ownerGenesis="1000000$DENOM,1000000$BASEDENOM"
valGenesis="1$DENOM,1$BASEDENOM"
valGenTx="1$DENOM"
delegate="1$DENOM"


# ownerGenesis="100000000000$DENOM,100000000000$BASEDENOM"
# valGenesis="100000000000$DENOM,100000000000$BASEDENOM"
# valGenTx="100000000000$DENOM"
# delegate="100000000000$DENOM"


nobled --home $CHAINDIR/$CHAINID --chain-id $CHAINID init $CHAINID 
sleep 1



nobled --home $CHAINDIR/$CHAINID $KEYRING keys add validator1 --output json > $CHAINDIR/$CHAINID/validator_seed1.json 2>&1
sleep 1
# nobled --home $CHAINDIR/$CHAINID $KEYRING keys add validator2 --output json > $CHAINDIR/$CHAINID/validator_seed2.json 2>&1
# sleep 1
# nobled --home $CHAINDIR/$CHAINID $KEYRING keys add validator3 --output json > $CHAINDIR/$CHAINID/validator_seed3.json 2>&1
# sleep 1
# nobled --home $CHAINDIR/$CHAINID $KEYRING keys add validator4 --output json > $CHAINDIR/$CHAINID/validator_seed4.json 2>&1
# sleep 1
# nobled --home $CHAINDIR/$CHAINID $KEYRING keys add validator5 --output json > $CHAINDIR/$CHAINID/validator_seed5.json 2>&1
# sleep 1
# nobled --home $CHAINDIR/$CHAINID $KEYRING keys add validator6 --output json > $CHAINDIR/$CHAINID/validator_seed6.json 2>&1
sleep 1
nobled --home $CHAINDIR/$CHAINID $KEYRING keys add owner --output json > $CHAINDIR/$CHAINID/owner_seed.json 2>&1
sleep 1
nobled --home $CHAINDIR/$CHAINID $KEYRING add-genesis-account $(nobled --home $CHAINDIR/$CHAINID $KEYRING keys show owner -a) $ownerGenesis 
sleep 1
nobled --home $CHAINDIR/$CHAINID $KEYRING add-genesis-account $(nobled --home $CHAINDIR/$CHAINID $KEYRING keys show validator1 -a) $valGenesis
sleep 1
# nobled --home $CHAINDIR/$CHAINID $KEYRING add-genesis-account $(nobled --home $CHAINDIR/$CHAINID $KEYRING keys show validator2 -a) $valGenesis
# sleep 1
# nobled --home $CHAINDIR/$CHAINID $KEYRING add-genesis-account $(nobled --home $CHAINDIR/$CHAINID $KEYRING keys show validator3 -a) $valGenesis
# sleep 1
# nobled --home $CHAINDIR/$CHAINID $KEYRING add-genesis-account $(nobled --home $CHAINDIR/$CHAINID $KEYRING keys show validator4 -a) $valGenesis
# sleep 1
# nobled --home $CHAINDIR/$CHAINID $KEYRING add-genesis-account $(nobled --home $CHAINDIR/$CHAINID $KEYRING keys show validator5 -a) $valGenesis
# sleep 1
# nobled --home $CHAINDIR/$CHAINID $KEYRING add-genesis-account $(nobled --home $CHAINDIR/$CHAINID $KEYRING keys show validator6 -a) $valGenesis
# sleep 1

mkdir $CHAINDIR/$CHAINID/config/gentx

nobled --home $CHAINDIR/$CHAINID $KEYRING gentx validator1 $valGenTx --chain-id $CHAINID --min-self-delegation 1 --output-document $CHAINDIR/$CHAINID/config/gentx/gentx1.json
sleep 1
# nobled --home $CHAINDIR/$CHAINID $KEYRING gentx validator2 $valGenTx --chain-id $CHAINID --min-self-delegation 1 --output-document $CHAINDIR/$CHAINID/config/gentx/gentx2.json
# sleep 1
# nobled --home $CHAINDIR/$CHAINID $KEYRING gentx validator3 $valGenTx --chain-id $CHAINID --min-self-delegation 1 --output-document $CHAINDIR/$CHAINID/config/gentx/gentx3.json
# sleep 1
# nobled --home $CHAINDIR/$CHAINID $KEYRING gentx validator4 $valGenTx --chain-id $CHAINID --min-self-delegation 1 --output-document $CHAINDIR/$CHAINID/config/gentx/gentx4.json
# sleep 1
# nobled --home $CHAINDIR/$CHAINID $KEYRING gentx validator5 $valGenTx --chain-id $CHAINID --min-self-delegation 1 --output-document $CHAINDIR/$CHAINID/config/gentx/gentx5.json
# sleep 1
# nobled --home $CHAINDIR/$CHAINID $KEYRING gentx validator6 $valGenTx --chain-id $CHAINID --min-self-delegation 1 --output-document $CHAINDIR/$CHAINID/config/gentx/gentx6.json
# sleep 1

nobled --home $CHAINDIR/$CHAINID collect-gentxs
sleep 1

OWNER=$(jq '.address' $CHAINDIR/$CHAINID/owner_seed.json)

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
  sed -i 's/"bond_denom": "stake"/"bond_denom": "'"$DENOM"'"/g' $CHAINDIR/$CHAINID/config/genesis.json
  sed -i 's/owner": null/owner": { "address": '"$OWNER"' }/g' $CHAINDIR/$CHAINID/config/genesis.json
  sed -i 's/mintingDenom": null/mintingDenom": { "denom": "'$MINTING_BASEDENOM'" }/g' $CHAINDIR/$CHAINID/config/genesis.json
  sed -i 's/paused": null/paused": { "paused": false }/g' $CHAINDIR/$CHAINID/config/genesis.json
  sed -i 's/"denom_metadata": \[]/"denom_metadata": [ { "display": "rupee", "base": "urupee", "name": "rupee", "symbol": "rupee", "denom_units": [ { "denom": "urupee", "aliases": [ "microrupee" ], "exponent": "0" }, { "denom": "mrupee", "aliases": [ "milirupee" ], "exponent": "3" }, { "denom": "rupee", "aliases": null, "exponent": "6" } ] } ]/g' $CHAINDIR/$CHAINID/config/genesis.json
  sed -i 's/"authority": ""/"authority": '"$OWNER"'/g' $CHAINDIR/$CHAINID/config/genesis.json

else
  sed -i '' 's#"tcp://127.0.0.1:26657"#"tcp://0.0.0.0:'"$RPCPORT"'"#g' $CHAINDIR/$CHAINID/config/config.toml
  sed -i '' 's#"tcp://0.0.0.0:26656"#"tcp://0.0.0.0:'"$P2PPORT"'"#g' $CHAINDIR/$CHAINID/config/config.toml
  sed -i '' 's#"localhost:6060"#"localhost:'"$P2PPORT"'"#g' $CHAINDIR/$CHAINID/config/config.toml
  sed -i '' 's/timeout_commit = "5s"/timeout_commit = "1s"/g' $CHAINDIR/$CHAINID/config/config.toml
  sed -i '' 's/timeout_propose = "3s"/timeout_propose = "1s"/g' $CHAINDIR/$CHAINID/config/config.toml
  sed -i '' 's/index_all_keys = false/index_all_keys = true/g' $CHAINDIR/$CHAINID/config/config.toml
  sed -i '' 's/"bond_denom": "stake"/"bond_denom": "'"$DENOM"'"/g' $CHAINDIR/$CHAINID/config/genesis.json
  sed -i '' 's/owner": null/owner": { "address": '"$OWNER"' }/g' $CHAINDIR/$CHAINID/config/genesis.json
  sed -i '' 's/mintingDenom": null/mintingDenom": { "denom": "'$MINTING_BASEDENOM'" }/g' $CHAINDIR/$CHAINID/config/genesis.json
  sed -i '' 's/paused": null/paused": { "paused": false }/g' $CHAINDIR/$CHAINID/config/genesis.json
  sed -i '' 's/"denom_metadata": \[]/"denom_metadata": [ { "display": "rupee", "base": "urupee", "name": "rupee", "symbol": "rupee", "denom_units": [ { "denom": "urupee", "aliases": [ "microrupee" ], "exponent": "0" }, { "denom": "mrupee", "aliases": [ "milirupee" ], "exponent": "3" }, { "denom": "rupee", "aliases": null, "exponent": "6" } ] } ]/g' $CHAINDIR/$CHAINID/config/genesis.json
  sed -i '' 's/"authority": ""/"authority": '"$OWNER"'/g' $CHAINDIR/$CHAINID/config/genesis.json

fi

echo "Starting node!!!"

nobled --home $CHAINDIR/$CHAINID start --pruning=nothing --grpc-web.enable=false --grpc.address="0.0.0.0:$GRPCPORT" > $CHAINDIR/$CHAINID.log 2>&1 &
sleep 2

echo "Started!!!"

# nobled --home $CHAINDIR/$CHAINID tx staking delegate $KEYRING $(jq -r '.body.messages[].validator_address' $CHAINDIR/$CHAINID/config/gentx/gentx1.json) $delegate --from owner
# sleep 1
# nobled --home $CHAINDIR/$CHAINID tx staking delegate $KEYRING $(jq -r '.body.messages[].validator_address' $CHAINDIR/$CHAINID/config/gentx/gentx2.json) $delegate --from owner
# sleep 1
# nobled --home $CHAINDIR/$CHAINID tx staking delegate $KEYRING $(jq -r '.body.messages[].validator_address' $CHAINDIR/$CHAINID/config/gentx/gentx3.json) $delegate --from owner
# sleep 1
# nobled --home $CHAINDIR/$CHAINID tx staking delegate $KEYRING $(jq -r '.body.messages[].validator_address' $CHAINDIR/$CHAINID/config/gentx/gentx4.json) $delegate --from owner
# sleep 1
# nobled --home $CHAINDIR/$CHAINID tx staking delegate $KEYRING $(jq -r '.body.messages[].validator_address' $CHAINDIR/$CHAINID/config/gentx/gentx5.json) $delegate --from owner
# sleep 1
# nobled --home $CHAINDIR/$CHAINID tx staking delegate $KEYRING $(jq -r '.body.messages[].validator_address' $CHAINDIR/$CHAINID/config/gentx/gentx6.json) $delegate --from owner