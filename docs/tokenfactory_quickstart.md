# Tokenfactory Quick Start

### The goal of this document is to run through the necessary commands to mint a tokenfactory asset.

The steps below assume the following:
- Just the "owner" account was set at genesis (This is a mandatory step to start the chain).
- The keys are named as follows (this is relevant for the `--from` flag):
    - Owner -> owner
    - Master Minter -> masterminter
    - Minter Controller -> mintercontroller
    - Minter -> minter
- Each wallet being used (except Owner and Alice) is pre-funded (making them known to the network) or included in genesis.


---

1. The `owner` account is set at genesis. Use this `owner` account to select a `Master Minter`.

```
nobled tx tokenfactory update-master-minter <MASTER-MINTERS's ADDRESS> --from owner
```

2. Use the `Master Minter` account to assign a `Minter Controller` to a `Minter`.

```
nobled tx tokenfactory configure-minter-controller <MINTER-CONTROLLER ADDRESS> <MINTER ADDRESS> --from masterminter
```

3. Use the `Minter Controller` account to assign the minter an allowance they are able to mint (ex: 1000ustake).
```
nobled tx tokenfactory configure-minter <MINTER ADDRESS> 1000ustake --from mintercontroller
```

4. Mint the asset into a user's (Alice's) wallet.
```
nobled tx tokenfactory mint <ALICE ADDRESS> 500ustake --from minter
``` 
