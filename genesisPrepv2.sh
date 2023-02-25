#!/bin/bash

# !! PLEASE ONLY RUN THIS SCRIPT IN A TESTING ENVIRONMENT !!
#   THIS SCRIPT: 
#     - Stops/starts noble binary
#     - Adds/deletes noble keys
#     

# This script is meant for quick experimentation of the usdcfactory Module and Noble Chain functionality. 
#   - Starts Noble chain
#   - Delagates and funds all privledged accounts relating to the usdcfactory Module
#   - The "owner" account is both the usdcfactory Owner AND the Param Authority

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

HOME=$CHAINDIR/$CHAINID

MINTING_BASEDENOM="uusdc"

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
if ! mkdir -p $HOME 2>/dev/null; then
    echo "Failed to create chain folder. Aborting..."
    exit 1
fi

AUTHGEN="0$DENOM,1000000000000$BASEDENOM"
valGenesis="0$DENOM,1000000$BASEDENOM"
valGenTx="1000000$BASEDENOM"
ZERO="0$DENOM,0$BASEDENOM"
SWAGGEN="0noble,1000unoble"


nobled --home $HOME --chain-id $CHAINID init $CHAINID 
sleep 1


nobled --home $HOME $KEYRING keys add paramAuth --output json > $HOME/paramAuth_seed.json 2>&1
sleep 1
nobled --home $HOME $KEYRING keys add swag --output json > $HOME/swag_seed.json 2>&1
sleep 1
nobled --home $HOME $KEYRING keys add owner --output json > $HOME/owner_seed.json 2>&1
sleep 1
nobled --home $HOME $KEYRING keys add masterMinter --output json > $HOME/masterMinter_seed.json 2>&1
sleep 1
nobled --home $HOME $KEYRING keys add pauser --output json > $HOME/pauser_seed.json 2>&1
sleep 1
nobled --home $HOME $KEYRING keys add blacklister --output json > $HOME/blacklister_seed.json 2>&1
sleep 1
nobled --home $HOME $KEYRING keys add validator1 --output json > $HOME/validator_seed1.json 2>&1
sleep 1
nobled --home $HOME $KEYRING keys add mintercontroller --output json > $HOME/mintercontroller_seed1.json 2>&1
sleep 1
nobled --home $HOME $KEYRING keys add minter --output json > $HOME/minter_seed1.json 2>&1
sleep 1

nobled --home $HOME $KEYRING add-genesis-account $(nobled --home $HOME $KEYRING keys show validator1 -a) $valGenesis --vesting-amount $valGenesis  --vesting-end-time 253380229199
sleep 1

nobled --home $HOME $KEYRING add-genesis-account $(nobled --home $HOME $KEYRING keys show paramAuth -a) $AUTHGEN
sleep 1
nobled --home $HOME $KEYRING add-genesis-account $(nobled --home $HOME $KEYRING keys show swag -a) $SWAGGEN
sleep 1
nobled --home $HOME $KEYRING add-genesis-account $(nobled --home $HOME $KEYRING keys show owner -a) $ZERO
sleep 1
nobled --home $HOME $KEYRING add-genesis-account $(nobled --home $HOME $KEYRING keys show masterMinter -a) $ZERO
sleep 1
nobled --home $HOME $KEYRING add-genesis-account $(nobled --home $HOME $KEYRING keys show pauser -a) $ZERO
sleep 1
nobled --home $HOME $KEYRING add-genesis-account $(nobled --home $HOME $KEYRING keys show blacklister -a) $ZERO
sleep 1
nobled --home $HOME $KEYRING add-genesis-account $(nobled --home $HOME $KEYRING keys show mintercontroller -a) $ZERO
sleep 1
nobled --home $HOME $KEYRING add-genesis-account $(nobled --home $HOME $KEYRING keys show minter -a) $ZERO
sleep 1

mkdir $HOME/config/gentx

nobled --home $HOME $KEYRING gentx validator1 $valGenTx --chain-id $CHAINID --min-self-delegation 1 --output-document $HOME/config/gentx/gentx1.json
sleep 1

nobled --home $HOME collect-gentxs
sleep 1

PARAMAUTH=$(jq '.address' $HOME/paramAuth_seed.json)
OWNER=$(jq '.address' $HOME/owner_seed.json)
MASTERMINTER=$(jq '.address' $HOME/masterMinter_seed.json)
PAUSER=$(jq '.address' $HOME/pauser_seed.json)
BLACKLISTER=$(jq '.address' $HOME/blacklister_seed.json)

# Check platform
platform='unknown'
unamestr=`uname`
if [ "$unamestr" = 'Linux' ]; then
   platform='linux'
fi

# Set proper defaults and change ports (use a different sed for Mac or Linux)
echo "Change settings in config.toml and genesis.json files..."
if [ $platform = 'linux' ]; then
  sed -i 's#"tcp://127.0.0.1:26657"#"tcp://0.0.0.0:'"$RPCPORT"'"#g' $HOME/config/config.toml
  sed -i 's#"tcp://0.0.0.0:26656"#"tcp://0.0.0.0:'"$P2PPORT"'"#g' $HOME/config/config.toml
  sed -i 's#"localhost:6060"#"localhost:'"$P2PPORT"'"#g' $HOME/config/config.toml
  sed -i 's/timeout_commit = "5s"/timeout_commit = "1s"/g' $HOME/config/config.toml
  sed -i 's/timeout_propose = "3s"/timeout_propose = "1s"/g' $HOME/config/config.toml
  sed -i 's/index_all_keys = false/index_all_keys = true/g' $HOME/config/config.toml
  sed -i 's/"bond_denom": "stake"/"bond_denom": "'"$BASEDENOM"'"/g' $HOME/config/genesis.json
  sed -i 's/owner": null/owner": { "address": '"$OWNER"' }/g' $HOME/config/genesis.json
  sed -i 's/masterMinter": null/masterMinter": { "address": '"$MASTERMINTER"' }/g' $HOME/config/genesis.json
  sed -i 's/pauser": null/pauser": { "address": '"$PAUSER"' }/g' $HOME/config/genesis.json
  sed -i 's/blacklister": null/blacklister": { "address": '"$BLACKLISTER"' }/g' $HOME/config/genesis.json
  sed -i 's/mintingDenom": null/mintingDenom": { "denom": "'$MINTING_BASEDENOM'" }/g' $HOME/config/genesis.json
  sed -i 's/paused": null/paused": { "paused": false }/g' $HOME/config/genesis.json
  sed -i 's/"denom_metadata": \[]/"denom_metadata": [ { "display": "noble", "base": "unoble", "name": "noble", "symbol": "noble", "denom_units": [ { "denom": "unoble", "aliases": [ "micronoble" ], "exponent": "0" }, { "denom": "mnoble", "aliases": [ "milinoble" ], "exponent": "3" }, { "denom": "noble", "aliases": null, "exponent": "6" }]}, { "display": "stake", "base": "ustake", "name": "stake", "symbol": "stake", "denom_units": [ { "denom": "ustake", "aliases": [ "microstake" ], "exponent": "0" }, { "denom": "mstake", "aliases": [ "milistake" ], "exponent": "3" }, { "denom": "stake", "aliases": null, "exponent": "6" } ] }, { "display": "uusdc", "base": "uusdc", "name": "usdc", "symbol": "usdc", "denom_units": [ { "denom": "uusdc", "aliases": [ "microusdc" ], "exponent": "0" }, { "denom": "musdc", "aliases": [ "miliusdc" ], "exponent": "3" }, { "denom": "usdc", "aliases": null, "exponent": "6" } ] } ]/g' $HOME/config/genesis.json
  sed -i 's/"authority": ""/"authority": '"$PARAMAUTH"'/g' $HOME/config/genesis.json

else
  sed -i '' 's#"tcp://127.0.0.1:26657"#"tcp://0.0.0.0:'"$RPCPORT"'"#g' $HOME/config/config.toml
  sed -i '' 's#"tcp://0.0.0.0:26656"#"tcp://0.0.0.0:'"$P2PPORT"'"#g' $HOME/config/config.toml
  sed -i '' 's#"localhost:6060"#"localhost:'"$P2PPORT"'"#g' $HOME/config/config.toml
  sed -i '' 's/timeout_commit = "5s"/timeout_commit = "1s"/g' $HOME/config/config.toml
  sed -i '' 's/timeout_propose = "3s"/timeout_propose = "1s"/g' $HOME/config/config.toml
  sed -i '' 's/index_all_keys = false/index_all_keys = true/g' $HOME/config/config.toml
  sed -i '' 's/"bond_denom": "stake"/"bond_denom": "'"$BASEDENOM"'"/g' $HOME/config/genesis.json
  sed -i '' 's/owner": null/owner": { "address": '"$OWNER"' }/g' $HOME/config/genesis.json
  sed -i '' 's/masterMinter": null/masterMinter": { "address": '"$MASTERMINTER"' }/g' $HOME/config/genesis.json
  sed -i '' 's/pauser": null/pauser": { "address": '"$PAUSER"' }/g' $HOME/config/genesis.json
  sed -i '' 's/blacklister": null/blacklister": { "address": '"$BLACKLISTER"' }/g' $HOME/config/genesis.json
  sed -i '' 's/mintingDenom": null/mintingDenom": { "denom": "'$MINTING_BASEDENOM'" }/g' $HOME/config/genesis.json
  sed -i '' 's/paused": null/paused": { "paused": false }/g' $HOME/config/genesis.json
  sed -i '' 's/"denom_metadata": \[]/"denom_metadata": [ { "display": "noble", "base": "unoble", "name": "noble", "symbol": "noble", "denom_units": [ { "denom": "unoble", "aliases": [ "micronoble" ], "exponent": "0" }, { "denom": "mnoble", "aliases": [ "milinoble" ], "exponent": "3" }, { "denom": "noble", "aliases": null, "exponent": "6" }]}, { "display": "stake", "base": "ustake", "name": "stake", "symbol": "stake", "denom_units": [ { "denom": "ustake", "aliases": [ "microstake" ], "exponent": "0" }, { "denom": "mstake", "aliases": [ "milistake" ], "exponent": "3" }, { "denom": "stake", "aliases": null, "exponent": "6" } ] }, { "display": "uusdc", "base": "uusdc", "name": "usdc", "symbol": "usdc", "denom_units": [ { "denom": "uusdc", "aliases": [ "microusdc" ], "exponent": "0" }, { "denom": "musdc", "aliases": [ "miliusdc" ], "exponent": "3" }, { "denom": "usdc", "aliases": null, "exponent": "6" } ] } ]/g' $HOME/config/genesis.json
  sed -i '' 's/"authority": ""/"authority": '"$PARAMAUTH"'/g' $HOME/config/genesis.json

fi

# nobled --home $HOME start --pruning=nothing --grpc-web.enable=false --grpc.address="0.0.0.0:$GRPCPORT" > $HOME.log 2>&1 &

echo "Done!"
