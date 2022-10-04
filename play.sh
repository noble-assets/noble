# run ignite c serve -r

sleep 2
nobled tx tokenfactory update-master-minter $(nobled keys show masterminter -a) --from owner -y
sleep 2
nobled tx tokenfactory configure-minter $(nobled keys show minter -a) 1000usdc --from masterminter -y
sleep 2
nobled tx tokenfactory mint $(nobled keys show user -a) 100usdc --from minter -y
sleep 2
nobled q bank balances $(nobled keys show user -a)
sleep 2
nobled q tokenfactory list-minters
sleep 2
nobled tx tokenfactory mint $(nobled keys show user -a) 99999999999999usdc --from minter -y
sleep 2
nobled tx tokenfactory update-blacklister $(nobled keys show blacklister -a) --from owner -y
sleep 2
nobled tx tokenfactory blacklist $(nobled keys show minter -a) --from blacklister -y
sleep 2
nobled tx tokenfactory mint $(nobled keys show user -a) 100usdc --from minter -y
sleep 2
nobled tx tokenfactory unblacklist $(nobled keys show minter -a) --from blacklister -y
sleep 2
nobled tx tokenfactory mint $(nobled keys show user -a) 100usdc --from minter -y
sleep 2
nobled q bank balances $(nobled keys show user -a)
sleep 2
nobled tx tokenfactory blacklist $(nobled keys show user -a) --from blacklister -y
sleep 2
nobled tx tokenfactory mint $(nobled keys show user -a) 100usdc --from minter -y
sleep 2
nobled tx tokenfactory unblacklist $(nobled keys show user -a) --from blacklister -y
sleep 2
nobled tx tokenfactory mint $(nobled keys show user -a) 100usdc --from minter -y
sleep 2
nobled tx tokenfactory update-pauser $(nobled keys show pauser -a) --from owner -y
sleep 2
nobled tx tokenfactory pause --from pauser -y
sleep 2
nobled tx tokenfactory mint $(nobled keys show user -a) 100usdc --from minter -y
sleep 2
nobled tx bank send $(nobled keys show user -a) $(nobled keys show alice -a) 100usdc --from user -y
sleep 2
nobled tx tokenfactory unpause --from pauser -y
sleep 2
nobled tx bank send $(nobled keys show user -a) $(nobled keys show alice -a) 100usdc --from user -y

# nobled tx tokenfactory mint $(nobled keys show user -a) 100usdc --from minter -y
# sleep 2
# nobled q bank balances $(nobled keys show user -a)
# sleep 2
# nobled tx tokenfactory mint $(nobled keys show minter -a) 100usdc --from minter -y
# sleep 2
# nobled tx tokenfactory burn 100usdc --from minter -y
# sleep 2
# nobled q bank balances $(nobled keys show user -a)



# {"body":{"messages":[{"@type":"/ibc.applications.transfer.v1.MsgTransfer","source_port":"transfer","source_channel":"channel-0","token":{"denom":"stake","amount":"100"},"sender":"cosmos154wpvzcw47ymgkpcpklfcf4tc8rf6mksfreutd","receiver":"cosmos1vpka0rrdffqc09la7rgkvg29m6hjygd8gl2yvz","timeout_height":{"revision_number":"0","revision_height":"1350"},"timeout_timestamp":"1663247744886911000"}],"memo":"","timeout_height":"0","extension_options":[],"non_critical_extension_options":[]},"auth_info":{"signer_infos":[],"fee":{"amount":[],"gas_limit":"200000","payer":"","granter":""},"tip":null},"signatures":[]}

# nobled tx ibc-transfer transfer transfer channel-0 cosmos1vpka0rrdffqc09la7rgkvg29m6hjygd8gl2yvz 100stake --from alice

# ignite relayer configure -a \
#   --source-rpc "http://0.0.0.0:26657" \
#   --source-faucet "http://0.0.0.0:4500" \
#   --source-port "transfer" \
#   --source-version "ics20-1" \
#   --source-gasprice "0.0000025stake" \
#   --source-prefix "cosmos" \
#   --source-gaslimit 300000 \
#   --target-rpc "http://0.0.0.0:26659" \
#   --target-faucet "http://0.0.0.0:4501" \
#   --target-port "transfer" \
#   --target-version "ics20-1" \
#   --target-gasprice "0.0000025stake" \
#   --target-prefix "cosmos" \
#   --target-gaslimit 300000