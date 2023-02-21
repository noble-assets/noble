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

HOME=$CHAINDIR/$CHAINID

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
if ! mkdir -p $HOME 2>/dev/null; then
    echo "Failed to create chain folder. Aborting..."
    exit 1
fi

AUTHGEN="1000000000000$DENOM,1000000000000$BASEDENOM"
valGenesis="0$DENOM,1000000$BASEDENOM"
valGenTx="1000000$BASEDENOM"
ZERO="0$DENOM,0$BASEDENOM"


nobled --home $HOME --chain-id $CHAINID init $CHAINID 
sleep 1


nobled --home $HOME $KEYRING keys add paramAuth --output json > $HOME/paramAuth_seed.json 2>&1
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
nobled --home $HOME $KEYRING keys add validator2 --output json > $HOME/validator_seed2.json 2>&1
sleep 1
nobled --home $HOME $KEYRING keys add validator3 --output json > $HOME/validator_seed3.json 2>&1
sleep 1
nobled --home $HOME $KEYRING keys add validator4 --output json > $HOME/validator_seed4.json 2>&1
sleep 1
nobled --home $HOME $KEYRING keys add validator5 --output json > $HOME/validator_seed5.json 2>&1
sleep 1
nobled --home $HOME $KEYRING keys add validator6 --output json > $HOME/validator_seed6.json 2>&1
sleep 1

nobled --home $HOME $KEYRING add-genesis-account $(nobled --home $HOME $KEYRING keys show paramAuth -a) $AUTHGEN
sleep 1
nobled --home $HOME $KEYRING add-genesis-account $(nobled --home $HOME $KEYRING keys show owner -a) $ZERO
sleep 1
nobled --home $HOME $KEYRING add-genesis-account $(nobled --home $HOME $KEYRING keys show masterMinter -a) $ZERO
sleep 1
nobled --home $HOME $KEYRING add-genesis-account $(nobled --home $HOME $KEYRING keys show pauser -a) $ZERO
sleep 1
nobled --home $HOME $KEYRING add-genesis-account $(nobled --home $HOME $KEYRING keys show blacklister -a) $ZERO
sleep 1
nobled --home $HOME $KEYRING add-genesis-account $(nobled --home $HOME $KEYRING keys show validator1 -a) $valGenesis --vesting-amount $valGenesis  --vesting-end-time 32507807081 # Normal val
sleep 1
nobled --home $HOME $KEYRING add-genesis-account $(nobled --home $HOME $KEYRING keys show validator2 -a) "0$DENOM,1000000$BASEDENOM" --vesting-amount "0$DENOM,1000000$BASEDENOM"  --vesting-end-time 32507807081 
sleep 1
nobled --home $HOME $KEYRING add-genesis-account $(nobled --home $HOME $KEYRING keys show validator3 -a) "0$DENOM,1000010$BASEDENOM" --vesting-amount "0$DENOM,1000000$BASEDENOM"  --vesting-end-time 32507807081
sleep 1
nobled --home $HOME $KEYRING add-genesis-account $(nobled --home $HOME $KEYRING keys show validator4 -a) "0$DENOM,2000000$BASEDENOM" 
sleep 1
# nobled --home $HOME $KEYRING add-genesis-account $(nobled --home $HOME $KEYRING keys show validator5 -a) $valGenesis --vesting-amount $valGenesis  --vesting-end-time 32507807081
# sleep 1
# nobled --home $HOME $KEYRING add-genesis-account $(nobled --home $HOME $KEYRING keys show validator6 -a) $valGenesis --vesting-amount $valGenesis  --vesting-end-time 32507807081
# sleep 1

mkdir $HOME/config/gentx

nobled --home $HOME $KEYRING gentx validator1 $valGenTx --chain-id $CHAINID --min-self-delegation 1 --output-document $HOME/config/gentx/gentx1.json
sleep 1
# nobled --home $HOME $KEYRING gentx validator2 $valGenTx --chain-id $CHAINID --min-self-delegation 1 --pubkey $(nobled --home $HOME/temp_inits/init2 tendermint show-validator) --output-document $HOME/config/gentx/gentx2.json
# sleep 1
# nobled --home $HOME $KEYRING gentx validator3 $valGenTx --chain-id $CHAINID --min-self-delegation 1 --pubkey $(nobled --home $HOME/temp_inits/init3 tendermint show-validator) --output-document $HOME/config/gentx/gentx3.json
# sleep 1
# nobled --home $HOME $KEYRING gentx validator4 $valGenTx --chain-id $CHAINID --min-self-delegation 1 --pubkey $(nobled --home $HOME/temp_inits/init4 tendermint show-validator) --output-document $HOME/config/gentx/gentx4.json
# sleep 1
# nobled --home $HOME $KEYRING gentx validator5 $valGenTx --chain-id $CHAINID --min-self-delegation 1 --pubkey $(nobled --home $HOME/temp_inits/init5 tendermint show-validator) --output-document $HOME/config/gentx/gentx5.json
# sleep 1
# nobled --home $HOME $KEYRING gentx validator6 $valGenTx --chain-id $CHAINID --min-self-delegation 1 --pubkey $(nobled --home $HOME/temp_inits/init6 tendermint show-validator) --output-document $HOME/config/gentx/gentx6.json
# sleep 1

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
  sed -i 's/"denom_metadata": \[]/"denom_metadata": [ { "display": "stake", "base": "ustake", "name": "stake", "symbol": "stake", "denom_units": [ { "denom": "ustake", "aliases": [ "microstake" ], "exponent": "0" }, { "denom": "mstake", "aliases": [ "milistake" ], "exponent": "3" }]}, { "denom": "stake", "aliases": null, "exponent": "6" } ] }, { "display": "rupee", "base": "urupee", "name": "rupee", "symbol": "rupee", "denom_units": [ { "denom": "urupee", "aliases": [ "microrupee" ], "exponent": "0" }, { "denom": "mrupee", "aliases": [ "milirupee" ], "exponent": "3" }, { "denom": "rupee", "aliases": null, "exponent": "6" } ] } ]/g' $HOME/config/genesis.json
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
  sed -i '' 's/"denom_metadata": \[]/"denom_metadata": [ { "display": "stake", "base": "ustake", "name": "stake", "symbol": "stake", "denom_units": [ { "denom": "ustake", "aliases": [ "microstake" ], "exponent": "0" }, { "denom": "mstake", "aliases": [ "milistake" ], "exponent": "3" }]}, { "denom": "stake", "aliases": null, "exponent": "6" } ] }, { "display": "rupee", "base": "urupee", "name": "rupee", "symbol": "rupee", "denom_units": [ { "denom": "urupee", "aliases": [ "microrupee" ], "exponent": "0" }, { "denom": "mrupee", "aliases": [ "milirupee" ], "exponent": "3" }, { "denom": "rupee", "aliases": null, "exponent": "6" } ] } ]/g' $HOME/config/genesis.json
  sed -i '' 's/"authority": ""/"authority": '"$PARAMAUTH"'/g' $HOME/config/genesis.json

fi

nobled --home $HOME start --pruning=nothing --grpc-web.enable=false --grpc.address="0.0.0.0:$GRPCPORT" > $HOME.log 2>&1 &

echo "Done!"
