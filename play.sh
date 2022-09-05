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
