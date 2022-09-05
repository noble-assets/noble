export MASTER_MINTER_ADDRESS=$(nobled keys show masterminter -a)
export MINTER_ADDRESS=$(nobled keys show minter -a)
export USER_ADDRESS=$(nobled keys show user -a)

sleep 3
nobled tx tokenfactory update-master-minter $MASTER_MINTER_ADDRESS --from owner -y
sleep 3
nobled tx tokenfactory configure-minter $MINTER_ADDRESS 1000usdc --from masterminter -y
sleep 3
nobled tx tokenfactory mint $USER_ADDRESS 100usdc --from minter -y
sleep 3
nobled q bank balances $USER_ADDRESS
sleep 3
nobled q tokenfactory list-minters
sleep 3
nobled tx tokenfactory mint $USER_ADDRESS 99999999999999usdc --from minter -y