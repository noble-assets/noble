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
coins="1000$DENOM,0$BASEDENOM"
coins0="1$DENOM"
delegate="1$DENOM"


nobled --home $CHAINDIR/$CHAINID --chain-id $CHAINID init $CHAINID 
sleep 1
nobled --home $CHAINDIR/$CHAINID keys add validator1 $KEYRING --output json > $CHAINDIR/$CHAINID/validator_seed1.json 2>&1
sleep 1
nobled --home $CHAINDIR/$CHAINID keys add validator2 $KEYRING --output json > $CHAINDIR/$CHAINID/validator_seed2.json 2>&1
sleep 1
nobled --home $CHAINDIR/$CHAINID keys add validator3 $KEYRING --output json > $CHAINDIR/$CHAINID/validator_seed3.json 2>&1
sleep 1
nobled --home $CHAINDIR/$CHAINID keys add validator4 $KEYRING --output json > $CHAINDIR/$CHAINID/validator_seed4.json 2>&1
sleep 1
nobled --home $CHAINDIR/$CHAINID keys add validator5 $KEYRING --output json > $CHAINDIR/$CHAINID/validator_seed5.json 2>&1
sleep 1
nobled --home $CHAINDIR/$CHAINID keys add validator6 $KEYRING --output json > $CHAINDIR/$CHAINID/validator_seed6.json 2>&1
sleep 1
nobled --home $CHAINDIR/$CHAINID keys add owner $KEYRING --output json > $CHAINDIR/$CHAINID/owner_seed.json 2>&1
sleep 1
nobled --home $CHAINDIR/$CHAINID add-genesis-account $(nobled --home $CHAINDIR/$CHAINID keys $KEYRING show owner -a) $coins 
sleep 1
nobled --home $CHAINDIR/$CHAINID add-genesis-account $(nobled --home $CHAINDIR/$CHAINID keys $KEYRING show validator1 -a) $coins0
sleep 1
nobled --home $CHAINDIR/$CHAINID add-genesis-account $(nobled --home $CHAINDIR/$CHAINID keys $KEYRING show validator2 -a) $coins0
sleep 1
nobled --home $CHAINDIR/$CHAINID add-genesis-account $(nobled --home $CHAINDIR/$CHAINID keys $KEYRING show validator3 -a) $coins0
sleep 1
nobled --home $CHAINDIR/$CHAINID add-genesis-account $(nobled --home $CHAINDIR/$CHAINID keys $KEYRING show validator4 -a) $coins0
sleep 1
nobled --home $CHAINDIR/$CHAINID add-genesis-account $(nobled --home $CHAINDIR/$CHAINID keys $KEYRING show validator5 -a) $coins0
sleep 1
nobled --home $CHAINDIR/$CHAINID add-genesis-account $(nobled --home $CHAINDIR/$CHAINID keys $KEYRING show validator6 -a) $coins0
sleep 1

mkdir $CHAINDIR/$CHAINID/gentx

nobled --home $CHAINDIR/$CHAINID gentx validator1 $coins0 $KEYRING --chain-id $CHAINID --min-self-delegation 1 --output-document $CHAINDIR/$CHAINID/gentx/gentx1.json
sleep 1
nobled --home $CHAINDIR/$CHAINID gentx validator2 $coins0 $KEYRING --chain-id $CHAINID --min-self-delegation 1 --output-document $CHAINDIR/$CHAINID/gentx/gentx2.json
sleep 1
nobled --home $CHAINDIR/$CHAINID gentx validator3 $coins0 $KEYRING --chain-id $CHAINID --min-self-delegation 1 --output-document $CHAINDIR/$CHAINID/gentx/gentx3.json
sleep 1
nobled --home $CHAINDIR/$CHAINID gentx validator4 $coins0 $KEYRING --chain-id $CHAINID --min-self-delegation 1 --output-document $CHAINDIR/$CHAINID/gentx/gentx4.json
sleep 1
nobled --home $CHAINDIR/$CHAINID gentx validator5 $coins0 $KEYRING --chain-id $CHAINID --min-self-delegation 1 --output-document $CHAINDIR/$CHAINID/gentx/gentx5.json
sleep 1
nobled --home $CHAINDIR/$CHAINID gentx validator6 $coins0 $KEYRING --chain-id $CHAINID --min-self-delegation 1 --output-document $CHAINDIR/$CHAINID/gentx/gentx6.json
sleep 1

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
  sed -i '' 's/owner": null/owner": { "address": '"$OWNER"' }/g' $CHAINDIR/$CHAINID/config/genesis.json
  sed -i '' 's/mintingDenom": null/mintingDenom": { "denom": "'$MINTING_BASEDENOM'" }/g' $CHAINDIR/$CHAINID/config/genesis.json
  sed -i '' 's/paused": null/paused": { "paused": false }/g' $CHAINDIR/$CHAINID/config/genesis.json
  sed -i '' 's/"denom_metadata": \[]/"denom_metadata": [ { "display": "rupee", "base": "urupee", "name": "rupee", "symbol": "rupee", "denom_units": [ { "denom": "urupee", "aliases": [ "microrupee" ], "exponent": "0" }, { "denom": "mrupee", "aliases": [ "milirupee" ], "exponent": "3" }, { "denom": "rupee", "aliases": null, "exponent": "6" } ] } ]/g' $CHAINDIR/$CHAINID/config/genesis.json
  sed -i '' 's/"authority": ""/"authority": '"$OWNER"'/g' $CHAINDIR/$CHAINID/config/genesis.json

fi

# Delete old keys if they exhist
KEYS=("owner" "masterminter" "mintercontroller" "minter" "blacklister" "pauser" "user")
for KEY in ${KEYS[@]}
do 
    if nobled keys show $KEY > /dev/null 2>&1; then
        nobled keys delete $KEY -y
    else
        continue
    fi
done 

nobled --home $CHAINDIR/$CHAINID tx staking delegate $(nobled --home $CHAINDIR/$CHAINID keys $KEYRING show validator1 -a) $delegate --from owner
sleep 1
nobled --home $CHAINDIR/$CHAINID tx staking delegate $(nobled --home $CHAINDIR/$CHAINID keys $KEYRING show validator2 -a) $delegate --from owner
sleep 1
nobled --home $CHAINDIR/$CHAINID tx staking delegate $(nobled --home $CHAINDIR/$CHAINID keys $KEYRING show validator3 -a) $delegate --from owner
sleep 1
nobled --home $CHAINDIR/$CHAINID tx staking delegate $(nobled --home $CHAINDIR/$CHAINID keys $KEYRING show validator4 -a) $delegate --from owner
sleep 1
nobled --home $CHAINDIR/$CHAINID tx staking delegate $(nobled --home $CHAINDIR/$CHAINID keys $KEYRING show validator5 -a) $delegate --from owner
sleep 1
nobled --home $CHAINDIR/$CHAINID tx staking delegate $(nobled --home $CHAINDIR/$CHAINID keys $KEYRING show validator6 -a) $delegate --from owner