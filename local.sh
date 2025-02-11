alias nobled=./build/nobled

for arg in "$@"
do
    case $arg in
        -r|--reset)
        rm -rf .duke
        shift
        ;;
    esac
done

if ! [ -f .duke/data/priv_validator_state.json ]; then
  nobled init validator --chain-id "duke-1" --home .duke &> /dev/null

  nobled keys add validator --home .duke --keyring-backend test &> /dev/null
  nobled genesis add-genesis-account validator 1000000ustake --home .duke --keyring-backend test
  AUTHORITY=$(nobled keys add authority --home .duke --keyring-backend test --output json | jq .address)
  nobled genesis add-genesis-account authority 4000000ustake --home .duke --keyring-backend test
  nobled genesis add-genesis-account noble1cyyzpxplxdzkeea7kwsydadg87357qnah9s9cv 1000000uusdc --home .duke --keyring-backend test

  TEMP=.duke/genesis.json
  touch $TEMP && jq '.app_state.authority.owner = '$AUTHORITY'' .duke/config/genesis.json > $TEMP && mv $TEMP .duke/config/genesis.json
  touch $TEMP && jq '.app_state.bank.denom_metadata += [{ "description": "Circle USD Coin", "denom_units": [{ "denom": "uusdc", "exponent": 0, "aliases": ["microusdc"] }, { "denom": "usdc", "exponent": 6 }], "base": "uusdc", "display": "usdc", "name": "Circle USD Coin", "symbol": "USDC" }]' .duke/config/genesis.json > $TEMP && mv $TEMP .duke/config/genesis.json
  touch $TEMP && jq '.app_state.bank.denom_metadata += [{ "description": "Ondo US Dollar Yield", "denom_units": [{ "denom": "ausdy", "exponent": 0, "aliases": ["attousdy"] }, { "denom": "usdy", "exponent": 18 }], "base": "ausdy", "display": "usdy", "name": "Ondo US Dollar Yield", "symbol": "USDY" }]' .duke/config/genesis.json > $TEMP && mv $TEMP .duke/config/genesis.json
  touch $TEMP && jq '.app_state.bank.denom_metadata += [{ "description": "Hashnote US Yield Coin", "denom_units": [{ "denom": "uusyc", "exponent": 0, "aliases": ["microusyc"] }, { "denom": "usyc", "exponent": 6 }], "base": "uusyc", "display": "usyc", "name": "Hashnote US Yield Coin", "symbol": "USYC" }]' .duke/config/genesis.json > $TEMP && mv $TEMP .duke/config/genesis.json
  touch $TEMP && jq '.app_state.bank.denom_metadata += [{ "description": "Monerium EUR emoney", "denom_units": [{ "denom": "ueure", "exponent": 0, "aliases": ["microeure"] }, { "denom": "eure", "exponent": 6 }], "base": "ueure", "display": "eure", "name": "Monerium EUR emoney", "symbol": "EURe" }]' .duke/config/genesis.json > $TEMP && mv $TEMP .duke/config/genesis.json
  # TODO: This is a dummy value to ensure successful genesis, replace with real value!
  touch $TEMP && jq '.app_state.hyperlane.params.domain = "42"' .duke/config/genesis.json > $TEMP && mv $TEMP .duke/config/genesis.json
  touch $TEMP && jq '.app_state."fiat-tokenfactory".mintingDenom = { "denom": "uusdc" }' .duke/config/genesis.json > $TEMP && mv $TEMP .duke/config/genesis.json
  touch $TEMP && jq '.app_state."fiat-tokenfactory".paused.paused = false' .duke/config/genesis.json > $TEMP && mv $TEMP .duke/config/genesis.json
  touch $TEMP && jq '.app_state.staking.params.bond_denom = "ustake"' .duke/config/genesis.json > $TEMP && mv $TEMP .duke/config/genesis.json
  touch $TEMP && jq '.app_state.wormhole.config.chain_id = 4009' .duke/config/genesis.json > $TEMP && mv $TEMP .duke/config/genesis.json
  touch $TEMP && jq '.app_state.wormhole.config.gov_chain = 1' .duke/config/genesis.json > $TEMP && mv $TEMP .duke/config/genesis.json
  touch $TEMP && jq '.app_state.wormhole.config.gov_address = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQ="' .duke/config/genesis.json > $TEMP && mv $TEMP .duke/config/genesis.json
  touch $TEMP && jq '.app_state.wormhole.guardian_sets = {"0":{"addresses":["vvpCnVfNGLf4pNkaLamrSvBdD74="],"expiration_time":0}}' .duke/config/genesis.json > $TEMP && mv $TEMP .duke/config/genesis.json

  nobled genesis gentx validator 1000000ustake --chain-id "duke-1" --home .duke --keyring-backend test &> /dev/null
  nobled genesis collect-gentxs --home .duke &> /dev/null
fi

nobled start --home .duke
