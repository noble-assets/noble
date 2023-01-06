# noble
**noble** is a blockchain built using Cosmos SDK and Tendermint

## Get started

Install [go](https://go.dev/dl/)

## Build and install to go bin path

```
make install
```

## Initialize config

Come up with a moniker for your node, then run:

```
nobled init $MONIKER
```

## Access Control

|                                | **Admin** | **Owner** | **Minter** | **Master Minter** | **Minter Controller** | **Pauser** | **Blacklister** | **Is Paused<br>(Actions Allowed)** |
|--------------------------------|:---------:|:---------:|:----------:|:-----------------:|:---------------------:|:----------:|:---------------:|:--------------------------------:|
| **Blacklist**                  |           |           |            |                   |                       |            |        x        |                 x                |
| **Unblacklist**                |           |           |            |                   |                       |            |        x        |                 x                |
| **Burn**                       |           |           |      x     |                   |                       |            |                 |                                  |
| **Mint**                       |           |           |      x     |                   |                       |            |                 |                                  |
| **Change Admin**               |     x     |           |            |                   |                       |            |                 |                 x                |
| **Configure Mint Controller**  |           |           |            |         x         |                       |            |                 |                 x                |
| **Configure Minter allowance** |           |           |            |                   |           x           |            |                 |                 x                |
| **Pause**                      |           |           |            |                   |                       |      x     |                 |                 x                |
| **Unpause**                    |           |           |            |                   |                       |      x     |                 |                 x                |
| **Remove Minter Controller**   |           |           |            |         x         |                       |            |                 |                 x                |
| **Remove Minter**              |           |           |            |                   |                       |            |                 |                 x                |
| **Update Blacklister**         |           |     x     |            |                   |                       |            |                 |                 x                |
| **Update Master Minter**       |           |     x     |            |                   |                       |            |                 |                 x                |
| **Update Owner**               |           |     x     |            |                   |                       |            |                 |                 x                |
| **Update Pauser**              |           |     x     |            |                   |                       |            |                 |                 x                |
| **Transer Tokens**             |     x     |     x     |      x     |         x         |           x           |      x     |        x        |                                  |
 
 
## Launch with genesis file or run as standalone chain

To launch as a consumer chain, download and save shared genesis file to `~/.noble/config/genesis.json`. Additionally add peering information (`persistent_peers` or `seeds`) to `~/.noble/config/config.toml`

To instead launch as a standalone, single node chain, run:

```
nobled add-consumer-section
```

## Launch node

```
nobled start
```
