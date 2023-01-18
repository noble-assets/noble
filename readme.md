# noble
**noble** is a blockchain built using Cosmos SDK and Tendermint

It is meant to be run as a Consumer Chain using [Interchain Security](https://github.com/cosmos/interchain-security) (ICS).

Noble chain includes the Tokenfactory Module which allows for the minting of generic assets. 

## Tokenfactory

The Tokenfactory Module allows generic assets to be minted and controlled by privileged accounts.

This module was built according to the [Centre](https://github.com/centrehq) specs [HERE](https://github.com/centrehq/centre-tokens/blob/master/doc/tokendesign.md#removing-minters)

The Access Control table below shows the functionality tied to each privileged account.

## Access Control

|                                | **Owner** | **Minter** | **Master Minter** | **Minter Controller** | **Pauser** | **Blacklister** | **Is Paused<br>(Actions Allowed)** |
|--------------------------------|:---------:|:----------:|:-----------------:|:---------------------:|:----------:|:---------------:|:--------------------------------:|
| **Blacklist**                  |           |            |                   |                       |            |        x        |                 x                |
| **Unblacklist**                |           |            |                   |                       |            |        x        |                 x                |
| **Burn**                       |           |      x     |                   |                       |            |                 |                                  |
| **Mint**                       |           |      x     |                   |                       |            |                 |                                  |
| **Configure Mint Controller**  |           |            |         x         |                       |            |                 |                 x                |
| **Configure Minter allowance** |           |            |                   |           x           |            |                 |                 x                |
| **Pause**                      |           |            |                   |                       |      x     |                 |                 x                |
| **Unpause**                    |           |            |                   |                       |      x     |                 |                 x                |
| **Remove Minter Controller**   |           |            |         x         |                       |            |                 |                 x                |
| **Remove Minter**              |           |            |                   |           x           |            |                 |                 x                |
| **Update Blacklister**         |     x     |            |                   |                       |            |                 |                 x                |
| **Update Master Minter**       |     x     |            |                   |                       |            |                 |                 x                |
| **Update Owner**               |     x     |            |                   |                       |            |                 |                 x                |
| **Update Pauser**              |     x     |            |                   |                       |            |                 |                 x                |
| **Transer Tokens**             |     x     |      x     |         x         |           x           |      x     |        x        |                                  |
 
 
## Build and install to go bin path

[go](https://go.dev/dl/) will be needed for this install

```
make install
```

## Initialize config

Come up with a moniker for your node, then run:

```
nobled init $MONIKER
```

## Launch with genesis file or run as standalone chain

To launch as a consumer chain, download and save shared genesis file to [TODO: Path to Genesis]. Additionally add peering information (`persistent_peers` or `seeds`) to `~/.noble/config/config.toml`

## Launch node

```
nobled start
```

## Testing

Testing Noble as a standalone chain (without Provider) is possible by running the below command prior to starting node:

```
nobled add-consumer-section
```

To quickly spin up a standalone noble chain and setup all privileged accounts, refer to[ play.sh](play.sh)

