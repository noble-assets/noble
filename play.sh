#!/bin/bash

# !! PLEASE ONLY RUN THIS SCRIPT IN A TESTING ENVIRONMENT !!
#   THIS SCRIPT: 
#     - Stops/starts noble binary    

# This script is meant for quick experimentation of the Tokenfactory Module and Noble Chain functionality. 
#   - Starts Noble chain
#   - Delagates and funds all privledged accounts relating to the Tokenfactory Module
#   - The "owner" account is both the Tokenfactory Owner AND the Param Authority

killall nobled
[ -d "play_sh" ] && rm -r "play_sh"


BINARY="nobled"
CHAINID="noble-1"
CHAINDIR="play_sh"
RPCPORT="26657"
P2PPORT="26656"
PROFPORT="6060"
GRPCPORT="9090"
DENOM="noble"
BASEDENOM="unoble"

KEYRING="--keyring-backend test"

# tokenfactory
MINTING_DENOM='token'
MINTING_BASEDENOM="u$MINTING_DENOM"


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

redirect nobled --home $CHAINDIR/$CHAINID --chain-id $CHAINID init $CHAINID 
sleep 1
nobled --home $CHAINDIR/$CHAINID $KEYRING keys add validator --output json > $CHAINDIR/$CHAINID/validator_seed.json 2>&1
sleep 1
nobled --home $CHAINDIR/$CHAINID $KEYRING keys add owner --output json > $CHAINDIR/$CHAINID/key_seed.json 2>&1
sleep 1
redirect nobled --home $CHAINDIR/$CHAINID $KEYRING add-genesis-account $(nobled --home $CHAINDIR/$CHAINID keys $KEYRING show owner -a) $coins 
sleep 1
redirect nobled --home $CHAINDIR/$CHAINID $KEYRING add-genesis-account $(nobled --home $CHAINDIR/$CHAINID keys $KEYRING show validator -a) $coins 
sleep 1
redirect nobled --home $CHAINDIR/$CHAINID $KEYRING gentx validator $delegate --chain-id $CHAINID
sleep 1
redirect nobled --home $CHAINDIR/$CHAINID collect-gentxs 
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
  sed -i 's/"bond_denom": "stake"/"bond_denom": "'"$DENOM"'"/g' $CHAINDIR/$CHAINID/config/genesis.json
  sed -i 's/owner": null/owner": { "address": '"$OWNER"' }/g' $CHAINDIR/$CHAINID/config/genesis.json
  sed -i 's/mintingDenom": null/mintingDenom": { "denom": "'$MINTING_BASEDENOM'" }/g' $CHAINDIR/$CHAINID/config/genesis.json
  sed -i 's/paused": null/paused": { "paused": false }/g' $CHAINDIR/$CHAINID/config/genesis.json
  sed -i 's/"denom_metadata": \[]/"denom_metadata": [ { "display": "'$MINTING_DENOM'", "base": "'$MINTING_BASEDENOM'", "name": "'$MINTING_DENOM'", "symbol": "'$MINTING_DENOM'", "denom_units": [ { "denom": "'$MINTING_DENOM'", "aliases": [ "micro'$MINTING_BASEDENOM'" ], "exponent": "0" }, { "denom": "m'$MINTING_BASEDENOM'", "aliases": [ "mili'$MINTING_BASEDENOM'" ], "exponent": "3" }, { "denom": "'$MINTING_BASEDENOM'", "aliases": null, "exponent": "6" } ] } ]/g' $CHAINDIR/$CHAINID/config/genesis.json
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
  sed -i '' 's/"denom_metadata": \[]/"denom_metadata": [ { "display": "'$MINTING_DENOM'", "base": "'$MINTING_BASEDENOM'", "name": "'$MINTING_DENOM'", "symbol": "'$MINTING_DENOM'", "denom_units": [ { "denom": "'$MINTING_DENOM'", "aliases": [ "micro'$MINTING_BASEDENOM'" ], "exponent": "0" }, { "denom": "m'$MINTING_BASEDENOM'", "aliases": [ "mili'$MINTING_BASEDENOM'" ], "exponent": "3" }, { "denom": "'$MINTING_BASEDENOM'", "aliases": null, "exponent": "6" } ] } ]/g' $CHAINDIR/$CHAINID/config/genesis.json
  sed -i '' 's/"authority": ""/"authority": '"$OWNER"'/g' $CHAINDIR/$CHAINID/config/genesis.json

fi

# Start
nobled --home $CHAINDIR/$CHAINID start --pruning=nothing --grpc-web.enable=false --grpc.address="0.0.0.0:$GRPCPORT" > $CHAINDIR/$CHAINID.log 2>&1 &

OWNER_MN=$(jq .mnemonic $CHAINDIR/$CHAINID/key_seed.json)
OWNER_MN=$(echo $OWNER_MN | cut -d "\"" -f 2)

# Create/recover keys
sleep 2
nobled --home $CHAINDIR/$CHAINID $KEYRING keys add masterminter
nobled --home $CHAINDIR/$CHAINID $KEYRING keys add mintercontroller
nobled --home $CHAINDIR/$CHAINID $KEYRING keys add minter
nobled --home $CHAINDIR/$CHAINID $KEYRING keys add blacklister
nobled --home $CHAINDIR/$CHAINID $KEYRING keys add pauser
nobled --home $CHAINDIR/$CHAINID $KEYRING keys add user

# Fund accounts
nobled --home $CHAINDIR/$CHAINID $KEYRING tx bank send owner $(nobled --home $CHAINDIR/$CHAINID $KEYRING keys show masterminter -a) 50noble -y
sleep 2
nobled --home $CHAINDIR/$CHAINID $KEYRING tx bank send owner $(nobled --home $CHAINDIR/$CHAINID $KEYRING keys show mintercontroller -a) 50noble -y
sleep 2
nobled --home $CHAINDIR/$CHAINID $KEYRING tx bank send owner $(nobled --home $CHAINDIR/$CHAINID $KEYRING keys show minter -a) 50noble -y
sleep 2
nobled --home $CHAINDIR/$CHAINID $KEYRING tx bank send owner $(nobled --home $CHAINDIR/$CHAINID $KEYRING keys show blacklister -a) 50noble -y
sleep 2
nobled --home $CHAINDIR/$CHAINID $KEYRING tx bank send owner $(nobled --home $CHAINDIR/$CHAINID $KEYRING keys show pauser -a) 50noble -y
sleep 2
nobled --home $CHAINDIR/$CHAINID $KEYRING tx bank send owner $(nobled --home $CHAINDIR/$CHAINID $KEYRING keys show user -a) 50noble -y
sleep 2

# Delegate privledges
nobled --home $CHAINDIR/$CHAINID $KEYRING tx tokenfactory update-master-minter $(nobled --home $CHAINDIR/$CHAINID $KEYRING keys show masterminter -a) --from owner -y
sleep 2
nobled --home $CHAINDIR/$CHAINID $KEYRING tx tokenfactory configure-minter-controller $(nobled --home $CHAINDIR/$CHAINID $KEYRING keys show mintercontroller -a) $(nobled --home $CHAINDIR/$CHAINID $KEYRING keys show minter -a) --from masterminter -y
sleep 2
nobled --home $CHAINDIR/$CHAINID $KEYRING tx tokenfactory configure-minter $(nobled --home $CHAINDIR/$CHAINID $KEYRING keys show minter -a) 1000$MINTING_BASEDENOM --from mintercontroller -y
sleep 2
nobled --home $CHAINDIR/$CHAINID $KEYRING tx tokenfactory update-blacklister $(nobled --home $CHAINDIR/$CHAINID $KEYRING keys show blacklister -a) --from owner -y
sleep 2
nobled --home $CHAINDIR/$CHAINID $KEYRING tx tokenfactory update-pauser $(nobled --home $CHAINDIR/$CHAINID $KEYRING keys show pauser -a) --from owner -y
