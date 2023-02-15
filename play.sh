#!/bin/bash

# !! PLEASE ONLY RUN THIS SCRIPT IN A TESTING ENVIRONMENT !!
#   THIS SCRIPT: 
#     - Stops/starts noble binary
#     - Adds/deletes noble keys
#     

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

MINTING_BASEDENOM="utoken"

KEYRING="--keyring-backend=test"
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
nobled --home $CHAINDIR/$CHAINID keys add validator $KEYRING --output json > $CHAINDIR/$CHAINID/validator_seed.json 2>&1
sleep 1
nobled --home $CHAINDIR/$CHAINID keys add owner $KEYRING --output json > $CHAINDIR/$CHAINID/key_seed.json 2>&1
sleep 1
redirect nobled --home $CHAINDIR/$CHAINID add-genesis-account $(nobled --home $CHAINDIR/$CHAINID keys $KEYRING show owner -a) $coins 
sleep 1
redirect nobled --home $CHAINDIR/$CHAINID add-genesis-account $(nobled --home $CHAINDIR/$CHAINID keys $KEYRING show validator -a) $coins 
sleep 1
redirect nobled --home $CHAINDIR/$CHAINID gentx validator $delegate $KEYRING --chain-id $CHAINID
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
  sed -i 's/owner": null/owner": { "address": '"$OWNER"' }/g' $CHAINDIR/$CHAINID/config/genesis.json
  sed -i 's/mintingDenom": null/mintingDenom": { "denom": "'$MINTING_BASEDENOM'" }/g' $CHAINDIR/$CHAINID/config/genesis.json
  sed -i 's/paused": null/paused": { "paused": false }/g' $CHAINDIR/$CHAINID/config/genesis.json
  sed -i 's/"denom_metadata": \[]/"denom_metadata": [ { "display": "token", "base": "utoken", "name": "Token", "symbol": "Token", "denom_units": [ { "denom": "utoken", "aliases": [ "microtoken" ], "exponent": "0" }, { "denom": "mtoken", "aliases": [ "militoken" ], "exponent": "3" }, { "denom": "token", "aliases": null, "exponent": "6" } ] } ]/g' $CHAINDIR/$CHAINID/config/genesis.json
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
  sed -i '' 's/"denom_metadata": \[]/"denom_metadata": [ { "display": "token", "base": "utoken", "name": "Token", "symbol": "Token", "denom_units": [ { "denom": "utoken", "aliases": [ "microtoken" ], "exponent": "0" }, { "denom": "mtoken", "aliases": [ "militoken" ], "exponent": "3" }, { "denom": "token", "aliases": null, "exponent": "6" } ] } ]/g' $CHAINDIR/$CHAINID/config/genesis.json
  sed -i '' 's/"authority": ""/"authority": '"$OWNER"'/g' $CHAINDIR/$CHAINID/config/genesis.json

fi

# Test command that allows Noble to be run without Provider chain
redirect nobled --home $CHAINDIR/$CHAINID add-consumer-section

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

# Start
nobled --home $CHAINDIR/$CHAINID start --pruning=nothing --grpc-web.enable=false --grpc.address="0.0.0.0:$GRPCPORT" > $CHAINDIR/$CHAINID.log 2>&1 &

OWNER_MN=$(jq .mnemonic $CHAINDIR/$CHAINID/key_seed.json)
OWNER_MN=$(echo $OWNER_MN | cut -d "\"" -f 2)

# Create/recover keys
echo $OWNER_MN | nobled --home $CHAINDIR/$CHAINID keys add owner --recover 
sleep 2
nobled --home $CHAINDIR/$CHAINID keys add masterminter
sleep 2
nobled --home $CHAINDIR/$CHAINID keys add mintercontroller
sleep 2
nobled --home $CHAINDIR/$CHAINID keys add minter
sleep 2
nobled --home $CHAINDIR/$CHAINID keys add blacklister
sleep 2
nobled --home $CHAINDIR/$CHAINID keys add pauser
sleep 2
nobled --home $CHAINDIR/$CHAINID keys add user
sleep 2

# Fund accounts
nobled --home $CHAINDIR/$CHAINID tx bank send owner $(nobled keys show masterminter -a) 50noble -y
sleep 2
nobled --home $CHAINDIR/$CHAINID tx bank send owner $(nobled keys show mintercontroller -a) 50noble -y
sleep 2
nobled --home $CHAINDIR/$CHAINID tx bank send owner $(nobled keys show minter -a) 50noble -y
sleep 2
nobled --home $CHAINDIR/$CHAINID tx bank send owner $(nobled keys show blacklister -a) 50noble -y
sleep 2
nobled --home $CHAINDIR/$CHAINID tx bank send owner $(nobled keys show pauser -a) 50noble -y
sleep 2
nobled --home $CHAINDIR/$CHAINID tx bank send owner $(nobled keys show user -a) 50noble -y
sleep 2

# Delegate privledges
nobled --home $CHAINDIR/$CHAINID tx tokenfactory update-master-minter $(nobled keys show masterminter -a) --from owner -y
sleep 2
nobled --home $CHAINDIR/$CHAINID tx tokenfactory configure-minter-controller $(nobled keys show mintercontroller -a) $(nobled keys show minter -a) --from masterminter -y
sleep 2
nobled --home $CHAINDIR/$CHAINID tx tokenfactory configure-minter $(nobled keys show minter -a) 1000utoken --from mintercontroller -y
sleep 2
nobled --home $CHAINDIR/$CHAINID tx tokenfactory update-blacklister $(nobled keys show blacklister -a) --from owner -y
sleep 2
nobled --home $CHAINDIR/$CHAINID tx tokenfactory update-pauser $(nobled keys show pauser -a) --from owner -y
